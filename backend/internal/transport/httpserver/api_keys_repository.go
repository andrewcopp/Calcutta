package httpserver

import (
	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type APIKey = dbadapters.APIKey

type APIKeysRepository = dbadapters.APIKeysRepository

func NewAPIKeysRepository(pool *pgxpool.Pool) *APIKeysRepository {
	return dbadapters.NewAPIKeysRepository(pool)
}
