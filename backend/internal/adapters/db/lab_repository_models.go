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
	"github.com/jackc/pgx/v5/pgxpool"
)

// LabRepository provides database access for lab.* tables.
type LabRepository struct {
	pool *pgxpool.Pool
}

// NewLabRepository creates a new lab repository.
func NewLabRepository(pool *pgxpool.Pool) *LabRepository {
	return &LabRepository{pool: pool}
}

// ListInvestmentModels returns investment models matching the filter.
func (r *LabRepository) ListInvestmentModels(ctx context.Context, filter models.LabListModelsFilter, page models.LabPagination) ([]models.InvestmentModel, error) {

	query := `
		SELECT
			im.id::text,
			im.name,
			im.kind,
			im.params_json::text,
			im.notes,
			im.created_at,
			im.updated_at,
			(SELECT COUNT(*) FROM lab.entries e WHERE e.investment_model_id = im.id AND e.deleted_at IS NULL)::int AS n_entries,
			(SELECT COUNT(*) FROM lab.evaluations ev JOIN lab.entries e ON e.id = ev.entry_id WHERE e.investment_model_id = im.id AND ev.deleted_at IS NULL AND e.deleted_at IS NULL)::int AS n_evaluations
		FROM lab.investment_models im
		WHERE im.deleted_at IS NULL
	`
	args := []any{}
	argIdx := 1

	if filter.Kind != nil && *filter.Kind != "" {
		query += ` AND im.kind = $` + strconv.Itoa(argIdx)
		args = append(args, *filter.Kind)
		argIdx++
	}

	query += ` ORDER BY im.created_at DESC`
	query += ` LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, page.Limit, page.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing investment models: %w", err)
	}
	defer rows.Close()

	out := make([]models.InvestmentModel, 0)
	for rows.Next() {
		var m models.InvestmentModel
		var paramsStr string
		if err := rows.Scan(&m.ID, &m.Name, &m.Kind, &paramsStr, &m.Notes, &m.CreatedAt, &m.UpdatedAt, &m.NEntries, &m.NEvaluations); err != nil {
			return nil, fmt.Errorf("scanning investment model: %w", err)
		}
		m.ParamsJSON = json.RawMessage(paramsStr)
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating investment models: %w", err)
	}
	return out, nil
}

// GetInvestmentModel returns a single investment model by ID.
func (r *LabRepository) GetInvestmentModel(ctx context.Context, id string) (*models.InvestmentModel, error) {

	query := `
		SELECT
			im.id::text,
			im.name,
			im.kind,
			im.params_json::text,
			im.notes,
			im.created_at,
			im.updated_at,
			(SELECT COUNT(*) FROM lab.entries e WHERE e.investment_model_id = im.id AND e.deleted_at IS NULL)::int AS n_entries,
			(SELECT COUNT(*) FROM lab.evaluations ev JOIN lab.entries e ON e.id = ev.entry_id WHERE e.investment_model_id = im.id AND ev.deleted_at IS NULL AND e.deleted_at IS NULL)::int AS n_evaluations
		FROM lab.investment_models im
		WHERE im.id = $1::uuid AND im.deleted_at IS NULL
	`

	var m models.InvestmentModel
	var paramsStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(&m.ID, &m.Name, &m.Kind, &paramsStr, &m.Notes, &m.CreatedAt, &m.UpdatedAt, &m.NEntries, &m.NEvaluations)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "investment_model", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("getting investment model %s: %w", id, err)
	}
	m.ParamsJSON = json.RawMessage(paramsStr)
	return &m, nil
}

// GetModelLeaderboard returns the model leaderboard.
func (r *LabRepository) GetModelLeaderboard(ctx context.Context) ([]models.LabLeaderboardEntry, error) {

	query := `
		SELECT
			investment_model_id::text,
			model_name,
			model_kind,
			n_entries::int,
			n_entries_with_predictions::int,
			n_evaluations::int,
			n_calcuttas_with_entries::int,
			n_calcuttas_with_evaluations::int,
			avg_mean_payout,
			avg_median_payout,
			avg_p_top1,
			avg_p_in_money,
			first_eval_at,
			last_eval_at
		FROM lab.model_leaderboard
		ORDER BY avg_mean_payout DESC NULLS LAST
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying model leaderboard: %w", err)
	}
	defer rows.Close()

	out := make([]models.LabLeaderboardEntry, 0)
	for rows.Next() {
		var e models.LabLeaderboardEntry
		if err := rows.Scan(
			&e.InvestmentModelID,
			&e.ModelName,
			&e.ModelKind,
			&e.NEntries,
			&e.NEntriesWithPredictions,
			&e.NEvaluations,
			&e.NCalcuttasWithEntries,
			&e.NCalcuttasWithEvaluations,
			&e.AvgMeanPayout,
			&e.AvgMedianPayout,
			&e.AvgPTop1,
			&e.AvgPInMoney,
			&e.FirstEvalAt,
			&e.LastEvalAt,
		); err != nil {
			return nil, fmt.Errorf("scanning leaderboard entry: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating leaderboard entries: %w", err)
	}
	return out, nil
}
