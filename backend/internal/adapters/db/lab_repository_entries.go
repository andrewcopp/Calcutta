package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
)

// GetEntryRaw returns a single entry with raw data (bids, predictions, teams, pool budget)
// before any enrichment calculations. The caller is responsible for enrichment.
func (r *LabRepository) GetEntryRaw(ctx context.Context, id string) (*models.LabEntryRaw, error) {

	// Fetch the basic entry details plus tournament_id for team lookup.
	query := `
		SELECT
			e.id::text,
			e.investment_model_id::text,
			e.calcutta_id::text,
			e.game_outcome_kind,
			e.game_outcome_params_json::text,
			e.optimizer_kind,
			e.optimizer_params_json::text,
			e.starting_state_key,
			e.predictions_json::text,
			e.bids_json::text,
			e.created_at,
			e.updated_at,
			im.name AS model_name,
			im.kind AS model_kind,
			c.name AS calcutta_name,
			c.tournament_id::text,
			(SELECT COUNT(*) FROM lab.evaluations ev WHERE ev.entry_id = e.id AND ev.deleted_at IS NULL)::int AS n_evaluations
		FROM lab.entries e
		JOIN lab.investment_models im ON im.id = e.investment_model_id
		JOIN core.calcuttas c ON c.id = e.calcutta_id
		WHERE e.id = $1::uuid AND e.deleted_at IS NULL
	`

	var result models.LabEntryRaw
	var (
		tournamentID                                      string
		gameOutcomeParamsStr, optimizerParamsStr, bidsStr string
		predictionsStr                                    *string
	)

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&result.ID, &result.InvestmentModelID, &result.CalcuttaID,
		&result.GameOutcomeKind, &gameOutcomeParamsStr, &result.OptimizerKind, &optimizerParamsStr,
		&result.StartingStateKey, &predictionsStr, &bidsStr, &result.CreatedAt, &result.UpdatedAt,
		&result.ModelName, &result.ModelKind, &result.CalcuttaName, &tournamentID, &result.NEvaluations,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "entry", ID: id}
	}
	if err != nil {
		return nil, err
	}

	result.GameOutcomeParamsJSON = json.RawMessage(gameOutcomeParamsStr)
	result.OptimizerParamsJSON = json.RawMessage(optimizerParamsStr)

	// Parse predictions (if present).
	result.HasPredictions = predictionsStr != nil && *predictionsStr != ""
	if result.HasPredictions {
		if err := json.Unmarshal([]byte(*predictionsStr), &result.Predictions); err != nil {
			return nil, err
		}
	}

	// Parse raw bids.
	if err := json.Unmarshal([]byte(bidsStr), &result.Bids); err != nil {
		return nil, err
	}

	// Load team info for all teams in this tournament.
	teamMap, err := r.loadTeamMap(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	result.Teams = teamMap

	// Load total pool budget.
	totalPoolBudget, err := r.loadTotalPoolBudget(ctx, result.CalcuttaID)
	if err != nil {
		return nil, err
	}
	result.TotalPoolBudget = totalPoolBudget

	return &result, nil
}

// GetEntryIDByModelAndCalcutta resolves an entry ID from a model name, calcutta ID, and starting state key.
func (r *LabRepository) GetEntryIDByModelAndCalcutta(ctx context.Context, modelName, calcuttaID, startingStateKey string) (string, error) {
	if startingStateKey == "" {
		return "", errors.New("startingStateKey is required")
	}

	var entryID string
	err := r.pool.QueryRow(ctx, `
		SELECT e.id::text
		FROM lab.entries e
		JOIN lab.investment_models im ON im.id = e.investment_model_id AND im.deleted_at IS NULL
		WHERE im.name = $1
			AND e.calcutta_id = $2::uuid
			AND e.starting_state_key = $3
			AND e.deleted_at IS NULL
		ORDER BY e.created_at DESC
		LIMIT 1
	`, modelName, calcuttaID, startingStateKey).Scan(&entryID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", &apperrors.NotFoundError{Resource: "entry", ID: modelName + "/" + calcuttaID}
	}
	if err != nil {
		return "", err
	}
	return entryID, nil
}

// loadTeamMap returns team metadata keyed by team ID for a tournament.
func (r *LabRepository) loadTeamMap(ctx context.Context, tournamentID string) (map[string]models.LabTeamInfo, error) {
	teamQuery := `
		SELECT t.id::text, s.name, t.seed, t.region
		FROM core.teams t
		JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
		WHERE t.tournament_id = $1::uuid AND t.deleted_at IS NULL
	`
	rows, err := r.pool.Query(ctx, teamQuery, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teamMap := make(map[string]models.LabTeamInfo)
	for rows.Next() {
		var tid, name, region string
		var seed int
		if err := rows.Scan(&tid, &name, &seed, &region); err != nil {
			return nil, err
		}
		teamMap[tid] = models.LabTeamInfo{Name: name, Seed: seed, Region: region}
	}
	return teamMap, rows.Err()
}

// loadTotalPoolBudget returns the total pool budget for a calcutta.
func (r *LabRepository) loadTotalPoolBudget(ctx context.Context, calcuttaID string) (int, error) {
	var totalPoolBudget int
	poolBudgetQuery := `
		SELECT c.budget_points * COUNT(e.id)::int
		FROM core.calcuttas c
		LEFT JOIN core.entries e ON e.calcutta_id = c.id AND e.deleted_at IS NULL
		WHERE c.id = $1::uuid AND c.deleted_at IS NULL
		GROUP BY c.budget_points
	`
	if err := r.pool.QueryRow(ctx, poolBudgetQuery, calcuttaID).Scan(&totalPoolBudget); err != nil {
		return 0, fmt.Errorf("failed to load total pool budget for calcutta %s: %w", calcuttaID, err)
	}
	if totalPoolBudget <= 0 {
		return 0, fmt.Errorf("total pool budget is non-positive (%d) for calcutta %s", totalPoolBudget, calcuttaID)
	}
	return totalPoolBudget, nil
}
