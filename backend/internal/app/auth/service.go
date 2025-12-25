package auth

import (
	"context"
	"errors"
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/andrewcopp/Calcutta/backend/pkg/services"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Result struct {
	User             *models.User
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type Service struct {
	userRepo   *dbadapters.UserRepository
	authRepo   *dbadapters.AuthRepository
	authzRepo  *dbadapters.AuthorizationRepository
	tokenMgr   *coreauth.TokenManager
	refreshTTL time.Duration
}

func New(userRepo *dbadapters.UserRepository, authRepo *dbadapters.AuthRepository, authzRepo *dbadapters.AuthorizationRepository, tokenMgr *coreauth.TokenManager, refreshTTL time.Duration) *Service {
	return &Service{userRepo: userRepo, authRepo: authRepo, authzRepo: authzRepo, tokenMgr: tokenMgr, refreshTTL: refreshTTL}
}

func (s *Service) Login(ctx context.Context, email, password, userAgent, ipAddress string, now time.Time) (*Result, error) {
	if s.tokenMgr == nil {
		return nil, errors.New("auth is not configured")
	}
	if now.IsZero() {
		now = time.Now()
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &services.NotFoundError{Resource: "user", ID: email}
	}
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, &services.NotFoundError{Resource: "user", ID: email}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, &services.NotFoundError{Resource: "user", ID: email}
	}

	refreshToken, err := coreauth.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshHash := coreauth.HashRefreshToken(refreshToken)
	expiresAt := now.Add(s.refreshTTL)

	sessionID, err := s.authRepo.CreateSession(ctx, user.ID, refreshHash, userAgent, ipAddress, expiresAt)
	if err != nil {
		return nil, err
	}

	accessToken, _, err := s.tokenMgr.IssueAccessToken(user.ID, sessionID, now)
	if err != nil {
		return nil, err
	}

	return &Result{User: user, AccessToken: accessToken, RefreshToken: refreshToken, RefreshExpiresAt: expiresAt}, nil
}

func (s *Service) Signup(ctx context.Context, email, firstName, lastName, password, userAgent, ipAddress string, now time.Time) (*Result, error) {
	if s.tokenMgr == nil {
		return nil, errors.New("auth is not configured")
	}
	if now.IsZero() {
		now = time.Now()
	}

	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, &services.AlreadyExistsError{Resource: "user", Field: "email", Value: email}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashStr := string(hash)

	user := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		PasswordHash: &hashStr,
		Created:      now,
		Updated:      now,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	if s.authzRepo != nil {
		_ = s.authzRepo.EnsureBootstrapAdmin(ctx, user.ID)
	}

	refreshToken, err := coreauth.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshHash := coreauth.HashRefreshToken(refreshToken)
	expiresAt := now.Add(s.refreshTTL)

	sessionID, err := s.authRepo.CreateSession(ctx, user.ID, refreshHash, userAgent, ipAddress, expiresAt)
	if err != nil {
		return nil, err
	}

	accessToken, _, err := s.tokenMgr.IssueAccessToken(user.ID, sessionID, now)
	if err != nil {
		return nil, err
	}

	return &Result{User: user, AccessToken: accessToken, RefreshToken: refreshToken, RefreshExpiresAt: expiresAt}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string, now time.Time) (*Result, error) {
	if s.tokenMgr == nil {
		return nil, errors.New("auth is not configured")
	}
	if refreshToken == "" {
		return nil, errors.New("refresh token missing")
	}
	if now.IsZero() {
		now = time.Now()
	}

	oldHash := coreauth.HashRefreshToken(refreshToken)
	sess, err := s.authRepo.GetSessionByRefreshTokenHash(ctx, oldHash)
	if err != nil {
		return nil, err
	}
	if sess == nil || sess.RevokedAt != nil || now.After(sess.ExpiresAt) {
		return nil, errors.New("invalid refresh token")
	}

	newToken, err := coreauth.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	newHash := coreauth.HashRefreshToken(newToken)
	newExpiresAt := now.Add(s.refreshTTL)
	if err := s.authRepo.RotateRefreshToken(ctx, sess.ID, newHash, newExpiresAt); err != nil {
		return nil, err
	}

	accessToken, _, err := s.tokenMgr.IssueAccessToken(sess.UserID, sess.ID, now)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid refresh token")
	}

	return &Result{User: user, AccessToken: accessToken, RefreshToken: newToken, RefreshExpiresAt: newExpiresAt}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	h := coreauth.HashRefreshToken(refreshToken)
	sess, err := s.authRepo.GetSessionByRefreshTokenHash(ctx, h)
	if err != nil {
		return err
	}
	if sess == nil {
		return nil
	}
	return s.authRepo.RevokeSession(ctx, sess.ID)
}
