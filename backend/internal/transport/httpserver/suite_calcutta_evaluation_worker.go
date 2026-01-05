package httpserver

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	reb "github.com/andrewcopp/Calcutta/backend/internal/features/recommended_entry_bids"
	appsimulatetournaments "github.com/andrewcopp/Calcutta/backend/internal/features/simulate_tournaments"
	appsimulatedcalcutta "github.com/andrewcopp/Calcutta/backend/internal/features/simulated_calcutta"
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
	SuiteID          string
	CalcuttaID       string
	GameOutcomeRunID string
	MarketShareRunID string
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

	row := &suiteCalcuttaEvaluationRow{}
	q := `
		WITH candidate AS (
			SELECT id
			FROM derived.suite_calcutta_evaluations
			WHERE deleted_at IS NULL
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
		UPDATE derived.suite_calcutta_evaluations r
		SET status = 'running',
			claimed_at = $1,
			claimed_by = $3,
			error_message = NULL
		FROM candidate
		WHERE r.id = candidate.id
		RETURNING
			r.id,
			r.suite_id,
			r.calcutta_id,
			r.game_outcome_run_id,
			r.market_share_run_id,
			r.optimizer_key,
			r.n_sims,
			r.seed,
			r.starting_state_key,
			r.excluded_entry_name
	`

	var excluded *string
	if err := tx.QueryRow(ctx, q,
		pgtype.Timestamptz{Time: now, Valid: true},
		pgtype.Timestamptz{Time: staleBefore, Valid: true},
		workerID,
	).Scan(
		&row.ID,
		&row.SuiteID,
		&row.CalcuttaID,
		&row.GameOutcomeRunID,
		&row.MarketShareRunID,
		&row.OptimizerKey,
		&row.NSims,
		&row.Seed,
		&row.StartingStateKey,
		&excluded,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
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

	runKey := uuid.NewString()

	excluded := ""
	if req.ExcludedEntry != nil {
		excluded = *req.ExcludedEntry
	}

	log.Printf("suite_eval_worker start worker_id=%s eval_id=%s suite_id=%s calcutta_id=%s run_key=%s game_outcome_run_id=%s market_share_run_id=%s starting_state_key=%s excluded_entry_name=%q",
		workerID,
		req.ID,
		req.SuiteID,
		req.CalcuttaID,
		runKey,
		req.GameOutcomeRunID,
		req.MarketShareRunID,
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

	rebSvc := reb.New(s.pool)
	msRunID := req.MarketShareRunID
	optimizerKey := ""
	if req.OptimizerKey != nil {
		optimizerKey = *req.OptimizerKey
	}
	if optimizerKey == "" {
		optimizerKey = s.resolveSuiteOptimizerKey(ctx, req.SuiteID, "minlp_v1")
	}
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

	evalSvc := appsimulatedcalcutta.New(s.pool)
	evalRunID, err := evalSvc.CalculateSimulatedCalcuttaForEvaluationRun(
		ctx,
		req.CalcuttaID,
		runKey,
		excluded,
		&simRes.TournamentSimulationBatchID,
	)
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
			error_message = NULL,
			updated_at = NOW()
		WHERE id = $1::uuid
	`,
		req.ID,
		optimizerKey,
		nSims,
		seed,
		genRes.StrategyGenerationRunID,
		evalRunID,
		ourRank,
		ourMeanNormalizedPayout,
		ourMedianNormalizedPayout,
		ourPTop1,
		ourPInMoney,
		totalSimulations,
	)
	if err != nil {
		log.Printf("Error updating suite calcutta evaluation %s to succeeded: %v", req.ID, err)
		return false
	}

	log.Printf("suite_eval_worker success worker_id=%s eval_id=%s run_key=%s strategy_generation_run_id=%s calcutta_evaluation_run_id=%s",
		workerID,
		req.ID,
		runKey,
		genRes.StrategyGenerationRunID,
		evalRunID,
	)
	return true
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
	_, e := s.pool.Exec(ctx, `
		UPDATE derived.suite_calcutta_evaluations
		SET status = 'failed',
			error_message = $2,
			updated_at = NOW()
		WHERE id = $1::uuid
	`, evaluationID, msg)
	if e != nil {
		log.Printf("Error marking suite calcutta evaluation %s failed: %v (original error: %v)", evaluationID, e, err)
	}
}
