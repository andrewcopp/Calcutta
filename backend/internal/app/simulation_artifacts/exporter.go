package simulation_artifacts

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) exportArtifacts(ctx context.Context, simulationRunID, runKey, calcuttaEvaluationRunID string) error {
	if s == nil || s.pool == nil {
		return nil
	}
	if strings.TrimSpace(s.artifactsDir) == "" {
		return nil
	}
	if strings.TrimSpace(simulationRunID) == "" {
		return nil
	}
	if strings.TrimSpace(calcuttaEvaluationRunID) == "" {
		return nil
	}

	baseDir := filepath.Join(s.artifactsDir, "simulation", simulationRunID)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return fmt.Errorf("create_artifacts_dir_failed: %w", err)
	}

	perfPath := filepath.Join(baseDir, "entry_performance.v1.jsonl")
	if _, err := s.exportEntryPerformanceJSONL(ctx, calcuttaEvaluationRunID, perfPath); err != nil {
		return fmt.Errorf("export_entry_performance_failed: %w", err)
	}

	return nil
}

func (s *Service) exportEntryPerformanceJSONL(ctx context.Context, calcuttaEvaluationRunID, outPath string) (bool, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			ep.entry_name,
			COALESCE(ep.mean_normalized_payout, 0.0)::double precision,
			COALESCE(ep.median_normalized_payout, 0.0)::double precision,
			COALESCE(ep.p_top1, 0.0)::double precision,
			COALESCE(ep.p_in_money, 0.0)::double precision
		FROM derived.entry_performance ep
		WHERE ep.calcutta_evaluation_run_id = $1::uuid
			AND ep.deleted_at IS NULL
		ORDER BY ep.entry_name ASC
	`, calcuttaEvaluationRunID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	f, err := os.Create(outPath)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	bw := bufio.NewWriter(f)
	defer func() { _ = bw.Flush() }()

	count := 0
	for rows.Next() {
		var entryName string
		var mean, median, pTop1, pInMoney float64
		if err := rows.Scan(&entryName, &mean, &median, &pTop1, &pInMoney); err != nil {
			return false, err
		}
		b, err := json.Marshal(map[string]any{
			"entry_name":               entryName,
			"mean_normalized_payout":   mean,
			"median_normalized_payout": median,
			"p_top1":                   pTop1,
			"p_in_money":               pInMoney,
		})
		if err != nil {
			return false, err
		}
		if _, err := bw.Write(append(b, '\n')); err != nil {
			return false, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

