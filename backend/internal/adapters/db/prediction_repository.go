package db

import (
	"context"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PredictionRepository provides database access for prediction batch data.
type PredictionRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

// NewPredictionRepository creates a new prediction repository.
func NewPredictionRepository(pool *pgxpool.Pool) *PredictionRepository {
	return &PredictionRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *PredictionRepository) ListBatches(ctx context.Context, tournamentID string) ([]models.PredictionBatch, error) {
	rows, err := r.q.ListPredictionBatches(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("listing prediction batches: %w", err)
	}
	out := make([]models.PredictionBatch, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.PredictionBatch{
			ID:                   row.ID,
			ProbabilitySourceKey: row.ProbabilitySourceKey,
			ThroughRound:         int(row.ThroughRound),
			CreatedAt:            row.CreatedAt.Time,
		})
	}
	return out, nil
}

func (r *PredictionRepository) GetLatestBatch(ctx context.Context, tournamentID string) (*models.PredictionBatch, bool, error) {
	row, err := r.q.GetLatestPredictionBatch(ctx, tournamentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("getting latest prediction batch: %w", err)
	}
	return &models.PredictionBatch{
		ID:                   row.ID,
		ProbabilitySourceKey: row.ProbabilitySourceKey,
		ThroughRound:         int(row.ThroughRound),
		CreatedAt:            row.CreatedAt.Time,
	}, true, nil
}

func (r *PredictionRepository) GetBatchSummary(ctx context.Context, batchID string) (*models.PredictionBatch, error) {
	row, err := r.q.GetPredictionBatchByID(ctx, batchID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("batch not found: %s", batchID)
		}
		return nil, fmt.Errorf("getting batch summary: %w", err)
	}
	return &models.PredictionBatch{
		ID:                   row.ID,
		ProbabilitySourceKey: row.ProbabilitySourceKey,
		ThroughRound:         int(row.ThroughRound),
		CreatedAt:            row.CreatedAt.Time,
	}, nil
}

func (r *PredictionRepository) GetTeamValues(ctx context.Context, batchID string) ([]models.PredictedTeamValue, error) {
	rows, err := r.q.GetPredictedTeamValues(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("querying team values: %w", err)
	}
	out := make([]models.PredictedTeamValue, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.PredictedTeamValue{
			TeamID:               row.TeamID,
			ActualPoints:         row.ActualPoints,
			ExpectedPoints:       row.ExpectedPoints,
			VariancePoints:       row.VariancePoints,
			StdPoints:            row.StdPoints,
			PRound1:              row.PRound1,
			PRound2:              row.PRound2,
			PRound3:              row.PRound3,
			PRound4:              row.PRound4,
			PRound5:              row.PRound5,
			PRound6:              row.PRound6,
			PRound7:              row.PRound7,
			FavoritesTotalPoints: row.FavoritesTotalPoints,
		})
	}
	return out, nil
}

func (r *PredictionRepository) LoadTeams(ctx context.Context, tournamentID string) ([]models.PredictionTeamInput, error) {
	rows, err := r.q.GetTeamsWithKenpomForPrediction(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("querying teams: %w", err)
	}
	out := make([]models.PredictionTeamInput, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.PredictionTeamInput{
			ID:        row.TID,
			Seed:      int(row.Seed),
			Region:    row.Region,
			KenPomNet: row.KenpomNet,
			Wins:      int(row.Wins),
			Byes:      int(row.Byes),
		})
	}
	return out, nil
}

func (r *PredictionRepository) LoadScoringRules(ctx context.Context, tournamentID string) ([]scoring.Rule, error) {
	rows, err := r.q.GetScoringRulesForTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("querying scoring rules: %w", err)
	}
	out := make([]scoring.Rule, 0, len(rows))
	for _, row := range rows {
		out = append(out, scoring.Rule{
			WinIndex:      int(row.PsrWinIndex),
			PointsAwarded: int(row.PsrPointsAwarded),
		})
	}
	return out, nil
}

func (r *PredictionRepository) LoadFinalFourConfig(ctx context.Context, tournamentID string) (*models.FinalFourConfig, error) {
	row, err := r.q.GetFinalFourConfig(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("querying final four config: %w", err)
	}
	cfg := &models.FinalFourConfig{}
	if row.FinalFourTopLeft != nil {
		cfg.TopLeftRegion = *row.FinalFourTopLeft
	}
	if row.FinalFourBottomLeft != nil {
		cfg.BottomLeftRegion = *row.FinalFourBottomLeft
	}
	if row.FinalFourTopRight != nil {
		cfg.TopRightRegion = *row.FinalFourTopRight
	}
	if row.FinalFourBottomRight != nil {
		cfg.BottomRightRegion = *row.FinalFourBottomRight
	}
	cfg.ApplyDefaults()
	return cfg, nil
}

func (r *PredictionRepository) ListEligibleTournamentsForBackfill(ctx context.Context) ([]string, error) {
	ids, err := r.q.ListEligibleTournamentsForBackfill(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing eligible tournaments: %w", err)
	}
	return ids, nil
}

func (r *PredictionRepository) StorePredictions(
	ctx context.Context,
	tournamentID string,
	probSourceKey string,
	specJSON []byte,
	values []models.PredictedTeamValue,
	throughRound int,
) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	batchID, err := qtx.CreatePredictionBatch(ctx, sqlc.CreatePredictionBatchParams{
		TournamentID:         tournamentID,
		ProbabilitySourceKey: probSourceKey,
		GameOutcomeSpecJson:  specJSON,
		ThroughRound:         int32(throughRound),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create batch: %w", err)
	}

	bulkParams := make([]sqlc.BulkCreatePredictedTeamValuesParams, len(values))
	for i, v := range values {
		bulkParams[i] = sqlc.BulkCreatePredictedTeamValuesParams{
			PredictionBatchID:    batchID,
			TournamentID:         tournamentID,
			TeamID:               v.TeamID,
			ActualPoints:         &v.ActualPoints,
			ExpectedPoints:       v.ExpectedPoints,
			VariancePoints:       &v.VariancePoints,
			StdPoints:            &v.StdPoints,
			PRound1:              &v.PRound1,
			PRound2:              &v.PRound2,
			PRound3:              &v.PRound3,
			PRound4:              &v.PRound4,
			PRound5:              &v.PRound5,
			PRound6:              &v.PRound6,
			PRound7:              &v.PRound7,
			FavoritesTotalPoints: &v.FavoritesTotalPoints,
		}
	}
	if _, err := qtx.BulkCreatePredictedTeamValues(ctx, bulkParams); err != nil {
		return "", fmt.Errorf("failed to bulk insert team values: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("committing predictions: %w", err)
	}

	return batchID, nil
}

func (r *PredictionRepository) PruneOldBatchesForCheckpoint(ctx context.Context, tournamentID string, throughRound int, keepN int) (int64, error) {
	return r.q.PruneOldBatchesForCheckpoint(ctx, sqlc.PruneOldBatchesForCheckpointParams{
		TournamentID: tournamentID,
		KeepN:        int32(keepN),
		ThroughRound: int32(throughRound),
	})
}
