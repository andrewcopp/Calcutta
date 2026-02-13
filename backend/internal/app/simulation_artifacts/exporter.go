package simulation_artifacts

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type artifactExportResult struct {
	ArtifactKind  string
	SchemaVersion string
	StorageURI    string
	RowCount      int
}

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

	results := make([]artifactExportResult, 0, 2)

	perfPath := filepath.Join(baseDir, "entry_performance.v1.jsonl")
	if res, ok, err := s.exportEntryPerformanceJSONL(ctx, calcuttaEvaluationRunID, perfPath); err != nil {
		return fmt.Errorf("export_entry_performance_failed: %w", err)
	} else if ok {
		results = append(results, res)
	}

	outcomesPath := filepath.Join(baseDir, "entry_simulation_outcomes.v1.jsonl")
	if res, ok, err := s.exportEntrySimulationOutcomesJSONL(ctx, calcuttaEvaluationRunID, outcomesPath); err != nil {
		return fmt.Errorf("export_entry_simulation_outcomes_failed: %w", err)
	} else if ok {
		results = append(results, res)
	}

	var runKeyParam any
	if strings.TrimSpace(runKey) != "" {
		runKeyParam = runKey
	} else {
		runKeyParam = nil
	}

	for _, res := range results {
		summary := map[string]any{
			"rowCount": res.RowCount,
		}
		summaryJSON, _ := json.Marshal(summary)
		_, err := s.pool.Exec(ctx, `
			INSERT INTO derived.run_artifacts (
				run_kind,
				run_id,
				run_key,
				artifact_kind,
				schema_version,
				storage_uri,
				summary_json
			)
			VALUES ('simulation', $1::uuid, $2::uuid, $3, $4, $5, $6::jsonb)
			ON CONFLICT (run_kind, run_id, artifact_kind) WHERE deleted_at IS NULL
			DO UPDATE
			SET run_key = EXCLUDED.run_key,
				schema_version = EXCLUDED.schema_version,
				storage_uri = EXCLUDED.storage_uri,
				summary_json = EXCLUDED.summary_json,
				updated_at = NOW(),
				deleted_at = NULL
		`, simulationRunID, runKeyParam, res.ArtifactKind, res.SchemaVersion, res.StorageURI, summaryJSON)
		if err != nil {
			return fmt.Errorf("upsert_run_artifact_failed kind=%s: %w", res.ArtifactKind, err)
		}
	}

	return nil
}

func (s *Service) exportEntryPerformanceJSONL(ctx context.Context, calcuttaEvaluationRunID, outPath string) (artifactExportResult, bool, error) {
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
		return artifactExportResult{}, false, err
	}
	defer rows.Close()

	f, err := os.Create(outPath)
	if err != nil {
		return artifactExportResult{}, false, err
	}
	defer func() { _ = f.Close() }()

	bw := bufio.NewWriter(f)
	defer func() { _ = bw.Flush() }()

	count := 0
	for rows.Next() {
		var entryName string
		var mean, median, pTop1, pInMoney float64
		if err := rows.Scan(&entryName, &mean, &median, &pTop1, &pInMoney); err != nil {
			return artifactExportResult{}, false, err
		}
		b, err := json.Marshal(map[string]any{
			"entry_name":               entryName,
			"mean_normalized_payout":   mean,
			"median_normalized_payout": median,
			"p_top1":                   pTop1,
			"p_in_money":               pInMoney,
		})
		if err != nil {
			return artifactExportResult{}, false, err
		}
		if _, err := bw.Write(append(b, '\n')); err != nil {
			return artifactExportResult{}, false, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return artifactExportResult{}, false, err
	}

	if count == 0 {
		return artifactExportResult{}, false, nil
	}

	abs, _ := filepath.Abs(outPath)
	u := (&url.URL{Scheme: "file", Path: abs}).String()
	return artifactExportResult{ArtifactKind: "entry_performance_jsonl", SchemaVersion: "v1", StorageURI: u, RowCount: count}, true, nil
}

func (s *Service) exportEntrySimulationOutcomesJSONL(ctx context.Context, calcuttaEvaluationRunID, outPath string) (artifactExportResult, bool, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			eo.entry_name,
			eo.sim_id::int,
			COALESCE(eo.points_scored, 0.0)::double precision,
			COALESCE(eo.payout_cents, 0)::int,
			COALESCE(eo.rank, 0)::int
		FROM derived.entry_simulation_outcomes eo
		WHERE eo.calcutta_evaluation_run_id = $1::uuid
			AND eo.deleted_at IS NULL
		ORDER BY eo.entry_name ASC, eo.sim_id ASC
	`, calcuttaEvaluationRunID)
	if err != nil {
		return artifactExportResult{}, false, err
	}
	defer rows.Close()

	f, err := os.Create(outPath)
	if err != nil {
		return artifactExportResult{}, false, err
	}
	defer func() { _ = f.Close() }()

	bw := bufio.NewWriter(f)
	defer func() { _ = bw.Flush() }()

	count := 0
	for rows.Next() {
		var entryName string
		var simID, payoutCents, rank int
		var pointsScored float64
		if err := rows.Scan(&entryName, &simID, &pointsScored, &payoutCents, &rank); err != nil {
			return artifactExportResult{}, false, err
		}
		b, err := json.Marshal(map[string]any{
			"entry_name":    entryName,
			"sim_id":        simID,
			"points_scored": pointsScored,
			"payout_cents":  payoutCents,
			"rank":          rank,
		})
		if err != nil {
			return artifactExportResult{}, false, err
		}
		if _, err := bw.Write(append(b, '\n')); err != nil {
			return artifactExportResult{}, false, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return artifactExportResult{}, false, err
	}

	if count == 0 {
		return artifactExportResult{}, false, nil
	}

	abs, _ := filepath.Abs(outPath)
	u := (&url.URL{Scheme: "file", Path: abs}).String()
	return artifactExportResult{ArtifactKind: "entry_simulation_outcomes_jsonl", SchemaVersion: "v1", StorageURI: u, RowCount: count}, true, nil
}
