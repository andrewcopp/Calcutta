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
	if status == "succeeded" || status == "failed" || status == "cancelled" || status == "partial" {
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
