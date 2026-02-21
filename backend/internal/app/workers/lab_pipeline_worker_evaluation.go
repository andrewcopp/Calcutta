package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	appcalcuttaevaluations "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta_evaluations"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func (w *LabPipelineWorker) processEvaluationJob(ctx context.Context, workerID string, job *labPipelineJob, params labPipelineJobParams) bool {
	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 0.7, "evaluation", "Running simulation")

	start := time.Now()

	// Get entry bids and calcutta_id
	var bidsJSON []byte
	var calcuttaID string
	err := w.pool.QueryRow(ctx, `
		SELECT bids_json, calcutta_id::text FROM lab.entries WHERE id = $1::uuid AND deleted_at IS NULL
	`, params.EntryID).Scan(&bidsJSON, &calcuttaID)
	if err != nil {
		w.failLabPipelineJob(ctx, job, errors.New("failed to get entry bids: "+err.Error()))
		return false
	}

	// Parse bids into map for evaluation
	type bidEntry struct {
		TeamID    string `json:"teamId"`
		BidPoints int    `json:"bidPoints"`
	}
	var bids []bidEntry
	if err := json.Unmarshal(bidsJSON, &bids); err != nil {
		w.failLabPipelineJob(ctx, job, errors.New("failed to parse bids: "+err.Error()))
		return false
	}

	labEntryBids := make(map[string]int)
	for _, b := range bids {
		if b.BidPoints > 0 {
			labEntryBids[b.TeamID] = b.BidPoints
		}
	}

	if len(labEntryBids) == 0 {
		w.failLabPipelineJob(ctx, job, errors.New("entry has no bids to evaluate"))
		return false
	}

	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 0.75, "evaluation", "Running "+fmt.Sprintf("%d", params.NSims)+" simulations")

	// Run evaluation using calcutta_evaluations service
	evalService := appcalcuttaevaluations.New(w.pool,
		appcalcuttaevaluations.WithTournamentResolver(dbadapters.NewTournamentQueryRepository(w.pool)),
	)
	result, err := evalService.EvaluateLabEntry(ctx, calcuttaID, labEntryBids, params.ExcludedEntryName)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("evaluation failed: %w", err))
		return false
	}

	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 0.95, "evaluation", "Saving results")

	// Extract lab strategy rank from results
	var ourRank int
	for _, entry := range result.AllEntryResults {
		if entry.EntryName == models.LabStrategyEntryName {
			ourRank = entry.Rank
			break
		}
	}

	// Create or update lab.evaluations row with results
	var evaluationID string
	err = w.pool.QueryRow(ctx, `
		INSERT INTO lab.evaluations (entry_id, n_sims, seed, mean_normalized_payout, median_normalized_payout, p_top1, p_in_money, our_rank)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (entry_id, n_sims, seed) WHERE deleted_at IS NULL
		DO UPDATE SET
			mean_normalized_payout = EXCLUDED.mean_normalized_payout,
			median_normalized_payout = EXCLUDED.median_normalized_payout,
			p_top1 = EXCLUDED.p_top1,
			p_in_money = EXCLUDED.p_in_money,
			our_rank = EXCLUDED.our_rank,
			updated_at = NOW()
		RETURNING id::text
	`, params.EntryID, result.NSims, params.Seed, result.MeanNormalizedPayout, result.MedianNormalizedPayout, result.PTop1, result.PInMoney, ourRank).Scan(&evaluationID)
	if err != nil {
		w.failLabPipelineJob(ctx, job, errors.New("failed to save evaluation: "+err.Error()))
		return false
	}

	// Save per-entry results
	if len(result.AllEntryResults) == 0 {
		w.failLabPipelineJob(ctx, job, errors.New("evaluation produced no entry results"))
		return false
	}
	{
		// Delete existing results for this evaluation (in case of re-run)
		if _, err := w.pool.Exec(ctx, `
			DELETE FROM lab.evaluation_entry_results WHERE evaluation_id = $1::uuid
		`, evaluationID); err != nil {
			slog.Warn("lab_pipeline_worker delete_old_entry_results", "error", err)
		}

		// Insert all entry results
		for _, entry := range result.AllEntryResults {
			_, err := w.pool.Exec(ctx, `
				INSERT INTO lab.evaluation_entry_results (evaluation_id, entry_name, mean_normalized_payout, p_top1, p_in_money, rank)
				VALUES ($1::uuid, $2, $3, $4, $5, $6)
			`, evaluationID, entry.EntryName, entry.MeanPayout, entry.PTop1, entry.PInMoney, entry.Rank)
			if err != nil {
				w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to save entry result for %s: %w", entry.EntryName, err))
				return false
			}
		}
	}

	dur := time.Since(start)

	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 1.0, "evaluation", "Evaluation complete")
	w.succeedLabPipelineJob(ctx, job)

	// Update calcutta run as completed
	if _, err := w.pool.Exec(ctx, `
		UPDATE lab.pipeline_calcutta_runs
		SET stage = 'completed', status = 'succeeded', progress = 1.0,
		    evaluation_id = $2::uuid, finished_at = NOW(), updated_at = NOW()
		WHERE id = $1::uuid
	`, params.PipelineCalcuttaRunID, evaluationID); err != nil {
		slog.Warn("lab_pipeline_worker update_calcutta_run_completed", "error", err)
	}

	slog.Info("lab_pipeline_worker evaluation_success", "worker_id", workerID, "run_id", job.RunID, "evaluation_id", evaluationID, "n_sims", result.NSims, "mean_payout", result.MeanNormalizedPayout, "p_top1", result.PTop1, "dur_ms", dur.Milliseconds())
	return true
}
