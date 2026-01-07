package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	appsimulatetournaments "github.com/andrewcopp/Calcutta/backend/internal/app/simulate_tournaments"
	appsimulatedcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/simulated_calcutta"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultEntryEvalWorkerPollInterval = 2 * time.Second
	defaultEntryEvalWorkerStaleAfter   = 30 * time.Minute
)

type entryEvaluationRequestRow struct {
	ID               string
	RunKey           string
	CalcuttaID       string
	EntryCandidateID string
	ExcludedEntry    *string
	StartingStateKey string
	NSims            int
	Seed             int
	ExperimentKey    string
	RequestSource    string
}

func (s *Server) RunEntryEvaluationWorker(ctx context.Context) {
	s.RunEntryEvaluationWorkerWithOptions(ctx, defaultEntryEvalWorkerPollInterval, defaultEntryEvalWorkerStaleAfter)
}

func (s *Server) RunEntryEvaluationWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if s.pool == nil {
		log.Printf("Entry evaluation worker disabled: database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultEntryEvalWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultEntryEvalWorkerStaleAfter
	}

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	workerID := os.Getenv("HOSTNAME")
	if workerID == "" {
		workerID = "entry-eval-worker"
	}

	claimedCount := int64(0)
	succeededCount := int64(0)
	failedCount := int64(0)
	lastStatsAt := time.Now().UTC()
	lastStatsClaimed := int64(0)
	lastStatsSucceeded := int64(0)
	lastStatsFailed := int64(0)

	var totalJobDuration time.Duration
	var totalSimDuration time.Duration
	var totalEvalDuration time.Duration

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			req, ok, err := s.claimNextEntryEvaluationRequest(ctx, workerID, staleAfter)
			if err != nil {
				log.Printf("Error claiming next entry evaluation request: %v", err)
				continue
			}
			if !ok {
				if time.Since(lastStatsAt) >= 60*time.Second {
					deltaClaimed := claimedCount - lastStatsClaimed
					deltaSucceeded := succeededCount - lastStatsSucceeded
					deltaFailed := failedCount - lastStatsFailed

					avgJobMs := int64(0)
					avgSimMs := int64(0)
					avgEvalMs := int64(0)
					if claimedCount > 0 {
						avgJobMs = int64(totalJobDuration.Milliseconds()) / claimedCount
						avgSimMs = int64(totalSimDuration.Milliseconds()) / claimedCount
						avgEvalMs = int64(totalEvalDuration.Milliseconds()) / claimedCount
					}

					log.Printf("entry_eval_worker stats worker_id=%s claimed_total=%d succeeded_total=%d failed_total=%d claimed_1m=%d succeeded_1m=%d failed_1m=%d avg_job_ms=%d avg_sim_ms=%d avg_eval_ms=%d",
						workerID,
						claimedCount,
						succeededCount,
						failedCount,
						deltaClaimed,
						deltaSucceeded,
						deltaFailed,
						avgJobMs,
						avgSimMs,
						avgEvalMs,
					)

					lastStatsAt = time.Now().UTC()
					lastStatsClaimed = claimedCount
					lastStatsSucceeded = succeededCount
					lastStatsFailed = failedCount
				}
				continue
			}

			claimedCount++
			jobStart := time.Now()
			simDur, evalDur, ok := s.processEntryEvaluationRequest(ctx, workerID, req)
			totalJobDuration += time.Since(jobStart)
			totalSimDuration += simDur
			totalEvalDuration += evalDur
			if ok {
				succeededCount++
			} else {
				failedCount++
			}
		}
	}
}

func (s *Server) claimNextEntryEvaluationRequest(ctx context.Context, workerID string, staleAfter time.Duration) (*entryEvaluationRequestRow, bool, error) {
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
			WHERE run_kind = 'entry_evaluation'
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

	row := &entryEvaluationRequestRow{}
	q2 := `
		UPDATE derived.entry_evaluation_requests r
		SET status = 'running',
			claimed_at = $1,
			claimed_by = $3,
			error_message = NULL
		WHERE r.id = $2::uuid
		RETURNING
			r.id,
			r.run_key,
			r.calcutta_id,
			r.entry_candidate_id,
			r.excluded_entry_name,
			r.starting_state_key,
			r.n_sims,
			r.seed,
			COALESCE(r.experiment_key, ''::text) AS experiment_key,
			COALESCE(r.request_source, ''::text) AS request_source
	`
	var excluded *string
	if err := tx.QueryRow(ctx, q2,
		pgtype.Timestamptz{Time: now, Valid: true},
		runID,
		workerID,
	).Scan(
		&row.ID,
		&row.RunKey,
		&row.CalcuttaID,
		&row.EntryCandidateID,
		&excluded,
		&row.StartingStateKey,
		&row.NSims,
		&row.Seed,
		&row.ExperimentKey,
		&row.RequestSource,
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

func (s *Server) processEntryEvaluationRequest(ctx context.Context, workerID string, req *entryEvaluationRequestRow) (time.Duration, time.Duration, bool) {
	if req == nil {
		return 0, 0, false
	}

	runKey := req.RunKey
	if runKey == "" {
		runKey = uuid.NewString()
	}

	excluded := ""
	if req.ExcludedEntry != nil {
		excluded = *req.ExcludedEntry
	}

	log.Printf("entry_eval_worker start worker_id=%s request_id=%s run_key=%s calcutta_id=%s entry_candidate_id=%s experiment_key=%s request_source=%s starting_state_key=%s n_sims=%d seed=%d excluded_entry_name=%q",
		workerID,
		req.ID,
		runKey,
		req.CalcuttaID,
		req.EntryCandidateID,
		req.ExperimentKey,
		req.RequestSource,
		req.StartingStateKey,
		req.NSims,
		req.Seed,
		excluded,
	)

	s.updateRunJobProgress(ctx, "entry_evaluation", req.ID, 0.05, "start", "Starting entry evaluation")

	year, err := s.resolveSeasonYearByCalcuttaID(ctx, req.CalcuttaID)
	if err != nil {
		s.updateRunJobProgress(ctx, "entry_evaluation", req.ID, 1.0, "failed", err.Error())
		s.failEntryEvaluationRequest(ctx, req.ID, err)
		log.Printf("entry_eval_worker fail worker_id=%s request_id=%s run_key=%s err=%v", workerID, req.ID, runKey, err)
		return 0, 0, false
	}

	s.updateRunJobProgress(ctx, "entry_evaluation", req.ID, 0.15, "simulate", "Simulating tournaments")

	simSvc := appsimulatetournaments.New(s.pool)
	simStart := time.Now()
	simRes, err := simSvc.Run(ctx, appsimulatetournaments.RunParams{
		Season:               year,
		NSims:                req.NSims,
		Seed:                 req.Seed,
		Workers:              0,
		BatchSize:            500,
		ProbabilitySourceKey: "entry_eval_worker",
		StartingStateKey:     req.StartingStateKey,
	})
	simDur := time.Since(simStart)
	if err != nil {
		s.updateRunJobProgress(ctx, "entry_evaluation", req.ID, 1.0, "failed", err.Error())
		s.failEntryEvaluationRequest(ctx, req.ID, err)
		log.Printf("entry_eval_worker fail worker_id=%s request_id=%s run_key=%s phase=simulate err=%v", workerID, req.ID, runKey, err)
		return simDur, 0, false
	}

	s.updateRunJobProgress(ctx, "entry_evaluation", req.ID, 0.65, "evaluate", "Evaluating entry candidate")

	evalSvc := appsimulatedcalcutta.New(s.pool)
	evalStart := time.Now()
	evalRunID, err := evalSvc.CalculateSimulatedCalcuttaForEntryCandidate(
		ctx,
		req.CalcuttaID,
		runKey,
		excluded,
		&simRes.TournamentSimulationBatchID,
		req.EntryCandidateID,
	)
	evalDur := time.Since(evalStart)
	if err != nil {
		s.updateRunJobProgress(ctx, "entry_evaluation", req.ID, 1.0, "failed", err.Error())
		s.failEntryEvaluationRequest(ctx, req.ID, err)
		log.Printf("entry_eval_worker fail worker_id=%s request_id=%s run_key=%s phase=evaluate err=%v", workerID, req.ID, runKey, err)
		return simDur, evalDur, false
	}

	s.updateRunJobProgress(ctx, "entry_evaluation", req.ID, 0.95, "persist", "Persisting results")

	_, err = s.pool.Exec(ctx, `
		UPDATE derived.entry_evaluation_requests
		SET status = 'succeeded',
			evaluation_run_id = $2::uuid,
			error_message = NULL,
			updated_at = NOW()
		WHERE id = $1::uuid
	`, req.ID, evalRunID)
	if err != nil {
		log.Printf("Error updating entry evaluation request %s to succeeded: %v", req.ID, err)
		return simDur, evalDur, false
	}

	_, _ = s.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'succeeded',
			finished_at = NOW(),
			error_message = NULL,
			updated_at = NOW()
		WHERE run_kind = 'entry_evaluation'
			AND run_id = $1::uuid
	`, req.ID)

	s.updateRunJobProgress(ctx, "entry_evaluation", req.ID, 1.0, "succeeded", "Completed")

	summary := map[string]any{
		"status":            "succeeded",
		"requestId":         req.ID,
		"runKey":            runKey,
		"calcuttaId":        req.CalcuttaID,
		"entryCandidateId":  req.EntryCandidateID,
		"startingStateKey":  req.StartingStateKey,
		"nSims":             req.NSims,
		"seed":              req.Seed,
		"experimentKey":     req.ExperimentKey,
		"requestSource":     req.RequestSource,
		"evaluationRunId":   evalRunID,
		"simulationBatchId": simRes.TournamentSimulationBatchID,
		"simMs":             simDur.Milliseconds(),
		"evalMs":            evalDur.Milliseconds(),
		"excludedEntryName": excluded,
	}
	summaryJSON, jerr := json.Marshal(summary)
	if jerr == nil {
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
			VALUES ('entry_evaluation', $1::uuid, $2::uuid, 'metrics', 'v1', NULL, $3::jsonb)
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

	log.Printf("entry_eval_worker success worker_id=%s request_id=%s run_key=%s evaluation_run_id=%s sim_ms=%d eval_ms=%d",
		workerID,
		req.ID,
		runKey,
		evalRunID,
		simDur.Milliseconds(),
		evalDur.Milliseconds(),
	)
	return simDur, evalDur, true
}

func (s *Server) failEntryEvaluationRequest(ctx context.Context, requestID string, err error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	s.updateRunJobProgress(ctx, "entry_evaluation", requestID, 1.0, "failed", msg)
	var runKey *string
	_, e := s.pool.Exec(ctx, `
		UPDATE derived.entry_evaluation_requests
		SET status = 'failed',
			error_message = $2,
			updated_at = NOW()
		WHERE id = $1::uuid
	`, requestID, msg)
	if e != nil {
		log.Printf("Error marking entry evaluation request %s failed: %v (original error: %v)", requestID, e, err)
	}

	_ = s.pool.QueryRow(ctx, `
		SELECT run_key::text
		FROM derived.entry_evaluation_requests
		WHERE id = $1::uuid
		LIMIT 1
	`, requestID).Scan(&runKey)

	_, _ = s.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = $2,
			updated_at = NOW()
		WHERE run_kind = 'entry_evaluation'
			AND run_id = $1::uuid
	`, requestID, msg)

	failureSummary := map[string]any{
		"status":       "failed",
		"requestId":    requestID,
		"runKey":       runKey,
		"errorMessage": msg,
	}
	failureSummaryJSON, jerr := json.Marshal(failureSummary)
	if jerr == nil {
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
			VALUES ('entry_evaluation', $1::uuid, $2::uuid, 'metrics', 'v1', NULL, $3::jsonb)
			ON CONFLICT (run_kind, run_id, artifact_kind) WHERE deleted_at IS NULL
			DO UPDATE
			SET run_key = EXCLUDED.run_key,
				schema_version = EXCLUDED.schema_version,
				storage_uri = EXCLUDED.storage_uri,
				summary_json = EXCLUDED.summary_json,
				updated_at = NOW(),
				deleted_at = NULL
		`, requestID, runKeyParam, failureSummaryJSON)
	}
}

func (s *Server) resolveSeasonYearByCalcuttaID(ctx context.Context, calcuttaID string) (int, error) {
	var year int
	q := `
		SELECT seas.year
		FROM core.calcuttas c
		JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		JOIN core.seasons seas ON seas.id = t.season_id AND seas.deleted_at IS NULL
		WHERE c.id = $1::uuid
			AND c.deleted_at IS NULL
		LIMIT 1
	`
	if err := s.pool.QueryRow(ctx, q, calcuttaID).Scan(&year); err != nil {
		return 0, err
	}
	return year, nil
}
