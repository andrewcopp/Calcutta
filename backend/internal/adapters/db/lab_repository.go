package db

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/app/lab"
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
func (r *LabRepository) ListInvestmentModels(filter lab.ListModelsFilter, page lab.Pagination) ([]lab.InvestmentModel, error) {
	ctx := context.Background()

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
		query += ` AND im.kind = $` + labItoa(argIdx)
		args = append(args, *filter.Kind)
		argIdx++
	}

	query += ` ORDER BY im.created_at DESC`
	query += ` LIMIT $` + labItoa(argIdx) + ` OFFSET $` + labItoa(argIdx+1)
	args = append(args, page.Limit, page.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]lab.InvestmentModel, 0)
	for rows.Next() {
		var m lab.InvestmentModel
		var paramsStr string
		if err := rows.Scan(&m.ID, &m.Name, &m.Kind, &paramsStr, &m.Notes, &m.CreatedAt, &m.UpdatedAt, &m.NEntries, &m.NEvaluations); err != nil {
			return nil, err
		}
		m.ParamsJSON = json.RawMessage(paramsStr)
		out = append(out, m)
	}
	return out, rows.Err()
}

// GetInvestmentModel returns a single investment model by ID.
func (r *LabRepository) GetInvestmentModel(id string) (*lab.InvestmentModel, error) {
	ctx := context.Background()

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

	var m lab.InvestmentModel
	var paramsStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(&m.ID, &m.Name, &m.Kind, &paramsStr, &m.Notes, &m.CreatedAt, &m.UpdatedAt, &m.NEntries, &m.NEvaluations)
	if err == pgx.ErrNoRows {
		return nil, &apperrors.NotFoundError{Resource: "investment_model", ID: id}
	}
	if err != nil {
		return nil, err
	}
	m.ParamsJSON = json.RawMessage(paramsStr)
	return &m, nil
}

// GetModelLeaderboard returns the model leaderboard.
func (r *LabRepository) GetModelLeaderboard() ([]lab.LeaderboardEntry, error) {
	ctx := context.Background()

	query := `
		SELECT
			investment_model_id::text,
			model_name,
			model_kind,
			n_entries::int,
			n_evaluations::int,
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
		return nil, err
	}
	defer rows.Close()

	out := make([]lab.LeaderboardEntry, 0)
	for rows.Next() {
		var e lab.LeaderboardEntry
		if err := rows.Scan(
			&e.InvestmentModelID,
			&e.ModelName,
			&e.ModelKind,
			&e.NEntries,
			&e.NEvaluations,
			&e.AvgMeanPayout,
			&e.AvgMedianPayout,
			&e.AvgPTop1,
			&e.AvgPInMoney,
			&e.FirstEvalAt,
			&e.LastEvalAt,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListEntries returns entries matching the filter.
func (r *LabRepository) ListEntries(filter lab.ListEntriesFilter, page lab.Pagination) ([]lab.EntryDetail, error) {
	ctx := context.Background()

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
		JOIN core.calcuttas c ON c.id = e.calcutta_id
		WHERE e.deleted_at IS NULL
	`
	args := []any{}
	argIdx := 1

	if filter.InvestmentModelID != nil && *filter.InvestmentModelID != "" {
		query += ` AND e.investment_model_id = $` + labItoa(argIdx) + `::uuid`
		args = append(args, *filter.InvestmentModelID)
		argIdx++
	}
	if filter.CalcuttaID != nil && *filter.CalcuttaID != "" {
		query += ` AND e.calcutta_id = $` + labItoa(argIdx) + `::uuid`
		args = append(args, *filter.CalcuttaID)
		argIdx++
	}
	if filter.StartingStateKey != nil && *filter.StartingStateKey != "" {
		query += ` AND e.starting_state_key = $` + labItoa(argIdx)
		args = append(args, *filter.StartingStateKey)
		argIdx++
	}

	query += ` ORDER BY e.created_at DESC`
	query += ` LIMIT $` + labItoa(argIdx) + ` OFFSET $` + labItoa(argIdx+1)
	args = append(args, page.Limit, page.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]lab.EntryDetail, 0)
	for rows.Next() {
		var e lab.EntryDetail
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
			return nil, err
		}
		e.GameOutcomeParamsJSON = json.RawMessage(gameOutcomeParamsStr)
		e.OptimizerParamsJSON = json.RawMessage(optimizerParamsStr)
		e.BidsJSON = json.RawMessage(bidsStr)
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetEntry returns a single entry by ID with full details.
func (r *LabRepository) GetEntry(id string) (*lab.EntryDetail, error) {
	ctx := context.Background()

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
		JOIN core.calcuttas c ON c.id = e.calcutta_id
		WHERE e.id = $1::uuid AND e.deleted_at IS NULL
	`

	var e lab.EntryDetail
	var gameOutcomeParamsStr, optimizerParamsStr, bidsStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
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
	)
	if err == pgx.ErrNoRows {
		return nil, &apperrors.NotFoundError{Resource: "entry", ID: id}
	}
	if err != nil {
		return nil, err
	}
	e.GameOutcomeParamsJSON = json.RawMessage(gameOutcomeParamsStr)
	e.OptimizerParamsJSON = json.RawMessage(optimizerParamsStr)
	e.BidsJSON = json.RawMessage(bidsStr)
	return &e, nil
}

// ListEvaluations returns evaluations matching the filter.
func (r *LabRepository) ListEvaluations(filter lab.ListEvaluationsFilter, page lab.Pagination) ([]lab.EvaluationDetail, error) {
	ctx := context.Background()

	query := `
		SELECT
			ev.id::text,
			ev.entry_id::text,
			ev.n_sims,
			ev.seed,
			ev.mean_normalized_payout,
			ev.median_normalized_payout,
			ev.p_top1,
			ev.p_in_money,
			ev.our_rank,
			ev.simulated_calcutta_id::text,
			ev.created_at,
			ev.updated_at,
			im.name AS model_name,
			im.kind AS model_kind,
			c.name AS calcutta_name,
			e.starting_state_key
		FROM lab.evaluations ev
		JOIN lab.entries e ON e.id = ev.entry_id
		JOIN lab.investment_models im ON im.id = e.investment_model_id
		JOIN core.calcuttas c ON c.id = e.calcutta_id
		WHERE ev.deleted_at IS NULL AND e.deleted_at IS NULL
	`
	args := []any{}
	argIdx := 1

	if filter.EntryID != nil && *filter.EntryID != "" {
		query += ` AND ev.entry_id = $` + labItoa(argIdx) + `::uuid`
		args = append(args, *filter.EntryID)
		argIdx++
	}
	if filter.InvestmentModelID != nil && *filter.InvestmentModelID != "" {
		query += ` AND e.investment_model_id = $` + labItoa(argIdx) + `::uuid`
		args = append(args, *filter.InvestmentModelID)
		argIdx++
	}
	if filter.CalcuttaID != nil && *filter.CalcuttaID != "" {
		query += ` AND e.calcutta_id = $` + labItoa(argIdx) + `::uuid`
		args = append(args, *filter.CalcuttaID)
		argIdx++
	}

	query += ` ORDER BY ev.mean_normalized_payout DESC NULLS LAST, ev.created_at DESC`
	query += ` LIMIT $` + labItoa(argIdx) + ` OFFSET $` + labItoa(argIdx+1)
	args = append(args, page.Limit, page.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]lab.EvaluationDetail, 0)
	for rows.Next() {
		var ev lab.EvaluationDetail
		if err := rows.Scan(
			&ev.ID,
			&ev.EntryID,
			&ev.NSims,
			&ev.Seed,
			&ev.MeanNormalizedPayout,
			&ev.MedianNormalizedPayout,
			&ev.PTop1,
			&ev.PInMoney,
			&ev.OurRank,
			&ev.SimulatedCalcuttaID,
			&ev.CreatedAt,
			&ev.UpdatedAt,
			&ev.ModelName,
			&ev.ModelKind,
			&ev.CalcuttaName,
			&ev.StartingStateKey,
		); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

// GetEvaluation returns a single evaluation by ID with full details.
func (r *LabRepository) GetEvaluation(id string) (*lab.EvaluationDetail, error) {
	ctx := context.Background()

	query := `
		SELECT
			ev.id::text,
			ev.entry_id::text,
			ev.n_sims,
			ev.seed,
			ev.mean_normalized_payout,
			ev.median_normalized_payout,
			ev.p_top1,
			ev.p_in_money,
			ev.our_rank,
			ev.simulated_calcutta_id::text,
			ev.created_at,
			ev.updated_at,
			im.name AS model_name,
			im.kind AS model_kind,
			c.name AS calcutta_name,
			e.starting_state_key
		FROM lab.evaluations ev
		JOIN lab.entries e ON e.id = ev.entry_id
		JOIN lab.investment_models im ON im.id = e.investment_model_id
		JOIN core.calcuttas c ON c.id = e.calcutta_id
		WHERE ev.id = $1::uuid AND ev.deleted_at IS NULL AND e.deleted_at IS NULL
	`

	var ev lab.EvaluationDetail
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ev.ID,
		&ev.EntryID,
		&ev.NSims,
		&ev.Seed,
		&ev.MeanNormalizedPayout,
		&ev.MedianNormalizedPayout,
		&ev.PTop1,
		&ev.PInMoney,
		&ev.OurRank,
		&ev.SimulatedCalcuttaID,
		&ev.CreatedAt,
		&ev.UpdatedAt,
		&ev.ModelName,
		&ev.ModelKind,
		&ev.CalcuttaName,
		&ev.StartingStateKey,
	)
	if err == pgx.ErrNoRows {
		return nil, &apperrors.NotFoundError{Resource: "evaluation", ID: id}
	}
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

// labItoa converts int to string for building parameterized queries.
func labItoa(i int) string {
	return strconv.Itoa(i)
}
