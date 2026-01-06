package recommended_entry_bids

import (
	appreb "github.com/andrewcopp/Calcutta/backend/internal/app/recommended_entry_bids"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service = appreb.Service

type GenerateParams = appreb.GenerateParams

type GenerateFromPredictionsParams = appreb.GenerateFromPredictionsParams

type ExpectedTeam = appreb.ExpectedTeam

type GenerateResult = appreb.GenerateResult

func New(pool *pgxpool.Pool) *Service {
	return appreb.New(pool)
}
