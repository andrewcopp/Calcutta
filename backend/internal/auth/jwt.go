package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type TokenManager struct {
	secret    []byte
	accessTTL time.Duration
}

type AccessTokenClaims struct {
	Sub string `json:"sub"`
	Sid string `json:"sid"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

func NewTokenManager(secret string, accessTTL time.Duration) (*TokenManager, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}
	if accessTTL <= 0 {
		return nil, errors.New("access token ttl must be > 0")
	}
	return &TokenManager{secret: []byte(secret), accessTTL: accessTTL}, nil
}

func (m *TokenManager) IssueAccessToken(userID, sessionID string, now time.Time) (string, *AccessTokenClaims, error) {
	if strings.TrimSpace(userID) == "" {
		return "", nil, errors.New("user id required")
	}
	if strings.TrimSpace(sessionID) == "" {
		return "", nil, errors.New("session id required")
	}
	if now.IsZero() {
		now = time.Now()
	}

	claims := &AccessTokenClaims{
		Sub: userID,
		Sid: sessionID,
		Iat: now.Unix(),
		Exp: now.Add(m.accessTTL).Unix(),
	}

	headerJSON, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", nil, err
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", nil, err
	}

	enc := base64.RawURLEncoding
	header := enc.EncodeToString(headerJSON)
	payload := enc.EncodeToString(payloadJSON)
	signingInput := header + "." + payload

	sig := signHS256([]byte(signingInput), m.secret)
	token := signingInput + "." + enc.EncodeToString(sig)
	return token, claims, nil
}

func (m *TokenManager) VerifyAccessToken(token string, now time.Time) (*AccessTokenClaims, error) {
	if now.IsZero() {
		now = time.Now()
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	enc := base64.RawURLEncoding
	signingInput := parts[0] + "." + parts[1]
	sig, err := enc.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid token signature encoding")
	}

	expected := signHS256([]byte(signingInput), m.secret)
	if !hmac.Equal(sig, expected) {
		return nil, errors.New("invalid token signature")
	}

	payloadJSON, err := enc.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid token payload encoding")
	}

	var claims AccessTokenClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, errors.New("invalid token payload")
	}

	if claims.Sub == "" || claims.Sid == "" {
		return nil, errors.New("invalid token claims")
	}
	if claims.Exp == 0 {
		return nil, errors.New("missing exp")
	}
	if now.Unix() >= claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

func signHS256(data, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	_, _ = h.Write(data)
	return h.Sum(nil)
}
