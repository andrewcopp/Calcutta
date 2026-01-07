package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	appcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"
	reb "github.com/andrewcopp/Calcutta/backend/internal/features/recommended_entry_bids"
	appsimulatetournaments "github.com/andrewcopp/Calcutta/backend/internal/features/simulate_tournaments"
	appsimulatedcalcutta "github.com/andrewcopp/Calcutta/backend/internal/features/simulated_calcutta"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultSuiteEvalWorkerPollInterval = 2 * time.Second
	defaultSuiteEvalWorkerStaleAfter   = 30 * time.Minute
)

type suiteCalcuttaEvaluationRow struct {
	ID               string
	RunKey           string
	SuiteExecutionID *string
	SuiteID          string
	CalcuttaID       string
	GameOutcomeRunID string
	MarketShareRunID string
	StrategyGenRunID *string
	OptimizerKey     *string
	NSims            *int
	Seed             *int
	StartingStateKey string
	ExcludedEntry    *string
}

func (s *Server) RunSuiteCalcuttaEvaluationWorker(ctx context.Context) {
	s.RunSuiteCalcuttaEvaluationWorkerWithOptions(ctx, defaultSuiteEvalWorkerPollInterval, defaultSuiteEvalWorkerStaleAfter)
}

func (s *Server) RunSuiteCalcuttaEvaluationWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if s.pool == nil {
		log.Printf("suite eval worker disabled: database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultSuiteEvalWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultSuiteEvalWorkerStaleAfter
	}

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	workerID := os.Getenv("HOSTNAME")
	if workerID == "" {
		workerID = "suite-eval-worker"
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			req, ok, err := s.claimNextSuiteCalcuttaEvaluation(ctx, workerID, staleAfter)
			if err != nil {
				log.Printf("Error claiming next suite calcutta evaluation: %v", err)
				continue
			}
			if !ok {
				continue
			}

			_ = s.processSuiteCalcuttaEvaluation(ctx, workerID, req)
		}
	}
}

func (s *Server) claimNextSuiteCalcuttaEvaluation(ctx context.Context, workerID string, staleAfter time.Duration) (*suiteCalcuttaEvaluationRow, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	now := time.Now().UTC()
	staleBefore := now.Add(-staleAfter)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	var runID string
	q := `
		WITH candidate AS (
			SELECT id
			FROM derived.run_jobs
			WHERE run_kind = 'simulation'
				AND (
					status = 'queued'
					OR (
						status = 'running'
						AND claimed_at IS NOT NULL
						AND claimed_at < $2
					)
				)
			ORDER BY created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE derived.run_jobs j
		SET status = 'running',
			attempt = j.attempt + 1,
			claimed_at = $1,
			claimed_by = $3,
			started_at = COALESCE(j.started_at, $1),
			finished_at = NULL,
			error_message = NULL,
			updated_at = NOW()
		FROM candidate
		WHERE j.id = candidate.id
		RETURNING j.run_id::text
	`
	if err := tx.QueryRow(ctx, q,
		pgtype.Timestamptz{Time: now, Valid: true},
		pgtype.Timestamptz{Time: staleBefore, Valid: true},
		workerID,
	).Scan(&runID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	row := &suiteCalcuttaEvaluationRow{}
	q2 := `
		UPDATE derived.suite_calcutta_evaluations r
		SET status = 'running',
			claimed_at = $1,
			claimed_by = $3,
			error_message = NULL
		WHERE r.id = $2::uuid
		RETURNING
			r.id,
			r.run_key,
			r.suite_execution_id,
			r.suite_id,
			r.calcutta_id,
			r.game_outcome_run_id,
			r.market_share_run_id,
			r.strategy_generation_run_id,
			r.optimizer_key,
			r.n_sims,
			r.seed,
			r.starting_state_key,
			r.excluded_entry_name
	`

	var excluded *string
	if err := tx.QueryRow(ctx, q2,
		pgtype.Timestamptz{Time: now, Valid: true},
		runID,
		workerID,
	).Scan(
		&row.ID,
		&row.RunKey,
		&row.SuiteExecutionID,
		&row.SuiteID,
		&row.CalcuttaID,
		&row.GameOutcomeRunID,
		&row.MarketShareRunID,
		&row.StrategyGenRunID,
		&row.OptimizerKey,
		&row.NSims,
		&row.Seed,
		&row.StartingStateKey,
		&excluded,
	); err != nil {
		return nil, false, err
	}
	row.ExcludedEntry = excluded

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	committed = true

	return row, true, nil
}

func (s *Server) processSuiteCalcuttaEvaluation(ctx context.Context, workerID string, req *suiteCalcuttaEvaluationRow) bool {
	if req == nil {
		return false
	}

	runKey := req.RunKey
	if runKey == "" {
		runKey = uuid.NewString()
	}

	strategyGenRunID := ""
	if req.StrategyGenRunID != nil {
		strategyGenRunID = *req.StrategyGenRunID
	}
	usingExistingStrategyGenRun := strategyGenRunID != ""

	excluded := ""
	if req.ExcludedEntry != nil {
		excluded = *req.ExcludedEntry
	}

	log.Printf("suite_eval_worker start worker_id=%s eval_id=%s suite_id=%s calcutta_id=%s run_key=%s game_outcome_run_id=%s market_share_run_id=%s strategy_generation_run_id=%s starting_state_key=%s excluded_entry_name=%q",
		workerID,
		req.ID,
		req.SuiteID,
		req.CalcuttaID,
		runKey,
		req.GameOutcomeRunID,
		req.MarketShareRunID,
		strategyGenRunID,
		req.StartingStateKey,
		excluded,
	)

	year, err := s.resolveSeasonYearByCalcuttaID(ctx, req.CalcuttaID)
	if err != nil {
		s.failSuiteCalcuttaEvaluation(ctx, req.ID, err)
		return false
	}

	simSvc := appsimulatetournaments.New(s.pool)
	goRunID := req.GameOutcomeRunID
	nSims := 0
	if req.NSims != nil {
		nSims = *req.NSims
	}
	if nSims <= 0 {
		nSims = s.resolveSuiteNSims(ctx, req.SuiteID, 10000)
	}
	seed := 0
	if req.Seed != nil {
		seed = *req.Seed
	}
	if seed == 0 {
		seed = s.resolveSuiteSeed(ctx, req.SuiteID, 42)
	}
	simRes, err := simSvc.Run(ctx, appsimulatetournaments.RunParams{
		Season:               year,
		NSims:                nSims,
		Seed:                 seed,
		Workers:              0,
		BatchSize:            500,
		ProbabilitySourceKey: "suite_eval_worker",
		StartingStateKey:     req.StartingStateKey,
		GameOutcomeRunID:     &goRunID,
	})
	if err != nil {
		s.failSuiteCalcuttaEvaluation(ctx, req.ID, err)
		return false
	}

	optimizerKey := ""
	if req.OptimizerKey != nil {
		optimizerKey = *req.OptimizerKey
	}
	if optimizerKey == "" {
		optimizerKey = s.resolveSuiteOptimizerKey(ctx, req.SuiteID, "minlp_v1")
	}

	if !usingExistingStrategyGenRun {
		rebSvc := reb.New(s.pool)
		msRunID := req.MarketShareRunID
		genRes, err := rebSvc.GenerateAndWrite(ctx, reb.GenerateParams{
			CalcuttaID:            req.CalcuttaID,
			RunKey:                runKey,
			Name:                  "suite_eval_worker",
			OptimizerKey:          optimizerKey,
			MarketShareRunID:      &msRunID,
			SimulatedTournamentID: &simRes.TournamentSimulationBatchID,
		})
		if err != nil {
			s.failSuiteCalcuttaEvaluation(ctx, req.ID, err)
			return false
		}
		strategyGenRunID = genRes.StrategyGenerationRunID
	}

	evalSvc := appsimulatedcalcutta.New(s.pool)
	evalRunID := ""
	if usingExistingStrategyGenRun {
		evalRunID, err = evalSvc.CalculateSimulatedCalcuttaForStrategyGenerationRun(
			ctx,
			req.CalcuttaID,
			runKey,
			excluded,
			&simRes.TournamentSimulationBatchID,
			strategyGenRunID,
		)
	} else {
		evalRunID, err = evalSvc.CalculateSimulatedCalcuttaForEvaluationRun(
			ctx,
			req.CalcuttaID,
			runKey,
			excluded,
			&simRes.TournamentSimulationBatchID,
		)
	}
	if err != nil {
		s.failSuiteCalcuttaEvaluation(ctx, req.ID, err)
		return false
	}

	var (
		ourRank                   *int
		ourMeanNormalizedPayout   *float64
		ourMedianNormalizedPayout *float64
		ourPTop1                  *float64
		ourPInMoney               *float64
		totalSimulations          *int
		realizedFinishPosition    *int
		realizedIsTied            *bool
		realizedInTheMoney        *bool
		realizedPayoutCents       *int
		realizedTotalPoints       *float64
	)
	if evalRunID != "" {
		var (
			rank                   int
			meanNormalizedPayout   float64
			medianNormalizedPayout float64
			pTop1                  float64
			pInMoney               float64
			totalSims              int
		)
		err := s.pool.QueryRow(ctx, `
			WITH ranked AS (
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
				), 0)::int as total_simulations
			FROM ranked r
			WHERE r.entry_name IN ('Our Strategy', 'our_strategy', 'Out Strategy')
			ORDER BY r.rank ASC
			LIMIT 1
		`, evalRunID).Scan(
			&rank,
			&meanNormalizedPayout,
			&medianNormalizedPayout,
			&pTop1,
			&pInMoney,
			&totalSims,
		)
		if err == nil {
			ourRank = &rank
			ourMeanNormalizedPayout = &meanNormalizedPayout
			ourMedianNormalizedPayout = &medianNormalizedPayout
			ourPTop1 = &pTop1
			ourPInMoney = &pInMoney
			totalSimulations = &totalSims
		}
	}

	if realized, ok, err := s.computeRealizedFinishForStrategyGenerationRun(ctx, req.CalcuttaID, strategyGenRunID); err != nil {
		log.Printf("Error computing realized finish for eval_id=%s: %v", req.ID, err)
	} else if ok {
		realizedFinishPosition = &realized.FinishPosition
		realizedIsTied = &realized.IsTied
		realizedInTheMoney = &realized.InTheMoney
		realizedPayoutCents = &realized.PayoutCents
		realizedTotalPoints = &realized.TotalPoints
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE derived.suite_calcutta_evaluations
		SET status = 'succeeded',
			optimizer_key = $2,
			n_sims = $3,
			seed = $4,
			strategy_generation_run_id = $5::uuid,
			calcutta_evaluation_run_id = $6::uuid,
			our_rank = $7,
			our_mean_normalized_payout = $8,
			our_median_normalized_payout = $9,
			our_p_top1 = $10,
			our_p_in_money = $11,
			total_simulations = $12,
			realized_finish_position = $13,
			realized_is_tied = $14,
			realized_in_the_money = $15,
			realized_payout_cents = $16,
			realized_total_points = $17,
			error_message = NULL,
			updated_at = NOW()
		WHERE id = $1::uuid
	`,
		req.ID,
		optimizerKey,
		nSims,
		seed,
		strategyGenRunID,
		evalRunID,
		ourRank,
		ourMeanNormalizedPayout,
		ourMedianNormalizedPayout,
		ourPTop1,
		ourPInMoney,
		totalSimulations,
		realizedFinishPosition,
		realizedIsTied,
		realizedInTheMoney,
		realizedPayoutCents,
		realizedTotalPoints,
	)
	if err != nil {
		log.Printf("Error updating suite calcutta evaluation %s to succeeded: %v", req.ID, err)
		return false
	}

	_, _ = s.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'succeeded',
			finished_at = NOW(),
			error_message = NULL,
			updated_at = NOW()
		WHERE run_kind = 'simulation'
			AND run_id = $1::uuid
	`, req.ID)

	summary := map[string]any{
		"status":                    "succeeded",
		"evaluationId":              req.ID,
		"runKey":                    runKey,
		"optimizerKey":              optimizerKey,
		"nSims":                     nSims,
		"seed":                      seed,
		"strategyGenerationRunId":   strategyGenRunID,
		"calcuttaEvaluationRunId":   evalRunID,
		"ourRank":                   ourRank,
		"ourMeanNormalizedPayout":   ourMeanNormalizedPayout,
		"ourMedianNormalizedPayout": ourMedianNormalizedPayout,
		"ourPTop1":                  ourPTop1,
		"ourPInMoney":               ourPInMoney,
		"totalSimulations":          totalSimulations,
		"realizedFinishPosition":    realizedFinishPosition,
		"realizedIsTied":            realizedIsTied,
		"realizedInTheMoney":        realizedInTheMoney,
		"realizedPayoutCents":       realizedPayoutCents,
		"realizedTotalPoints":       realizedTotalPoints,
	}
	summaryJSON, err := json.Marshal(summary)
	if err == nil {
		var runKeyParam any
		if runKey != "" {
			runKeyParam = runKey
		} else {
			runKeyParam = nil
		}
		_, _ = s.pool.Exec(ctx, `
			INSERT INTO derived.run_artifacts (
				run_kind,
				run_id,
				run_key,
				artifact_kind,
				schema_version,
				storage_uri,
				summary_json
			)
			VALUES ('simulation', $1::uuid, $2::uuid, 'metrics', 'v1', NULL, $3::jsonb)
			ON CONFLICT (run_kind, run_id, artifact_kind) WHERE deleted_at IS NULL
			DO UPDATE
			SET run_key = EXCLUDED.run_key,
				schema_version = EXCLUDED.schema_version,
				storage_uri = EXCLUDED.storage_uri,
				summary_json = EXCLUDED.summary_json,
				updated_at = NOW(),
				deleted_at = NULL
		`, req.ID, runKeyParam, summaryJSON)
	}

	log.Printf("suite_eval_worker success worker_id=%s eval_id=%s run_key=%s strategy_generation_run_id=%s calcutta_evaluation_run_id=%s",
		workerID,
		req.ID,
		runKey,
		strategyGenRunID,
		evalRunID,
	)

	if req.SuiteExecutionID != nil && *req.SuiteExecutionID != "" {
		s.updateSuiteExecutionStatus(ctx, *req.SuiteExecutionID)
	}
	return true
}

type realizedFinishResult struct {
	FinishPosition int
	IsTied         bool
	InTheMoney     bool
	PayoutCents    int
	TotalPoints    float64
}

func (s *Server) computeRealizedFinishForStrategyGenerationRun(ctx context.Context, calcuttaID string, strategyGenerationRunID string) (*realizedFinishResult, bool, error) {
	if calcuttaID == "" || strategyGenerationRunID == "" {
		return nil, false, nil
	}

	// Load payouts
	payoutRows, err := s.pool.Query(ctx, `
		SELECT position::int, amount_cents::int
		FROM core.payouts
		WHERE calcutta_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY position ASC
	`, calcuttaID)
	if err != nil {
		return nil, false, err
	}
	defer payoutRows.Close()

	payouts := make([]*models.CalcuttaPayout, 0)
	for payoutRows.Next() {
		var pos int
		var cents int
		if err := payoutRows.Scan(&pos, &cents); err != nil {
			return nil, false, err
		}
		payouts = append(payouts, &models.CalcuttaPayout{CalcuttaID: calcuttaID, Position: pos, AmountCents: cents})
	}
	if err := payoutRows.Err(); err != nil {
		return nil, false, err
	}

	// Load team points (actual wins/byes)
	teamRows, err := s.pool.Query(ctx, `
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
		return nil, false, err
	}
	defer teamRows.Close()

	teamPoints := make(map[string]float64)
	for teamRows.Next() {
		var teamID string
		var pts float64
		if err := teamRows.Scan(&teamID, &pts); err != nil {
			return nil, false, err
		}
		teamPoints[teamID] = pts
	}
	if err := teamRows.Err(); err != nil {
		return nil, false, err
	}

	// Load real entries and their bids
	rows, err := s.pool.Query(ctx, `
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
		return nil, false, err
	}
	defer rows.Close()

	entryByID := make(map[string]*models.CalcuttaEntry)
	entryBids := make(map[string]map[string]float64)
	existingTotalBid := make(map[string]float64)
	for rows.Next() {
		var entryID, name, teamID string
		var createdAt time.Time
		var bid int
		if err := rows.Scan(&entryID, &name, &createdAt, &teamID, &bid); err != nil {
			return nil, false, err
		}
		if _, ok := entryByID[entryID]; !ok {
			entryByID[entryID] = &models.CalcuttaEntry{ID: entryID, Name: name, CalcuttaID: calcuttaID, Created: createdAt}
			entryBids[entryID] = make(map[string]float64)
		}
		entryBids[entryID][teamID] += float64(bid)
		existingTotalBid[teamID] += float64(bid)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	// Load our strategy bids
	ourBids := make(map[string]float64)
	ourRows, err := s.pool.Query(ctx, `
		SELECT team_id::text, bid_points::int
		FROM derived.recommended_entry_bids
		WHERE strategy_generation_run_id = $1::uuid
			AND deleted_at IS NULL
	`, strategyGenerationRunID)
	if err != nil {
		return nil, false, err
	}
	defer ourRows.Close()
	for ourRows.Next() {
		var teamID string
		var bid int
		if err := ourRows.Scan(&teamID, &bid); err != nil {
			return nil, false, err
		}
		ourBids[teamID] += float64(bid)
	}
	if err := ourRows.Err(); err != nil {
		return nil, false, err
	}
	if len(ourBids) == 0 {
		return nil, false, nil
	}

	// Compute points for each real entry under the new total bids (including ours)
	entries := make([]*models.CalcuttaEntry, 0, len(entryByID)+1)
	for entryID, e := range entryByID {
		total := 0.0
		for teamID, bid := range entryBids[entryID] {
			pts, ok := teamPoints[teamID]
			if !ok {
				continue
			}
			den := existingTotalBid[teamID] + ourBids[teamID]
			if den <= 0 {
				continue
			}
			total += pts * (bid / den)
		}
		e.TotalPoints = total
		entries = append(entries, e)
	}

	ourID := "our_strategy"
	ourCreated := time.Now()
	ourTotal := 0.0
	for teamID, bid := range ourBids {
		pts, ok := teamPoints[teamID]
		if !ok {
			continue
		}
		den := existingTotalBid[teamID] + bid
		if den <= 0 {
			continue
		}
		ourTotal += pts * (bid / den)
	}
	entries = append(entries, &models.CalcuttaEntry{ID: ourID, Name: "Our Strategy", CalcuttaID: calcuttaID, TotalPoints: ourTotal, Created: ourCreated})

	_, results := appcalcutta.ComputeEntryPlacementsAndPayouts(entries, payouts)
	res, ok := results[ourID]
	if !ok {
		return nil, false, nil
	}

	out := &realizedFinishResult{
		FinishPosition: res.FinishPosition,
		IsTied:         res.IsTied,
		InTheMoney:     res.InTheMoney,
		PayoutCents:    res.PayoutCents,
		TotalPoints:    ourTotal,
	}
	return out, true, nil
}

func (s *Server) updateSuiteExecutionStatus(ctx context.Context, suiteExecutionID string) {
	if suiteExecutionID == "" {
		return
	}
	_, _ = s.pool.Exec(ctx, `
		WITH agg AS (
			SELECT
				SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END)::int AS failed,
				SUM(CASE WHEN status IN ('queued', 'running') THEN 1 ELSE 0 END)::int AS pending
			FROM derived.suite_calcutta_evaluations
			WHERE suite_execution_id = $1::uuid
				AND deleted_at IS NULL
		)
		UPDATE derived.suite_executions e
		SET status = CASE
			WHEN a.failed > 0 THEN 'failed'
			WHEN a.pending > 0 THEN 'running'
			ELSE 'succeeded'
		END,
			error_message = CASE
			WHEN a.failed > 0 THEN COALESCE((
				SELECT error_message
				FROM derived.suite_calcutta_evaluations
				WHERE suite_execution_id = $1::uuid
					AND status = 'failed'
					AND error_message IS NOT NULL
					AND deleted_at IS NULL
				LIMIT 1
			), e.error_message)
			ELSE NULL
		END,
			updated_at = NOW()
		FROM agg a
		WHERE e.id = $1::uuid
			AND e.deleted_at IS NULL
	`, suiteExecutionID)
}

func (s *Server) resolveSuiteNSims(ctx context.Context, suiteID string, fallback int) int {
	var n int
	if err := s.pool.QueryRow(ctx, `
		SELECT n_sims
		FROM derived.suites
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, suiteID).Scan(&n); err != nil {
		return fallback
	}
	if n <= 0 {
		return fallback
	}
	return n
}

func (s *Server) resolveSuiteSeed(ctx context.Context, suiteID string, fallback int) int {
	var seed int
	if err := s.pool.QueryRow(ctx, `
		SELECT seed
		FROM derived.suites
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, suiteID).Scan(&seed); err != nil {
		return fallback
	}
	if seed == 0 {
		return fallback
	}
	return seed
}

func (s *Server) resolveSuiteOptimizerKey(ctx context.Context, suiteID string, fallback string) string {
	var key string
	if err := s.pool.QueryRow(ctx, `
		SELECT optimizer_key
		FROM derived.suites
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, suiteID).Scan(&key); err != nil {
		return fallback
	}
	if key == "" {
		return fallback
	}
	return key
}

func (s *Server) failSuiteCalcuttaEvaluation(ctx context.Context, evaluationID string, err error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	var runKey *string
	var suiteExecutionID *string
	e := s.pool.QueryRow(ctx, `
		UPDATE derived.suite_calcutta_evaluations
		SET status = 'failed',
			error_message = $2,
			updated_at = NOW()
		WHERE id = $1::uuid
		RETURNING run_key::text, suite_execution_id
	`, evaluationID, msg).Scan(&runKey, &suiteExecutionID)
	if e != nil {
		log.Printf("Error marking suite calcutta evaluation %s failed: %v (original error: %v)", evaluationID, e, err)
		return
	}

	_, _ = s.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = $2,
			updated_at = NOW()
		WHERE run_kind = 'simulation'
			AND run_id = $1::uuid
	`, evaluationID, msg)

	failureSummary := map[string]any{
		"status":       "failed",
		"evaluationId": evaluationID,
		"runKey":       runKey,
		"errorMessage": msg,
	}
	failureSummaryJSON, err := json.Marshal(failureSummary)
	if err == nil {
		var runKeyParam any
		if runKey != nil && *runKey != "" {
			runKeyParam = *runKey
		} else {
			runKeyParam = nil
		}
		_, _ = s.pool.Exec(ctx, `
			INSERT INTO derived.run_artifacts (
				run_kind,
				run_id,
				run_key,
				artifact_kind,
				schema_version,
				storage_uri,
				summary_json
			)
			VALUES ('simulation', $1::uuid, $2::uuid, 'metrics', 'v1', NULL, $3::jsonb)
			ON CONFLICT (run_kind, run_id, artifact_kind) WHERE deleted_at IS NULL
			DO UPDATE
			SET run_key = EXCLUDED.run_key,
				schema_version = EXCLUDED.schema_version,
				storage_uri = EXCLUDED.storage_uri,
				summary_json = EXCLUDED.summary_json,
				updated_at = NOW(),
				deleted_at = NULL
		`, evaluationID, runKeyParam, failureSummaryJSON)
	}

	if suiteExecutionID != nil && *suiteExecutionID != "" {
		s.updateSuiteExecutionStatus(ctx, *suiteExecutionID)
	}
}
