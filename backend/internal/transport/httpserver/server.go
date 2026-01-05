package httpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
	appbootstrap "github.com/andrewcopp/Calcutta/backend/internal/app/bootstrap"
	"github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	app          *app.App
	authRepo     *AuthRepository
	authzRepo    *AuthorizationRepository
	userRepo     *UserRepository
	apiKeysRepo  *APIKeysRepository
	tokenManager *auth.TokenManager
	cognitoJWT   *cognitoJWTVerifier
	pool         *pgxpool.Pool
	cfg          platform.Config
	emailSender  platform.EmailSender
}

func NewServer(pool *pgxpool.Pool, cfg platform.Config) (*Server, error) {
	authRepo := NewAuthRepository(pool)
	authzRepo := NewAuthorizationRepository(pool)
	userRepo := NewUserRepository(pool)
	apiKeysRepo := NewAPIKeysRepository(pool)

	a, tm, err := appbootstrap.NewApp(pool, cfg, authRepo, authzRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize app: %w", err)
	}

	var cognitoJWT *cognitoJWTVerifier
	if cfg.AuthMode == "cognito" {
		created, err := newCognitoJWTVerifier(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize cognito jwt verifier: %w", err)
		}
		cognitoJWT = created
	}

	var emailSender platform.EmailSender
	if strings.TrimSpace(cfg.SMTPHost) != "" && strings.TrimSpace(cfg.SMTPFromEmail) != "" {
		emailSender = &platform.SMTPSender{
			Host:      cfg.SMTPHost,
			Port:      cfg.SMTPPort,
			Username:  cfg.SMTPUsername,
			Password:  cfg.SMTPPassword,
			FromEmail: cfg.SMTPFromEmail,
			FromName:  cfg.SMTPFromName,
			StartTLS:  cfg.SMTPStartTLS,
		}
	}

	return &Server{
		app:          a,
		authRepo:     authRepo,
		authzRepo:    authzRepo,
		userRepo:     userRepo,
		apiKeysRepo:  apiKeysRepo,
		tokenManager: tm,
		cognitoJWT:   cognitoJWT,
		pool:         pool,
		cfg:          cfg,
		emailSender:  emailSender,
	}, nil
}

func (s *Server) bootstrapAdmin(ctx context.Context) error {
	email := strings.TrimSpace(s.cfg.BootstrapAdminEmail)
	if email == "" {
		return nil
	}
	if s.authzRepo == nil {
		return fmt.Errorf("authorization repository is not configured")
	}
	if s.userRepo == nil {
		return fmt.Errorf("user repository is not configured")
	}

	password := s.cfg.BootstrapAdminPassword
	if strings.TrimSpace(password) == "" && s.cfg.AuthMode != "cognito" {
		return fmt.Errorf("BOOTSTRAP_ADMIN_PASSWORD must be set when BOOTSTRAP_ADMIN_EMAIL is set")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user == nil {
		var passwordHash *string
		if strings.TrimSpace(password) != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			h := string(hash)
			passwordHash = &h
		}
		user = &models.User{
			ID:           uuid.New().String(),
			Email:        email,
			FirstName:    "Admin",
			LastName:     "User",
			PasswordHash: passwordHash,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return err
		}
	} else if (user.PasswordHash == nil || strings.TrimSpace(*user.PasswordHash) == "") && strings.TrimSpace(password) != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		h := string(hash)
		user.PasswordHash = &h
		if err := s.userRepo.Update(ctx, user); err != nil {
			return err
		}
	}

	return s.authzRepo.GrantGlobalAdmin(ctx, user.ID)
}
