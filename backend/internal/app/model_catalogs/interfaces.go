package model_catalogs

import (
	"context"

	reb "github.com/andrewcopp/Calcutta/backend/internal/app/recommended_entry_bids"
	simt "github.com/andrewcopp/Calcutta/backend/internal/app/simulate_tournaments"
)

type AdvancementModelInterface interface {
	Descriptor() ModelDescriptor
}

type MarketShareModelInterface interface {
	Descriptor() ModelDescriptor
}

type EntryOptimizerInterface interface {
	Descriptor() ModelDescriptor
	AllocateBids(teams []reb.Team, params reb.AllocationParams) (reb.AllocationResult, error)
}

type SimulationModelInterface interface {
	Descriptor() ModelDescriptor
	Run(ctx context.Context, p simt.RunParams) (*simt.RunResult, error)
}
