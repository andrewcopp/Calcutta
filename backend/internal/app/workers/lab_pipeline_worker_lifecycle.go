package workers

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func (w *LabPipelineWorker) checkPipelineCompletion(ctx context.Context, pipelineRunID string) {
	// Check if all calcutta runs are complete
	var pending, running, failed, succeeded int
	err := w.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'running'),
			COUNT(*) FILTER (WHERE status = 'failed'),
			COUNT(*) FILTER (WHERE status = 'succeeded')
		FROM lab.pipeline_calcutta_runs
		WHERE pipeline_run_id = $1::uuid
	`, pipelineRunID).Scan(&pending, &running, &failed, &succeeded)
	if err != nil {
		slog.Warn("lab_pipeline_worker check_completion", "error", err)
		return
	}

	if pending > 0 || running > 0 {
		return // Still processing
	}

	// All done - update pipeline status
	status := "succeeded"
	if failed > 0 && succeeded == 0 {
		status = "failed"
	} else if failed > 0 {
		status = "partial"
	}

	if _, err := w.pool.Exec(ctx, `
		UPDATE lab.pipeline_runs
		SET status = $2, finished_at = NOW(), updated_at = NOW()
		WHERE id = $1::uuid AND status = 'running'
	`, pipelineRunID, status); err != nil {
		slog.Warn("lab_pipeline_worker update_pipeline_status", "error", err)
	}

	slog.Info("lab_pipeline_worker pipeline_complete", "pipeline_run", pipelineRunID, "status", status, "succeeded", succeeded, "failed", failed)
}

func (w *LabPipelineWorker) updateProgress(ctx context.Context, runKind, runID, pcrID string, percent float64, phase, message string) {
	if w.progress != nil {
		w.progress.Update(ctx, runKind, runID, percent, phase, message)
	}

	// Also update pipeline_calcutta_runs progress
	if pcrID != "" {
		if _, err := w.pool.Exec(ctx, `
			UPDATE lab.pipeline_calcutta_runs
			SET progress = $2, progress_message = $3, updated_at = NOW()
			WHERE id = $1::uuid
		`, pcrID, percent, message); err != nil {
			slog.Warn("lab_pipeline_worker update_progress", "error", err)
		}
	}
}

func (w *LabPipelineWorker) succeedLabPipelineJob(ctx context.Context, job *labPipelineJob) {
	if _, err := w.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'succeeded', finished_at = NOW(), error_message = NULL, updated_at = NOW()
		WHERE run_kind = $1 AND run_id = $2::uuid
	`, job.RunKind, job.RunID); err != nil {
		slog.Warn("lab_pipeline_worker succeed_job", "error", err)
	}
}

func (w *LabPipelineWorker) failLabPipelineJob(ctx context.Context, job *labPipelineJob, err error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}

	if _, execErr := w.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed', finished_at = NOW(), error_message = $3, updated_at = NOW()
		WHERE run_kind = $1 AND run_id = $2::uuid
	`, job.RunKind, job.RunID, msg); execErr != nil {
		slog.Warn("lab_pipeline_worker fail_job", "error", execErr)
	}

	// Also update pipeline_calcutta_runs
	var params labPipelineJobParams
	if unmarshalErr := json.Unmarshal(job.Params, &params); unmarshalErr == nil && params.PipelineCalcuttaRunID != "" {
		if _, execErr := w.pool.Exec(ctx, `
			UPDATE lab.pipeline_calcutta_runs
			SET status = 'failed', error_message = $2, finished_at = NOW(), updated_at = NOW()
			WHERE id = $1::uuid
		`, params.PipelineCalcuttaRunID, msg); execErr != nil {
			slog.Warn("lab_pipeline_worker fail_calcutta_run", "error", execErr)
		}
	}

	if w.progress != nil {
		w.progress.Update(ctx, job.RunKind, job.RunID, 1.0, "failed", msg)
	}
}

// requeueLabPipelineJob puts a job back into the queue with a delay.
// Used when a job can't complete yet (e.g. waiting for simulation data).
func (w *LabPipelineWorker) requeueLabPipelineJob(ctx context.Context, job *labPipelineJob, delay time.Duration) {
	retryAfter := time.Now().Add(delay)
	if err := w.claimer.Requeue(ctx, job.RunKind, job.RunID, retryAfter); err != nil {
		slog.Warn("lab_pipeline_worker requeue_job", "run_kind", job.RunKind, "run_id", job.RunID, "error", err)
	}
}

func (w *LabPipelineWorker) resolvePythonScript(relativePath string) string {
	// Prefer DATA_SCIENCE_DIR env var for deterministic resolution in cloud.
	if base := os.Getenv("DATA_SCIENCE_DIR"); base != "" {
		abs := filepath.Join(base, relativePath)
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}

	candidates := []string{
		relativePath,
		"../" + relativePath,
		"../../" + relativePath,
	}

	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	return ""
}
