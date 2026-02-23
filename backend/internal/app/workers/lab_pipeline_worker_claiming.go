package workers

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/jobqueue"
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
	if staleAfter <= 0 {
		staleAfter = defaultLabPipelineWorkerStaleAfter
	}

	kinds := []string{
		jobqueue.KindLabPredictions,
		jobqueue.KindLabOptimization,
		jobqueue.KindLabEvaluation,
	}
	j, err := w.claimer.ClaimNext(ctx, kinds, workerID, w.cfg.RunJobsMaxAttempts, staleAfter)
	if err != nil {
		return nil, false, err
	}
	if j == nil {
		return nil, false, nil
	}

	return &labPipelineJob{
		RunID:     j.RunID,
		RunKey:    j.RunKey,
		RunKind:   j.RunKind,
		Params:    j.Params,
		ClaimedAt: j.ClaimedAt,
	}, true, nil
}
