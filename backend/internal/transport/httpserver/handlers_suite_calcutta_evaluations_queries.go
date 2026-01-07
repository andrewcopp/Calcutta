package httpserver

import (
	"context"
)

func (s *Server) loadSuiteCalcuttaEvaluations(
	ctx context.Context,
	calcuttaID string,
	suiteID string,
	suiteExecutionID string,
	limit int,
	offset int,
) ([]suiteCalcuttaEvaluationListItem, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			r.id,
			r.simulation_run_batch_id,
			r.cohort_id,
			COALESCE(c.name, '') AS suite_name,
			COALESCE(r.optimizer_key, c.optimizer_key, '') AS optimizer_key,
			COALESCE(r.n_sims, c.n_sims, 0) AS n_sims,
			COALESCE(r.seed, c.seed, 0) AS seed,
			r.our_rank,
			r.our_mean_normalized_payout,
			r.our_median_normalized_payout,
			r.our_p_top1,
			r.our_p_in_money,
			r.total_simulations,
			r.calcutta_id,
			r.game_outcome_run_id,
			r.market_share_run_id,
			r.strategy_generation_run_id,
			r.calcutta_evaluation_run_id,
			r.realized_finish_position,
			r.realized_is_tied,
			r.realized_in_the_money,
			r.realized_payout_cents,
			r.realized_total_points,
			r.starting_state_key,
			r.excluded_entry_name,
			r.status,
			r.claimed_at,
			r.claimed_by,
			r.error_message,
			r.created_at,
			r.updated_at
		FROM derived.simulation_runs r
		LEFT JOIN derived.synthetic_calcutta_cohorts c
			ON c.id = r.cohort_id
			AND c.deleted_at IS NULL
		WHERE r.deleted_at IS NULL
			AND ($1::uuid IS NULL OR r.calcutta_id = $1::uuid)
			AND ($2::uuid IS NULL OR r.cohort_id = $2::uuid)
			AND ($3::uuid IS NULL OR r.simulation_run_batch_id = $3::uuid)
		ORDER BY r.created_at DESC
		LIMIT $4::int
		OFFSET $5::int
	`, nullUUIDParam(calcuttaID), nullUUIDParam(suiteID), nullUUIDParam(suiteExecutionID), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]suiteCalcuttaEvaluationListItem, 0)
	for rows.Next() {
		var it suiteCalcuttaEvaluationListItem
		if err := rows.Scan(
			&it.ID,
			&it.SuiteExecutionID,
			&it.SuiteID,
			&it.SuiteName,
			&it.OptimizerKey,
			&it.NSims,
			&it.Seed,
			&it.OurRank,
			&it.OurMeanNormalizedPayout,
			&it.OurMedianNormalizedPayout,
			&it.OurPTop1,
			&it.OurPInMoney,
			&it.TotalSimulations,
			&it.CalcuttaID,
			&it.GameOutcomeRunID,
			&it.MarketShareRunID,
			&it.StrategyGenerationRunID,
			&it.CalcuttaEvaluationRunID,
			&it.RealizedFinishPosition,
			&it.RealizedIsTied,
			&it.RealizedInTheMoney,
			&it.RealizedPayoutCents,
			&it.RealizedTotalPoints,
			&it.StartingStateKey,
			&it.ExcludedEntryName,
			&it.Status,
			&it.ClaimedAt,
			&it.ClaimedBy,
			&it.ErrorMessage,
			&it.CreatedAt,
			&it.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Server) loadSuiteCalcuttaEvaluationByID(ctx context.Context, id string) (*suiteCalcuttaEvaluationListItem, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var it suiteCalcuttaEvaluationListItem
	err := s.pool.QueryRow(ctx, `
		SELECT
			r.id,
			r.simulation_run_batch_id,
			r.cohort_id,
			COALESCE(c.name, '') AS suite_name,
			COALESCE(r.optimizer_key, c.optimizer_key, '') AS optimizer_key,
			COALESCE(r.n_sims, c.n_sims, 0) AS n_sims,
			COALESCE(r.seed, c.seed, 0) AS seed,
			r.our_rank,
			r.our_mean_normalized_payout,
			r.our_median_normalized_payout,
			r.our_p_top1,
			r.our_p_in_money,
			r.total_simulations,
			r.calcutta_id,
			r.game_outcome_run_id,
			r.market_share_run_id,
			r.strategy_generation_run_id,
			r.calcutta_evaluation_run_id,
			r.realized_finish_position,
			r.realized_is_tied,
			r.realized_in_the_money,
			r.realized_payout_cents,
			r.realized_total_points,
			r.starting_state_key,
			r.excluded_entry_name,
			r.status,
			r.claimed_at,
			r.claimed_by,
			r.error_message,
			r.created_at,
			r.updated_at
		FROM derived.simulation_runs r
		LEFT JOIN derived.synthetic_calcutta_cohorts c
			ON c.id = r.cohort_id
			AND c.deleted_at IS NULL
		WHERE r.id = $1::uuid
			AND r.deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.SuiteExecutionID,
		&it.SuiteID,
		&it.SuiteName,
		&it.OptimizerKey,
		&it.NSims,
		&it.Seed,
		&it.OurRank,
		&it.OurMeanNormalizedPayout,
		&it.OurMedianNormalizedPayout,
		&it.OurPTop1,
		&it.OurPInMoney,
		&it.TotalSimulations,
		&it.CalcuttaID,
		&it.GameOutcomeRunID,
		&it.MarketShareRunID,
		&it.StrategyGenerationRunID,
		&it.CalcuttaEvaluationRunID,
		&it.RealizedFinishPosition,
		&it.RealizedIsTied,
		&it.RealizedInTheMoney,
		&it.RealizedPayoutCents,
		&it.RealizedTotalPoints,
		&it.StartingStateKey,
		&it.ExcludedEntryName,
		&it.Status,
		&it.ClaimedAt,
		&it.ClaimedBy,
		&it.ErrorMessage,
		&it.CreatedAt,
		&it.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &it, nil
}
