package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
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
	userRepo   ports.UserRepository
	authRepo   ports.AuthSessionRepository
	tokenMgr   *coreauth.TokenManager
	refreshTTL time.Duration
}

func New(userRepo ports.UserRepository, authRepo ports.AuthSessionRepository, tokenMgr *coreauth.TokenManager, refreshTTL time.Duration) *Service {
	return &Service{userRepo: userRepo, authRepo: authRepo, tokenMgr: tokenMgr, refreshTTL: refreshTTL}
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
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	if user == nil {
		return nil, &apperrors.UnauthorizedError{Message: "invalid credentials"}
	}
	ok, err := s.authRepo.IsUserActive(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("checking user active status: %w", err)
	}
	if !ok {
		return nil, &apperrors.UnauthorizedError{Message: "invalid credentials"}
	}
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, &apperrors.UnauthorizedError{Message: "invalid credentials"}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, &apperrors.UnauthorizedError{Message: "invalid credentials"}
	}

	refreshToken, err := coreauth.NewRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}
	refreshHash := coreauth.HashRefreshToken(refreshToken)
	expiresAt := now.Add(s.refreshTTL)

	sessionID, err := s.authRepo.CreateSession(ctx, user.ID, refreshHash, userAgent, ipAddress, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	accessToken, _, err := s.tokenMgr.IssueAccessToken(user.ID, sessionID, now)
	if err != nil {
		return nil, fmt.Errorf("issuing access token: %w", err)
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
		return nil, fmt.Errorf("checking existing user: %w", err)
	}
	if existing != nil {
		return nil, &apperrors.AlreadyExistsError{Resource: "user", Field: "email", Value: email}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}
	hashStr := string(hash)

	user := &models.User{
		ID:           uuid.New().String(),
		Email:        &email,
		FirstName:    firstName,
		LastName:     lastName,
		Status:       "active",
		PasswordHash: &hashStr,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	refreshToken, err := coreauth.NewRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}
	refreshHash := coreauth.HashRefreshToken(refreshToken)
	expiresAt := now.Add(s.refreshTTL)

	sessionID, err := s.authRepo.CreateSession(ctx, user.ID, refreshHash, userAgent, ipAddress, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	accessToken, _, err := s.tokenMgr.IssueAccessToken(user.ID, sessionID, now)
	if err != nil {
		return nil, fmt.Errorf("issuing access token: %w", err)
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
		return nil, fmt.Errorf("getting session by refresh token: %w", err)
	}
	if sess == nil || sess.RevokedAt != nil || now.After(sess.ExpiresAt) {
		return nil, &apperrors.UnauthorizedError{Message: "invalid refresh token"}
	}

	user, err := s.userRepo.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, fmt.Errorf("getting user by id: %w", err)
	}
	if user == nil {
		return nil, &apperrors.UnauthorizedError{Message: "invalid refresh token"}
	}
	ok, err := s.authRepo.IsUserActive(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("checking user active status: %w", err)
	}
	if !ok {
		return nil, &apperrors.UnauthorizedError{Message: "invalid refresh token"}
	}

	newToken, err := coreauth.NewRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}
	newHash := coreauth.HashRefreshToken(newToken)
	newExpiresAt := now.Add(s.refreshTTL)
	if err := s.authRepo.RotateRefreshToken(ctx, sess.ID, newHash, newExpiresAt); err != nil {
		return nil, fmt.Errorf("rotating refresh token: %w", err)
	}

	accessToken, _, err := s.tokenMgr.IssueAccessToken(sess.UserID, sess.ID, now)
	if err != nil {
		return nil, fmt.Errorf("issuing access token: %w", err)
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
		return fmt.Errorf("getting session by refresh token: %w", err)
	}
	if sess == nil {
		return nil
	}
	if err := s.authRepo.RevokeSession(ctx, sess.ID); err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}
	return nil
}
