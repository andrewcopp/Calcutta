package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

var _ ports.Authenticator = (*APIKeyAuthenticator)(nil)

// APIKeyAuthenticator validates bearer tokens as API keys via SHA-256 hash lookup.
type APIKeyAuthenticator struct {
	keys  ports.APIKeyReader
	users ports.UserRepository
	now   func() time.Time
}

func NewAPIKeyAuthenticator(keys ports.APIKeyReader, users ports.UserRepository) *APIKeyAuthenticator {
	return &APIKeyAuthenticator{keys: keys, users: users, now: time.Now}
}

func (a *APIKeyAuthenticator) Authenticate(ctx context.Context, token string) (*ports.AuthIdentity, error) {
	raw := token
	if strings.HasPrefix(raw, "mmk_") {
		raw = raw[4:]
	}
	sum := sha256.Sum256([]byte(raw))
	h := hex.EncodeToString(sum[:])

	k, err := a.keys.GetActiveByHash(ctx, h, a.now())
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, nil
	}

	user, err := a.users.GetByID(ctx, k.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Status != "active" {
		return nil, nil
	}

	return &ports.AuthIdentity{UserID: k.UserID}, nil
}
