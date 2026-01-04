package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepo struct {
	byEmail map[string]*models.User
	byID    map[string]*models.User
}

func (r *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.byEmail[email], nil
}

func (r *fakeUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	return r.byID[id], nil
}

func (r *fakeUserRepo) Create(ctx context.Context, user *models.User) error {
	panic("not implemented")
}

func (r *fakeUserRepo) Update(ctx context.Context, user *models.User) error {
	panic("not implemented")
}

type fakeAuthRepo struct {
	activeByUserID map[string]bool
	sessionByHash  map[string]*dbadapters.AuthSession
	rotateCalled   bool
}

func (r *fakeAuthRepo) CreateSession(ctx context.Context, userID, refreshTokenHash, userAgent, ipAddress string, expiresAt time.Time) (string, error) {
	panic("not implemented")
}

func (r *fakeAuthRepo) GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*dbadapters.AuthSession, error) {
	return r.sessionByHash[refreshTokenHash], nil
}

func (r *fakeAuthRepo) RotateRefreshToken(ctx context.Context, sessionID, newRefreshTokenHash string, newExpiresAt time.Time) error {
	r.rotateCalled = true
	return nil
}

func (r *fakeAuthRepo) RevokeSession(ctx context.Context, sessionID string) error {
	panic("not implemented")
}

func (r *fakeAuthRepo) IsUserActive(ctx context.Context, userID string) (bool, error) {
	return r.activeByUserID[userID], nil
}

func TestThatLoginRejectsInactiveUser(t *testing.T) {
	// GIVEN
	h, err := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	hStr := string(h)
	u := &models.User{ID: "u1", Email: "a@b.com", PasswordHash: &hStr}

	ur := &fakeUserRepo{byEmail: map[string]*models.User{"a@b.com": u}, byID: map[string]*models.User{"u1": u}}
	ar := &fakeAuthRepo{activeByUserID: map[string]bool{"u1": false}, sessionByHash: map[string]*dbadapters.AuthSession{}}

	tm, err := coreauth.NewTokenManager("secret", 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	svc := &Service{userRepo: ur, authRepo: ar, tokenMgr: tm, refreshTTL: time.Hour}

	// WHEN
	_, gotErr := svc.Login(context.Background(), "a@b.com", "pw", "", "", time.Now())

	// THEN
	var unauthorized *apperrors.UnauthorizedError
	if !errors.As(gotErr, &unauthorized) {
		t.Fatalf("expected unauthorized error, got %v", gotErr)
	}
}

func TestThatRefreshRejectsInactiveUser(t *testing.T) {
	// GIVEN
	u := &models.User{ID: "u1", Email: "a@b.com"}
	sess := &dbadapters.AuthSession{ID: "s1", UserID: "u1", RefreshTokenHash: "h", ExpiresAt: time.Now().Add(time.Hour)}

	refreshToken := "raw"
	h := coreauth.HashRefreshToken(refreshToken)

	ur := &fakeUserRepo{byEmail: map[string]*models.User{}, byID: map[string]*models.User{"u1": u}}
	ar := &fakeAuthRepo{activeByUserID: map[string]bool{"u1": false}, sessionByHash: map[string]*dbadapters.AuthSession{h: sess}}

	tm, err := coreauth.NewTokenManager("secret", 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	svc := &Service{userRepo: ur, authRepo: ar, tokenMgr: tm, refreshTTL: time.Hour}

	// WHEN
	_, gotErr := svc.Refresh(context.Background(), refreshToken, time.Now())

	// THEN
	var unauthorized *apperrors.UnauthorizedError
	if !errors.As(gotErr, &unauthorized) {
		t.Fatalf("expected unauthorized error, got %v", gotErr)
	}
}

func TestThatRefreshDoesNotRotateTokenForInactiveUser(t *testing.T) {
	// GIVEN
	u := &models.User{ID: "u1", Email: "a@b.com"}
	sess := &dbadapters.AuthSession{ID: "s1", UserID: "u1", RefreshTokenHash: "h", ExpiresAt: time.Now().Add(time.Hour)}

	refreshToken := "raw"
	h := coreauth.HashRefreshToken(refreshToken)

	ur := &fakeUserRepo{byEmail: map[string]*models.User{}, byID: map[string]*models.User{"u1": u}}
	ar := &fakeAuthRepo{activeByUserID: map[string]bool{"u1": false}, sessionByHash: map[string]*dbadapters.AuthSession{h: sess}}

	tm, err := coreauth.NewTokenManager("secret", 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	svc := &Service{userRepo: ur, authRepo: ar, tokenMgr: tm, refreshTTL: time.Hour}

	// WHEN
	_, _ = svc.Refresh(context.Background(), refreshToken, time.Now())

	// THEN
	if ar.rotateCalled {
		t.Fatalf("expected RotateRefreshToken not to be called")
	}
}
