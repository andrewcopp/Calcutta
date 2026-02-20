package workers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/prediction"
)

func (w *LabPipelineWorker) processPredictionsJob(ctx context.Context, workerID string, job *labPipelineJob, params labPipelineJobParams) bool {
	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 0.1, "predictions", "Generating market predictions")

	// Get model kind to determine which approach to use
	var modelKind string
	err := w.pool.QueryRow(ctx, `
		SELECT kind FROM lab.investment_models WHERE id = $1::uuid AND deleted_at IS NULL
	`, params.InvestmentModelID).Scan(&modelKind)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to get model kind: %w", err))
		return false
	}

	var entryID string
	start := time.Now()

	// Use Go-native predictions for naive_ev and oracle models
	// ML-based models (ridge, etc.) still use Python for market share prediction
	switch modelKind {
	case "naive_ev", "oracle":
		entryID, err = w.processGoPredictions(ctx, workerID, job, params, modelKind)
	default:
		entryID, err = w.processPythonPredictions(ctx, workerID, job, params, modelKind)
	}

	dur := time.Since(start)

	if err != nil {
		w.failLabPipelineJob(ctx, job, err)
		slog.Warn("lab_pipeline_worker predictions_fail", "worker_id", workerID, "run_id", job.RunID, "model_kind", modelKind, "dur_ms", dur.Milliseconds(), "error", err)
		return false
	}

	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 1.0, "predictions", "Predictions complete")

	// Mark job succeeded
	w.succeedLabPipelineJob(ctx, job)

	// Update calcutta run and enqueue next stage
	if _, err := w.pool.Exec(ctx, `
		UPDATE lab.pipeline_calcutta_runs
		SET entry_id = $2::uuid, stage = 'optimization', progress = 0.33, updated_at = NOW()
		WHERE id = $1::uuid
	`, params.PipelineCalcuttaRunID, entryID); err != nil {
		slog.Warn("lab_pipeline_worker update_calcutta_run_optimization", "error", err)
	}

	// Enqueue optimization job
	nextParams := params
	nextParams.EntryID = entryID
	nextParamsJSON, _ := json.Marshal(nextParams)

	var nextJobID string
	err = w.pool.QueryRow(ctx, `
		INSERT INTO derived.run_jobs (run_kind, run_id, run_key, params_json, status)
		VALUES ('lab_optimization', uuid_generate_v4(), $1::uuid, $2::jsonb, 'queued')
		RETURNING run_id::text
	`, params.PipelineCalcuttaRunID, nextParamsJSON).Scan(&nextJobID)
	if err != nil {
		slog.Warn("lab_pipeline_worker enqueue_optimization", "error", err)
	} else {
		if _, err := w.pool.Exec(ctx, `
			UPDATE lab.pipeline_calcutta_runs
			SET optimization_job_id = $2::uuid, updated_at = NOW()
			WHERE id = $1::uuid
		`, params.PipelineCalcuttaRunID, nextJobID); err != nil {
			slog.Warn("lab_pipeline_worker update_optimization_job_id", "error", err)
		}
	}

	slog.Info("lab_pipeline_worker predictions_success", "worker_id", workerID, "run_id", job.RunID, "entry_id", entryID, "model_kind", modelKind, "dur_ms", dur.Milliseconds())
	return true
}

// processGoPredictions generates predictions using the native Go prediction service.
// Used for naive_ev and oracle model kinds.
func (w *LabPipelineWorker) processGoPredictions(ctx context.Context, workerID string, job *labPipelineJob, params labPipelineJobParams, modelKind string) (string, error) {
	// Get tournament_id from calcutta
	var tournamentID string
	err := w.pool.QueryRow(ctx, `
		SELECT tournament_id::text FROM core.calcuttas WHERE id = $1::uuid AND deleted_at IS NULL
	`, params.CalcuttaID).Scan(&tournamentID)
	if err != nil {
		return "", fmt.Errorf("failed to get tournament_id: %w", err)
	}

	// Generate or get predictions using Go prediction service
	predSvc := prediction.New(w.pool)

	// Check if we already have predictions for this tournament
	batchID, found, err := predSvc.GetLatestBatchID(ctx, tournamentID)
	if err != nil {
		return "", fmt.Errorf("failed to check for existing predictions: %w", err)
	}

	// Generate new predictions if none exist
	if !found {
		result, err := predSvc.Run(ctx, prediction.RunParams{
			TournamentID:         tournamentID,
			ProbabilitySourceKey: "kenpom",
		})
		if err != nil {
			return "", fmt.Errorf("failed to generate predictions: %w", err)
		}
		batchID = result.BatchID
		slog.Info("lab_pipeline_worker generated_predictions", "worker_id", workerID, "batch_id", batchID, "team_count", result.TeamCount)
	}

	// Get expected points from predictions
	expectedPointsMap, err := predSvc.GetExpectedPointsMap(ctx, tournamentID)
	if err != nil {
		return "", fmt.Errorf("failed to get expected points: %w", err)
	}

	// For naive_ev: market share is proportional to expected points
	// For oracle: market share comes from actual bids (excluding specified entry)
	var marketShareMap map[string]float64

	if modelKind == "naive_ev" {
		// Naive EV: market share proportional to expected points
		totalExpected := 0.0
		for _, ep := range expectedPointsMap {
			totalExpected += ep
		}
		marketShareMap = make(map[string]float64, len(expectedPointsMap))
		for teamID, ep := range expectedPointsMap {
			if totalExpected > 0 {
				marketShareMap[teamID] = ep / totalExpected
			}
		}
	} else if modelKind == "oracle" {
		// Oracle: use actual market bids
		marketShareMap, err = w.getActualMarketShare(ctx, params.CalcuttaID, params.ExcludedEntryName)
		if err != nil {
			return "", fmt.Errorf("failed to get actual market share: %w", err)
		}
	}

	// Create lab entry with predictions
	entryID, err := w.createLabEntry(ctx, params, expectedPointsMap, marketShareMap)
	if err != nil {
		return "", fmt.Errorf("failed to create lab entry: %w", err)
	}

	return entryID, nil
}

// processPythonPredictions runs the Python prediction script for ML-based models.
func (w *LabPipelineWorker) processPythonPredictions(ctx context.Context, workerID string, job *labPipelineJob, params labPipelineJobParams, modelKind string) (string, error) {
	pythonBin := w.cfg.PythonBin
	scriptName := "data-science/scripts/generate_lab_predictions.py"

	scriptPath := w.resolvePythonScript(scriptName)
	if scriptPath == "" {
		return "", fmt.Errorf("predictions script not found: %s", scriptName)
	}

	cmd := exec.CommandContext(ctx, pythonBin, scriptPath,
		"--model-id", params.InvestmentModelID,
		"--calcutta-id", params.CalcuttaID,
		"--json-output",
	)
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", errors.New(errMsg)
	}

	// Parse output to get entry ID
	var result struct {
		OK      bool   `json:"ok"`
		EntryID string `json:"entry_id"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || !result.OK {
		errMsg := result.Error
		if errMsg == "" {
			errMsg = "predictions script returned error"
		}
		return "", errors.New(errMsg)
	}

	return result.EntryID, nil
}

// getActualMarketShare returns the actual market share for each team based on real bids.
func (w *LabPipelineWorker) getActualMarketShare(ctx context.Context, calcuttaID string, excludedEntryName string) (map[string]float64, error) {
	query := `
		SELECT b.team_id::text, SUM(b.bid_points)::float
		FROM core.bids b
		JOIN core.entries e ON e.id = b.entry_id AND e.deleted_at IS NULL
		WHERE e.calcutta_id = $1::uuid
			AND b.deleted_at IS NULL
			AND ($2 = '' OR e.name != $2)
		GROUP BY b.team_id
	`

	rows, err := w.pool.Query(ctx, query, calcuttaID, excludedEntryName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bidsByTeam := make(map[string]float64)
	totalBids := 0.0
	for rows.Next() {
		var teamID string
		var bids float64
		if err := rows.Scan(&teamID, &bids); err != nil {
			return nil, err
		}
		bidsByTeam[teamID] = bids
		totalBids += bids
	}

	// Convert to market share
	marketShare := make(map[string]float64, len(bidsByTeam))
	for teamID, bids := range bidsByTeam {
		if totalBids > 0 {
			marketShare[teamID] = bids / totalBids
		}
	}

	return marketShare, rows.Err()
}

// createLabEntry creates a lab.entries record with predictions.
func (w *LabPipelineWorker) createLabEntry(ctx context.Context, params labPipelineJobParams, expectedPointsMap map[string]float64, marketShareMap map[string]float64) (string, error) {
	// Build predictions JSON
	type predictionRow struct {
		TeamID               string  `json:"team_id"`
		PredictedMarketShare float64 `json:"predicted_market_share"`
		ExpectedPoints       float64 `json:"expected_points"`
	}

	var predictions []predictionRow
	for teamID, expectedPoints := range expectedPointsMap {
		marketShare := marketShareMap[teamID]
		predictions = append(predictions, predictionRow{
			TeamID:               teamID,
			PredictedMarketShare: marketShare,
			ExpectedPoints:       expectedPoints,
		})
	}

	predictionsJSON, err := json.Marshal(predictions)
	if err != nil {
		return "", fmt.Errorf("failed to marshal predictions: %w", err)
	}

	// Insert entry
	var entryID string
	err = w.pool.QueryRow(ctx, `
		INSERT INTO lab.entries (
			id, investment_model_id, calcutta_id,
			game_outcome_kind, game_outcome_params_json,
			optimizer_kind, optimizer_params_json,
			starting_state_key, predictions_json, bids_json
		)
		VALUES (
			uuid_generate_v4(), $1::uuid, $2::uuid,
			'kenpom', '{}'::jsonb,
			'pending', '{}'::jsonb,
			'post_first_four', $3::jsonb, '[]'::jsonb
		)
		ON CONFLICT (investment_model_id, calcutta_id, starting_state_key)
		WHERE deleted_at IS NULL
		DO UPDATE SET
			game_outcome_kind = EXCLUDED.game_outcome_kind,
			game_outcome_params_json = EXCLUDED.game_outcome_params_json,
			predictions_json = EXCLUDED.predictions_json,
			updated_at = NOW()
		RETURNING id::text
	`, params.InvestmentModelID, params.CalcuttaID, predictionsJSON).Scan(&entryID)
	if err != nil {
		return "", fmt.Errorf("failed to insert entry: %w", err)
	}

	return entryID, nil
}
