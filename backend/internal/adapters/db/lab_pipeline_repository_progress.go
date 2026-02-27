package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
)

// GetPipelineProgress returns detailed progress for a pipeline run.
func (r *LabRepository) GetPipelineProgress(ctx context.Context, pipelineRunID string) (*models.LabPipelineProgressResponse, error) {

	// Get pipeline run
	pipelineRun, err := r.GetPipelineRun(ctx, pipelineRunID)
	if err != nil {
		return nil, fmt.Errorf("getting pipeline run for progress %s: %w", pipelineRunID, err)
	}

	// Get model name
	var modelName string
	err = r.pool.QueryRow(ctx, `
		SELECT name FROM lab.investment_models WHERE id = $1::uuid
	`, pipelineRun.InvestmentModelID).Scan(&modelName)
	if err != nil {
		return nil, fmt.Errorf("getting model name for pipeline run %s: %w", pipelineRunID, err)
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
		JOIN core.pools c ON c.id = pcr.calcutta_id
		JOIN core.tournaments t ON t.id = c.tournament_id
		JOIN core.seasons s ON s.id = t.season_id
		LEFT JOIN lab.entries e ON e.id = pcr.entry_id AND e.deleted_at IS NULL
		LEFT JOIN lab.evaluations ev ON ev.id = pcr.evaluation_id AND ev.deleted_at IS NULL
		WHERE pcr.pipeline_run_id = $1::uuid
		ORDER BY s.year DESC
	`

	rows, err := r.pool.Query(ctx, query, pipelineRunID)
	if err != nil {
		return nil, fmt.Errorf("querying calcutta runs for pipeline %s: %w", pipelineRunID, err)
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
			return nil, fmt.Errorf("scanning calcutta progress: %w", err)
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
		return nil, fmt.Errorf("iterating calcutta progress for pipeline %s: %w", pipelineRunID, err)
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
		return nil, fmt.Errorf("getting model name for model %s: %w", modelID, err)
	}

	// Check for active pipeline run
	activePipeline, err := r.GetActivePipelineRun(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("getting active pipeline run for model %s: %w", modelID, err)
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
		FROM core.pools c
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
		return nil, fmt.Errorf("querying calcutta progress for model %s: %w", modelID, err)
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
			return nil, fmt.Errorf("scanning model calcutta progress: %w", err)
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
		return nil, fmt.Errorf("iterating model calcutta progress for model %s: %w", modelID, err)
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
