package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
)

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
		return nil, fmt.Errorf("listing evaluations: %w", err)
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
			return nil, fmt.Errorf("scanning evaluation: %w", err)
		}
		out = append(out, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating evaluations: %w", err)
	}
	return out, nil
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
		return nil, fmt.Errorf("getting evaluation %s: %w", id, err)
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
		return nil, fmt.Errorf("querying evaluation entry results for evaluation %s: %w", evaluationID, err)
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
			return nil, fmt.Errorf("scanning evaluation entry result: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating evaluation entry results: %w", err)
	}
	return out, nil
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
		return nil, fmt.Errorf("getting evaluation entry result %s: %w", entryResultID, err)
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
		return nil, fmt.Errorf("getting entry and calcutta for evaluation %s: %w", evaluationID, err)
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
		return nil, fmt.Errorf("querying bids for entry result %s: %w", entryResultID, err)
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
			return nil, fmt.Errorf("scanning entry bid: %w", err)
		}
		profile.TotalBidPoints += bid.BidPoints
		profile.Bids = append(profile.Bids, bid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating entry bids: %w", err)
	}

	return &profile, nil
}

// UpdateEvaluationSummary persists the computed summary JSON for an evaluation.
func (r *LabRepository) UpdateEvaluationSummary(ctx context.Context, evaluationID string, summaryJSON []byte) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE lab.evaluations SET summary_json = $2::jsonb WHERE id = $1::uuid`,
		evaluationID, summaryJSON,
	)
	if err != nil {
		return fmt.Errorf("updating evaluation summary %s: %w", evaluationID, err)
	}
	return nil
}

// GetBaselineEvaluation finds the naive_ev evaluation for the same calcutta
// and starting state key. Returns nil (no error) if none found.
func (r *LabRepository) GetBaselineEvaluation(ctx context.Context, calcuttaID, startingStateKey string) (*models.LabEvaluationDetail, error) {
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
		WHERE im.kind = 'naive_ev'
			AND e.calcutta_id = $1::uuid
			AND e.starting_state_key = $2
			AND ev.deleted_at IS NULL
			AND e.deleted_at IS NULL
			AND im.deleted_at IS NULL
		ORDER BY ev.created_at DESC
		LIMIT 1
	`

	var ev models.LabEvaluationDetail
	err := r.pool.QueryRow(ctx, query, calcuttaID, startingStateKey).Scan(
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
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting baseline evaluation for calcutta %s: %w", calcuttaID, err)
	}
	return &ev, nil
}

// GetEvaluationSummaryJSON returns the raw summary_json for an evaluation.
// Returns nil (no error) if the column is NULL.
func (r *LabRepository) GetEvaluationSummaryJSON(ctx context.Context, evaluationID string) ([]byte, error) {
	var raw []byte
	err := r.pool.QueryRow(ctx,
		`SELECT summary_json FROM lab.evaluations WHERE id = $1::uuid AND deleted_at IS NULL`,
		evaluationID,
	).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "evaluation", ID: evaluationID}
	}
	if err != nil {
		return nil, fmt.Errorf("getting evaluation summary %s: %w", evaluationID, err)
	}
	return raw, nil
}
