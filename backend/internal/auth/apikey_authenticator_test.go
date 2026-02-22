package auth_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

// --- stubs ---

type stubAPIKeyReader struct {
	key *models.APIKey
	err error
}

func (s *stubAPIKeyReader) GetActiveByHash(_ context.Context, _ string, _ time.Time) (*models.APIKey, error) {
	return s.key, s.err
}

// --- helpers ---

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// --- tests ---

func TestThatAPIKeyAuthReturnsNilNilForUnknownKey(t *testing.T) {
	// GIVEN no matching API key
	a := auth.NewAPIKeyAuthenticator(&stubAPIKeyReader{}, &stubUserRepo{user: activeUser("u1")})

	// WHEN authenticating with an unknown token
	identity, err := a.Authenticate(context.Background(), "unknown-api-key")

	// THEN it returns nil, nil
	if identity != nil || err != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", identity, err)
	}
}

func TestThatAPIKeyAuthReturnsIdentityForValidKey(t *testing.T) {
	// GIVEN a valid API key mapping to an active user
	key := &models.APIKey{ID: "k1", UserID: "u1"}
	a := auth.NewAPIKeyAuthenticator(
		&stubAPIKeyReader{key: key},
		&stubUserRepo{user: activeUser("u1")},
	)

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), "my-api-key")

	// THEN it returns the correct identity
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity == nil {
		t.Fatal("expected identity, got nil")
	}
	if identity.UserID != "u1" {
		t.Errorf("expected UserID u1, got %s", identity.UserID)
	}
}

func TestThatAPIKeyAuthReturnsNilNilForInactiveUser(t *testing.T) {
	// GIVEN a valid API key but the user is inactive
	key := &models.APIKey{ID: "k1", UserID: "u1"}
	a := auth.NewAPIKeyAuthenticator(
		&stubAPIKeyReader{key: key},
		&stubUserRepo{user: inactiveUser("u1")},
	)

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), "my-api-key")

	// THEN it returns nil, nil
	if identity != nil || err != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", identity, err)
	}
}

func TestThatAPIKeyAuthReturnsErrorForDBFailure(t *testing.T) {
	// GIVEN the API key lookup fails
	dbErr := errors.New("db connection failed")
	a := auth.NewAPIKeyAuthenticator(
		&stubAPIKeyReader{err: dbErr},
		&stubUserRepo{user: activeUser("u1")},
	)

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), "my-api-key")

	// THEN it returns the error
	if identity != nil {
		t.Errorf("expected nil identity, got %v", identity)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Verify interface compliance at compile time.
var _ ports.Authenticator = (*auth.APIKeyAuthenticator)(nil)
