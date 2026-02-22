package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------------------------
// Fake UserRepository
// ---------------------------------------------------------------------------

type fakeUserRepo struct {
	byEmail  map[string]*models.User
	byID     map[string]*models.User
	created  []*models.User
	createFn func(ctx context.Context, u *models.User) error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		byEmail: make(map[string]*models.User),
		byID:    make(map[string]*models.User),
	}
}

func (r *fakeUserRepo) GetByEmail(_ context.Context, email string) (*models.User, error) {
	return r.byEmail[email], nil
}

func (r *fakeUserRepo) GetByID(_ context.Context, id string) (*models.User, error) {
	return r.byID[id], nil
}

func (r *fakeUserRepo) GetByExternalProvider(_ context.Context, _, _ string) (*models.User, error) {
	return nil, nil
}

func (r *fakeUserRepo) Create(_ context.Context, u *models.User) error {
	if r.createFn != nil {
		return r.createFn(nil, u)
	}
	r.created = append(r.created, u)
	return nil
}

func (r *fakeUserRepo) Update(_ context.Context, _ *models.User) error {
	return nil
}

// ---------------------------------------------------------------------------
// Fake AuthSessionRepository
// ---------------------------------------------------------------------------

type fakeAuthRepo struct {
	activeUsers map[string]bool
	sessions    map[string]*models.AuthSession // keyed by refresh token hash
	sessionsID  map[string]*models.AuthSession // keyed by session ID
	revoked     []string
	rotated     []string
	nextSessID  string
	createErr   error
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{
		activeUsers: make(map[string]bool),
		sessions:    make(map[string]*models.AuthSession),
		sessionsID:  make(map[string]*models.AuthSession),
		nextSessID:  "sess-001",
	}
}

func (r *fakeAuthRepo) CreateSession(_ context.Context, userID, refreshTokenHash, userAgent, ipAddress string, expiresAt time.Time) (string, error) {
	if r.createErr != nil {
		return "", r.createErr
	}
	sess := &models.AuthSession{
		ID:               r.nextSessID,
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        expiresAt,
	}
	r.sessions[refreshTokenHash] = sess
	r.sessionsID[r.nextSessID] = sess
	return r.nextSessID, nil
}

func (r *fakeAuthRepo) GetSessionByID(_ context.Context, id string) (*models.AuthSession, error) {
	return r.sessionsID[id], nil
}

func (r *fakeAuthRepo) GetSessionByRefreshTokenHash(_ context.Context, hash string) (*models.AuthSession, error) {
	return r.sessions[hash], nil
}

func (r *fakeAuthRepo) RotateRefreshToken(_ context.Context, sessionID, newHash string, newExpiresAt time.Time) error {
	r.rotated = append(r.rotated, sessionID)
	return nil
}

func (r *fakeAuthRepo) RevokeSession(_ context.Context, sessionID string) error {
	r.revoked = append(r.revoked, sessionID)
	return nil
}

func (r *fakeAuthRepo) IsUserActive(_ context.Context, userID string) (bool, error) {
	return r.activeUsers[userID], nil
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

var fixedNow = time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)

func mustTokenManager(t *testing.T) *coreauth.TokenManager {
	t.Helper()
	tm, err := coreauth.NewTokenManager("test-secret-at-least-32-bytes-long", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to create TokenManager: %v", err)
	}
	return tm
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hash)
}

func newTestService(userRepo *fakeUserRepo, authRepo *fakeAuthRepo, tokenMgr *coreauth.TokenManager) *Service {
	return New(userRepo, authRepo, tokenMgr, 7*24*time.Hour)
}

func activeUser(t *testing.T, id, email, password string) *models.User {
	t.Helper()
	hash := mustHashPassword(t, password)
	return &models.User{
		ID:           id,
		Email:        &email,
		FirstName:    "Test",
		LastName:     "User",
		Status:       "active",
		PasswordHash: &hash,
		CreatedAt:    fixedNow.Add(-24 * time.Hour),
		UpdatedAt:    fixedNow.Add(-24 * time.Hour),
	}
}

// ---------------------------------------------------------------------------
// Login tests
// ---------------------------------------------------------------------------

func TestThatLoginReturnsUnauthorizedWhenEmailNotFound(t *testing.T) {
	// GIVEN a service with no users
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging in with a nonexistent email
	_, err := svc.Login(context.Background(), "nobody@example.com", "password", "agent", "127.0.0.1", fixedNow)

	// THEN an UnauthorizedError is returned
	var ue *apperrors.UnauthorizedError
	if !errors.As(err, &ue) {
		t.Errorf("expected UnauthorizedError, got %T: %v", err, err)
	}
}

func TestThatLoginReturnsUnauthorizedWhenUserIsInactive(t *testing.T) {
	// GIVEN a user that exists but is not active
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	user := activeUser(t, "user-1", "inactive@example.com", "password")
	userRepo.byEmail["inactive@example.com"] = user
	// authRepo.activeUsers does NOT include user-1
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging in
	_, err := svc.Login(context.Background(), "inactive@example.com", "password", "agent", "127.0.0.1", fixedNow)

	// THEN an UnauthorizedError is returned
	var ue *apperrors.UnauthorizedError
	if !errors.As(err, &ue) {
		t.Errorf("expected UnauthorizedError, got %T: %v", err, err)
	}
}

func TestThatLoginReturnsUnauthorizedWhenPasswordHashIsNil(t *testing.T) {
	// GIVEN a user with no password hash (e.g. external provider only)
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	email := "nopw@example.com"
	user := &models.User{ID: "user-1", Email: &email, Status: "active"}
	userRepo.byEmail[email] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging in with any password
	_, err := svc.Login(context.Background(), email, "anything", "agent", "127.0.0.1", fixedNow)

	// THEN an UnauthorizedError is returned
	var ue *apperrors.UnauthorizedError
	if !errors.As(err, &ue) {
		t.Errorf("expected UnauthorizedError, got %T: %v", err, err)
	}
}

func TestThatLoginReturnsUnauthorizedWhenPasswordHashIsEmpty(t *testing.T) {
	// GIVEN a user with an empty password hash
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	email := "empty@example.com"
	empty := ""
	user := &models.User{ID: "user-1", Email: &email, Status: "active", PasswordHash: &empty}
	userRepo.byEmail[email] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging in
	_, err := svc.Login(context.Background(), email, "anything", "agent", "127.0.0.1", fixedNow)

	// THEN an UnauthorizedError is returned
	var ue *apperrors.UnauthorizedError
	if !errors.As(err, &ue) {
		t.Errorf("expected UnauthorizedError, got %T: %v", err, err)
	}
}

func TestThatLoginReturnsUnauthorizedWhenPasswordIsWrong(t *testing.T) {
	// GIVEN a user with a valid password hash
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	user := activeUser(t, "user-1", "user@example.com", "correct-password")
	userRepo.byEmail["user@example.com"] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging in with the wrong password
	_, err := svc.Login(context.Background(), "user@example.com", "wrong-password", "agent", "127.0.0.1", fixedNow)

	// THEN an UnauthorizedError is returned
	var ue *apperrors.UnauthorizedError
	if !errors.As(err, &ue) {
		t.Errorf("expected UnauthorizedError, got %T: %v", err, err)
	}
}

func TestThatLoginReturnsResultOnSuccess(t *testing.T) {
	// GIVEN a valid active user with correct credentials
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	user := activeUser(t, "user-1", "user@example.com", "correct-password")
	userRepo.byEmail["user@example.com"] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging in with the correct password
	result, err := svc.Login(context.Background(), "user@example.com", "correct-password", "agent", "127.0.0.1", fixedNow)

	// THEN a non-nil result is returned without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestThatLoginReturnsNonEmptyAccessToken(t *testing.T) {
	// GIVEN a valid active user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	user := activeUser(t, "user-1", "user@example.com", "correct-password")
	userRepo.byEmail["user@example.com"] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging in
	result, err := svc.Login(context.Background(), "user@example.com", "correct-password", "agent", "127.0.0.1", fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the access token is non-empty
	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestThatLoginReturnsNonEmptyRefreshToken(t *testing.T) {
	// GIVEN a valid active user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	user := activeUser(t, "user-1", "user@example.com", "correct-password")
	userRepo.byEmail["user@example.com"] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging in
	result, err := svc.Login(context.Background(), "user@example.com", "correct-password", "agent", "127.0.0.1", fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the refresh token is non-empty
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestThatLoginReturnsCorrectUser(t *testing.T) {
	// GIVEN a valid active user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	user := activeUser(t, "user-1", "user@example.com", "correct-password")
	userRepo.byEmail["user@example.com"] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging in
	result, err := svc.Login(context.Background(), "user@example.com", "correct-password", "agent", "127.0.0.1", fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the returned user ID matches
	if result.User.ID != "user-1" {
		t.Errorf("expected user ID user-1, got %s", result.User.ID)
	}
}

func TestThatLoginSetsRefreshExpiresAtUsingTTL(t *testing.T) {
	// GIVEN a service with 7-day refresh TTL
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	user := activeUser(t, "user-1", "user@example.com", "correct-password")
	userRepo.byEmail["user@example.com"] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging in at fixedNow
	result, err := svc.Login(context.Background(), "user@example.com", "correct-password", "agent", "127.0.0.1", fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN refresh expiration is fixedNow + 7 days
	want := fixedNow.Add(7 * 24 * time.Hour)
	if !result.RefreshExpiresAt.Equal(want) {
		t.Errorf("expected refresh expiry %v, got %v", want, result.RefreshExpiresAt)
	}
}

func TestThatLoginReturnsErrorWhenTokenManagerIsNil(t *testing.T) {
	// GIVEN a service with nil token manager
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, nil)

	// WHEN logging in
	_, err := svc.Login(context.Background(), "user@example.com", "password", "agent", "127.0.0.1", fixedNow)

	// THEN an error about auth not configured is returned
	if err == nil || err.Error() != "auth is not configured" {
		t.Errorf("expected 'auth is not configured' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Signup tests
// ---------------------------------------------------------------------------

func TestThatSignupReturnsAlreadyExistsWhenEmailIsTaken(t *testing.T) {
	// GIVEN a user already exists with the given email
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	existing := activeUser(t, "user-1", "taken@example.com", "password")
	userRepo.byEmail["taken@example.com"] = existing
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN signing up with the same email
	_, err := svc.Signup(context.Background(), "taken@example.com", "New", "User", "password", "agent", "127.0.0.1", fixedNow)

	// THEN an AlreadyExistsError is returned
	var ae *apperrors.AlreadyExistsError
	if !errors.As(err, &ae) {
		t.Errorf("expected AlreadyExistsError, got %T: %v", err, err)
	}
}

func TestThatSignupReturnsResultOnSuccess(t *testing.T) {
	// GIVEN no existing user with the email
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	authRepo.activeUsers["ignored"] = true // not relevant here
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN signing up
	result, err := svc.Signup(context.Background(), "new@example.com", "First", "Last", "password", "agent", "127.0.0.1", fixedNow)

	// THEN a non-nil result is returned without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestThatSignupCreatesUserWithCorrectEmail(t *testing.T) {
	// GIVEN no existing user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN signing up
	result, err := svc.Signup(context.Background(), "new@example.com", "First", "Last", "password", "agent", "127.0.0.1", fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the created user has the correct email
	if result.User.Email == nil || *result.User.Email != "new@example.com" {
		t.Errorf("expected email new@example.com, got %v", result.User.Email)
	}
}

func TestThatSignupCreatesUserWithCorrectName(t *testing.T) {
	// GIVEN no existing user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN signing up
	result, err := svc.Signup(context.Background(), "new@example.com", "Alice", "Smith", "password", "agent", "127.0.0.1", fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the created user has the correct first and last name
	got := result.User.FirstName + " " + result.User.LastName
	if got != "Alice Smith" {
		t.Errorf("expected name 'Alice Smith', got %q", got)
	}
}

func TestThatSignupCreatesUserWithActiveStatus(t *testing.T) {
	// GIVEN no existing user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN signing up
	result, err := svc.Signup(context.Background(), "new@example.com", "First", "Last", "password", "agent", "127.0.0.1", fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the user status is "active"
	if result.User.Status != "active" {
		t.Errorf("expected status 'active', got %q", result.User.Status)
	}
}

func TestThatSignupCreatesUserWithHashedPassword(t *testing.T) {
	// GIVEN no existing user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN signing up with password "my-secret"
	result, err := svc.Signup(context.Background(), "new@example.com", "First", "Last", "my-secret", "agent", "127.0.0.1", fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the password hash verifies against the original password
	if result.User.PasswordHash == nil {
		t.Fatal("expected non-nil password hash")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*result.User.PasswordHash), []byte("my-secret")); err != nil {
		t.Errorf("password hash does not match original password: %v", err)
	}
}

func TestThatSignupReturnsNonEmptyAccessToken(t *testing.T) {
	// GIVEN no existing user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN signing up
	result, err := svc.Signup(context.Background(), "new@example.com", "First", "Last", "password", "agent", "127.0.0.1", fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the access token is non-empty
	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestThatSignupSetsTimestampsToNow(t *testing.T) {
	// GIVEN no existing user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN signing up at fixedNow
	result, err := svc.Signup(context.Background(), "new@example.com", "First", "Last", "password", "agent", "127.0.0.1", fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN both created_at and updated_at match fixedNow
	if !result.User.CreatedAt.Equal(fixedNow) || !result.User.UpdatedAt.Equal(fixedNow) {
		t.Errorf("expected timestamps %v, got created=%v updated=%v", fixedNow, result.User.CreatedAt, result.User.UpdatedAt)
	}
}

func TestThatSignupReturnsErrorWhenTokenManagerIsNil(t *testing.T) {
	// GIVEN a service with nil token manager
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, nil)

	// WHEN signing up
	_, err := svc.Signup(context.Background(), "new@example.com", "First", "Last", "password", "agent", "127.0.0.1", fixedNow)

	// THEN an error about auth not configured is returned
	if err == nil || err.Error() != "auth is not configured" {
		t.Errorf("expected 'auth is not configured' error, got: %v", err)
	}
}

func TestThatSignupPersistsUserViaRepository(t *testing.T) {
	// GIVEN no existing user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN signing up
	_, err := svc.Signup(context.Background(), "new@example.com", "First", "Last", "password", "agent", "127.0.0.1", fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the user repository received one Create call
	if len(userRepo.created) != 1 {
		t.Errorf("expected 1 created user, got %d", len(userRepo.created))
	}
}

// ---------------------------------------------------------------------------
// Refresh tests
// ---------------------------------------------------------------------------

func TestThatRefreshReturnsErrorWhenTokenIsEmpty(t *testing.T) {
	// GIVEN a service
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN refreshing with an empty token
	_, err := svc.Refresh(context.Background(), "", fixedNow)

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for empty refresh token")
	}
}

func TestThatRefreshReturnsUnauthorizedWhenSessionNotFound(t *testing.T) {
	// GIVEN a service with no sessions
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN refreshing with a token that has no matching session
	_, err := svc.Refresh(context.Background(), "nonexistent-token", fixedNow)

	// THEN an UnauthorizedError is returned
	var ue *apperrors.UnauthorizedError
	if !errors.As(err, &ue) {
		t.Errorf("expected UnauthorizedError, got %T: %v", err, err)
	}
}

func TestThatRefreshReturnsUnauthorizedWhenSessionIsRevoked(t *testing.T) {
	// GIVEN a session that has been revoked
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	token := "some-refresh-token"
	hash := coreauth.HashRefreshToken(token)
	revokedAt := fixedNow.Add(-1 * time.Hour)
	authRepo.sessions[hash] = &models.AuthSession{
		ID:               "sess-1",
		UserID:           "user-1",
		RefreshTokenHash: hash,
		ExpiresAt:        fixedNow.Add(24 * time.Hour),
		RevokedAt:        &revokedAt,
	}
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN refreshing with that token
	_, err := svc.Refresh(context.Background(), token, fixedNow)

	// THEN an UnauthorizedError is returned
	var ue *apperrors.UnauthorizedError
	if !errors.As(err, &ue) {
		t.Errorf("expected UnauthorizedError, got %T: %v", err, err)
	}
}

func TestThatRefreshReturnsUnauthorizedWhenSessionIsExpired(t *testing.T) {
	// GIVEN a session that has expired
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	token := "some-refresh-token"
	hash := coreauth.HashRefreshToken(token)
	authRepo.sessions[hash] = &models.AuthSession{
		ID:               "sess-1",
		UserID:           "user-1",
		RefreshTokenHash: hash,
		ExpiresAt:        fixedNow.Add(-1 * time.Hour), // expired 1 hour ago
	}
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN refreshing at fixedNow (after expiry)
	_, err := svc.Refresh(context.Background(), token, fixedNow)

	// THEN an UnauthorizedError is returned
	var ue *apperrors.UnauthorizedError
	if !errors.As(err, &ue) {
		t.Errorf("expected UnauthorizedError, got %T: %v", err, err)
	}
}

func TestThatRefreshReturnsUnauthorizedWhenUserNotFound(t *testing.T) {
	// GIVEN a valid session but the user no longer exists
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	token := "some-refresh-token"
	hash := coreauth.HashRefreshToken(token)
	authRepo.sessions[hash] = &models.AuthSession{
		ID:               "sess-1",
		UserID:           "user-gone",
		RefreshTokenHash: hash,
		ExpiresAt:        fixedNow.Add(24 * time.Hour),
	}
	// userRepo has no user with ID "user-gone"
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN refreshing
	_, err := svc.Refresh(context.Background(), token, fixedNow)

	// THEN an UnauthorizedError is returned
	var ue *apperrors.UnauthorizedError
	if !errors.As(err, &ue) {
		t.Errorf("expected UnauthorizedError, got %T: %v", err, err)
	}
}

func TestThatRefreshReturnsUnauthorizedWhenUserIsInactive(t *testing.T) {
	// GIVEN a valid session but the user is no longer active
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	token := "some-refresh-token"
	hash := coreauth.HashRefreshToken(token)
	authRepo.sessions[hash] = &models.AuthSession{
		ID:               "sess-1",
		UserID:           "user-1",
		RefreshTokenHash: hash,
		ExpiresAt:        fixedNow.Add(24 * time.Hour),
	}
	user := activeUser(t, "user-1", "user@example.com", "password")
	userRepo.byID["user-1"] = user
	// authRepo.activeUsers does NOT include user-1
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN refreshing
	_, err := svc.Refresh(context.Background(), token, fixedNow)

	// THEN an UnauthorizedError is returned
	var ue *apperrors.UnauthorizedError
	if !errors.As(err, &ue) {
		t.Errorf("expected UnauthorizedError, got %T: %v", err, err)
	}
}

func TestThatRefreshReturnsResultOnSuccess(t *testing.T) {
	// GIVEN a valid non-expired session with an active user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	token := "valid-refresh-token"
	hash := coreauth.HashRefreshToken(token)
	authRepo.sessions[hash] = &models.AuthSession{
		ID:               "sess-1",
		UserID:           "user-1",
		RefreshTokenHash: hash,
		ExpiresAt:        fixedNow.Add(24 * time.Hour),
	}
	user := activeUser(t, "user-1", "user@example.com", "password")
	userRepo.byID["user-1"] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN refreshing
	result, err := svc.Refresh(context.Background(), token, fixedNow)

	// THEN a non-nil result is returned without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestThatRefreshReturnsNewAccessToken(t *testing.T) {
	// GIVEN a valid non-expired session with an active user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	token := "valid-refresh-token"
	hash := coreauth.HashRefreshToken(token)
	authRepo.sessions[hash] = &models.AuthSession{
		ID:               "sess-1",
		UserID:           "user-1",
		RefreshTokenHash: hash,
		ExpiresAt:        fixedNow.Add(24 * time.Hour),
	}
	user := activeUser(t, "user-1", "user@example.com", "password")
	userRepo.byID["user-1"] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN refreshing
	result, err := svc.Refresh(context.Background(), token, fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN a non-empty access token is returned
	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestThatRefreshRotatesRefreshToken(t *testing.T) {
	// GIVEN a valid non-expired session with an active user
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	token := "valid-refresh-token"
	hash := coreauth.HashRefreshToken(token)
	authRepo.sessions[hash] = &models.AuthSession{
		ID:               "sess-1",
		UserID:           "user-1",
		RefreshTokenHash: hash,
		ExpiresAt:        fixedNow.Add(24 * time.Hour),
	}
	user := activeUser(t, "user-1", "user@example.com", "password")
	userRepo.byID["user-1"] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN refreshing
	_, err := svc.Refresh(context.Background(), token, fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the session's refresh token was rotated
	if len(authRepo.rotated) != 1 || authRepo.rotated[0] != "sess-1" {
		t.Errorf("expected rotation of sess-1, got %v", authRepo.rotated)
	}
}

func TestThatRefreshReturnsCorrectUser(t *testing.T) {
	// GIVEN a valid session for user-1
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	token := "valid-refresh-token"
	hash := coreauth.HashRefreshToken(token)
	authRepo.sessions[hash] = &models.AuthSession{
		ID:               "sess-1",
		UserID:           "user-1",
		RefreshTokenHash: hash,
		ExpiresAt:        fixedNow.Add(24 * time.Hour),
	}
	user := activeUser(t, "user-1", "user@example.com", "password")
	userRepo.byID["user-1"] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN refreshing
	result, err := svc.Refresh(context.Background(), token, fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the returned user ID matches
	if result.User.ID != "user-1" {
		t.Errorf("expected user ID user-1, got %s", result.User.ID)
	}
}

func TestThatRefreshReturnsErrorWhenTokenManagerIsNil(t *testing.T) {
	// GIVEN a service with nil token manager
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, nil)

	// WHEN refreshing
	_, err := svc.Refresh(context.Background(), "some-token", fixedNow)

	// THEN an error about auth not configured is returned
	if err == nil || err.Error() != "auth is not configured" {
		t.Errorf("expected 'auth is not configured' error, got: %v", err)
	}
}

func TestThatRefreshSetsNewExpiresAtUsingTTL(t *testing.T) {
	// GIVEN a valid non-expired session
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	token := "valid-refresh-token"
	hash := coreauth.HashRefreshToken(token)
	authRepo.sessions[hash] = &models.AuthSession{
		ID:               "sess-1",
		UserID:           "user-1",
		RefreshTokenHash: hash,
		ExpiresAt:        fixedNow.Add(24 * time.Hour),
	}
	user := activeUser(t, "user-1", "user@example.com", "password")
	userRepo.byID["user-1"] = user
	authRepo.activeUsers["user-1"] = true
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN refreshing at fixedNow
	result, err := svc.Refresh(context.Background(), token, fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the new refresh expiration is fixedNow + 7 days
	want := fixedNow.Add(7 * 24 * time.Hour)
	if !result.RefreshExpiresAt.Equal(want) {
		t.Errorf("expected refresh expiry %v, got %v", want, result.RefreshExpiresAt)
	}
}

// ---------------------------------------------------------------------------
// Logout tests
// ---------------------------------------------------------------------------

func TestThatLogoutReturnsNilWhenTokenIsEmpty(t *testing.T) {
	// GIVEN a service
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging out with an empty token
	err := svc.Logout(context.Background(), "")

	// THEN no error is returned
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestThatLogoutReturnsNilWhenSessionNotFound(t *testing.T) {
	// GIVEN a service with no sessions
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging out with a token that has no matching session
	err := svc.Logout(context.Background(), "unknown-token")

	// THEN no error is returned
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestThatLogoutDoesNotRevokeWhenSessionNotFound(t *testing.T) {
	// GIVEN a service with no sessions
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging out with a token that has no matching session
	_ = svc.Logout(context.Background(), "unknown-token")

	// THEN no sessions were revoked
	if len(authRepo.revoked) != 0 {
		t.Errorf("expected no revocations, got %v", authRepo.revoked)
	}
}

func TestThatLogoutRevokesExistingSession(t *testing.T) {
	// GIVEN a session exists for the refresh token
	userRepo := newFakeUserRepo()
	authRepo := newFakeAuthRepo()
	token := "my-refresh-token"
	hash := coreauth.HashRefreshToken(token)
	authRepo.sessions[hash] = &models.AuthSession{
		ID:               "sess-1",
		UserID:           "user-1",
		RefreshTokenHash: hash,
		ExpiresAt:        fixedNow.Add(24 * time.Hour),
	}
	svc := newTestService(userRepo, authRepo, mustTokenManager(t))

	// WHEN logging out
	err := svc.Logout(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the session is revoked
	if len(authRepo.revoked) != 1 || authRepo.revoked[0] != "sess-1" {
		t.Errorf("expected revocation of sess-1, got %v", authRepo.revoked)
	}
}
