package simulated_calcutta

import (
	appsimulatedcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/simulated_calcutta"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service = appsimulatedcalcutta.Service

type SimulationResult = appsimulatedcalcutta.SimulationResult

type EntryPerformance = appsimulatedcalcutta.EntryPerformance

type Entry = appsimulatedcalcutta.Entry

type TeamSimResult = appsimulatedcalcutta.TeamSimResult

func New(pool *pgxpool.Pool) *Service {
	return appsimulatedcalcutta.New(pool)
}
