package workers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (w *LabPipelineWorker) checkAndStartPendingPipelines(ctx context.Context) {
	// Find pending pipeline runs and enqueue their first jobs
	rows, err := w.pool.Query(ctx, `
		SELECT pr.id::text, pr.investment_model_id::text, pr.budget_points, pr.optimizer_kind,
		       pr.n_sims, pr.seed, pr.excluded_entry_name
		FROM lab.pipeline_runs pr
		WHERE pr.status = 'pending'
		ORDER BY pr.created_at ASC
		LIMIT 5
	`)
	if err != nil {
		slog.Warn("lab_pipeline_worker check_pending", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var pipelineRunID, modelID string
		var budgetPoints, nSims, seed int
		var optimizerKind string
		var excludedEntryName *string
		if err := rows.Scan(&pipelineRunID, &modelID, &budgetPoints, &optimizerKind, &nSims, &seed, &excludedEntryName); err != nil {
			slog.Warn("lab_pipeline_worker scan", "error", err)
			continue
		}

		// Update pipeline to running
		_, err := w.pool.Exec(ctx, `
			UPDATE lab.pipeline_runs
			SET status = 'running', started_at = NOW(), updated_at = NOW()
			WHERE id = $1::uuid AND status = 'pending'
		`, pipelineRunID)
		if err != nil {
			slog.Warn("lab_pipeline_worker update_running", "error", err)
			continue
		}

		// Get all calcutta runs for this pipeline and enqueue prediction jobs
		calcuttaRows, err := w.pool.Query(ctx, `
			SELECT id::text, calcutta_id::text
			FROM lab.pipeline_calcutta_runs
			WHERE pipeline_run_id = $1::uuid AND status = 'pending'
		`, pipelineRunID)
		if err != nil {
			slog.Warn("lab_pipeline_worker get_calcuttas", "error", err)
			continue
		}

		for calcuttaRows.Next() {
			var pcrID, calcuttaID string
			if err := calcuttaRows.Scan(&pcrID, &calcuttaID); err != nil {
				continue
			}

			// Create job params
			params := labPipelineJobParams{
				PipelineRunID:         pipelineRunID,
				PipelineCalcuttaRunID: pcrID,
				InvestmentModelID:     modelID,
				CalcuttaID:            calcuttaID,
				BudgetPoints:          budgetPoints,
				OptimizerKind:         optimizerKind,
				NSims:                 nSims,
				Seed:                  seed,
			}
			if excludedEntryName != nil {
				params.ExcludedEntryName = *excludedEntryName
			}
			paramsJSON, err := json.Marshal(params)
			if err != nil {
				slog.Warn("lab_pipeline_worker marshal_params", "error", err)
				continue
			}

			// Enqueue prediction job
			var jobID string
			err = w.pool.QueryRow(ctx, `
				INSERT INTO derived.run_jobs (run_kind, run_id, run_key, params_json, status)
				VALUES ('lab_predictions', uuid_generate_v4(), $1::uuid, $2::jsonb, 'queued')
				RETURNING run_id::text
			`, pcrID, paramsJSON).Scan(&jobID)
			if err != nil {
				slog.Warn("lab_pipeline_worker enqueue_job", "error", err)
				continue
			}

			// Update calcutta run with job ID
			if _, err := w.pool.Exec(ctx, `
				UPDATE lab.pipeline_calcutta_runs
				SET predictions_job_id = $2::uuid, status = 'running', started_at = NOW(), updated_at = NOW()
				WHERE id = $1::uuid
			`, pcrID, jobID); err != nil {
				slog.Warn("lab_pipeline_worker update_calcutta_run_predictions", "error", err)
			}

			slog.Info("lab_pipeline_worker enqueued_predictions", "pipeline_run", pipelineRunID, "calcutta_run", pcrID, "job_id", jobID)
		}
		calcuttaRows.Close()
	}
}

func (w *LabPipelineWorker) claimNextLabPipelineJob(ctx context.Context, workerID string, staleAfter time.Duration) (*labPipelineJob, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	now := time.Now().UTC()
	maxAttempts := w.cfg.RunJobsMaxAttempts
	baseStaleSeconds := staleAfter.Seconds()
	if baseStaleSeconds <= 0 {
		baseStaleSeconds = defaultLabPipelineWorkerStaleAfter.Seconds()
	}

	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Fail stale jobs that exceeded max attempts
	if _, err := tx.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = COALESCE(error_message, 'max_attempts_exceeded'),
			updated_at = NOW()
		WHERE run_kind IN ('lab_predictions', 'lab_optimization', 'lab_evaluation')
			AND status = 'running'
			AND claimed_at IS NOT NULL
			AND claimed_at < ($1::timestamptz - make_interval(secs => ($2 * POWER(2, GREATEST(attempt - 1, 0)))))
			AND attempt >= $3
	`, pgtype.Timestamptz{Time: now, Valid: true}, baseStaleSeconds, maxAttempts); err != nil {
		return nil, false, err
	}

	// Claim next job
	q := `
		WITH candidate AS (
			SELECT id
			FROM derived.run_jobs
			WHERE run_kind IN ('lab_predictions', 'lab_optimization', 'lab_evaluation')
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
		RETURNING j.run_id::text, j.run_key::text, j.run_kind, j.params_json::text
	`

	job := &labPipelineJob{}
	var paramsStr string
	if err := tx.QueryRow(ctx, q,
		pgtype.Timestamptz{Time: now, Valid: true},
		baseStaleSeconds,
		workerID,
		maxAttempts,
	).Scan(&job.RunID, &job.RunKey, &job.RunKind, &paramsStr); err != nil {
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
