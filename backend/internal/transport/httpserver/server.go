package httpserver

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app"
	appbootstrap "github.com/andrewcopp/Calcutta/backend/internal/app/bootstrap"
	"github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	app             *app.App
	authRepo        *dbadapters.AuthRepository
	authzRepo       *dbadapters.AuthorizationRepository
	userRepo        *dbadapters.UserRepository
	apiKeysRepo     *dbadapters.APIKeysRepository
	idempotencyRepo *dbadapters.IdempotencyRepository
	tokenManager    *auth.TokenManager
	cognitoJWT      *cognitoJWTVerifier
	pool            *pgxpool.Pool
	cfg             platform.Config
	emailSender     platform.EmailSender
	cookieSecure    bool
	cookieSameSite  http.SameSite
}

func NewServer(pool *pgxpool.Pool, cfg platform.Config) (*Server, error) {
	authRepo := dbadapters.NewAuthRepository(pool)
	authzRepo := dbadapters.NewAuthorizationRepository(pool)
	userRepo := dbadapters.NewUserRepository(pool)
	apiKeysRepo := dbadapters.NewAPIKeysRepository(pool)
	idempotencyRepo := dbadapters.NewIdempotencyRepository(pool)

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

	cookieSecure, cookieSameSite := computeCookieSettings(cfg)

	return &Server{
		app:             a,
		authRepo:        authRepo,
		authzRepo:       authzRepo,
		userRepo:        userRepo,
		apiKeysRepo:     apiKeysRepo,
		idempotencyRepo: idempotencyRepo,
		tokenManager:    tm,
		cognitoJWT:      cognitoJWT,
		pool:            pool,
		cfg:             cfg,
		emailSender:     emailSender,
		cookieSecure:    cookieSecure,
		cookieSameSite:  cookieSameSite,
	}, nil
}

func computeCookieSettings(cfg platform.Config) (secure bool, sameSite http.SameSite) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "production"
	}

	// Default: secure in production, not in development
	secure = env != "development"
	if secure {
		sameSite = http.SameSiteNoneMode
	} else {
		sameSite = http.SameSiteLaxMode
	}

	// Override from config if explicitly set
	if cfg.CookieSecure != nil {
		secure = *cfg.CookieSecure
	}

	switch cfg.CookieSameSite {
	case "none":
		sameSite = http.SameSiteNoneMode
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "strict":
		sameSite = http.SameSiteStrictMode
	}

	// SameSite=None requires Secure
	if sameSite == http.SameSiteNoneMode && !secure {
		sameSite = http.SameSiteLaxMode
	}

	return secure, sameSite
}
