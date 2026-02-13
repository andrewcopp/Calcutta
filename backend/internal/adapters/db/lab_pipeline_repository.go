package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/app/lab"
	"github.com/jackc/pgx/v5"
)

// CreatePipelineRun creates a new pipeline run.
func (r *LabRepository) CreatePipelineRun(run *lab.PipelineRun) (*lab.PipelineRun, error) {
	ctx := context.Background()

	query := `
		INSERT INTO lab.pipeline_runs (
			investment_model_id,
			target_calcutta_ids,
			budget_points,
			optimizer_kind,
			n_sims,
			seed,
			excluded_entry_name,
			status
		) VALUES (
			$1::uuid,
			$2::uuid[],
			$3,
			$4,
			$5,
			$6,
			$7,
			$8
		)
		RETURNING
			id::text,
			investment_model_id::text,
			target_calcutta_ids::text[],
			budget_points,
			optimizer_kind,
			n_sims,
			seed,
			excluded_entry_name,
			status,
			started_at,
			finished_at,
			error_message,
			created_at,
			updated_at
	`

	var result lab.PipelineRun
	var targetIDs []string
	err := r.pool.QueryRow(ctx, query,
		run.InvestmentModelID,
		run.TargetCalcuttaIDs,
		run.BudgetPoints,
		run.OptimizerKind,
		run.NSims,
		run.Seed,
		run.ExcludedEntryName,
		run.Status,
	).Scan(
		&result.ID,
		&result.InvestmentModelID,
		&targetIDs,
		&result.BudgetPoints,
		&result.OptimizerKind,
		&result.NSims,
		&result.Seed,
		&result.ExcludedEntryName,
		&result.Status,
		&result.StartedAt,
		&result.FinishedAt,
		&result.ErrorMessage,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	result.TargetCalcuttaIDs = targetIDs
	return &result, nil
}

// GetPipelineRun returns a pipeline run by ID.
func (r *LabRepository) GetPipelineRun(id string) (*lab.PipelineRun, error) {
	ctx := context.Background()

	query := `
		SELECT
			id::text,
			investment_model_id::text,
			target_calcutta_ids::text[],
			budget_points,
			optimizer_kind,
			n_sims,
			seed,
			excluded_entry_name,
			status,
			started_at,
			finished_at,
			error_message,
			created_at,
			updated_at
		FROM lab.pipeline_runs
		WHERE id = $1::uuid
	`

	var result lab.PipelineRun
	var targetIDs []string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.InvestmentModelID,
		&targetIDs,
		&result.BudgetPoints,
		&result.OptimizerKind,
		&result.NSims,
		&result.Seed,
		&result.ExcludedEntryName,
		&result.Status,
		&result.StartedAt,
		&result.FinishedAt,
		&result.ErrorMessage,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, &apperrors.NotFoundError{Resource: "pipeline_run", ID: id}
	}
	if err != nil {
		return nil, err
	}
	result.TargetCalcuttaIDs = targetIDs
	return &result, nil
}

// UpdatePipelineRunStatus updates the status of a pipeline run.
func (r *LabRepository) UpdatePipelineRunStatus(id string, status string, errorMessage *string) error {
	ctx := context.Background()

	var finishedAt *time.Time
	var startedAt *time.Time
	now := time.Now()

	if status == "running" {
		startedAt = &now
	}
	if status == "succeeded" || status == "failed" || status == "cancelled" {
		finishedAt = &now
	}

	query := `
		UPDATE lab.pipeline_runs
		SET status = $2,
			started_at = COALESCE($3, started_at),
			finished_at = COALESCE($4, finished_at),
			error_message = $5,
			updated_at = NOW()
		WHERE id = $1::uuid
	`

	tag, err := r.pool.Exec(ctx, query, id, status, startedAt, finishedAt, errorMessage)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return &apperrors.NotFoundError{Resource: "pipeline_run", ID: id}
	}
	return nil
}

// ListPipelineRuns returns pipeline runs, optionally filtered by model ID and status.
func (r *LabRepository) ListPipelineRuns(modelID *string, status *string, limit int) ([]lab.PipelineRun, error) {
	ctx := context.Background()

	query := `
		SELECT
			id::text,
			investment_model_id::text,
			target_calcutta_ids::text[],
			budget_points,
			optimizer_kind,
			n_sims,
			seed,
			excluded_entry_name,
			status,
			started_at,
			finished_at,
			error_message,
			created_at,
			updated_at
		FROM lab.pipeline_runs
		WHERE 1=1
	`
	args := []any{}
	argIdx := 1

	if modelID != nil && *modelID != "" {
		query += fmt.Sprintf(" AND investment_model_id = $%d::uuid", argIdx)
		args = append(args, *modelID)
		argIdx++
	}
	if status != nil && *status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *status)
		argIdx++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []lab.PipelineRun
	for rows.Next() {
		var run lab.PipelineRun
		var targetIDs []string
		if err := rows.Scan(
			&run.ID,
			&run.InvestmentModelID,
			&targetIDs,
			&run.BudgetPoints,
			&run.OptimizerKind,
			&run.NSims,
			&run.Seed,
			&run.ExcludedEntryName,
			&run.Status,
			&run.StartedAt,
			&run.FinishedAt,
			&run.ErrorMessage,
			&run.CreatedAt,
			&run.UpdatedAt,
		); err != nil {
			return nil, err
		}
		run.TargetCalcuttaIDs = targetIDs
		results = append(results, run)
	}
	return results, rows.Err()
}

// GetActivePipelineRun returns the most recent running or pending pipeline run for a model.
func (r *LabRepository) GetActivePipelineRun(modelID string) (*lab.PipelineRun, error) {
	ctx := context.Background()

	query := `
		SELECT
			id::text,
			investment_model_id::text,
			target_calcutta_ids::text[],
			budget_points,
			optimizer_kind,
			n_sims,
			seed,
			excluded_entry_name,
			status,
			started_at,
			finished_at,
			error_message,
			created_at,
			updated_at
		FROM lab.pipeline_runs
		WHERE investment_model_id = $1::uuid
			AND status IN ('pending', 'running')
		ORDER BY created_at DESC
		LIMIT 1
	`

	var result lab.PipelineRun
	var targetIDs []string
	err := r.pool.QueryRow(ctx, query, modelID).Scan(
		&result.ID,
		&result.InvestmentModelID,
		&targetIDs,
		&result.BudgetPoints,
		&result.OptimizerKind,
		&result.NSims,
		&result.Seed,
		&result.ExcludedEntryName,
		&result.Status,
		&result.StartedAt,
		&result.FinishedAt,
		&result.ErrorMessage,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	result.TargetCalcuttaIDs = targetIDs
	return &result, nil
}

// CreatePipelineCalcuttaRuns creates calcutta run records for a pipeline.
func (r *LabRepository) CreatePipelineCalcuttaRuns(pipelineRunID string, calcuttaIDs []string) error {
	ctx := context.Background()

	if len(calcuttaIDs) == 0 {
		return nil
	}

	// Build multi-row insert
	values := make([]string, 0, len(calcuttaIDs))
	args := []any{pipelineRunID}
	for i, cid := range calcuttaIDs {
		values = append(values, fmt.Sprintf("($1::uuid, $%d::uuid)", i+2))
		args = append(args, cid)
	}

	query := `
		INSERT INTO lab.pipeline_calcutta_runs (pipeline_run_id, calcutta_id)
		VALUES ` + strings.Join(values, ", ")

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// GetPipelineCalcuttaRuns returns all calcutta runs for a pipeline.
func (r *LabRepository) GetPipelineCalcuttaRuns(pipelineRunID string) ([]lab.PipelineCalcuttaRun, error) {
	ctx := context.Background()

	query := `
		SELECT
			id::text,
			pipeline_run_id::text,
			calcutta_id::text,
			entry_id::text,
			stage,
			status,
			progress,
			progress_message,
			predictions_job_id::text,
			optimization_job_id::text,
			evaluation_job_id::text,
			evaluation_id::text,
			error_message,
			started_at,
			finished_at,
			created_at,
			updated_at
		FROM lab.pipeline_calcutta_runs
		WHERE pipeline_run_id = $1::uuid
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, pipelineRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []lab.PipelineCalcuttaRun
	for rows.Next() {
		var run lab.PipelineCalcuttaRun
		if err := rows.Scan(
			&run.ID,
			&run.PipelineRunID,
			&run.CalcuttaID,
			&run.EntryID,
			&run.Stage,
			&run.Status,
			&run.Progress,
			&run.ProgressMessage,
			&run.PredictionsJobID,
			&run.OptimizationJobID,
			&run.EvaluationJobID,
			&run.EvaluationID,
			&run.ErrorMessage,
			&run.StartedAt,
			&run.FinishedAt,
			&run.CreatedAt,
			&run.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, run)
	}
	return results, rows.Err()
}

// UpdatePipelineCalcuttaRun updates fields on a pipeline calcutta run.
func (r *LabRepository) UpdatePipelineCalcuttaRun(id string, updates map[string]interface{}) error {
	ctx := context.Background()

	if len(updates) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(updates))
	args := []any{id}
	argIdx := 2

	for col, val := range updates {
		switch col {
		case "entry_id", "predictions_job_id", "optimization_job_id", "evaluation_job_id", "evaluation_id":
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d::uuid", col, argIdx))
		case "progress":
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d::double precision", col, argIdx))
		case "started_at", "finished_at":
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d::timestamptz", col, argIdx))
		default:
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argIdx))
		}
		args = append(args, val)
		argIdx++
	}

	query := fmt.Sprintf(`
		UPDATE lab.pipeline_calcutta_runs
		SET %s, updated_at = NOW()
		WHERE id = $1::uuid
	`, strings.Join(setClauses, ", "))

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return &apperrors.NotFoundError{Resource: "pipeline_calcutta_run", ID: id}
	}
	return nil
}

// GetPipelineProgress returns detailed progress for a pipeline run.
func (r *LabRepository) GetPipelineProgress(pipelineRunID string) (*lab.PipelineProgressResponse, error) {
	ctx := context.Background()

	// Get pipeline run
	pipelineRun, err := r.GetPipelineRun(pipelineRunID)
	if err != nil {
		return nil, err
	}

	// Get model name
	var modelName string
	err = r.pool.QueryRow(ctx, `
		SELECT name FROM lab.investment_models WHERE id = $1::uuid
	`, pipelineRun.InvestmentModelID).Scan(&modelName)
	if err != nil {
		return nil, err
	}

	// Get calcutta runs with calcutta details
	query := `
		SELECT
			pcr.id::text,
			pcr.calcutta_id::text,
			c.name AS calcutta_name,
			s.year AS calcutta_year,
			pcr.stage,
			pcr.status,
			pcr.progress,
			pcr.progress_message,
			pcr.entry_id::text,
			pcr.evaluation_id::text,
			pcr.error_message,
			CASE WHEN e.predictions_json IS NOT NULL AND e.predictions_json::text != 'null' AND e.predictions_json::text != '' THEN true ELSE false END AS has_predictions,
			CASE WHEN e.id IS NOT NULL THEN true ELSE false END AS has_entry,
			CASE WHEN ev.id IS NOT NULL THEN true ELSE false END AS has_evaluation,
			ev.mean_normalized_payout
		FROM lab.pipeline_calcutta_runs pcr
		JOIN core.calcuttas c ON c.id = pcr.calcutta_id
		JOIN core.tournaments t ON t.id = c.tournament_id
		JOIN core.seasons s ON s.id = t.season_id
		LEFT JOIN lab.entries e ON e.id = pcr.entry_id AND e.deleted_at IS NULL
		LEFT JOIN lab.evaluations ev ON ev.id = pcr.evaluation_id AND ev.deleted_at IS NULL
		WHERE pcr.pipeline_run_id = $1::uuid
		ORDER BY s.year DESC
	`

	rows, err := r.pool.Query(ctx, query, pipelineRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calcuttas []lab.CalcuttaProgressResponse
	var summary lab.PipelineProgressSummary
	var payoutSum float64
	var payoutCount int

	for rows.Next() {
		var c lab.CalcuttaProgressResponse
		var meanPayout *float64
		if err := rows.Scan(
			&c.CalcuttaID, &c.CalcuttaID, &c.CalcuttaName, &c.CalcuttaYear,
			&c.Stage, &c.Status, &c.Progress, &c.ProgressMessage,
			&c.EntryID, &c.EvaluationID, &c.ErrorMessage,
			&c.HasPredictions, &c.HasEntry, &c.HasEvaluation,
			&meanPayout,
		); err != nil {
			return nil, err
		}
		c.MeanPayout = meanPayout

		summary.TotalCalcuttas++
		if c.HasPredictions {
			summary.PredictionsCount++
		}
		if c.HasEntry {
			summary.EntriesCount++
		}
		if c.HasEvaluation {
			summary.EvaluationsCount++
		}
		if c.Status == "failed" {
			summary.FailedCount++
		}
		if meanPayout != nil {
			payoutSum += *meanPayout
			payoutCount++
		}

		calcuttas = append(calcuttas, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if payoutCount > 0 {
		avgPayout := payoutSum / float64(payoutCount)
		summary.AvgMeanPayout = &avgPayout
	}

	return &lab.PipelineProgressResponse{
		ID:                pipelineRunID,
		InvestmentModelID: pipelineRun.InvestmentModelID,
		ModelName:         modelName,
		Status:            pipelineRun.Status,
		StartedAt:         pipelineRun.StartedAt,
		FinishedAt:        pipelineRun.FinishedAt,
		ErrorMessage:      pipelineRun.ErrorMessage,
		Summary:           summary,
		Calcuttas:         calcuttas,
	}, nil
}

// GetModelPipelineProgress returns the pipeline progress for a model, including existing artifacts.
func (r *LabRepository) GetModelPipelineProgress(modelID string) (*lab.ModelPipelineProgress, error) {
	ctx := context.Background()

	// Get model name
	var modelName string
	err := r.pool.QueryRow(ctx, `
		SELECT name FROM lab.investment_models WHERE id = $1::uuid AND deleted_at IS NULL
	`, modelID).Scan(&modelName)
	if err == pgx.ErrNoRows {
		return nil, &apperrors.NotFoundError{Resource: "investment_model", ID: modelID}
	}
	if err != nil {
		return nil, err
	}

	// Check for active pipeline run
	activePipeline, err := r.GetActivePipelineRun(modelID)
	if err != nil {
		return nil, err
	}

	// Get historical calcuttas with their entry/evaluation status for this model
	query := `
		SELECT
			c.id::text AS calcutta_id,
			c.name AS calcutta_name,
			s.year AS calcutta_year,
			e.id::text AS entry_id,
			ev.id::text AS evaluation_id,
			CASE WHEN e.predictions_json IS NOT NULL AND e.predictions_json::text != 'null' AND e.predictions_json::text != '' THEN true ELSE false END AS has_predictions,
			CASE WHEN e.id IS NOT NULL THEN true ELSE false END AS has_entry,
			CASE WHEN ev.id IS NOT NULL THEN true ELSE false END AS has_evaluation,
			ev.mean_normalized_payout,
			pcr.stage,
			pcr.status,
			pcr.progress,
			pcr.progress_message,
			pcr.error_message
		FROM core.calcuttas c
		JOIN core.tournaments t ON t.id = c.tournament_id
		JOIN core.seasons s ON s.id = t.season_id
		LEFT JOIN lab.entries e ON e.calcutta_id = c.id
			AND e.investment_model_id = $1::uuid
			AND e.starting_state_key = 'post_first_four'
			AND e.deleted_at IS NULL
		LEFT JOIN lab.evaluations ev ON ev.entry_id = e.id AND ev.deleted_at IS NULL
		LEFT JOIN lab.pipeline_calcutta_runs pcr ON pcr.calcutta_id = c.id
			AND pcr.pipeline_run_id = (
				SELECT id FROM lab.pipeline_runs
				WHERE investment_model_id = $1::uuid AND status IN ('pending', 'running')
				ORDER BY created_at DESC LIMIT 1
			)
		WHERE c.deleted_at IS NULL
			AND t.deleted_at IS NULL
			AND s.year < EXTRACT(YEAR FROM NOW())
		ORDER BY s.year DESC
	`

	rows, err := r.pool.Query(ctx, query, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calcuttas []lab.CalcuttaProgressResponse
	var totalCalcuttas, predictionsCount, entriesCount, evaluationsCount int
	var payoutSum float64
	var payoutCount int

	for rows.Next() {
		var c lab.CalcuttaProgressResponse
		var stage, status *string
		var progress *float64
		if err := rows.Scan(
			&c.CalcuttaID, &c.CalcuttaName, &c.CalcuttaYear,
			&c.EntryID, &c.EvaluationID,
			&c.HasPredictions, &c.HasEntry, &c.HasEvaluation,
			&c.MeanPayout,
			&stage, &status, &progress, &c.ProgressMessage, &c.ErrorMessage,
		); err != nil {
			return nil, err
		}

		// Set stage/status/progress from active pipeline run if present
		if stage != nil {
			c.Stage = *stage
		} else {
			c.Stage = "completed"
		}
		if status != nil {
			c.Status = *status
		} else if c.HasEvaluation {
			c.Status = "succeeded"
		} else {
			c.Status = "pending"
		}
		if progress != nil {
			c.Progress = *progress
		} else if c.HasEvaluation {
			c.Progress = 1.0
		}

		totalCalcuttas++
		if c.HasPredictions {
			predictionsCount++
		}
		if c.HasEntry {
			entriesCount++
		}
		if c.HasEvaluation {
			evaluationsCount++
		}
		if c.MeanPayout != nil {
			payoutSum += *c.MeanPayout
			payoutCount++
		}

		calcuttas = append(calcuttas, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := &lab.ModelPipelineProgress{
		ModelID:          modelID,
		ModelName:        modelName,
		TotalCalcuttas:   totalCalcuttas,
		PredictionsCount: predictionsCount,
		EntriesCount:     entriesCount,
		EvaluationsCount: evaluationsCount,
		Calcuttas:        calcuttas,
	}

	if activePipeline != nil {
		result.ActivePipelineRunID = &activePipeline.ID
	}

	if payoutCount > 0 {
		avgPayout := payoutSum / float64(payoutCount)
		result.AvgMeanPayout = &avgPayout
	}

	return result, nil
}

// GetHistoricalCalcuttaIDs returns all historical calcutta IDs (years before current year).
func (r *LabRepository) GetHistoricalCalcuttaIDs() ([]string, error) {
	ctx := context.Background()

	query := `
		SELECT c.id::text
		FROM core.calcuttas c
		JOIN core.tournaments t ON t.id = c.tournament_id
		JOIN core.seasons s ON s.id = t.season_id
		WHERE c.deleted_at IS NULL
			AND t.deleted_at IS NULL
			AND s.year < EXTRACT(YEAR FROM NOW())
		ORDER BY s.year DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// EnqueuePipelineJob creates a run_job for a pipeline stage.
func (r *LabRepository) EnqueuePipelineJob(runKind string, pipelineCalcuttaRunID string, params map[string]interface{}) (string, error) {
	ctx := context.Background()

	paramsJSON := "{}"
	if params != nil {
		// Simple JSON encoding
		parts := make([]string, 0, len(params))
		for k, v := range params {
			switch val := v.(type) {
			case string:
				parts = append(parts, fmt.Sprintf(`"%s":"%s"`, k, val))
			case int:
				parts = append(parts, fmt.Sprintf(`"%s":%d`, k, val))
			case float64:
				parts = append(parts, fmt.Sprintf(`"%s":%f`, k, val))
			case bool:
				parts = append(parts, fmt.Sprintf(`"%s":%t`, k, val))
			}
		}
		paramsJSON = "{" + strings.Join(parts, ",") + "}"
	}

	var jobID string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO derived.run_jobs (run_kind, run_key, params_json, status)
		VALUES ($1, $2::uuid, $3::jsonb, 'queued')
		RETURNING run_id::text
	`, runKind, pipelineCalcuttaRunID, paramsJSON).Scan(&jobID)
	if err != nil {
		return "", err
	}
	return jobID, nil
}
