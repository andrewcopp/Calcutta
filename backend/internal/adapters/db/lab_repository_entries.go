package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
)

// ListEntries returns entries matching the filter.
func (r *LabRepository) ListEntries(ctx context.Context, filter models.LabListEntriesFilter, page models.LabPagination) ([]models.LabEntryDetail, error) {

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
			e.bids_json::text,
			e.created_at,
			e.updated_at,
			im.name AS model_name,
			im.kind AS model_kind,
			c.name AS calcutta_name,
			(SELECT COUNT(*) FROM lab.evaluations ev WHERE ev.entry_id = e.id AND ev.deleted_at IS NULL)::int AS n_evaluations
		FROM lab.entries e
		JOIN lab.investment_models im ON im.id = e.investment_model_id
		JOIN core.pools c ON c.id = e.calcutta_id
		WHERE e.deleted_at IS NULL
	`
	args := []any{}
	argIdx := 1

	if filter.InvestmentModelID != nil && *filter.InvestmentModelID != "" {
		query += ` AND e.investment_model_id = $` + strconv.Itoa(argIdx) + `::uuid`
		args = append(args, *filter.InvestmentModelID)
		argIdx++
	}
	if filter.CalcuttaID != nil && *filter.CalcuttaID != "" {
		query += ` AND e.calcutta_id = $` + strconv.Itoa(argIdx) + `::uuid`
		args = append(args, *filter.CalcuttaID)
		argIdx++
	}
	if filter.StartingStateKey != nil && *filter.StartingStateKey != "" {
		query += ` AND e.starting_state_key = $` + strconv.Itoa(argIdx)
		args = append(args, *filter.StartingStateKey)
		argIdx++
	}

	query += ` ORDER BY e.created_at DESC`
	query += ` LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, page.Limit, page.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing lab entries: %w", err)
	}
	defer rows.Close()

	out := make([]models.LabEntryDetail, 0)
	for rows.Next() {
		var e models.LabEntryDetail
		var gameOutcomeParamsStr, optimizerParamsStr, bidsStr string
		if err := rows.Scan(
			&e.ID,
			&e.InvestmentModelID,
			&e.CalcuttaID,
			&e.GameOutcomeKind,
			&gameOutcomeParamsStr,
			&e.OptimizerKind,
			&optimizerParamsStr,
			&e.StartingStateKey,
			&bidsStr,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.ModelName,
			&e.ModelKind,
			&e.CalcuttaName,
			&e.NEvaluations,
		); err != nil {
			return nil, fmt.Errorf("scanning lab entry: %w", err)
		}
		e.GameOutcomeParamsJSON = json.RawMessage(gameOutcomeParamsStr)
		e.OptimizerParamsJSON = json.RawMessage(optimizerParamsStr)
		e.BidsJSON = json.RawMessage(bidsStr)
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating lab entries: %w", err)
	}
	return out, nil
}

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
		JOIN core.pools c ON c.id = e.calcutta_id
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
		return nil, fmt.Errorf("getting lab entry %s: %w", id, err)
	}

	result.GameOutcomeParamsJSON = json.RawMessage(gameOutcomeParamsStr)
	result.OptimizerParamsJSON = json.RawMessage(optimizerParamsStr)

	// Parse predictions (if present).
	result.HasPredictions = predictionsStr != nil && *predictionsStr != ""
	if result.HasPredictions {
		if err := json.Unmarshal([]byte(*predictionsStr), &result.Predictions); err != nil {
			return nil, fmt.Errorf("unmarshalling predictions for entry %s: %w", id, err)
		}
	}

	// Parse raw bids.
	if err := json.Unmarshal([]byte(bidsStr), &result.Bids); err != nil {
		return nil, fmt.Errorf("unmarshalling bids for entry %s: %w", id, err)
	}

	// Load team info for all teams in this tournament.
	teamMap, err := r.loadTeamMap(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("loading team map for entry %s: %w", id, err)
	}
	result.Teams = teamMap

	// Load total pool budget.
	totalPoolBudget, err := r.loadTotalPoolBudget(ctx, result.CalcuttaID)
	if err != nil {
		return nil, fmt.Errorf("loading total pool budget for entry %s: %w", id, err)
	}
	result.TotalPoolBudget = totalPoolBudget

	return &result, nil
}

// GetEntryIDByModelAndCalcutta resolves an entry ID from a model name, calcutta ID, and starting state key.
func (r *LabRepository) GetEntryIDByModelAndCalcutta(ctx context.Context, modelName, calcuttaID, startingStateKey string) (string, error) {
	var entryID string

	query := `
		SELECT e.id::text
		FROM lab.entries e
		JOIN lab.investment_models im ON im.id = e.investment_model_id AND im.deleted_at IS NULL
		WHERE im.name = $1
			AND e.calcutta_id = $2::uuid
			AND e.deleted_at IS NULL`
	args := []any{modelName, calcuttaID}

	if startingStateKey != "" {
		query += ` AND e.starting_state_key = $3`
		args = append(args, startingStateKey)
	}

	query += ` ORDER BY e.created_at DESC LIMIT 1`

	err := r.pool.QueryRow(ctx, query, args...).Scan(&entryID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", &apperrors.NotFoundError{Resource: "entry", ID: modelName + "/" + calcuttaID}
	}
	if err != nil {
		return "", fmt.Errorf("getting entry by model %s and calcutta %s: %w", modelName, calcuttaID, err)
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
		return nil, fmt.Errorf("querying teams for tournament %s: %w", tournamentID, err)
	}
	defer rows.Close()

	teamMap := make(map[string]models.LabTeamInfo)
	for rows.Next() {
		var tid, name, region string
		var seed int
		if err := rows.Scan(&tid, &name, &seed, &region); err != nil {
			return nil, fmt.Errorf("scanning team info: %w", err)
		}
		teamMap[tid] = models.LabTeamInfo{Name: name, Seed: seed, Region: region}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating teams for tournament %s: %w", tournamentID, err)
	}
	return teamMap, nil
}

// loadTotalPoolBudget returns the total pool budget for a calcutta.
func (r *LabRepository) loadTotalPoolBudget(ctx context.Context, calcuttaID string) (int, error) {
	var totalPoolBudget int
	poolBudgetQuery := `
		SELECT c.budget_credits * COUNT(p.id)::int
		FROM core.pools c
		LEFT JOIN core.portfolios p ON p.pool_id = c.id AND p.deleted_at IS NULL
		WHERE c.id = $1::uuid AND c.deleted_at IS NULL
		GROUP BY c.budget_credits
	`
	if err := r.pool.QueryRow(ctx, poolBudgetQuery, calcuttaID).Scan(&totalPoolBudget); err != nil {
		return 0, fmt.Errorf("failed to load total pool budget for calcutta %s: %w", calcuttaID, err)
	}
	if totalPoolBudget <= 0 {
		return 0, fmt.Errorf("total pool budget is non-positive (%d) for calcutta %s", totalPoolBudget, calcuttaID)
	}
	return totalPoolBudget, nil
}
