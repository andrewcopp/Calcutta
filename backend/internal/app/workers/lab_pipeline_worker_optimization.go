package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/recommended_entry_bids"
)

func (w *LabPipelineWorker) processOptimizationJob(ctx context.Context, workerID string, job *labPipelineJob, params labPipelineJobParams) bool {
	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 0.4, "optimization", "Optimizing bids with DP allocator")

	start := time.Now()

	// Fetch predictions from database
	var predictionsJSON []byte
	err := w.pool.QueryRow(ctx, `
		SELECT predictions_json FROM lab.entries
		WHERE id = $1::uuid AND deleted_at IS NULL
	`, params.EntryID).Scan(&predictionsJSON)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to fetch predictions: %w", err))
		return false
	}
	if len(predictionsJSON) == 0 {
		w.failLabPipelineJob(ctx, job, errors.New("entry has no predictions - cannot optimize"))
		return false
	}

	// Parse predictions
	type prediction struct {
		TeamID               string  `json:"teamId"`
		PredictedMarketShare float64 `json:"predictedMarketShare"`
		ExpectedPoints       float64 `json:"expectedPoints"`
	}
	var predictions []prediction
	if err := json.Unmarshal(predictionsJSON, &predictions); err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to parse predictions: %w", err))
		return false
	}
	if len(predictions) == 0 {
		w.failLabPipelineJob(ctx, job, errors.New("predictions array is empty - cannot optimize"))
		return false
	}

	// Get calcutta rules for constraints
	var minTeams, maxTeams, maxPerTeam int32
	err = w.pool.QueryRow(ctx, `
		SELECT min_teams, max_teams, max_bid
		FROM core.calcuttas
		WHERE id = $1::uuid AND deleted_at IS NULL
	`, params.CalcuttaID).Scan(&minTeams, &maxTeams, &maxPerTeam)
	if err != nil {
		slog.Error("lab_pipeline_worker failed to load calcutta constraints", "calcutta_id", params.CalcuttaID, "error", err)
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to load calcutta constraints: %w", err))
		return false
	}

	// Get total pool budget (number of entries x budget per entry)
	// This is needed because predicted_market_share is a fraction of the TOTAL pool
	var totalPoolBudget int
	err = w.pool.QueryRow(ctx, `
		SELECT c.budget_points * COUNT(e.id)::int
		FROM core.calcuttas c
		LEFT JOIN core.entries e ON e.calcutta_id = c.id AND e.deleted_at IS NULL
		WHERE c.id = $1::uuid AND c.deleted_at IS NULL
		GROUP BY c.budget_points
	`, params.CalcuttaID).Scan(&totalPoolBudget)
	if err != nil {
		slog.Error("lab_pipeline_worker failed to load total pool budget", "calcutta_id", params.CalcuttaID, "error", err)
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to load total pool budget: %w", err))
		return false
	}
	if totalPoolBudget <= 0 {
		slog.Error("lab_pipeline_worker total pool budget is non-positive", "calcutta_id", params.CalcuttaID, "total_pool_budget", totalPoolBudget)
		w.failLabPipelineJob(ctx, job, fmt.Errorf("total pool budget is non-positive: %d", totalPoolBudget))
		return false
	}

	// Build teams for allocator
	budgetPoints := params.BudgetPoints
	if budgetPoints <= 0 {
		budgetPoints = 100
	}

	teams := make([]recommended_entry_bids.Team, len(predictions))
	for i, pred := range predictions {
		teams[i] = recommended_entry_bids.Team{
			ID:             pred.TeamID,
			ExpectedPoints: pred.ExpectedPoints,
			MarketPoints:   pred.PredictedMarketShare * float64(totalPoolBudget),
		}
	}

	// Run the Go DP allocator
	allocParams := recommended_entry_bids.AllocationParams{
		BudgetPoints: budgetPoints,
		MinTeams:     int(minTeams),
		MaxTeams:     int(maxTeams),
		MinBidPoints: 1,
		MaxBidPoints: int(maxPerTeam),
	}
	result, err := recommended_entry_bids.AllocateBids(teams, allocParams)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("allocator failed: %w", err))
		return false
	}

	// FAIL FAST: Validate the allocation
	totalBid := 0
	for _, bid := range result.Bids {
		totalBid += bid
	}
	numTeams := len(result.Bids)

	// Strict validation - no silent fallbacks
	if totalBid > budgetPoints {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("CRITICAL: allocator violated budget constraint: total=%d > budget=%d", totalBid, budgetPoints))
		return false
	}
	if numTeams > 0 && numTeams < int(minTeams) {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("CRITICAL: allocator violated min_teams constraint: count=%d < min=%d", numTeams, minTeams))
		return false
	}
	if numTeams > int(maxTeams) {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("CRITICAL: allocator violated max_teams constraint: count=%d > max=%d", numTeams, maxTeams))
		return false
	}
	for teamID, bid := range result.Bids {
		if bid > int(maxPerTeam) {
			w.failLabPipelineJob(ctx, job, fmt.Errorf("CRITICAL: allocator violated max_per_team constraint: team=%s bid=%d > max=%d", teamID, bid, maxPerTeam))
			return false
		}
	}

	// Build bids JSON with expected ROI
	type bidRow struct {
		TeamID      string  `json:"teamId"`
		BidPoints   int     `json:"bidPoints"`
		ExpectedROI float64 `json:"expectedRoi"`
	}
	bids := make([]bidRow, 0, len(result.Bids))
	for _, pred := range predictions {
		bid, ok := result.Bids[pred.TeamID]
		if !ok || bid == 0 {
			continue
		}
		marketCost := pred.PredictedMarketShare * float64(budgetPoints)
		expectedROI := 0.0
		if (marketCost + float64(bid)) > 0 {
			expectedROI = pred.ExpectedPoints / (marketCost + float64(bid))
		}
		bids = append(bids, bidRow{
			TeamID:      pred.TeamID,
			BidPoints:   bid,
			ExpectedROI: expectedROI,
		})
	}

	bidsJSON, err := json.Marshal(bids)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to marshal bids: %w", err))
		return false
	}

	// Save bids to database
	optimizerParams := map[string]interface{}{
		"budget_points": budgetPoints,
		"min_teams":     minTeams,
		"max_teams":     maxTeams,
		"max_per_team":  maxPerTeam,
		"min_bid":       1,
	}
	optimizerParamsJSON, _ := json.Marshal(optimizerParams)

	_, err = w.pool.Exec(ctx, `
		UPDATE lab.entries
		SET bids_json = $2::jsonb,
			optimizer_kind = 'dp',
			optimizer_params_json = $3::jsonb,
			updated_at = NOW()
		WHERE id = $1::uuid
	`, params.EntryID, bidsJSON, optimizerParamsJSON)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to save bids: %w", err))
		return false
	}

	dur := time.Since(start)

	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 1.0, "optimization", "Optimization complete")
	w.succeedLabPipelineJob(ctx, job)

	// Update calcutta run and enqueue evaluation
	if _, err := w.pool.Exec(ctx, `
		UPDATE lab.pipeline_calcutta_runs
		SET stage = 'evaluation', progress = 0.66, updated_at = NOW()
		WHERE id = $1::uuid
	`, params.PipelineCalcuttaRunID); err != nil {
		slog.Warn("lab_pipeline_worker update_calcutta_run_evaluation", "error", err)
	}

	// Enqueue evaluation job
	nextParamsJSON, _ := json.Marshal(params)
	var nextJobID string
	err = w.pool.QueryRow(ctx, `
		INSERT INTO derived.run_jobs (run_kind, run_id, run_key, params_json, status)
		VALUES ('lab_evaluation', uuid_generate_v4(), $1::uuid, $2::jsonb, 'queued')
		RETURNING run_id::text
	`, params.PipelineCalcuttaRunID, nextParamsJSON).Scan(&nextJobID)
	if err != nil {
		slog.Warn("lab_pipeline_worker enqueue_evaluation", "error", err)
	} else {
		if _, err := w.pool.Exec(ctx, `
			UPDATE lab.pipeline_calcutta_runs
			SET evaluation_job_id = $2::uuid, updated_at = NOW()
			WHERE id = $1::uuid
		`, params.PipelineCalcuttaRunID, nextJobID); err != nil {
			slog.Warn("lab_pipeline_worker update_evaluation_job_id", "error", err)
		}
	}

	slog.Info("lab_pipeline_worker optimization_success", "worker_id", workerID, "run_id", job.RunID, "teams", numTeams, "total_bid", totalBid, "dur_ms", dur.Milliseconds())
	return true
}
