package workers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	pgo "github.com/andrewcopp/Calcutta/backend/internal/app/predicted_game_outcomes"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultGameOutcomeWorkerPollInterval = 2 * time.Second
	defaultGameOutcomeWorkerStaleAfter   = 30 * time.Minute
)

type GameOutcomeWorker struct {
	pool     *pgxpool.Pool
	progress ProgressWriter
}

func NewGameOutcomeWorker(pool *pgxpool.Pool, progress ProgressWriter) *GameOutcomeWorker {
	if progress == nil {
		progress = NewDBProgressWriter(pool)
	}
	return &GameOutcomeWorker{pool: pool, progress: progress}
}

type gameOutcomeJob struct {
	RunID     string
	RunKey    string
	Params    json.RawMessage
	ClaimedAt time.Time
}

func (w *GameOutcomeWorker) Run(ctx context.Context) {
	w.RunWithOptions(ctx, defaultGameOutcomeWorkerPollInterval, defaultGameOutcomeWorkerStaleAfter)
}

func (w *GameOutcomeWorker) RunWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if w == nil || w.pool == nil {
		log.Printf("Game outcome worker disabled: database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultGameOutcomeWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultGameOutcomeWorkerStaleAfter
	}

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	workerID := os.Getenv("HOSTNAME")
	if workerID == "" {
		workerID = "game-outcome-worker"
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			job, ok, err := w.claimNextGameOutcomeJob(ctx, workerID, staleAfter)
			if err != nil {
				log.Printf("Error claiming next game outcome job: %v", err)
				continue
			}
			if !ok {
				continue
			}
			_ = w.processGameOutcomeJob(ctx, workerID, job)
		}
	}
}

func (w *GameOutcomeWorker) claimNextGameOutcomeJob(ctx context.Context, workerID string, staleAfter time.Duration) (*gameOutcomeJob, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	now := time.Now().UTC()
	maxAttempts := resolveRunJobsMaxAttempts(5)
	baseStaleSeconds := staleAfter.Seconds()
	if baseStaleSeconds <= 0 {
		baseStaleSeconds = defaultGameOutcomeWorkerStaleAfter.Seconds()
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

	job := &gameOutcomeJob{}
	if _, err := tx.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = COALESCE(error_message, 'max_attempts_exceeded'),
			updated_at = NOW()
		WHERE run_kind = 'game_outcome'
			AND status = 'running'
			AND claimed_at IS NOT NULL
			AND claimed_at < ($1::timestamptz - make_interval(secs => ($2 * POWER(2, GREATEST(attempt - 1, 0)))))
			AND attempt >= $3
	`, pgtype.Timestamptz{Time: now, Valid: true}, baseStaleSeconds, maxAttempts); err != nil {
		return nil, false, err
	}

	q := `
		WITH candidate AS (
			SELECT id
			FROM derived.run_jobs
			WHERE run_kind = 'game_outcome'
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
		RETURNING j.run_id::text, j.run_key::text, j.params_json::text
	`

	var paramsStr string
	if err := tx.QueryRow(ctx, q,
		pgtype.Timestamptz{Time: now, Valid: true},
		baseStaleSeconds,
		workerID,
		maxAttempts,
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

func (w *GameOutcomeWorker) processGameOutcomeJob(ctx context.Context, workerID string, job *gameOutcomeJob) bool {
	if job == nil {
		return false
	}

	log.Printf("game_outcome_worker start worker_id=%s run_id=%s run_key=%s", workerID, job.RunID, job.RunKey)
	w.updateRunJobProgress(ctx, "game_outcome", job.RunID, 0.05, "start", "Starting game outcome job")
	w.updateRunJobProgress(ctx, "game_outcome", job.RunID, 0.25, "running", "Generating predicted game outcomes")

	svc := pgo.New(w.pool)
	start := time.Now()
	tournamentID, nRows, err := svc.GenerateAndWriteToExistingRun(ctx, job.RunID)
	dur := time.Since(start)
	if err != nil {
		w.failGameOutcomeJob(ctx, job, err)
		log.Printf("game_outcome_worker fail worker_id=%s run_id=%s run_key=%s dur_ms=%d err=%v", workerID, job.RunID, job.RunKey, dur.Milliseconds(), err)
		return false
	}

	_, _ = w.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'succeeded',
			finished_at = NOW(),
			error_message = NULL,
			updated_at = NOW()
		WHERE run_kind = 'game_outcome'
			AND run_id = $1::uuid
	`, job.RunID)

	w.updateRunJobProgress(ctx, "game_outcome", job.RunID, 1.0, "succeeded", "Completed")

	summary := map[string]any{
		"status":       "succeeded",
		"runId":        job.RunID,
		"runKey":       job.RunKey,
		"tournamentId": tournamentID,
		"rowsInserted": nRows,
		"durationMs":   dur.Milliseconds(),
	}
	if len(job.Params) > 0 {
		var p any
		if err := json.Unmarshal(job.Params, &p); err == nil {
			summary["params"] = p
		}
	}
	summaryJSON, jerr := json.Marshal(summary)
	if jerr == nil {
		var runKeyParam any
		if job.RunKey != "" {
			runKeyParam = job.RunKey
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
			VALUES ('game_outcome', $1::uuid, $2::uuid, 'metrics', 'v1', NULL, $3::jsonb)
			ON CONFLICT (run_kind, run_id, artifact_kind) WHERE deleted_at IS NULL
			DO UPDATE
			SET run_key = EXCLUDED.run_key,
				schema_version = EXCLUDED.schema_version,
				storage_uri = EXCLUDED.storage_uri,
				summary_json = EXCLUDED.summary_json,
				updated_at = NOW(),
				deleted_at = NULL
		`, job.RunID, runKeyParam, summaryJSON)
	}

	log.Printf("game_outcome_worker success worker_id=%s run_id=%s run_key=%s rows=%d dur_ms=%d", workerID, job.RunID, job.RunKey, nRows, dur.Milliseconds())
	return true
}

func (w *GameOutcomeWorker) failGameOutcomeJob(ctx context.Context, job *gameOutcomeJob, err error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	if job != nil {
		w.updateRunJobProgress(ctx, "game_outcome", job.RunID, 1.0, "failed", msg)
	}

	_, _ = w.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = $2,
			updated_at = NOW()
		WHERE run_kind = 'game_outcome'
			AND run_id = $1::uuid
	`, job.RunID, msg)

	failureSummary := map[string]any{
		"status":       "failed",
		"runId":        job.RunID,
		"runKey":       job.RunKey,
		"errorMessage": msg,
	}
	if len(job.Params) > 0 {
		var p any
		if err := json.Unmarshal(job.Params, &p); err == nil {
			failureSummary["params"] = p
		}
	}
	failureSummaryJSON, jerr := json.Marshal(failureSummary)
	if jerr == nil {
		var runKeyParam any
		if job.RunKey != "" {
			runKeyParam = job.RunKey
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
			VALUES ('game_outcome', $1::uuid, $2::uuid, 'metrics', 'v1', NULL, $3::jsonb)
			ON CONFLICT (run_kind, run_id, artifact_kind) WHERE deleted_at IS NULL
			DO UPDATE
			SET run_key = EXCLUDED.run_key,
				schema_version = EXCLUDED.schema_version,
				storage_uri = EXCLUDED.storage_uri,
				summary_json = EXCLUDED.summary_json,
				updated_at = NOW(),
				deleted_at = NULL
		`, job.RunID, runKeyParam, failureSummaryJSON)
	}
}

func (w *GameOutcomeWorker) updateRunJobProgress(ctx context.Context, runKind string, runID string, percent float64, phase string, message string) {
	if w == nil || w.progress == nil {
		return
	}
	w.progress.Update(ctx, runKind, runID, percent, phase, message)
}
