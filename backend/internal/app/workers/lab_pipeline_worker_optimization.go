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

type optimizationPrediction struct {
	TeamID               string  `json:"teamId"`
	PredictedMarketShare float64 `json:"predictedMarketShare"`
	ExpectedPoints       float64 `json:"expectedPoints"`
}

type optimizationBidRow struct {
	TeamID      string  `json:"teamId"`
	BidPoints   int     `json:"bidPoints"`
	ExpectedROI float64 `json:"expectedRoi"`
}

type optimizationConstraints struct {
	MinTeams        int32
	MaxTeams        int32
	MaxPerTeam      int32
	TotalPoolBudget int
}

func (w *LabPipelineWorker) processOptimizationJob(ctx context.Context, workerID string, job *labPipelineJob, params labPipelineJobParams) bool {
	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 0.4, "optimization", "Optimizing bids with DP allocator")

	start := time.Now()

	predictions, err := w.fetchAndParsePredictions(ctx, params.EntryID)
	if err != nil {
		w.failLabPipelineJob(ctx, job, err)
		return false
	}

	constraints, err := w.fetchOptimizationConstraints(ctx, params.CalcuttaID)
	if err != nil {
		w.failLabPipelineJob(ctx, job, err)
		return false
	}

	budgetPoints := params.BudgetPoints
	if budgetPoints <= 0 {
		budgetPoints = 100
	}

	teams := make([]recommended_entry_bids.Team, len(predictions))
	for i, pred := range predictions {
		teams[i] = recommended_entry_bids.Team{
			ID:             pred.TeamID,
			ExpectedPoints: pred.ExpectedPoints,
			MarketPoints:   pred.PredictedMarketShare * float64(constraints.TotalPoolBudget),
		}
	}

	allocParams := recommended_entry_bids.AllocationParams{
		BudgetPoints: budgetPoints,
		MinTeams:     int(constraints.MinTeams),
		MaxTeams:     int(constraints.MaxTeams),
		MinBidPoints: 1,
		MaxBidPoints: int(constraints.MaxPerTeam),
	}
	result, err := recommended_entry_bids.AllocateBids(teams, allocParams)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("allocator failed: %w", err))
		return false
	}

	if err := validateAllocation(result.Bids, budgetPoints, constraints); err != nil {
		w.failLabPipelineJob(ctx, job, err)
		return false
	}

	bidsJSON, err := buildBidsJSON(predictions, result.Bids, budgetPoints)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to marshal bids: %w", err))
		return false
	}

	if err := w.persistOptimizationResult(ctx, params, bidsJSON, budgetPoints, constraints); err != nil {
		w.failLabPipelineJob(ctx, job, err)
		return false
	}

	dur := time.Since(start)
	numTeams := len(result.Bids)
	totalBid := 0
	for _, bid := range result.Bids {
		totalBid += bid
	}

	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 1.0, "optimization", "Optimization complete")
	w.succeedLabPipelineJob(ctx, job)

	w.enqueueEvaluationJob(ctx, job, params)

	slog.Info("lab_pipeline_worker optimization_success", "worker_id", workerID, "run_id", job.RunID, "teams", numTeams, "total_bid", totalBid, "dur_ms", dur.Milliseconds())
	return true
}

func (w *LabPipelineWorker) fetchAndParsePredictions(ctx context.Context, entryID string) ([]optimizationPrediction, error) {
	var predictionsJSON []byte
	err := w.pool.QueryRow(ctx, `
		SELECT predictions_json FROM lab.entries
		WHERE id = $1::uuid AND deleted_at IS NULL
	`, entryID).Scan(&predictionsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch predictions: %w", err)
	}
	if len(predictionsJSON) == 0 {
		return nil, errors.New("entry has no predictions - cannot optimize")
	}

	var predictions []optimizationPrediction
	if err := json.Unmarshal(predictionsJSON, &predictions); err != nil {
		return nil, fmt.Errorf("failed to parse predictions: %w", err)
	}
	if len(predictions) == 0 {
		return nil, errors.New("predictions array is empty - cannot optimize")
	}

	return predictions, nil
}

func (w *LabPipelineWorker) fetchOptimizationConstraints(ctx context.Context, calcuttaID string) (optimizationConstraints, error) {
	var c optimizationConstraints

	err := w.pool.QueryRow(ctx, `
		SELECT min_teams, max_teams, max_investment_credits
		FROM core.pools
		WHERE id = $1::uuid AND deleted_at IS NULL
	`, calcuttaID).Scan(&c.MinTeams, &c.MaxTeams, &c.MaxPerTeam)
	if err != nil {
		slog.Error("lab_pipeline_worker failed to load calcutta constraints", "calcutta_id", calcuttaID, "error", err)
		return c, fmt.Errorf("failed to load calcutta constraints: %w", err)
	}

	err = w.pool.QueryRow(ctx, `
		SELECT c.budget_credits * COUNT(p.id)::int
		FROM core.pools c
		LEFT JOIN core.portfolios p ON p.pool_id = c.id AND p.deleted_at IS NULL
		WHERE c.id = $1::uuid AND c.deleted_at IS NULL
		GROUP BY c.budget_credits
	`, calcuttaID).Scan(&c.TotalPoolBudget)
	if err != nil {
		slog.Error("lab_pipeline_worker failed to load total pool budget", "calcutta_id", calcuttaID, "error", err)
		return c, fmt.Errorf("failed to load total pool budget: %w", err)
	}
	if c.TotalPoolBudget <= 0 {
		slog.Error("lab_pipeline_worker total pool budget is non-positive", "calcutta_id", calcuttaID, "total_pool_budget", c.TotalPoolBudget)
		return c, fmt.Errorf("total pool budget is non-positive: %d", c.TotalPoolBudget)
	}

	return c, nil
}

func validateAllocation(bids map[string]int, budgetPoints int, constraints optimizationConstraints) error {
	totalBid := 0
	for _, bid := range bids {
		totalBid += bid
	}
	numTeams := len(bids)

	if totalBid > budgetPoints {
		return fmt.Errorf("CRITICAL: allocator violated budget constraint: total=%d > budget=%d", totalBid, budgetPoints)
	}
	if numTeams > 0 && numTeams < int(constraints.MinTeams) {
		return fmt.Errorf("CRITICAL: allocator violated min_teams constraint: count=%d < min=%d", numTeams, constraints.MinTeams)
	}
	if numTeams > int(constraints.MaxTeams) {
		return fmt.Errorf("CRITICAL: allocator violated max_teams constraint: count=%d > max=%d", numTeams, constraints.MaxTeams)
	}
	for teamID, bid := range bids {
		if bid > int(constraints.MaxPerTeam) {
			return fmt.Errorf("CRITICAL: allocator violated max_per_team constraint: team=%s bid=%d > max=%d", teamID, bid, constraints.MaxPerTeam)
		}
	}

	return nil
}

func buildBidsJSON(predictions []optimizationPrediction, bids map[string]int, budgetPoints int) ([]byte, error) {
	rows := make([]optimizationBidRow, 0, len(bids))
	for _, pred := range predictions {
		bid, ok := bids[pred.TeamID]
		if !ok || bid == 0 {
			continue
		}
		marketCost := pred.PredictedMarketShare * float64(budgetPoints)
		expectedROI := 0.0
		if (marketCost + float64(bid)) > 0 {
			expectedROI = pred.ExpectedPoints / (marketCost + float64(bid))
		}
		rows = append(rows, optimizationBidRow{
			TeamID:      pred.TeamID,
			BidPoints:   bid,
			ExpectedROI: expectedROI,
		})
	}

	return json.Marshal(rows)
}

func (w *LabPipelineWorker) persistOptimizationResult(ctx context.Context, params labPipelineJobParams, bidsJSON []byte, budgetPoints int, constraints optimizationConstraints) error {
	optimizerParams := map[string]interface{}{
		"budget_points": budgetPoints,
		"min_teams":     constraints.MinTeams,
		"max_teams":     constraints.MaxTeams,
		"max_per_team":  constraints.MaxPerTeam,
		"min_bid":       1,
	}
	optimizerParamsJSON, _ := json.Marshal(optimizerParams)

	_, err := w.pool.Exec(ctx, `
		UPDATE lab.entries
		SET bids_json = $2::jsonb,
			optimizer_kind = 'dp',
			optimizer_params_json = $3::jsonb,
			updated_at = NOW()
		WHERE id = $1::uuid
	`, params.EntryID, bidsJSON, optimizerParamsJSON)
	if err != nil {
		return fmt.Errorf("failed to save bids: %w", err)
	}

	return nil
}

func (w *LabPipelineWorker) enqueueEvaluationJob(ctx context.Context, job *labPipelineJob, params labPipelineJobParams) {
	if _, err := w.pool.Exec(ctx, `
		UPDATE lab.pipeline_calcutta_runs
		SET stage = 'evaluation', progress = 0.66, updated_at = NOW()
		WHERE id = $1::uuid
	`, params.PipelineCalcuttaRunID); err != nil {
		slog.Warn("lab_pipeline_worker update_calcutta_run_evaluation", "error", err)
	}

	nextParamsJSON, _ := json.Marshal(params)
	var nextJobID string
	err := w.pool.QueryRow(ctx, `
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
}
