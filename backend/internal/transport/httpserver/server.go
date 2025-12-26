package httpserver

import (
	"database/sql"
	"log"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
	appbootstrap "github.com/andrewcopp/Calcutta/backend/internal/app/bootstrap"
	"github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	app               *app.App
	authRepo          *AuthRepository
	authzRepo         *AuthorizationRepository
	apiKeysRepo       *APIKeysRepository
	tokenManager      *auth.TokenManager
	pool              *pgxpool.Pool
	cfg               platform.Config
	bundleImportQueue chan string
}

func NewServer(db *sql.DB, pool *pgxpool.Pool, cfg platform.Config) *Server {
	authRepo := NewAuthRepository(pool)
	authzRepo := NewAuthorizationRepository(pool)
	apiKeysRepo := NewAPIKeysRepository(pool)

	a, tm, err := appbootstrap.NewApp(db, pool, cfg, authRepo, authzRepo)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	return &Server{
		app:               a,
		authRepo:          authRepo,
		authzRepo:         authzRepo,
		apiKeysRepo:       apiKeysRepo,
		tokenManager:      tm,
		pool:              pool,
		cfg:               cfg,
		bundleImportQueue: make(chan string, 32),
	}
}
