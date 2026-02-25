package ports

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type TournamentDataLoader interface {
	LoadTeams(ctx context.Context, tournamentID string) ([]models.PredictionTeamInput, error)
	LoadScoringRules(ctx context.Context, tournamentID string) ([]scoring.Rule, error)
	LoadFinalFourConfig(ctx context.Context, tournamentID string) (*models.FinalFourConfig, error)
}

type PredictionBatchReader interface {
	ListBatches(ctx context.Context, tournamentID string) ([]models.PredictionBatch, error)
	GetLatestBatch(ctx context.Context, tournamentID string) (*models.PredictionBatch, bool, error)
	GetBatchSummary(ctx context.Context, batchID string) (*models.PredictionBatch, error)
	GetTeamValues(ctx context.Context, batchID string) ([]models.PredictedTeamValue, error)
	ListEligibleTournamentsForBackfill(ctx context.Context) ([]string, error)
}

type PredictionBatchWriter interface {
	StorePredictions(ctx context.Context, tournamentID string, probSourceKey string, specJSON []byte, values []models.PredictedTeamValue, throughRound int) (string, error)
	PruneOldBatchesForCheckpoint(ctx context.Context, tournamentID string, throughRound int, keepN int) (int64, error)
}

type PredictionRepository interface {
	TournamentDataLoader
	PredictionBatchReader
	PredictionBatchWriter
}
