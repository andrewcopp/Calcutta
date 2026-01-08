package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	appcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"
	reb "github.com/andrewcopp/Calcutta/backend/internal/app/recommended_entry_bids"
	appsimulatetournaments "github.com/andrewcopp/Calcutta/backend/internal/app/simulate_tournaments"
	appsimulatedcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/simulated_calcutta"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultSimulationWorkerPollInterval = 2 * time.Second
	defaultSimulationWorkerStaleAfter   = 30 * time.Minute
)

type simulationRunRow struct {
	ID                      string
	RunKey                  string
	SimulationBatchID       *string
	CohortID                string
	CalcuttaID              string
	GameOutcomeRunID        string
	MarketShareRunID        string
	StrategyGenerationRunID *string
	OptimizerKey            *string
	NSims                   *int
	Seed                    *int
	StartingStateKey        string
	ExcludedEntry           *string
}

func (s *Server) RunSimulationWorker(ctx context.Context) {
	s.RunSimulationWorkerWithOptions(ctx, defaultSimulationWorkerPollInterval, defaultSimulationWorkerStaleAfter)
}

func (s *Server) RunSimulationWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if s.pool == nil {
		log.Printf("simulation worker disabled: database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultSimulationWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultSimulationWorkerStaleAfter
	}

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	workerID := os.Getenv("HOSTNAME")
	if workerID == "" {
		workerID = "simulation-worker"
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			req, ok, err := s.claimNextSimulationRun(ctx, workerID, staleAfter)
			if err != nil {
				log.Printf("Error claiming next simulation run: %v", err)
				continue
			}
			if !ok {
				continue
			}

			_ = s.processSimulationRun(ctx, workerID, req)
		}
	}
}

func (s *Server) claimNextSimulationRun(ctx context.Context, workerID string, staleAfter time.Duration) (*simulationRunRow, bool, error) {
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

	row := &simulationRunRow{}
	q2 := `
		UPDATE derived.simulation_runs r
		SET status = 'running',
			claimed_at = $1,
			claimed_by = $3,
			error_message = NULL
		WHERE r.id = $2::uuid
		RETURNING
			r.id,
			r.run_key::text,
			r.simulation_run_batch_id::text,
			r.cohort_id::text,
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
		&row.SimulationBatchID,
		&row.CohortID,
		&row.CalcuttaID,
		&row.GameOutcomeRunID,
		&row.MarketShareRunID,
		&row.StrategyGenerationRunID,
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

func (s *Server) processSimulationRun(ctx context.Context, workerID string, req *simulationRunRow) bool {
	if req == nil {
		return false
	}

	runKey := req.RunKey
	if runKey == "" {
		runKey = uuid.NewString()
	}

	strategyGenRunID := ""
	if req.StrategyGenerationRunID != nil {
		strategyGenRunID = *req.StrategyGenerationRunID
	}
	usingExistingStrategyGenRun := strategyGenRunID != ""

	excluded := ""
	if req.ExcludedEntry != nil {
		excluded = *req.ExcludedEntry
	}

	log.Printf("simulation_worker start worker_id=%s run_id=%s cohort_id=%s calcutta_id=%s run_key=%s game_outcome_run_id=%s market_share_run_id=%s strategy_generation_run_id=%s starting_state_key=%s excluded_entry_name=%q",
		workerID,
		req.ID,
		req.CohortID,
		req.CalcuttaID,
		runKey,
		req.GameOutcomeRunID,
		req.MarketShareRunID,
		strategyGenRunID,
		req.StartingStateKey,
		excluded,
	)

	s.updateRunJobProgress(ctx, "simulation", req.ID, 0.05, "start", "Starting simulation run")

	year, err := s.resolveSeasonYearByCalcuttaID(ctx, req.CalcuttaID)
	if err != nil {
		s.updateRunJobProgress(ctx, "simulation", req.ID, 1.0, "failed", err.Error())
		s.failSimulationRun(ctx, req.ID, err)
		return false
	}

	s.updateRunJobProgress(ctx, "simulation", req.ID, 0.15, "simulate", "Simulating tournaments")

	simSvc := appsimulatetournaments.New(s.pool)
	goRunID := req.GameOutcomeRunID
	nSims := 0
	if req.NSims != nil {
		nSims = *req.NSims
	}
	if nSims <= 0 {
		nSims = s.resolveCohortNSims(ctx, req.CohortID, 10000)
	}
	seed := 0
	if req.Seed != nil {
		seed = *req.Seed
	}
	if seed == 0 {
		seed = s.resolveCohortSeed(ctx, req.CohortID, 42)
	}
	simRes, err := simSvc.Run(ctx, appsimulatetournaments.RunParams{
		Season:               year,
		NSims:                nSims,
		Seed:                 seed,
		Workers:              0,
		BatchSize:            500,
		ProbabilitySourceKey: "simulation_worker",
		StartingStateKey:     req.StartingStateKey,
		GameOutcomeRunID:     &goRunID,
	})
	if err != nil {
		s.updateRunJobProgress(ctx, "simulation", req.ID, 1.0, "failed", err.Error())
		s.failSimulationRun(ctx, req.ID, err)
		return false
	}

	strategyMsg := "Generating strategy"
	if usingExistingStrategyGenRun {
		strategyMsg = "Using existing strategy generation run"
	}
	s.updateRunJobProgress(ctx, "simulation", req.ID, 0.55, "strategy", strategyMsg)

	optimizerKey := ""
	if req.OptimizerKey != nil {
		optimizerKey = *req.OptimizerKey
	}
	if optimizerKey == "" {
		optimizerKey = s.resolveCohortOptimizerKey(ctx, req.CohortID, "minlp_v1")
	}

	if !usingExistingStrategyGenRun {
		rebSvc := reb.New(s.pool)
		msRunID := req.MarketShareRunID
		genRes, err := rebSvc.GenerateAndWrite(ctx, reb.GenerateParams{
			CalcuttaID:            req.CalcuttaID,
			RunKey:                runKey,
			Name:                  "simulation_worker",
			OptimizerKey:          optimizerKey,
			MarketShareRunID:      &msRunID,
			SimulatedTournamentID: &simRes.TournamentSimulationBatchID,
		})
		if err != nil {
			s.updateRunJobProgress(ctx, "simulation", req.ID, 1.0, "failed", err.Error())
			s.failSimulationRun(ctx, req.ID, err)
			return false
		}
		strategyGenRunID = genRes.StrategyGenerationRunID
	}

	evalSvc := appsimulatedcalcutta.New(s.pool)
	evalRunID := ""
	s.updateRunJobProgress(ctx, "simulation", req.ID, 0.75, "evaluate", "Evaluating calcutta")
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
		s.updateRunJobProgress(ctx, "simulation", req.ID, 1.0, "failed", err.Error())
		s.failSimulationRun(ctx, req.ID, err)
		return false
	}

	var focusSnapshotEntryID *string
	var focusEntryName *string
	if evalRunID != "" && strategyGenRunID != "" {
		var seID string
		var name string
		err := s.pool.QueryRow(ctx, `
			WITH cer AS (
				SELECT calcutta_snapshot_id
				FROM derived.calcutta_evaluation_runs
				WHERE id = $1::uuid
					AND deleted_at IS NULL
				LIMIT 1
			),
			target AS (
				SELECT team_id, bid_points::int
				FROM derived.recommended_entry_bids
				WHERE strategy_generation_run_id = $2::uuid
					AND deleted_at IS NULL
			),
			candidate AS (
				SELECT cse.id::text AS id, cse.display_name
				FROM core.calcutta_snapshot_entries cse
				WHERE cse.calcutta_snapshot_id = (SELECT calcutta_snapshot_id FROM cer)
					AND cse.is_synthetic = true
					AND cse.deleted_at IS NULL
					AND NOT EXISTS (
						(SELECT team_id, bid_points FROM target)
						EXCEPT
						(SELECT cset.team_id, cset.bid_points::int
						 FROM core.calcutta_snapshot_entry_teams cset
						 WHERE cset.calcutta_snapshot_entry_id = cse.id
							AND cset.deleted_at IS NULL)
					)
					AND NOT EXISTS (
						(SELECT cset.team_id, cset.bid_points::int
						 FROM core.calcutta_snapshot_entry_teams cset
						 WHERE cset.calcutta_snapshot_entry_id = cse.id
							AND cset.deleted_at IS NULL)
						EXCEPT
						(SELECT team_id, bid_points FROM target)
					)
				LIMIT 1
			)
			SELECT id, display_name
			FROM candidate
		`, evalRunID, strategyGenRunID).Scan(&seID, &name)
		if err == nil {
			focusSnapshotEntryID = &seID
			focusEntryName = &name
		}
	}

	if evalRunID != "" {
		_, _ = s.pool.Exec(ctx, `
			UPDATE derived.synthetic_calcuttas sc
			SET calcutta_snapshot_id = cer.calcutta_snapshot_id,
				focus_strategy_generation_run_id = $3::uuid,
				focus_entry_name = COALESCE(NULLIF(sc.focus_entry_name, ''), $4::text),
				updated_at = NOW()
			FROM derived.simulation_runs sr
			JOIN derived.calcutta_evaluation_runs cer
				ON cer.id = $2::uuid
				AND cer.deleted_at IS NULL
			WHERE sr.id = $1::uuid
				AND sr.synthetic_calcutta_id = sc.id
				AND sc.deleted_at IS NULL
		`, req.ID, evalRunID, nullUUIDParam(strategyGenRunID), focusEntryName)
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
			WHERE r.entry_name = $2::text
			ORDER BY r.rank ASC
			LIMIT 1
		`, evalRunID, focusEntryName).Scan(
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
		UPDATE derived.simulation_runs
		SET status = 'succeeded',
			optimizer_key = $2,
			n_sims = $3,
			seed = $4,
			strategy_generation_run_id = $5::uuid,
			calcutta_evaluation_run_id = $6::uuid,
			focus_snapshot_entry_id = $7::uuid,
			our_rank = $8,
			our_mean_normalized_payout = $9,
			our_median_normalized_payout = $10,
			our_p_top1 = $11,
			our_p_in_money = $12,
			total_simulations = $13,
			realized_finish_position = $14,
			realized_is_tied = $15,
			realized_in_the_money = $16,
			realized_payout_cents = $17,
			realized_total_points = $18,
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
		nullUUIDParamPtr(focusSnapshotEntryID),
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
		log.Printf("Error updating simulation run %s to succeeded: %v", req.ID, err)
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

	s.updateRunJobProgress(ctx, "simulation", req.ID, 1.0, "succeeded", "Completed")

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

	log.Printf("simulation_worker success worker_id=%s run_id=%s run_key=%s strategy_generation_run_id=%s calcutta_evaluation_run_id=%s",
		workerID,
		req.ID,
		runKey,
		strategyGenRunID,
		evalRunID,
	)

	if req.SimulationBatchID != nil && *req.SimulationBatchID != "" {
		s.updateSimulationBatchStatus(ctx, *req.SimulationBatchID)
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

func (s *Server) updateSimulationBatchStatus(ctx context.Context, simulationBatchID string) {
	if simulationBatchID == "" {
		return
	}
	_, _ = s.pool.Exec(ctx, `
		WITH agg AS (
			SELECT
				SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END)::int AS failed,
				SUM(CASE WHEN status IN ('queued', 'running') THEN 1 ELSE 0 END)::int AS pending
			FROM derived.simulation_runs
			WHERE simulation_run_batch_id = $1::uuid
				AND deleted_at IS NULL
		)
		UPDATE derived.simulation_run_batches e
		SET status = CASE
			WHEN a.failed > 0 THEN 'failed'
			WHEN a.pending > 0 THEN 'running'
			ELSE 'succeeded'
		END,
			error_message = CASE
			WHEN a.failed > 0 THEN COALESCE((
				SELECT error_message
				FROM derived.simulation_runs
				WHERE simulation_run_batch_id = $1::uuid
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
	`, simulationBatchID)
}

func (s *Server) resolveCohortNSims(ctx context.Context, cohortID string, fallback int) int {
	var n int
	if err := s.pool.QueryRow(ctx, `
		SELECT n_sims
		FROM derived.synthetic_calcutta_cohorts
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, cohortID).Scan(&n); err != nil {
		return fallback
	}
	if n <= 0 {
		return fallback
	}
	return n
}

func (s *Server) resolveCohortSeed(ctx context.Context, cohortID string, fallback int) int {
	var seed int
	if err := s.pool.QueryRow(ctx, `
		SELECT seed
		FROM derived.synthetic_calcutta_cohorts
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, cohortID).Scan(&seed); err != nil {
		return fallback
	}
	if seed == 0 {
		return fallback
	}
	return seed
}

func (s *Server) resolveCohortOptimizerKey(ctx context.Context, cohortID string, fallback string) string {
	var key string
	if err := s.pool.QueryRow(ctx, `
		SELECT optimizer_key
		FROM derived.synthetic_calcutta_cohorts
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, cohortID).Scan(&key); err != nil {
		return fallback
	}
	if key == "" {
		return fallback
	}
	return key
}

func (s *Server) failSimulationRun(ctx context.Context, evaluationID string, err error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	s.updateRunJobProgress(ctx, "simulation", evaluationID, 1.0, "failed", msg)
	var runKey *string
	var simulationBatchID *string
	e := s.pool.QueryRow(ctx, `
		UPDATE derived.simulation_runs
		SET status = 'failed',
			error_message = $2,
			updated_at = NOW()
		WHERE id = $1::uuid
		RETURNING run_key::text, simulation_run_batch_id::text
	`, evaluationID, msg).Scan(&runKey, &simulationBatchID)
	if e != nil {
		log.Printf("Error marking simulation run %s failed: %v (original error: %v)", evaluationID, e, err)
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

	if simulationBatchID != nil && *simulationBatchID != "" {
		s.updateSimulationBatchStatus(ctx, *simulationBatchID)
	}
}
