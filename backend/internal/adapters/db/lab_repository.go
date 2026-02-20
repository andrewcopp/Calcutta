package db

import (
	"context"
	"encoding/json"
	"errors"
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
		return nil, err
	}
	defer rows.Close()

	out := make([]models.InvestmentModel, 0)
	for rows.Next() {
		var m models.InvestmentModel
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
		return nil, err
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
		return nil, err
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
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

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
		JOIN core.calcuttas c ON c.id = e.calcutta_id
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
		return nil, err
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
			return nil, err
		}
		e.GameOutcomeParamsJSON = json.RawMessage(gameOutcomeParamsStr)
		e.OptimizerParamsJSON = json.RawMessage(optimizerParamsStr)
		e.BidsJSON = json.RawMessage(bidsStr)
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListEvaluations returns evaluations matching the filter.
func (r *LabRepository) ListEvaluations(ctx context.Context, filter models.LabListEvaluationsFilter, page models.LabPagination) ([]models.LabEvaluationDetail, error) {

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
			ev.created_at,
			ev.updated_at,
			im.name AS model_name,
			im.kind AS model_kind,
			c.id::text AS calcutta_id,
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
		query += ` AND ev.entry_id = $` + strconv.Itoa(argIdx) + `::uuid`
		args = append(args, *filter.EntryID)
		argIdx++
	}
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

	query += ` ORDER BY ev.mean_normalized_payout DESC NULLS LAST, ev.created_at DESC`
	query += ` LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, page.Limit, page.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.LabEvaluationDetail, 0)
	for rows.Next() {
		var ev models.LabEvaluationDetail
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
			&ev.CreatedAt,
			&ev.UpdatedAt,
			&ev.ModelName,
			&ev.ModelKind,
			&ev.CalcuttaID,
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
func (r *LabRepository) GetEvaluation(ctx context.Context, id string) (*models.LabEvaluationDetail, error) {

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
			ev.created_at,
			ev.updated_at,
			im.name AS model_name,
			im.kind AS model_kind,
			c.id::text AS calcutta_id,
			c.name AS calcutta_name,
			e.starting_state_key
		FROM lab.evaluations ev
		JOIN lab.entries e ON e.id = ev.entry_id
		JOIN lab.investment_models im ON im.id = e.investment_model_id
		JOIN core.calcuttas c ON c.id = e.calcutta_id
		WHERE ev.id = $1::uuid AND ev.deleted_at IS NULL AND e.deleted_at IS NULL
	`

	var ev models.LabEvaluationDetail
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
		&ev.CreatedAt,
		&ev.UpdatedAt,
		&ev.ModelName,
		&ev.ModelKind,
		&ev.CalcuttaID,
		&ev.CalcuttaName,
		&ev.StartingStateKey,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "evaluation", ID: id}
	}
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

// GetEvaluationEntryResults returns per-entry results for an evaluation.
func (r *LabRepository) GetEvaluationEntryResults(ctx context.Context, evaluationID string) ([]models.LabEvaluationEntryResult, error) {

	query := `
		SELECT
			id,
			entry_name,
			mean_normalized_payout,
			p_top1,
			p_in_money,
			rank
		FROM lab.evaluation_entry_results
		WHERE evaluation_id = $1::uuid
		ORDER BY rank ASC
	`

	rows, err := r.pool.Query(ctx, query, evaluationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.LabEvaluationEntryResult, 0)
	for rows.Next() {
		var e models.LabEvaluationEntryResult
		if err := rows.Scan(
			&e.ID,
			&e.EntryName,
			&e.MeanNormalizedPayout,
			&e.PTop1,
			&e.PInMoney,
			&e.Rank,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetEvaluationEntryProfile returns detailed profile for an entry in an evaluation.
func (r *LabRepository) GetEvaluationEntryProfile(ctx context.Context, entryResultID string) (*models.LabEvaluationEntryProfile, error) {

	// First, get the entry result and evaluation_id from lab.evaluation_entry_results
	var profile models.LabEvaluationEntryProfile
	var evaluationID string
	err := r.pool.QueryRow(ctx, `
		SELECT
			entry_name,
			mean_normalized_payout,
			p_top1,
			p_in_money,
			rank,
			evaluation_id
		FROM lab.evaluation_entry_results
		WHERE id = $1::uuid
	`, entryResultID).Scan(
		&profile.EntryName,
		&profile.MeanNormalizedPayout,
		&profile.PTop1,
		&profile.PInMoney,
		&profile.Rank,
		&evaluationID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "evaluation_entry_result", ID: entryResultID}
	}
	if err != nil {
		return nil, err
	}

	// Get the lab entry_id and calcutta_id from the evaluation
	var labEntryID string
	var calcuttaID string
	err = r.pool.QueryRow(ctx, `
		SELECT ev.entry_id, e.calcutta_id
		FROM lab.evaluations ev
		JOIN lab.entries e ON e.id = ev.entry_id
		WHERE ev.id = $1::uuid
			AND ev.deleted_at IS NULL
	`, evaluationID).Scan(&labEntryID, &calcuttaID)
	if err != nil {
		return nil, err
	}

	profile.Bids = make([]models.LabEvaluationEntryBid, 0)
	profile.TotalBidPoints = 0

	var rows pgx.Rows

	if profile.EntryName == models.LabStrategyEntryName {
		// Get bids from lab.entries.bids_json
		query := `
			WITH entry_bids AS (
				SELECT
					(bid->>'team_id')::uuid AS team_id,
					(bid->>'bid_points')::int AS bid_points
				FROM lab.entries e,
					jsonb_array_elements(e.bids_json) AS bid
				WHERE e.id = $1::uuid
					AND e.deleted_at IS NULL
			),
			total_pool AS (
				SELECT
					cet.team_id,
					SUM(cet.bid_points)::float AS total_bid
				FROM core.entry_teams cet
				JOIN core.entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
				WHERE ce.calcutta_id = $2::uuid
					AND cet.deleted_at IS NULL
				GROUP BY cet.team_id
			)
			SELECT
				eb.team_id,
				s.name AS school_name,
				t.seed,
				t.region,
				eb.bid_points,
				COALESCE(eb.bid_points::float / NULLIF(COALESCE(tp.total_bid, 0) + eb.bid_points, 0), 0) AS ownership
			FROM entry_bids eb
			JOIN core.teams t ON t.id = eb.team_id AND t.deleted_at IS NULL
			JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
			LEFT JOIN total_pool tp ON tp.team_id = eb.team_id
			WHERE eb.bid_points > 0
			ORDER BY t.seed ASC, s.name ASC
		`
		rows, err = r.pool.Query(ctx, query, labEntryID, calcuttaID)
	} else {
		// Get bids from core.entries + core.entry_teams
		query := `
			WITH entry_bids AS (
				SELECT
					cet.team_id,
					cet.bid_points
				FROM core.entry_teams cet
				JOIN core.entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
				WHERE ce.calcutta_id = $1::uuid
					AND ce.name = $2
					AND cet.deleted_at IS NULL
			),
			total_pool AS (
				SELECT
					cet.team_id,
					SUM(cet.bid_points)::float AS total_bid
				FROM core.entry_teams cet
				JOIN core.entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
				WHERE ce.calcutta_id = $1::uuid
					AND cet.deleted_at IS NULL
				GROUP BY cet.team_id
			)
			SELECT
				eb.team_id,
				s.name AS school_name,
				t.seed,
				t.region,
				eb.bid_points,
				COALESCE(eb.bid_points::float / NULLIF(tp.total_bid, 0), 0) AS ownership
			FROM entry_bids eb
			JOIN core.teams t ON t.id = eb.team_id AND t.deleted_at IS NULL
			JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
			LEFT JOIN total_pool tp ON tp.team_id = eb.team_id
			WHERE eb.bid_points > 0
			ORDER BY t.seed ASC, s.name ASC
		`
		rows, err = r.pool.Query(ctx, query, calcuttaID, profile.EntryName)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profile.Bids = make([]models.LabEvaluationEntryBid, 0)
	profile.TotalBidPoints = 0
	for rows.Next() {
		var bid models.LabEvaluationEntryBid
		if err := rows.Scan(
			&bid.TeamID,
			&bid.SchoolName,
			&bid.Seed,
			&bid.Region,
			&bid.BidPoints,
			&bid.Ownership,
		); err != nil {
			return nil, err
		}
		profile.TotalBidPoints += bid.BidPoints
		profile.Bids = append(profile.Bids, bid)
	}

	return &profile, rows.Err()
}
