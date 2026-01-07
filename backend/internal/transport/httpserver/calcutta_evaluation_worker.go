package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	appsimulatedcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/simulated_calcutta"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultCalcuttaEvalWorkerPollInterval = 2 * time.Second
	defaultCalcuttaEvalWorkerStaleAfter   = 30 * time.Minute
)

type calcuttaEvalJob struct {
	RunID     string
	RunKey    string
	Params    json.RawMessage
	ClaimedAt time.Time
}

func (s *Server) RunCalcuttaEvaluationWorker(ctx context.Context) {
	s.RunCalcuttaEvaluationWorkerWithOptions(ctx, defaultCalcuttaEvalWorkerPollInterval, defaultCalcuttaEvalWorkerStaleAfter)
}

func (s *Server) RunCalcuttaEvaluationWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if s.pool == nil {
		log.Printf("calcutta evaluation worker disabled: database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultCalcuttaEvalWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultCalcuttaEvalWorkerStaleAfter
	}

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	workerID := os.Getenv("HOSTNAME")
	if workerID == "" {
		workerID = "calcutta-eval-worker"
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			job, ok, err := s.claimNextCalcuttaEvaluationJob(ctx, workerID, staleAfter)
			if err != nil {
				log.Printf("Error claiming next calcutta evaluation job: %v", err)
				continue
			}
			if !ok {
				continue
			}
			_ = s.processCalcuttaEvaluationJob(ctx, workerID, job)
		}
	}
}

func (s *Server) claimNextCalcuttaEvaluationJob(ctx context.Context, workerID string, staleAfter time.Duration) (*calcuttaEvalJob, bool, error) {
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

	job := &calcuttaEvalJob{}
	q := `
		WITH candidate AS (
			SELECT id
			FROM derived.run_jobs
			WHERE run_kind = 'calcutta_evaluation'
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
		RETURNING j.run_id::text, j.run_key::text, j.params_json::text
	`

	var paramsStr string
	if err := tx.QueryRow(ctx, q,
		pgtype.Timestamptz{Time: now, Valid: true},
		pgtype.Timestamptz{Time: staleBefore, Valid: true},
		workerID,
	).Scan(&job.RunID, &job.RunKey, &paramsStr); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	job.ClaimedAt = now
	job.Params = json.RawMessage([]byte(paramsStr))

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	committed = true

	return job, true, nil
}

func (s *Server) processCalcuttaEvaluationJob(ctx context.Context, workerID string, job *calcuttaEvalJob) bool {
	if job == nil {
		return false
	}

	log.Printf("calcutta_eval_worker start worker_id=%s run_id=%s run_key=%s", workerID, job.RunID, job.RunKey)
	s.updateRunJobProgress(ctx, "calcutta_evaluation", job.RunID, 0.05, "start", "Starting calcutta evaluation job")
	s.updateRunJobProgress(ctx, "calcutta_evaluation", job.RunID, 0.25, "running", "Evaluating calcutta")

	svc := appsimulatedcalcutta.New(s.pool)
	start := time.Now()
	res, err := svc.EvaluateExistingCalcuttaEvaluationRun(ctx, job.RunID)
	dur := time.Since(start)
	if err != nil {
		s.failCalcuttaEvaluationJob(ctx, job, err)
		log.Printf("calcutta_eval_worker fail worker_id=%s run_id=%s run_key=%s dur_ms=%d err=%v", workerID, job.RunID, job.RunKey, dur.Milliseconds(), err)
		return false
	}

	_, _ = s.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'succeeded',
			finished_at = NOW(),
			error_message = NULL,
			updated_at = NOW()
		WHERE run_kind = 'calcutta_evaluation'
			AND run_id = $1::uuid
	`, job.RunID)

	s.updateRunJobProgress(ctx, "calcutta_evaluation", job.RunID, 1.0, "succeeded", "Completed")

	summary := map[string]any{
		"status":                  "succeeded",
		"calcuttaEvaluationRunId": res.CalcuttaEvaluationRunID,
		"runId":                   job.RunID,
		"runKey":                  job.RunKey,
		"nSims":                   res.NSims,
		"nEntries":                res.NEntries,
		"durationMs":              dur.Milliseconds(),
	}
	summaryJSON, jerr := json.Marshal(summary)
	if jerr == nil {
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
			VALUES ('calcutta_evaluation', $1::uuid, $2::uuid, 'metrics', 'v1', NULL, $3::jsonb)
			ON CONFLICT (run_kind, run_id, artifact_kind) WHERE deleted_at IS NULL
			DO UPDATE
			SET run_key = EXCLUDED.run_key,
				schema_version = EXCLUDED.schema_version,
				storage_uri = EXCLUDED.storage_uri,
				summary_json = EXCLUDED.summary_json,
				updated_at = NOW(),
				deleted_at = NULL
		`, job.RunID, job.RunKey, summaryJSON)
	}

	log.Printf("calcutta_eval_worker success worker_id=%s run_id=%s run_key=%s n_sims=%d n_entries=%d dur_ms=%d",
		workerID,
		job.RunID,
		job.RunKey,
		res.NSims,
		res.NEntries,
		dur.Milliseconds(),
	)
	return true
}

func (s *Server) failCalcuttaEvaluationJob(ctx context.Context, job *calcuttaEvalJob, err error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	if job != nil {
		s.updateRunJobProgress(ctx, "calcutta_evaluation", job.RunID, 1.0, "failed", msg)
	}

	_, _ = s.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = $2,
			updated_at = NOW()
		WHERE run_kind = 'calcutta_evaluation'
			AND run_id = $1::uuid
	`, job.RunID, msg)

	failureSummary := map[string]any{
		"status":       "failed",
		"runId":        job.RunID,
		"runKey":       job.RunKey,
		"errorMessage": msg,
	}
	failureSummaryJSON, jerr := json.Marshal(failureSummary)
	if jerr == nil {
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
			VALUES ('calcutta_evaluation', $1::uuid, $2::uuid, 'metrics', 'v1', NULL, $3::jsonb)
			ON CONFLICT (run_kind, run_id, artifact_kind) WHERE deleted_at IS NULL
			DO UPDATE
			SET run_key = EXCLUDED.run_key,
				schema_version = EXCLUDED.schema_version,
				storage_uri = EXCLUDED.storage_uri,
				summary_json = EXCLUDED.summary_json,
				updated_at = NOW(),
				deleted_at = NULL
		`, job.RunID, job.RunKey, failureSummaryJSON)
	}
}
