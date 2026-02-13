package workers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	appcalcuttaevaluations "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta_evaluations"
	appsimulatetournaments "github.com/andrewcopp/Calcutta/backend/internal/app/simulate_tournaments"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_artifacts"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultSimulationWorkerPollInterval = 2 * time.Second
	defaultSimulationWorkerStaleAfter   = 30 * time.Minute
)

type SimulationWorker struct {
	pool        *pgxpool.Pool
	progress    ProgressWriter
	artifactSvc *simulation_artifacts.Service
}

func NewSimulationWorker(pool *pgxpool.Pool, progress ProgressWriter, artifactsDir string) *SimulationWorker {
	if progress == nil {
		progress = NewDBProgressWriter(pool)
	}
	return &SimulationWorker{
		pool:        pool,
		progress:    progress,
		artifactSvc: simulation_artifacts.New(pool, artifactsDir),
	}
}

type simulationRunRow struct {
	ID                      string
	RunKey                  string
	SimulationBatchID       *string
	CohortID                string
	CalcuttaID              *string
	SimulatedCalcuttaID     *string
	GameOutcomeRunID        *string
	GameOutcomeSpec         *simulation_game_outcomes.Spec
	MarketShareRunID        *string
	StrategyGenerationRunID *string
	OptimizerKey            *string
	NSims                   *int
	Seed                    *int
	StartingStateKey        string
	ExcludedEntry           *string
}

func (w *SimulationWorker) Run(ctx context.Context) {
	w.RunWithOptions(ctx, defaultSimulationWorkerPollInterval, defaultSimulationWorkerStaleAfter)
}

func (w *SimulationWorker) RunWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if w == nil || w.pool == nil {
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
			req, ok, err := w.claimNextSimulationRun(ctx, workerID, staleAfter)
			if err != nil {
				log.Printf("Error claiming next simulation run: %v", err)
				continue
			}
			if !ok {
				continue
			}

			_ = w.processSimulationRun(ctx, workerID, req)
		}
	}
}

func (w *SimulationWorker) claimNextSimulationRun(ctx context.Context, workerID string, staleAfter time.Duration) (*simulationRunRow, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	now := time.Now().UTC()
	maxAttempts := resolveRunJobsMaxAttempts(5)
	baseStaleSeconds := staleAfter.Seconds()
	if baseStaleSeconds <= 0 {
		baseStaleSeconds = defaultSimulationWorkerStaleAfter.Seconds()
	}

	tx, err := w.pool.Begin(ctx)
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

	if _, err := tx.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = COALESCE(error_message, 'max_attempts_exceeded'),
			updated_at = NOW()
		WHERE run_kind = 'simulation'
			AND status = 'running'
			AND claimed_at IS NOT NULL
			AND claimed_at < ($1::timestamptz - make_interval(secs => ($2 * POWER(2, GREATEST(attempt - 1, 0)))))
			AND attempt >= $3
	`, pgtype.Timestamptz{Time: now, Valid: true}, baseStaleSeconds, maxAttempts); err != nil {
		return nil, false, err
	}

	var runID string
	q := `
		WITH candidate AS (
			SELECT id
			FROM derived.run_jobs
			WHERE run_kind = 'simulation'
				AND attempt < $4
				AND (
					status = 'queued'
					OR (
						status = 'running'
						AND claimed_at IS NOT NULL
						AND claimed_at < ($1::timestamptz - make_interval(secs => ($2 * POWER(2, GREATEST(attempt - 1, 0)))))
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
		baseStaleSeconds,
		workerID,
		maxAttempts,
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
			r.calcutta_id::text,
			r.simulated_calcutta_id::text,
			r.game_outcome_run_id::text,
			r.game_outcome_spec_json,
			r.market_share_run_id::text,
			r.strategy_generation_run_id,
			r.optimizer_key,
			r.n_sims,
			r.seed,
			r.starting_state_key,
			r.excluded_entry_name
	`

	var excluded *string
	var specRaw []byte
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
		&row.SimulatedCalcuttaID,
		&row.GameOutcomeRunID,
		&specRaw,
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
	if len(specRaw) > 0 {
		var spec simulation_game_outcomes.Spec
		if err := json.Unmarshal(specRaw, &spec); err != nil {
			return nil, false, err
		}
		spec.Normalize()
		row.GameOutcomeSpec = &spec
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	committed = true

	return row, true, nil
}

func (w *SimulationWorker) processSimulationRun(ctx context.Context, workerID string, req *simulationRunRow) bool {
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

	excluded := ""
	if req.ExcludedEntry != nil {
		excluded = *req.ExcludedEntry
	}

	calcuttaID := ""
	if req.CalcuttaID != nil {
		calcuttaID = *req.CalcuttaID
	}
	simulatedCalcuttaID := ""
	if req.SimulatedCalcuttaID != nil {
		simulatedCalcuttaID = *req.SimulatedCalcuttaID
	}
	marketShareRunID := ""
	if req.MarketShareRunID != nil {
		marketShareRunID = *req.MarketShareRunID
	}

	goRunID := ""
	if req.GameOutcomeRunID != nil {
		goRunID = *req.GameOutcomeRunID
	}
	log.Printf("simulation_worker start worker_id=%s run_id=%s cohort_id=%s calcutta_id=%s run_key=%s game_outcome_run_id=%s market_share_run_id=%s strategy_generation_run_id=%s starting_state_key=%s excluded_entry_name=%q",
		workerID,
		req.ID,
		req.CohortID,
		calcuttaID,
		runKey,
		goRunID,
		marketShareRunID,
		strategyGenRunID,
		req.StartingStateKey,
		excluded,
	)

	w.updateRunJobProgress(ctx, "simulation", req.ID, 0.05, "start", "Starting simulation run")

	year := 0
	var err error
	if strings.TrimSpace(simulatedCalcuttaID) != "" {
		var tournamentID string
		if err := w.pool.QueryRow(ctx, `
			SELECT tournament_id::text
			FROM derived.simulated_calcuttas
			WHERE id = $1::uuid
				AND deleted_at IS NULL
			LIMIT 1
		`, simulatedCalcuttaID).Scan(&tournamentID); err != nil {
			w.updateRunJobProgress(ctx, "simulation", req.ID, 1.0, "failed", err.Error())
			w.failSimulationRun(ctx, req.ID, err)
			return false
		}
		year, err = resolveSeasonYearByTournamentID(ctx, w.pool, tournamentID)
	} else {
		year, err = resolveSeasonYearByCalcuttaID(ctx, w.pool, calcuttaID)
	}
	if err != nil {
		w.updateRunJobProgress(ctx, "simulation", req.ID, 1.0, "failed", err.Error())
		w.failSimulationRun(ctx, req.ID, err)
		return false
	}

	w.updateRunJobProgress(ctx, "simulation", req.ID, 0.15, "simulate", "Simulating tournaments")

	simSvc := appsimulatetournaments.New(w.pool)
	spec := req.GameOutcomeSpec
	if spec == nil {
		tmp := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
		tmp.Normalize()
		spec = tmp
	}
	nSims := 0
	if req.NSims != nil {
		nSims = *req.NSims
	}
	if nSims <= 0 {
		nSims = resolveCohortNSims(ctx, w.pool, req.CohortID, 10000)
	}
	seed := 0
	if req.Seed != nil {
		seed = *req.Seed
	}
	if seed == 0 {
		seed = resolveCohortSeed(ctx, w.pool, req.CohortID, 42)
	}
	simRes, err := simSvc.Run(ctx, appsimulatetournaments.RunParams{
		Season:               year,
		NSims:                nSims,
		Seed:                 seed,
		Workers:              0,
		BatchSize:            500,
		ProbabilitySourceKey: "simulation_worker",
		StartingStateKey:     req.StartingStateKey,
		GameOutcomeRunID:     req.GameOutcomeRunID,
		GameOutcomeSpec:      spec,
	})
	if err != nil {
		w.updateRunJobProgress(ctx, "simulation", req.ID, 1.0, "failed", err.Error())
		w.failSimulationRun(ctx, req.ID, err)
		return false
	}

	evalSvc := appcalcuttaevaluations.New(w.pool)
	evalRunID := ""
	w.updateRunJobProgress(ctx, "simulation", req.ID, 0.75, "evaluate", "Evaluating calcutta")
	if strings.TrimSpace(simulatedCalcuttaID) != "" {
		evalRunID, err = evalSvc.CalculateSimulatedCalcuttaForSimulatedCalcutta(
			ctx,
			simulatedCalcuttaID,
			runKey,
			excluded,
			&simRes.TournamentSimulationBatchID,
		)
	} else {
		evalRunID, err = evalSvc.CalculateSimulatedCalcuttaForEvaluationRun(
			ctx,
			calcuttaID,
			runKey,
			excluded,
			&simRes.TournamentSimulationBatchID,
		)
	}
	if err != nil {
		w.updateRunJobProgress(ctx, "simulation", req.ID, 1.0, "failed", err.Error())
		w.failSimulationRun(ctx, req.ID, err)
		return false
	}

	optimizerKey := ""

	_, err = w.pool.Exec(ctx, `
		UPDATE derived.simulation_runs
		SET status = 'succeeded',
			optimizer_key = $2,
			n_sims = $3,
			seed = $4,
			strategy_generation_run_id = $5::uuid,
			calcutta_evaluation_run_id = $6::uuid,
			error_message = NULL,
			updated_at = NOW()
		WHERE id = $1::uuid
	`,
		req.ID,
		optimizerKey,
		nSims,
		seed,
		nullUUIDParam(strategyGenRunID),
		evalRunID,
	)
	if err != nil {
		log.Printf("Error updating simulation run %s to succeeded: %v", req.ID, err)
		return false
	}

	_, _ = w.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'succeeded',
			finished_at = NOW(),
			error_message = NULL,
			updated_at = NOW()
		WHERE run_kind = 'simulation'
			AND run_id = $1::uuid
	`, req.ID)

	w.updateRunJobProgress(ctx, "simulation", req.ID, 0.95, "artifacts", "Writing artifacts")
	if err := w.artifactSvc.ExportArtifacts(ctx, req.ID, runKey, evalRunID); err != nil {
		log.Printf("simulation_worker artifact_export_failed run_id=%s run_key=%s err=%v", req.ID, runKey, err)
	}

	w.updateRunJobProgress(ctx, "simulation", req.ID, 1.0, "succeeded", "Completed")

	summary := map[string]any{
		"status":                  "succeeded",
		"evaluationId":            req.ID,
		"runKey":                  runKey,
		"optimizerKey":            optimizerKey,
		"nSims":                   nSims,
		"seed":                    seed,
		"strategyGenerationRunId": strategyGenRunID,
		"calcuttaEvaluationRunId": evalRunID,
	}
	summaryJSON, err := json.Marshal(summary)
	if err == nil {
		var runKeyParam any
		if runKey != "" {
			runKeyParam = runKey
		} else {
			runKeyParam = nil
		}
		_, _ = w.pool.Exec(ctx, `
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
		w.artifactSvc.UpdateBatchStatus(ctx, *req.SimulationBatchID)
	}
	return true
}

func (w *SimulationWorker) failSimulationRun(ctx context.Context, evaluationID string, err error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	w.updateRunJobProgress(ctx, "simulation", evaluationID, 1.0, "failed", msg)
	var runKey *string
	var simulationBatchID *string
	e := w.pool.QueryRow(ctx, `
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

	_, _ = w.pool.Exec(ctx, `
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
		_, _ = w.pool.Exec(ctx, `
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
		w.artifactSvc.UpdateBatchStatus(ctx, *simulationBatchID)
	}
}

func (w *SimulationWorker) updateRunJobProgress(ctx context.Context, runKind string, runID string, percent float64, phase string, message string) {
	if w == nil || w.progress == nil {
		return
	}
	w.progress.Update(ctx, runKind, runID, percent, phase, message)
}
