package httpserver

import (
	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthorizationRepository = dbadapters.AuthorizationRepository

func NewAuthorizationRepository(pool *pgxpool.Pool) *AuthorizationRepository {
	return dbadapters.NewAuthorizationRepository(pool)
}
