package simulate_tournaments

import (
	appsim "github.com/andrewcopp/Calcutta/backend/internal/app/simulate_tournaments"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service = appsim.Service

type RunParams = appsim.RunParams

type RunResult = appsim.RunResult

func New(pool *pgxpool.Pool) *Service {
	return appsim.New(pool)
}
