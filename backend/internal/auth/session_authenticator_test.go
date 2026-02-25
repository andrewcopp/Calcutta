package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

// --- stubs ---

type stubSessionRepo struct {
	session *models.AuthSession
	err     error
}

func (s *stubSessionRepo) CreateSession(_ context.Context, _, _, _, _ string, _ time.Time) (string, error) {
	return "", nil
}
func (s *stubSessionRepo) GetSessionByID(_ context.Context, _ string) (*models.AuthSession, error) {
	return s.session, s.err
}
func (s *stubSessionRepo) GetSessionByRefreshTokenHash(_ context.Context, _ string) (*models.AuthSession, error) {
	return nil, nil
}
func (s *stubSessionRepo) RotateRefreshToken(_ context.Context, _, _ string, _ time.Time) error {
	return nil
}
func (s *stubSessionRepo) RevokeSession(_ context.Context, _ string) error { return nil }
func (s *stubSessionRepo) RevokeAllSessionsForUser(_ context.Context, _ string) error { return nil }
func (s *stubSessionRepo) IsUserActive(_ context.Context, _ string) (bool, error) {
	return false, nil
}

type stubUserRepo struct {
	user    *models.User
	err     error
	byExtFn func(provider, providerID string) (*models.User, error)
}

func (s *stubUserRepo) GetByEmail(_ context.Context, _ string) (*models.User, error) {
	return s.user, s.err
}
func (s *stubUserRepo) GetByID(_ context.Context, _ string) (*models.User, error) {
	return s.user, s.err
}
func (s *stubUserRepo) GetByExternalProvider(_ context.Context, provider, providerID string) (*models.User, error) {
	if s.byExtFn != nil {
		return s.byExtFn(provider, providerID)
	}
	return nil, nil
}
func (s *stubUserRepo) Create(_ context.Context, _ *models.User) error { return nil }
func (s *stubUserRepo) Update(_ context.Context, _ *models.User) error { return nil }

// --- helpers ---

func mustTokenManager(t *testing.T) *auth.TokenManager {
	t.Helper()
	tm, err := auth.NewTokenManager("test-secret-that-is-long-enough", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

func issueToken(t *testing.T, tm *auth.TokenManager, userID, sessionID string, now time.Time) string {
	t.Helper()
	tok, _, err := tm.IssueAccessToken(userID, sessionID, now)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func activeUser(id string) *models.User {
	return &models.User{ID: id, Status: "active"}
}

func inactiveUser(id string) *models.User {
	return &models.User{ID: id, Status: "inactive"}
}

func validSession(userID string, expiresAt time.Time) *models.AuthSession {
	return &models.AuthSession{ID: "sess-1", UserID: userID, ExpiresAt: expiresAt}
}

// --- tests ---

func TestThatSessionAuthReturnsNilNilForMalformedToken(t *testing.T) {
	// GIVEN a session authenticator
	tm := mustTokenManager(t)
	a := auth.NewSessionAuthenticator(tm, &stubSessionRepo{}, &stubUserRepo{user: activeUser("u1")})

	// WHEN authenticating a malformed token
	identity, err := a.Authenticate(context.Background(), "not-a-valid-token")

	// THEN it returns nil, nil (not my token)
	if identity != nil || err != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", identity, err)
	}
}

func TestThatSessionAuthReturnsNilNilForExpiredToken(t *testing.T) {
	// GIVEN a token that was issued in the past and has expired
	tm := mustTokenManager(t)
	pastTime := time.Now().Add(-1 * time.Hour)
	tok := issueToken(t, tm, "u1", "sess-1", pastTime)

	sess := validSession("u1", time.Now().Add(24*time.Hour))
	a := auth.NewSessionAuthenticator(tm, &stubSessionRepo{session: sess}, &stubUserRepo{user: activeUser("u1")})

	// WHEN authenticating the expired token
	identity, err := a.Authenticate(context.Background(), tok)

	// THEN it returns nil, nil
	if identity != nil || err != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", identity, err)
	}
}

func TestThatSessionAuthReturnsNilNilForRevokedSession(t *testing.T) {
	// GIVEN a valid token but the session has been revoked
	tm := mustTokenManager(t)
	now := time.Now()
	tok := issueToken(t, tm, "u1", "sess-1", now)

	revoked := now
	sess := &models.AuthSession{ID: "sess-1", UserID: "u1", ExpiresAt: now.Add(24 * time.Hour), RevokedAt: &revoked}
	a := auth.NewSessionAuthenticator(tm, &stubSessionRepo{session: sess}, &stubUserRepo{user: activeUser("u1")})

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), tok)

	// THEN it returns nil, nil
	if identity != nil || err != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", identity, err)
	}
}

func TestThatSessionAuthReturnsNilNilForExpiredSession(t *testing.T) {
	// GIVEN a valid token but the session itself has expired
	tm := mustTokenManager(t)
	now := time.Now()
	tok := issueToken(t, tm, "u1", "sess-1", now)

	sess := validSession("u1", now.Add(-1*time.Minute))
	a := auth.NewSessionAuthenticator(tm, &stubSessionRepo{session: sess}, &stubUserRepo{user: activeUser("u1")})

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), tok)

	// THEN it returns nil, nil
	if identity != nil || err != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", identity, err)
	}
}

func TestThatSessionAuthReturnsNilNilForInactiveUser(t *testing.T) {
	// GIVEN a valid token and session, but the user is inactive
	tm := mustTokenManager(t)
	now := time.Now()
	tok := issueToken(t, tm, "u1", "sess-1", now)

	sess := validSession("u1", now.Add(24*time.Hour))
	a := auth.NewSessionAuthenticator(tm, &stubSessionRepo{session: sess}, &stubUserRepo{user: inactiveUser("u1")})

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), tok)

	// THEN it returns nil, nil
	if identity != nil || err != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", identity, err)
	}
}

func TestThatSessionAuthReturnsIdentityForValidSession(t *testing.T) {
	// GIVEN a valid token, active session, and active user
	tm := mustTokenManager(t)
	now := time.Now()
	tok := issueToken(t, tm, "u1", "sess-1", now)

	sess := validSession("u1", now.Add(24*time.Hour))
	a := auth.NewSessionAuthenticator(tm, &stubSessionRepo{session: sess}, &stubUserRepo{user: activeUser("u1")})

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), tok)

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
	if identity.SessionID != "sess-1" {
		t.Errorf("expected SessionID sess-1, got %s", identity.SessionID)
	}
}

func TestThatSessionAuthReturnsErrorForDBFailure(t *testing.T) {
	// GIVEN a valid token but the session lookup fails
	tm := mustTokenManager(t)
	now := time.Now()
	tok := issueToken(t, tm, "u1", "sess-1", now)

	dbErr := errors.New("db connection failed")
	a := auth.NewSessionAuthenticator(tm, &stubSessionRepo{err: dbErr}, &stubUserRepo{user: activeUser("u1")})

	// WHEN authenticating
	identity, err := a.Authenticate(context.Background(), tok)

	// THEN it returns the error
	if identity != nil {
		t.Errorf("expected nil identity, got %v", identity)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Verify interface compliance at compile time.
var _ ports.Authenticator = (*auth.SessionAuthenticator)(nil)
