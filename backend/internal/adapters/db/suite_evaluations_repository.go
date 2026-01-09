package db

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/app/suite_evaluations"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SuiteEvaluationsRepository struct {
	pool *pgxpool.Pool
}

func NewSuiteEvaluationsRepository(pool *pgxpool.Pool) *SuiteEvaluationsRepository {
	return &SuiteEvaluationsRepository{pool: pool}
}

func (r *SuiteEvaluationsRepository) ListEvaluations(ctx context.Context, calcuttaID, cohortID, simulationBatchID *string, limit, offset int) ([]suite_evaluations.EvaluationListItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			r.id,
			r.simulation_run_batch_id,
			r.cohort_id,
			COALESCE(c.name, '') AS cohort_name,
			COALESCE(r.optimizer_key, c.optimizer_key, '') AS optimizer_key,
			COALESCE(r.n_sims, c.n_sims, 0) AS n_sims,
			COALESCE(r.seed, c.seed, 0) AS seed,
			r.our_rank,
			r.our_mean_normalized_payout,
			r.our_median_normalized_payout,
			r.our_p_top1,
			r.our_p_in_money,
			r.total_simulations,
			COALESCE(r.calcutta_id::text, '') AS calcutta_id,
			r.simulated_calcutta_id::text,
			r.game_outcome_run_id::text,
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
	`, uuidParamOrNil(calcuttaID), uuidParamOrNil(cohortID), uuidParamOrNil(simulationBatchID), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]suite_evaluations.EvaluationListItem, 0)
	for rows.Next() {
		var it suite_evaluations.EvaluationListItem
		if err := rows.Scan(
			&it.ID,
			&it.SimulationBatchID,
			&it.CohortID,
			&it.CohortName,
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
			&it.SimulatedCalcuttaID,
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

func (r *SuiteEvaluationsRepository) GetEvaluation(ctx context.Context, id string) (*suite_evaluations.EvaluationListItem, error) {
	var it suite_evaluations.EvaluationListItem
	err := r.pool.QueryRow(ctx, `
		SELECT
			r.id,
			r.simulation_run_batch_id,
			r.cohort_id,
			COALESCE(c.name, '') AS cohort_name,
			COALESCE(r.optimizer_key, c.optimizer_key, '') AS optimizer_key,
			COALESCE(r.n_sims, c.n_sims, 0) AS n_sims,
			COALESCE(r.seed, c.seed, 0) AS seed,
			r.our_rank,
			r.our_mean_normalized_payout,
			r.our_median_normalized_payout,
			r.our_p_top1,
			r.our_p_in_money,
			r.total_simulations,
			COALESCE(r.calcutta_id::text, '') AS calcutta_id,
			r.simulated_calcutta_id::text,
			r.game_outcome_run_id::text,
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
		&it.SimulationBatchID,
		&it.CohortID,
		&it.CohortName,
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
		&it.SimulatedCalcuttaID,
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, suite_evaluations.ErrSimulationNotFound
		}
		return nil, err
	}
	return &it, nil
}

func (r *SuiteEvaluationsRepository) GetSnapshotEntry(ctx context.Context, evalID, snapshotEntryID string) (*suite_evaluations.SnapshotEntry, error) {
	var calcuttaEvaluationRunID *string
	if err := r.pool.QueryRow(ctx, `
		SELECT calcutta_evaluation_run_id::text
		FROM derived.simulation_runs
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, evalID).Scan(&calcuttaEvaluationRunID); err != nil {
		return nil, err
	}
	if calcuttaEvaluationRunID == nil || strings.TrimSpace(*calcuttaEvaluationRunID) == "" {
		return nil, suite_evaluations.ErrEvaluationHasNoCalcuttaEvaluationRunID
	}

	var snapshotID string
	if err := r.pool.QueryRow(ctx, `
		SELECT calcutta_snapshot_id::text
		FROM derived.calcutta_evaluation_runs
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, *calcuttaEvaluationRunID).Scan(&snapshotID); err != nil {
		return nil, err
	}

	var displayName string
	var isSynthetic bool
	if err := r.pool.QueryRow(ctx, `
		SELECT display_name, is_synthetic
		FROM core.calcutta_snapshot_entries
		WHERE id = $1::uuid
			AND calcutta_snapshot_id = $2::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, snapshotEntryID, snapshotID).Scan(&displayName, &isSynthetic); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, suite_evaluations.ErrSnapshotEntryNotFoundForEvaluation
		}
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			t.id::text,
			s.name,
			COALESCE(t.seed, 0)::int,
			COALESCE(t.region, ''::text),
			cset.bid_points::int
		FROM core.calcutta_snapshot_entry_teams cset
		JOIN core.teams t ON t.id = cset.team_id AND t.deleted_at IS NULL
		JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
		WHERE cset.calcutta_snapshot_entry_id = $1::uuid
			AND cset.deleted_at IS NULL
		ORDER BY cset.bid_points DESC
	`, snapshotEntryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]suite_evaluations.SnapshotEntryTeam, 0)
	for rows.Next() {
		var t suite_evaluations.SnapshotEntryTeam
		if err := rows.Scan(&t.TeamID, &t.School, &t.Seed, &t.Region, &t.BidPoints); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &suite_evaluations.SnapshotEntry{SnapshotEntryID: snapshotEntryID, DisplayName: displayName, IsSynthetic: isSynthetic, Teams: teams}, nil
}

func (r *SuiteEvaluationsRepository) ListPortfolioBids(ctx context.Context, strategyGenerationRunID string) ([]suite_evaluations.PortfolioBid, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			t.id::text,
			s.name,
			COALESCE(t.seed, 0)::int,
			COALESCE(t.region, ''::text),
			reb.bid_points::int,
			COALESCE(reb.expected_roi, 0.0)::double precision
		FROM derived.strategy_generation_run_bids reb
		JOIN core.teams t ON t.id = reb.team_id AND t.deleted_at IS NULL
		JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
		WHERE reb.strategy_generation_run_id = $1::uuid
			AND reb.deleted_at IS NULL
		ORDER BY reb.bid_points DESC
	`, strategyGenerationRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]suite_evaluations.PortfolioBid, 0)
	for rows.Next() {
		var b suite_evaluations.PortfolioBid
		if err := rows.Scan(&b.TeamID, &b.SchoolName, &b.Seed, &b.Region, &b.BidPoints, &b.ExpectedROI); err != nil {
			return nil, err
		}
		items = append(items, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SuiteEvaluationsRepository) GetOurStrategyPerformance(ctx context.Context, calcuttaEvaluationRunID, evalID string) (*suite_evaluations.OurStrategyPerformance, error) {
	var tmp suite_evaluations.OurStrategyPerformance
	err := r.pool.QueryRow(ctx, `
		WITH focus AS (
			SELECT se.display_name
			FROM derived.simulation_runs sr
			JOIN core.calcutta_snapshot_entries se
				ON se.id = sr.focus_snapshot_entry_id
				AND se.deleted_at IS NULL
			WHERE sr.id = $2::uuid
				AND sr.deleted_at IS NULL
			LIMIT 1
		),
		ranked AS (
			SELECT
				ROW_NUMBER() OVER (ORDER BY COALESCE(ep.mean_normalized_payout, 0.0) DESC)::int AS rank,
				ep.entry_name,
				COALESCE(ep.mean_normalized_payout, 0.0)::double precision AS mean_normalized_payout,
				COALESCE(ep.median_normalized_payout, 0.0)::double precision AS median_normalized_payout,
				COALESCE(ep.p_top1, 0.0)::double precision AS p_top1,
				COALESCE(ep.p_in_money, 0.0)::double precision AS p_in_money
			FROM derived.entry_performance ep
			WHERE ep.calcutta_evaluation_run_id = $1::uuid
				AND ep.deleted_at IS NULL
		)
		SELECT
			r.rank,
			r.entry_name,
			r.mean_normalized_payout,
			r.median_normalized_payout,
			r.p_top1,
			r.p_in_money,
			COALESCE((
				SELECT st.n_sims::int
				FROM derived.calcutta_evaluation_runs cer
				JOIN derived.simulated_tournaments st
					ON st.id = cer.simulated_tournament_id
					AND st.deleted_at IS NULL
				WHERE cer.id = $1::uuid
					AND cer.deleted_at IS NULL
				LIMIT 1
			), 0)::int
		FROM ranked r
		WHERE r.entry_name = (SELECT display_name FROM focus)
		ORDER BY r.rank ASC
		LIMIT 1
	`, calcuttaEvaluationRunID, evalID).Scan(
		&tmp.Rank,
		&tmp.EntryName,
		&tmp.MeanNormalizedPayout,
		&tmp.MedianNormalizedPayout,
		&tmp.PTop1,
		&tmp.PInMoney,
		&tmp.TotalSimulations,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &tmp, nil
}

func (r *SuiteEvaluationsRepository) ListEntryPerformance(ctx context.Context, calcuttaEvaluationRunID string) ([]suite_evaluations.EntryPerformance, error) {
	rows, err := r.pool.Query(ctx, `
		WITH cer AS (
			SELECT calcutta_snapshot_id
			FROM derived.calcutta_evaluation_runs
			WHERE id = $1::uuid
				AND deleted_at IS NULL
			LIMIT 1
		),
		ranked AS (
			SELECT
				ROW_NUMBER() OVER (ORDER BY COALESCE(ep.mean_normalized_payout, 0.0) DESC)::int AS rank,
				ep.entry_name,
				COALESCE(ep.mean_normalized_payout, 0.0)::double precision AS mean_normalized_payout,
				COALESCE(ep.p_top1, 0.0)::double precision AS p_top1,
				COALESCE(ep.p_in_money, 0.0)::double precision AS p_in_money
			FROM derived.entry_performance ep
			WHERE ep.calcutta_evaluation_run_id = $1::uuid
				AND ep.deleted_at IS NULL
		)
		SELECT
			r.rank,
			r.entry_name,
			se.id::text,
			r.mean_normalized_payout,
			r.p_top1,
			r.p_in_money
		FROM ranked r
		LEFT JOIN core.calcutta_snapshot_entries se
			ON se.calcutta_snapshot_id = (SELECT calcutta_snapshot_id FROM cer)
			AND se.display_name = r.entry_name
			AND se.deleted_at IS NULL
		ORDER BY r.rank ASC
	`, calcuttaEvaluationRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]suite_evaluations.EntryPerformance, 0)
	for rows.Next() {
		var it suite_evaluations.EntryPerformance
		if err := rows.Scan(&it.Rank, &it.EntryName, &it.SnapshotEntryID, &it.MeanNormalizedPayout, &it.PTop1, &it.PInMoney); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SuiteEvaluationsRepository) GetSimulationBatchConfig(ctx context.Context, simulationBatchID string) (*suite_evaluations.SimulationBatchConfig, error) {
	var out suite_evaluations.SimulationBatchConfig
	if err := r.pool.QueryRow(ctx, `
		SELECT
			e.cohort_id::text,
			e.optimizer_key,
			e.n_sims,
			e.seed,
			e.starting_state_key,
			e.excluded_entry_name,
			COALESCE(s.game_outcomes_algorithm_id::text, ''::text),
			COALESCE(s.market_share_algorithm_id::text, ''::text),
			COALESCE(s.optimizer_key, ''::text),
			COALESCE(s.n_sims, 0)::int,
			COALESCE(s.seed, 0)::int
		FROM derived.simulation_run_batches e
		JOIN derived.synthetic_calcutta_cohorts s ON s.id = e.cohort_id AND s.deleted_at IS NULL
		WHERE e.id = $1::uuid
			AND e.deleted_at IS NULL
		LIMIT 1
	`, simulationBatchID).Scan(
		&out.CohortID,
		&out.OptimizerKey,
		&out.NSims,
		&out.Seed,
		&out.StartingStateKey,
		&out.ExcludedEntryName,
		&out.GameOutcomesAlgID,
		&out.MarketShareAlgID,
		&out.CohortOptimizerKey,
		&out.CohortNSims,
		&out.CohortSeed,
	); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *SuiteEvaluationsRepository) GetCohortOptimizerKey(ctx context.Context, cohortID string) (string, error) {
	v := ""
	_ = r.pool.QueryRow(ctx, `
		SELECT COALESCE(optimizer_key, '')
		FROM derived.synthetic_calcutta_cohorts
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, cohortID).Scan(&v)
	return v, nil
}

func (r *SuiteEvaluationsRepository) GetTournamentIDForCalcutta(ctx context.Context, calcuttaID string) (string, error) {
	tournamentID := ""
	if err := r.pool.QueryRow(ctx, `
		SELECT tournament_id::text
		FROM core.calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, calcuttaID).Scan(&tournamentID); err != nil {
		return "", err
	}
	return tournamentID, nil
}

func (r *SuiteEvaluationsRepository) GetTournamentIDForSimulatedCalcutta(ctx context.Context, simulatedCalcuttaID string) (string, error) {
	tournamentID := ""
	if err := r.pool.QueryRow(ctx, `
		SELECT tournament_id::text
		FROM derived.simulated_calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, simulatedCalcuttaID).Scan(&tournamentID); err != nil {
		return "", err
	}
	return tournamentID, nil
}

func (r *SuiteEvaluationsRepository) GetLatestGameOutcomeRunID(ctx context.Context, tournamentID, algorithmID string) (string, error) {
	resolved := ""
	_ = r.pool.QueryRow(ctx, `
		SELECT COALESCE((
			SELECT id::text
			FROM derived.game_outcome_runs
			WHERE tournament_id = $1::uuid
				AND algorithm_id = $2::uuid
				AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1
		), ''::text)
	`, tournamentID, algorithmID).Scan(&resolved)
	return resolved, nil
}

func (r *SuiteEvaluationsRepository) GetLatestMarketShareRunID(ctx context.Context, calcuttaID, algorithmID string) (string, error) {
	resolved := ""
	_ = r.pool.QueryRow(ctx, `
		SELECT COALESCE((
			SELECT id::text
			FROM derived.market_share_runs
			WHERE calcutta_id = $1::uuid
				AND algorithm_id = $2::uuid
				AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1
		), ''::text)
	`, calcuttaID, algorithmID).Scan(&resolved)
	return resolved, nil
}

func (r *SuiteEvaluationsRepository) UpsertSyntheticCalcutta(ctx context.Context, cohortID, calcuttaID string) (string, *string, *string, error) {
	var syntheticCalcuttaID string
	var syntheticSnapshotID *string
	var existingExcludedEntryName *string
	if err := r.pool.QueryRow(ctx, `
		INSERT INTO derived.synthetic_calcuttas (cohort_id, calcutta_id)
		VALUES ($1::uuid, $2::uuid)
		ON CONFLICT (cohort_id, calcutta_id) WHERE deleted_at IS NULL
		DO UPDATE SET updated_at = NOW(), deleted_at = NULL
		RETURNING id::text, calcutta_snapshot_id::text, excluded_entry_name
	`, cohortID, calcuttaID).Scan(&syntheticCalcuttaID, &syntheticSnapshotID, &existingExcludedEntryName); err != nil {
		return "", nil, nil, err
	}
	return syntheticCalcuttaID, syntheticSnapshotID, existingExcludedEntryName, nil
}

func (r *SuiteEvaluationsRepository) EnsureSyntheticSnapshot(ctx context.Context, syntheticCalcuttaID, calcuttaID string, excludedEntryName *string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	createdSnapshotID, err := createSyntheticCalcuttaSnapshot(ctx, tx, calcuttaID, excludedEntryName, "", nil)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE derived.synthetic_calcuttas
		SET calcutta_snapshot_id = $2::uuid,
			updated_at = NOW()
		WHERE id = $1::uuid
			AND deleted_at IS NULL
	`, syntheticCalcuttaID, createdSnapshotID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *SuiteEvaluationsRepository) CreateSimulationRun(ctx context.Context, p suite_evaluations.CreateSimulationParams, syntheticCalcuttaID string) (*suite_evaluations.CreateSimulationResult, error) {
	evalID := ""
	status := ""

	var execID any
	if p.SimulationRunBatchID != nil && *p.SimulationRunBatchID != "" {
		execID = *p.SimulationRunBatchID
	}
	var goRun any
	if p.GameOutcomeRunID != nil {
		goRun = *p.GameOutcomeRunID
	}
	var goSpec any
	if p.GameOutcomeSpec != nil {
		b, err := json.Marshal(p.GameOutcomeSpec)
		if err != nil {
			return nil, err
		}
		goSpec = b
	}
	var msRun any
	if p.MarketShareRunID != nil {
		msRun = *p.MarketShareRunID
	}
	var simulatedCalcutta any
	if p.SimulatedCalcuttaID != nil {
		simulatedCalcutta = *p.SimulatedCalcuttaID
	}
	var excluded any
	if p.ExcludedEntryName != nil {
		excluded = *p.ExcludedEntryName
	}
	var nSims any
	if p.NSims != nil {
		nSims = *p.NSims
	}
	var seed any
	if p.Seed != nil {
		seed = *p.Seed
	}

	q := `
		INSERT INTO derived.simulation_runs (
			simulation_run_batch_id,
			synthetic_calcutta_id,
			cohort_id,
			calcutta_id,
			simulated_calcutta_id,
			game_outcome_run_id,
			game_outcome_spec_json,
			market_share_run_id,
			optimizer_key,
			n_sims,
			seed,
			starting_state_key,
			excluded_entry_name
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid, $6::uuid, $7::jsonb, $8::uuid, $9, $10::int, $11::int, $12, $13::text)
		RETURNING id, status
	`

	var calcutta any
	if strings.TrimSpace(p.CalcuttaID) != "" {
		calcutta = p.CalcuttaID
	}
	var synth any
	if strings.TrimSpace(syntheticCalcuttaID) != "" {
		synth = syntheticCalcuttaID
	}

	if err := r.pool.QueryRow(ctx, q, execID, synth, p.CohortID, calcutta, simulatedCalcutta, goRun, goSpec, msRun, p.OptimizerKey, nSims, seed, p.StartingStateKey, excluded).Scan(&evalID, &status); err != nil {
		return nil, err
	}
	return &suite_evaluations.CreateSimulationResult{ID: evalID, Status: status}, nil
}

func (r *SuiteEvaluationsRepository) LoadPayouts(ctx context.Context, calcuttaID string) ([]*models.CalcuttaPayout, error) {
	payoutRows, err := r.pool.Query(ctx, `
		SELECT position::int, amount_cents::int
		FROM core.payouts
		WHERE calcutta_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY position ASC
	`, calcuttaID)
	if err != nil {
		return nil, err
	}
	defer payoutRows.Close()

	payouts := make([]*models.CalcuttaPayout, 0)
	for payoutRows.Next() {
		var pos int
		var cents int
		if err := payoutRows.Scan(&pos, &cents); err != nil {
			return nil, err
		}
		payouts = append(payouts, &models.CalcuttaPayout{CalcuttaID: calcuttaID, Position: pos, AmountCents: cents})
	}
	if err := payoutRows.Err(); err != nil {
		return nil, err
	}
	return payouts, nil
}

func (r *SuiteEvaluationsRepository) LoadTeamPoints(ctx context.Context, calcuttaID string) (map[string]float64, error) {
	teamRows, err := r.pool.Query(ctx, `
		WITH t AS (
			SELECT tournament_id
			FROM core.calcuttas
			WHERE id = $1::uuid
				AND deleted_at IS NULL
			LIMIT 1
		)
		SELECT
			team.id::text,
			core.calcutta_points_for_progress($1::uuid, team.wins, team.byes)::float8
		FROM core.teams team
		JOIN t ON t.tournament_id = team.tournament_id
		WHERE team.deleted_at IS NULL
	`, calcuttaID)
	if err != nil {
		return nil, err
	}
	defer teamRows.Close()

	teamPoints := make(map[string]float64)
	for teamRows.Next() {
		var teamID string
		var pts float64
		if err := teamRows.Scan(&teamID, &pts); err != nil {
			return nil, err
		}
		teamPoints[teamID] = pts
	}
	if err := teamRows.Err(); err != nil {
		return nil, err
	}
	return teamPoints, nil
}

func (r *SuiteEvaluationsRepository) LoadEntryBids(ctx context.Context, calcuttaID string) ([]suite_evaluations.EntryBidRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			e.id::text,
			e.name,
			e.created_at,
			et.team_id::text,
			et.bid_points::int
		FROM core.entries e
		JOIN core.entry_teams et ON et.entry_id = e.id AND et.deleted_at IS NULL
		WHERE e.calcutta_id = $1::uuid
			AND e.deleted_at IS NULL
		ORDER BY e.created_at ASC
	`, calcuttaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]suite_evaluations.EntryBidRow, 0)
	for rows.Next() {
		var r0 suite_evaluations.EntryBidRow
		if err := rows.Scan(&r0.EntryID, &r0.Name, &r0.CreatedAt, &r0.TeamID, &r0.BidPoints); err != nil {
			return nil, err
		}
		out = append(out, r0)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SuiteEvaluationsRepository) LoadStrategyBids(ctx context.Context, strategyGenerationRunID string) (map[string]float64, error) {
	ourBids := make(map[string]float64)
	ourRows, err := r.pool.Query(ctx, `
		SELECT team_id::text, bid_points::int
		FROM derived.strategy_generation_run_bids
		WHERE strategy_generation_run_id = $1::uuid
			AND deleted_at IS NULL
	`, strategyGenerationRunID)
	if err != nil {
		return nil, err
	}
	defer ourRows.Close()
	for ourRows.Next() {
		var teamID string
		var bid int
		if err := ourRows.Scan(&teamID, &bid); err != nil {
			return nil, err
		}
		ourBids[teamID] += float64(bid)
	}
	if err := ourRows.Err(); err != nil {
		return nil, err
	}
	return ourBids, nil
}
