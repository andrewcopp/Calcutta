package predicted_game_outcomes

import (
	apppgo "github.com/andrewcopp/Calcutta/backend/internal/app/predicted_game_outcomes"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service = apppgo.Service

type GenerateParams = apppgo.GenerateParams

func New(pool *pgxpool.Pool) *Service {
	return apppgo.New(pool)
}
