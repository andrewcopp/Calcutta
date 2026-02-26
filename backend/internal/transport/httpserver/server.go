package httpserver

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/cognito"
	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app"
	appbootstrap "github.com/andrewcopp/Calcutta/backend/internal/app/bootstrap"
	"github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	app             *app.App
	authenticator   ports.Authenticator
	authRepo        *dbadapters.AuthRepository
	authzRepo       *dbadapters.AuthorizationRepository
	userRepo        *dbadapters.UserRepository
	apiKeysRepo     *dbadapters.APIKeysRepository
	idempotencyRepo *dbadapters.IdempotencyRepository
	pool            *pgxpool.Pool
	cfg             platform.Config
	emailSender     platform.EmailSender
	cookieSecure    bool
	cookieSameSite  http.SameSite
	devMode         bool
	hasLocalAuth    bool
}

func NewServer(pool *pgxpool.Pool, cfg platform.Config) (*Server, error) {
	if cfg.AuthMode == "dev" && cfg.AppEnv != "" && cfg.AppEnv != "development" {
		return nil, fmt.Errorf("dev auth mode is not allowed in %s environment", cfg.AppEnv)
	}

	authRepo := dbadapters.NewAuthRepository(pool)
	authzRepo := dbadapters.NewAuthorizationRepository(pool)
	userRepo := dbadapters.NewUserRepository(pool)
	apiKeysRepo := dbadapters.NewAPIKeysRepository(pool)
	idempotencyRepo := dbadapters.NewIdempotencyRepository(pool)

	// Create token manager for non-cognito modes (needed by bootstrap for the auth service).
	var tm *auth.TokenManager
	if cfg.AuthMode != "cognito" {
		created, err := auth.NewTokenManager(cfg.JWTSecret, time.Duration(cfg.AccessTokenTTLSeconds)*time.Second)
		if err != nil {
			return nil, fmt.Errorf("failed to create token manager: %w", err)
		}
		tm = created
	}

	a, err := appbootstrap.NewApp(pool, cfg, authRepo, tm)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize app: %w", err)
	}

	// Build authenticator chain.
	var authenticators []ports.Authenticator
	if tm != nil {
		authenticators = append(authenticators, auth.NewSessionAuthenticator(tm, authRepo, userRepo))
	}
	if cfg.AuthMode == "cognito" {
		verifier, err := cognito.NewVerifier(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize cognito jwt verifier: %w", err)
		}
		authenticators = append(authenticators, cognito.NewAuthenticator(verifier, userRepo, cfg.CognitoAutoProvision, cfg.CognitoAllowUnprovisioned))
	}
	authenticators = append(authenticators, auth.NewAPIKeyAuthenticator(apiKeysRepo, userRepo))
	chain := auth.NewChainAuthenticator(authenticators...)

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
		authenticator:   chain,
		authRepo:        authRepo,
		authzRepo:       authzRepo,
		userRepo:        userRepo,
		apiKeysRepo:     apiKeysRepo,
		idempotencyRepo: idempotencyRepo,
		pool:            pool,
		cfg:             cfg,
		emailSender:     emailSender,
		cookieSecure:    cookieSecure,
		cookieSameSite:  cookieSameSite,
		devMode:         cfg.AuthMode == "dev",
		hasLocalAuth:    cfg.AuthMode != "cognito",
	}, nil
}

func computeCookieSettings(cfg platform.Config) (secure bool, sameSite http.SameSite) {
	env := cfg.AppEnv
	if env == "" {
		env = "production"
	}

	// Default: secure in production, not in development
	secure = env != "development"
	if secure {
		sameSite = http.SameSiteLaxMode
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
