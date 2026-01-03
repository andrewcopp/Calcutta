package auth

import (
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	appauth "github.com/andrewcopp/Calcutta/backend/internal/app/auth"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Service = appauth.Service

type Result = appauth.Result

func New(userRepo ports.UserRepository, authRepo *dbadapters.AuthRepository, authzRepo *dbadapters.AuthorizationRepository, tokenMgr *coreauth.TokenManager, refreshTTL time.Duration) *Service {
	return appauth.New(userRepo, authRepo, authzRepo, tokenMgr, refreshTTL)
}
