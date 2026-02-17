package workers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	appcalcuttaevaluations "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta_evaluations"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultCalcuttaEvalWorkerPollInterval = 2 * time.Second
	defaultCalcuttaEvalWorkerStaleAfter   = 30 * time.Minute
)

type CalcuttaEvaluationWorker struct {
	pool     *pgxpool.Pool
	progress ProgressWriter
}

func NewCalcuttaEvaluationWorker(pool *pgxpool.Pool, progress ProgressWriter) *CalcuttaEvaluationWorker {
	if progress == nil {
		progress = NewDBProgressWriter(pool)
	}
	return &CalcuttaEvaluationWorker{pool: pool, progress: progress}
}

type calcuttaEvalJob struct {
	RunID     string
	RunKey    string
	Params    json.RawMessage
	ClaimedAt time.Time
}

func (w *CalcuttaEvaluationWorker) Run(ctx context.Context) {
	w.RunWithOptions(ctx, defaultCalcuttaEvalWorkerPollInterval, defaultCalcuttaEvalWorkerStaleAfter)
}

func (w *CalcuttaEvaluationWorker) RunWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if w == nil || w.pool == nil {
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
			job, ok, err := w.claimNextCalcuttaEvaluationJob(ctx, workerID, staleAfter)
			if err != nil {
				log.Printf("Error claiming next calcutta evaluation job: %v", err)
				continue
			}
			if !ok {
				continue
			}
			_ = w.processCalcuttaEvaluationJob(ctx, workerID, job)
		}
	}
}

func (w *CalcuttaEvaluationWorker) claimNextCalcuttaEvaluationJob(ctx context.Context, workerID string, staleAfter time.Duration) (*calcuttaEvalJob, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	now := time.Now().UTC()
	maxAttempts := resolveRunJobsMaxAttempts(5)
	baseStaleSeconds := staleAfter.Seconds()
	if baseStaleSeconds <= 0 {
		baseStaleSeconds = defaultCalcuttaEvalWorkerStaleAfter.Seconds()
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

	job := &calcuttaEvalJob{}
	if _, err := tx.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = COALESCE(error_message, 'max_attempts_exceeded'),
			updated_at = NOW()
		WHERE run_kind = 'calcutta_evaluation'
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
			WHERE run_kind = 'calcutta_evaluation'
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

func (w *CalcuttaEvaluationWorker) processCalcuttaEvaluationJob(ctx context.Context, workerID string, job *calcuttaEvalJob) bool {
	if job == nil {
		return false
	}

	log.Printf("calcutta_eval_worker start worker_id=%s run_id=%s run_key=%s", workerID, job.RunID, job.RunKey)
	w.updateRunJobProgress(ctx, "calcutta_evaluation", job.RunID, 0.05, "start", "Starting calcutta evaluation job")
	w.updateRunJobProgress(ctx, "calcutta_evaluation", job.RunID, 0.25, "running", "Evaluating calcutta")

	svc := appcalcuttaevaluations.New(w.pool)
	start := time.Now()
	res, err := svc.EvaluateExistingCalcuttaEvaluationRun(ctx, job.RunID)
	dur := time.Since(start)
	if err != nil {
		w.failCalcuttaEvaluationJob(ctx, job, err)
		log.Printf("calcutta_eval_worker fail worker_id=%s run_id=%s run_key=%s dur_ms=%d err=%v", workerID, job.RunID, job.RunKey, dur.Milliseconds(), err)
		return false
	}

	_, _ = w.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'succeeded',
			finished_at = NOW(),
			error_message = NULL,
			updated_at = NOW()
		WHERE run_kind = 'calcutta_evaluation'
			AND run_id = $1::uuid
	`, job.RunID)

	w.updateRunJobProgress(ctx, "calcutta_evaluation", job.RunID, 1.0, "succeeded", "Completed")

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

func (w *CalcuttaEvaluationWorker) failCalcuttaEvaluationJob(ctx context.Context, job *calcuttaEvalJob, err error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	if job != nil {
		w.updateRunJobProgress(ctx, "calcutta_evaluation", job.RunID, 1.0, "failed", msg)
	}

	_, _ = w.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = $2,
			updated_at = NOW()
		WHERE run_kind = 'calcutta_evaluation'
			AND run_id = $1::uuid
	`, job.RunID, msg)
}

func (w *CalcuttaEvaluationWorker) updateRunJobProgress(ctx context.Context, runKind string, runID string, percent float64, phase string, message string) {
	if w == nil || w.progress == nil {
		return
	}
	w.progress.Update(ctx, runKind, runID, percent, phase, message)
}
