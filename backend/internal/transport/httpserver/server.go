package httpserver

import (
	"log"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
	appbootstrap "github.com/andrewcopp/Calcutta/backend/internal/app/bootstrap"
	"github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
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
}

func NewServer(pool *pgxpool.Pool, cfg platform.Config) *Server {
	authRepo := NewAuthRepository(pool)
	authzRepo := NewAuthorizationRepository(pool)
	userRepo := NewUserRepository(pool)
	apiKeysRepo := NewAPIKeysRepository(pool)

	a, tm, err := appbootstrap.NewApp(pool, cfg, authRepo, authzRepo)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	var cognitoJWT *cognitoJWTVerifier
	if cfg.AuthMode == "cognito" {
		created, err := newCognitoJWTVerifier(cfg)
		if err != nil {
			log.Fatalf("failed to initialize cognito jwt verifier: %v", err)
		}
		cognitoJWT = created
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
	}
}
