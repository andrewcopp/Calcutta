package httpserver

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
)

type cognitoJWTVerifier struct {
	cfg     platform.Config
	iss     string
	jwksURL string
	client  *http.Client

	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
	cacheTTL  time.Duration
}

type cognitoClaims struct {
	Sub        string `json:"sub"`
	Email      string `json:"email"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	TokenUse   string `json:"token_use"`
	Iss        string `json:"iss"`
	Aud        string `json:"aud"`
	ClientID   string `json:"client_id"`
	Exp        int64  `json:"exp"`
	Iat        int64  `json:"iat"`
}

type jwksDocument struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	Alg string   `json:"alg"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

func newCognitoJWTVerifier(cfg platform.Config) (*cognitoJWTVerifier, error) {
	if cfg.CognitoRegion == "" || cfg.CognitoUserPoolID == "" {
		return nil, errors.New("cognito config missing")
	}
	iss := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", cfg.CognitoRegion, cfg.CognitoUserPoolID)
	jwksURL := iss + "/.well-known/jwks.json"
	return &cognitoJWTVerifier{
		cfg:      cfg,
		iss:      iss,
		jwksURL:  jwksURL,
		client:   &http.Client{Timeout: 5 * time.Second},
		keys:     map[string]*rsa.PublicKey{},
		cacheTTL: 1 * time.Hour,
	}, nil
}

func (v *cognitoJWTVerifier) Verify(token string, now time.Time) (*cognitoClaims, error) {
	if now.IsZero() {
		now = time.Now()
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	enc := base64.RawURLEncoding
	headerJSON, err := enc.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("invalid jwt header")
	}
	var hdr jwtHeader
	if err := json.Unmarshal(headerJSON, &hdr); err != nil {
		return nil, errors.New("invalid jwt header")
	}
	if hdr.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported jwt alg %q", hdr.Alg)
	}
	if strings.TrimSpace(hdr.Kid) == "" {
		return nil, errors.New("missing kid")
	}

	payloadJSON, err := enc.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid jwt payload")
	}
	var claims cognitoClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, errors.New("invalid jwt payload")
	}

	if claims.Iss != v.iss {
		return nil, errors.New("invalid issuer")
	}
	if claims.Exp == 0 || now.Unix() >= claims.Exp {
		return nil, errors.New("token expired")
	}
	if claims.TokenUse != "id" {
		return nil, errors.New("invalid token_use")
	}
	if claims.Aud != v.cfg.CognitoAppClientID {
		return nil, errors.New("invalid audience")
	}
	if strings.TrimSpace(claims.Sub) == "" {
		return nil, errors.New("missing sub")
	}

	pub, err := v.getKey(hdr.Kid, now)
	if err != nil {
		return nil, err
	}

	signingInput := parts[0] + "." + parts[1]
	sig, err := enc.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid jwt signature")
	}
	h := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], sig); err != nil {
		return nil, errors.New("invalid jwt signature")
	}

	return &claims, nil
}

func (v *cognitoJWTVerifier) getKey(kid string, now time.Time) (*rsa.PublicKey, error) {
	v.mu.RLock()
	if k, ok := v.keys[kid]; ok && !v.cacheExpired(now) {
		v.mu.RUnlock()
		return k, nil
	}
	v.mu.RUnlock()

	if err := v.refreshKeys(now); err != nil {
		v.mu.RLock()
		k, ok := v.keys[kid]
		v.mu.RUnlock()
		if ok {
			return k, nil
		}
		return nil, err
	}

	v.mu.RLock()
	k, ok := v.keys[kid]
	v.mu.RUnlock()
	if !ok {
		return nil, errors.New("unknown kid")
	}
	return k, nil
}

func (v *cognitoJWTVerifier) cacheExpired(now time.Time) bool {
	if v.fetchedAt.IsZero() {
		return true
	}
	return now.Sub(v.fetchedAt) > v.cacheTTL
}

func (v *cognitoJWTVerifier) refreshKeys(now time.Time) error {
	v.mu.Lock()
	if !v.cacheExpired(now) {
		v.mu.Unlock()
		return nil
	}
	v.mu.Unlock()

	resp, err := v.client.Get(v.jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("jwks fetch failed: %s", resp.Status)
	}

	var doc jwksDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return err
	}

	keys := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, k := range doc.Keys {
		if k.Kid == "" {
			continue
		}
		pub, err := jwkToRSAPublicKey(k)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}

	v.mu.Lock()
	v.keys = keys
	v.fetchedAt = now
	v.mu.Unlock()
	return nil
}

func jwkToRSAPublicKey(k jwkKey) (*rsa.PublicKey, error) {
	if k.Kty != "RSA" {
		return nil, errors.New("unsupported kty")
	}
	if len(k.X5c) > 0 {
		certDER, err := base64.StdEncoding.DecodeString(k.X5c[0])
		if err == nil {
			cert, err := x509.ParseCertificate(certDER)
			if err == nil {
				if pub, ok := cert.PublicKey.(*rsa.PublicKey); ok {
					return pub, nil
				}
			}
		}
	}

	enc := base64.RawURLEncoding
	nBytes, err := enc.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := enc.DecodeString(k.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	if e == 0 {
		return nil, errors.New("invalid exponent")
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}
