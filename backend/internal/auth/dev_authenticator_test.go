package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/auth"
)

func TestThatDevAuthReturnsNilWhenHeaderMissing(t *testing.T) {
	// GIVEN a request with no X-Dev-User header
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	// WHEN calling DevAuthenticate
	identity, err := auth.DevAuthenticate(r, &stubUserRepo{user: activeUser("u1")})

	// THEN it returns nil, nil
	if identity != nil || err != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", identity, err)
	}
}

func TestThatDevAuthReturnsIdentityForActiveUser(t *testing.T) {
	// GIVEN a request with X-Dev-User header and an active user
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Dev-User", "u1")

	// WHEN calling DevAuthenticate
	identity, err := auth.DevAuthenticate(r, &stubUserRepo{user: activeUser("u1")})

	// THEN it returns the identity
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

func TestThatDevAuthReturnsNilForInactiveUser(t *testing.T) {
	// GIVEN a request with X-Dev-User header but the user is inactive
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Dev-User", "u1")

	// WHEN calling DevAuthenticate
	identity, err := auth.DevAuthenticate(r, &stubUserRepo{user: inactiveUser("u1")})

	// THEN it returns nil, nil
	if identity != nil || err != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", identity, err)
	}
}
