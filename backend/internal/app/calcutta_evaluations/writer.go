package calcutta_evaluations

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (s *Service) deleteSimulationOutcomes(ctx context.Context, runID string, calcuttaEvaluationRunID string) error {
	var err error
	if calcuttaEvaluationRunID != "" {
		_, err = s.pool.Exec(ctx, "DELETE FROM derived.entry_simulation_outcomes WHERE calcutta_evaluation_run_id = $1", calcuttaEvaluationRunID)
	} else {
		_, err = s.pool.Exec(ctx, "DELETE FROM derived.entry_simulation_outcomes WHERE run_id = $1", runID)
	}
	return err
}

func (s *Service) writeSimulationOutcomes(ctx context.Context, runID string, calcuttaEvaluationRunID string, results []SimulationResult) error {
	var err error
	if calcuttaEvaluationRunID != "" {
		_, err = s.pool.Exec(ctx, "DELETE FROM derived.entry_simulation_outcomes WHERE calcutta_evaluation_run_id = $1", calcuttaEvaluationRunID)
	} else {
		_, err = s.pool.Exec(ctx, "DELETE FROM derived.entry_simulation_outcomes WHERE run_id = $1", runID)
	}
	if err != nil {
		return err
	}

	// Batch insert new results
	batch := &pgx.Batch{}
	for _, r := range results {
		var evalID any
		if calcuttaEvaluationRunID != "" {
			evalID = calcuttaEvaluationRunID
		} else {
			evalID = nil
		}
		batch.Queue(`
			INSERT INTO derived.entry_simulation_outcomes (run_id, entry_name, sim_id, points_scored, payout_cents, rank, calcutta_evaluation_run_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, runID, r.EntryName, r.SimID, r.TotalPoints, r.PayoutCents, r.Rank, evalID)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(results); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) writePerformanceMetrics(ctx context.Context, runID string, calcuttaEvaluationRunID string, performance map[string]*EntryPerformance) error {
	var err error
	if calcuttaEvaluationRunID != "" {
		_, err = s.pool.Exec(ctx, "DELETE FROM derived.entry_performance WHERE calcutta_evaluation_run_id = $1", calcuttaEvaluationRunID)
	} else {
		_, err = s.pool.Exec(ctx, "DELETE FROM derived.entry_performance WHERE run_id = $1", runID)
	}
	if err != nil {
		return err
	}

	// Insert new performance metrics
	for _, p := range performance {
		var evalID any
		if calcuttaEvaluationRunID != "" {
			evalID = calcuttaEvaluationRunID
		} else {
			evalID = nil
		}
		_, err := s.pool.Exec(ctx, `
			INSERT INTO derived.entry_performance (run_id, entry_name, mean_normalized_payout, median_normalized_payout, p_top1, p_in_money, calcutta_evaluation_run_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, runID, p.EntryName, p.MeanPayout, p.MedianPayout, p.PTop1, p.PInMoney, evalID)
		if err != nil {
			return err
		}
	}

	return nil
}
