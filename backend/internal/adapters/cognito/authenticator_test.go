package cognito_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/cognito"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

// --- stubs ---

type stubVerifier struct {
	claims *cognito.Claims
	err    error
}

func (s *stubVerifier) Verify(_ string, _ time.Time) (*cognito.Claims, error) {
	return s.claims, s.err
}

type stubUserRepo struct {
	user      *models.User
	err       error
	byExtUser *models.User
	byExtErr  error
	created   *models.User
}

func (s *stubUserRepo) GetByEmail(_ context.Context, _ string) (*models.User, error) {
	return s.user, s.err
}
func (s *stubUserRepo) GetByID(_ context.Context, _ string) (*models.User, error) {
	return s.user, s.err
}
func (s *stubUserRepo) GetByExternalProvider(_ context.Context, _, _ string) (*models.User, error) {
	return s.byExtUser, s.byExtErr
}
func (s *stubUserRepo) GetByIDs(_ context.Context, _ []string) ([]*models.User, error) {
	return nil, nil
}
func (s *stubUserRepo) Create(_ context.Context, u *models.User) error {
	s.created = u
	return nil
}
func (s *stubUserRepo) Update(_ context.Context, _ *models.User) error { return nil }

// --- helpers ---

func validClaims() *cognito.Claims {
	return &cognito.Claims{
		Sub:        "cognito-sub-123",
		Email:      "test@example.com",
		GivenName:  "Test",
		FamilyName: "User",
	}
}

func activeUser(id string) *models.User {
	return &models.User{ID: id, Status: "active"}
}

func inactiveUser(id string) *models.User {
	return &models.User{ID: id, Status: "inactive"}
}

// --- tests ---

func TestThatCognitoAuthReturnsNilNilForInvalidToken(t *testing.T) {
	// GIVEN a verifier that rejects the token
	v := &stubVerifier{err: errors.New("invalid")}
	a := cognito.NewAuthenticator(v, &stubUserRepo{}, false, false)

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), "bad-token")

	// THEN it returns nil, nil
	if identity != nil || err != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", identity, err)
	}
}

func TestThatCognitoAuthReturnsIdentityForKnownUser(t *testing.T) {
	// GIVEN a valid token and a known user mapped to the cognito sub
	v := &stubVerifier{claims: validClaims()}
	users := &stubUserRepo{byExtUser: activeUser("u1")}
	a := cognito.NewAuthenticator(v, users, false, false)

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), "valid-token")

	// THEN it returns the user identity
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

func TestThatCognitoAuthAutoProvisionsNewUser(t *testing.T) {
	// GIVEN a valid token, no existing user, and auto-provision enabled
	v := &stubVerifier{claims: validClaims()}
	users := &stubUserRepo{} // byExtUser is nil
	a := cognito.NewAuthenticator(v, users, true, false)

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), "valid-token")

	// THEN it provisions a new user and returns its identity
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity == nil {
		t.Fatal("expected identity, got nil")
	}
	if users.created == nil {
		t.Fatal("expected user to be created")
	}
}

func TestThatCognitoAuthReturnsNilNilWhenNoProvisionAndUserMissing(t *testing.T) {
	// GIVEN a valid token but no user and auto-provision disabled
	v := &stubVerifier{claims: validClaims()}
	users := &stubUserRepo{}
	a := cognito.NewAuthenticator(v, users, false, false)

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), "valid-token")

	// THEN it returns nil, nil
	if identity != nil || err != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", identity, err)
	}
}

func TestThatCognitoAuthUsesUnprovisionedFallback(t *testing.T) {
	// GIVEN a valid token, no external provider user, but allowUnprovisioned and user exists by ID
	v := &stubVerifier{claims: validClaims()}
	users := &stubUserRepo{user: activeUser("cognito-sub-123")} // GetByID returns this
	a := cognito.NewAuthenticator(v, users, false, true)

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), "valid-token")

	// THEN it returns the identity using the cognito sub
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity == nil {
		t.Fatal("expected identity, got nil")
	}
	if identity.UserID != "cognito-sub-123" {
		t.Errorf("expected UserID cognito-sub-123, got %s", identity.UserID)
	}
}

func TestThatCognitoAuthReturnsErrorForDBFailure(t *testing.T) {
	// GIVEN a valid token but the external provider lookup fails
	v := &stubVerifier{claims: validClaims()}
	dbErr := errors.New("db connection failed")
	users := &stubUserRepo{byExtErr: dbErr}
	a := cognito.NewAuthenticator(v, users, false, false)

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), "valid-token")

	// THEN it returns the error
	if identity != nil {
		t.Errorf("expected nil identity, got %v", identity)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Verify interface compliance at compile time.
var _ ports.Authenticator = (*cognito.Authenticator)(nil)
