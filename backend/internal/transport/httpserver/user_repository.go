package httpserver

import (
	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository = dbadapters.UserRepository

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return dbadapters.NewUserRepository(pool)
}
