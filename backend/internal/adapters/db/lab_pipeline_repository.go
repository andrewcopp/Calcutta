package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
)

// CreatePipelineRun creates a new pipeline run.
func (r *LabRepository) CreatePipelineRun(ctx context.Context, run *models.LabPipelineRun) (*models.LabPipelineRun, error) {

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

	var result models.LabPipelineRun
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
func (r *LabRepository) GetPipelineRun(ctx context.Context, id string) (*models.LabPipelineRun, error) {

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

	var result models.LabPipelineRun
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
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "pipeline_run", ID: id}
	}
	if err != nil {
		return nil, err
	}
	result.TargetCalcuttaIDs = targetIDs
	return &result, nil
}

// UpdatePipelineRunStatus updates the status of a pipeline run.
func (r *LabRepository) UpdatePipelineRunStatus(ctx context.Context, id string, status string, errorMessage *string) error {

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

// GetActivePipelineRun returns the most recent running or pending pipeline run for a model.
func (r *LabRepository) GetActivePipelineRun(ctx context.Context, modelID string) (*models.LabPipelineRun, error) {

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

	var result models.LabPipelineRun
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
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	result.TargetCalcuttaIDs = targetIDs
	return &result, nil
}

// CreatePipelineCalcuttaRuns creates calcutta run records for a pipeline.
func (r *LabRepository) CreatePipelineCalcuttaRuns(ctx context.Context, pipelineRunID string, calcuttaIDs []string) error {

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

// GetPipelineProgress returns detailed progress for a pipeline run.
func (r *LabRepository) GetPipelineProgress(ctx context.Context, pipelineRunID string) (*models.LabPipelineProgressResponse, error) {

	// Get pipeline run
	pipelineRun, err := r.GetPipelineRun(ctx, pipelineRunID)
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

	var calcuttas []models.LabCalcuttaProgressResponse
	var summary models.LabPipelineProgressSummary
	var payoutSum float64
	var payoutCount int

	for rows.Next() {
		var c models.LabCalcuttaProgressResponse
		var meanPayout *float64
		if err := rows.Scan(
			&c.CalcuttaID, &c.CalcuttaName, &c.CalcuttaYear,
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

	return &models.LabPipelineProgressResponse{
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
func (r *LabRepository) GetModelPipelineProgress(ctx context.Context, modelID string) (*models.LabModelPipelineProgress, error) {

	// Get model name
	var modelName string
	err := r.pool.QueryRow(ctx, `
		SELECT name FROM lab.investment_models WHERE id = $1::uuid AND deleted_at IS NULL
	`, modelID).Scan(&modelName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "investment_model", ID: modelID}
	}
	if err != nil {
		return nil, err
	}

	// Check for active pipeline run
	activePipeline, err := r.GetActivePipelineRun(ctx, modelID)
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
			eer.rank AS our_rank,
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
		LEFT JOIN lab.evaluation_entry_results eer ON eer.evaluation_id = ev.id
			AND eer.entry_name = 'Our Strategy'
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

	var calcuttas []models.LabCalcuttaProgressResponse
	var totalCalcuttas, predictionsCount, entriesCount, evaluationsCount int
	var payoutSum float64
	var payoutCount int

	for rows.Next() {
		var c models.LabCalcuttaProgressResponse
		var stage, status *string
		var progress *float64
		if err := rows.Scan(
			&c.CalcuttaID, &c.CalcuttaName, &c.CalcuttaYear,
			&c.EntryID, &c.EvaluationID,
			&c.HasPredictions, &c.HasEntry, &c.HasEvaluation,
			&c.MeanPayout, &c.OurRank,
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

	result := &models.LabModelPipelineProgress{
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
func (r *LabRepository) GetHistoricalCalcuttaIDs(ctx context.Context) ([]string, error) {

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

// SoftDeleteModelArtifacts soft-deletes all entries and evaluations for a model.
// This is used when force_rerun=true to ensure fresh results.
func (r *LabRepository) SoftDeleteModelArtifacts(ctx context.Context, modelID string) error {

	// Use a transaction to ensure atomicity
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Hard-delete evaluation_entry_results (no deleted_at column)
	// These are linked to evaluations which are linked to entries
	_, err = tx.Exec(ctx, `
		DELETE FROM lab.evaluation_entry_results
		WHERE evaluation_id IN (
			SELECT ev.id FROM lab.evaluations ev
			JOIN lab.entries e ON e.id = ev.entry_id
			WHERE e.investment_model_id = $1::uuid
				AND e.deleted_at IS NULL
				AND ev.deleted_at IS NULL
		)
	`, modelID)
	if err != nil {
		return fmt.Errorf("failed to delete evaluation_entry_results: %w", err)
	}

	// 2. Soft-delete evaluations
	_, err = tx.Exec(ctx, `
		UPDATE lab.evaluations
		SET deleted_at = NOW()
		WHERE entry_id IN (
			SELECT id FROM lab.entries
			WHERE investment_model_id = $1::uuid
				AND deleted_at IS NULL
		)
		AND deleted_at IS NULL
	`, modelID)
	if err != nil {
		return fmt.Errorf("failed to soft-delete evaluations: %w", err)
	}

	// 3. Soft-delete entries
	_, err = tx.Exec(ctx, `
		UPDATE lab.entries
		SET deleted_at = NOW()
		WHERE investment_model_id = $1::uuid
			AND deleted_at IS NULL
	`, modelID)
	if err != nil {
		return fmt.Errorf("failed to soft-delete entries: %w", err)
	}

	// 4. Cancel any active pipeline runs for this model
	_, err = tx.Exec(ctx, `
		UPDATE lab.pipeline_runs
		SET status = 'cancelled',
			error_message = 'Superseded by force re-run',
			finished_at = NOW(),
			updated_at = NOW()
		WHERE investment_model_id = $1::uuid
			AND status IN ('pending', 'running')
	`, modelID)
	if err != nil {
		return fmt.Errorf("failed to cancel active pipeline runs: %w", err)
	}

	return tx.Commit(ctx)
}
