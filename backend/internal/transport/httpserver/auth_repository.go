package httpserver

import (
	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthSession = dbadapters.AuthSession
type AuthRepository = dbadapters.AuthRepository

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return dbadapters.NewAuthRepository(pool)
}
