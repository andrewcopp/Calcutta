package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type simulationRunToDelete struct {
	ID                    uuid.UUID
	RunKey                uuid.UUID
	CalcuttaEvaluationRun uuid.UUID
	FinishedAt            time.Time
}

func main() {
	platform.InitLogger()
	if err := run(); err != nil {
		slog.Error("cmd_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	retentionDays := flag.Int("retention-days", 30, "delete simulation runs completed more than N days ago")
	maxRuns := flag.Int("max-runs", 500, "max runs to delete in one execution")
	apply := flag.Bool("apply", false, "apply deletions (default: dry run)")
	artifactsDirFlag := flag.String("artifacts-dir", "", "artifact base directory (defaults to ARTIFACTS_DIR)")
	flag.Parse()

	if *retentionDays <= 0 {
		return fmt.Errorf("retention-days must be > 0")
	}
	if *maxRuns <= 0 {
		return fmt.Errorf("max-runs must be > 0")
	}

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return err
	}

	artifactsDir := strings.TrimSpace(*artifactsDirFlag)
	if artifactsDir == "" {
		artifactsDir = strings.TrimSpace(cfg.ArtifactsDir)
	}
	absArtifactsDir := ""
	if artifactsDir != "" {
		abs, err := filepath.Abs(artifactsDir)
		if err != nil {
			return fmt.Errorf("resolve_artifacts_dir_failed: %w", err)
		}
		absArtifactsDir = filepath.Clean(abs)
	}

	ctx := context.Background()
	pool, err := platform.OpenPGXPool(ctx, cfg, nil)
	if err != nil {
		return fmt.Errorf("db_connect_failed: %w", err)
	}
	defer pool.Close()

	cutoff := time.Now().UTC().Add(-time.Duration(*retentionDays) * 24 * time.Hour)
	slog.Info("retention_start", "cutoff", cutoff.Format(time.RFC3339), "apply", *apply, "max_runs", *maxRuns, "artifacts_dir", absArtifactsDir)

	runs, err := selectSimulationRunsToDelete(ctx, pool, cutoff, *maxRuns)
	if err != nil {
		return err
	}
	if len(runs) == 0 {
		slog.Info("retention_noop")
		return nil
	}

	runIDs := make([]uuid.UUID, 0, len(runs))
	runsByID := make(map[uuid.UUID]simulationRunToDelete, len(runs))
	evalIDs := make([]uuid.UUID, 0, len(runs))
	evalSet := make(map[uuid.UUID]struct{}, len(runs))
	for _, r := range runs {
		runIDs = append(runIDs, r.ID)
		runsByID[r.ID] = r
		if r.CalcuttaEvaluationRun != uuid.Nil {
			if _, ok := evalSet[r.CalcuttaEvaluationRun]; !ok {
				evalSet[r.CalcuttaEvaluationRun] = struct{}{}
				evalIDs = append(evalIDs, r.CalcuttaEvaluationRun)
			}
		}
	}

	evalIDsSafeToDelete := make(map[uuid.UUID]bool, len(evalIDs))
	for _, evalID := range evalIDs {
		ok, err := evalRunExclusiveToDeletionSet(ctx, pool, evalID, runIDs)
		if err != nil {
			return err
		}
		evalIDsSafeToDelete[evalID] = ok
		if !ok {
			slog.Warn("eval_run_shared_skip_outputs", "calcutta_evaluation_run_id", evalID.String())
		}
	}

	if !*apply {
		for _, r := range runs {
			slog.Info(
				"retention_candidate",
				"simulation_run_id", r.ID.String(),
				"run_key", r.RunKey.String(),
				"finished_at", r.FinishedAt.Format(time.RFC3339),
				"calcutta_evaluation_run_id", r.CalcuttaEvaluationRun.String(),
			)
		}
		slog.Info("retention_dry_run_complete", "candidates", len(runs))
		return nil
	}

	// 1) Delete DB rows in a transaction.
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("tx_begin_failed: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	deletedRuns := 0
	for _, r := range runs {
		if evalIDsSafeToDelete[r.CalcuttaEvaluationRun] {
			if _, err := tx.Exec(ctx, `
				DELETE FROM derived.entry_simulation_outcomes
				WHERE calcutta_evaluation_run_id = $1::uuid
			`, r.CalcuttaEvaluationRun); err != nil {
				return fmt.Errorf("delete_entry_simulation_outcomes_failed run_id=%s: %w", r.ID.String(), err)
			}
			if _, err := tx.Exec(ctx, `
				DELETE FROM derived.entry_performance
				WHERE calcutta_evaluation_run_id = $1::uuid
			`, r.CalcuttaEvaluationRun); err != nil {
				return fmt.Errorf("delete_entry_performance_failed run_id=%s: %w", r.ID.String(), err)
			}
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM derived.run_jobs
			WHERE run_kind = 'simulation'
				AND run_id = $1::uuid
			`, r.ID); err != nil {
			return fmt.Errorf("delete_run_jobs_failed run_id=%s: %w", r.ID.String(), err)
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM derived.simulation_runs
			WHERE id = $1::uuid
			`, r.ID); err != nil {
			return fmt.Errorf("delete_simulation_run_failed run_id=%s: %w", r.ID.String(), err)
		}

		deletedRuns++
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx_commit_failed: %w", err)
	}

	// 2) Delete artifact files after DB commit.
	for _, r := range runs {
		if absArtifactsDir != "" {
			runDir := filepath.Join(absArtifactsDir, "simulation", r.ID.String())
			ok, err := isPathUnderDir(runDir, absArtifactsDir)
			if err == nil && ok {
				_ = os.RemoveAll(runDir)
			}
		}
	}

	slog.Info("retention_complete", "deleted_runs", deletedRuns)
	return nil
}

func selectSimulationRunsToDelete(ctx context.Context, q querier, cutoff time.Time, limit int) ([]simulationRunToDelete, error) {
	rows, err := q.Query(ctx, `
		SELECT
			sr.id,
			sr.run_key,
			COALESCE(sr.calcutta_evaluation_run_id, '00000000-0000-0000-0000-000000000000'::uuid) AS calcutta_evaluation_run_id,
			COALESCE(rj.finished_at, sr.updated_at, sr.created_at) AS finished_at
		FROM derived.simulation_runs sr
		JOIN derived.run_jobs rj
			ON rj.run_kind = 'simulation'
			AND rj.run_id = sr.id
		WHERE sr.deleted_at IS NULL
			AND rj.status IN ('succeeded', 'failed')
			AND COALESCE(rj.finished_at, sr.updated_at, sr.created_at) < $1
		ORDER BY COALESCE(rj.finished_at, sr.updated_at, sr.created_at) ASC
		LIMIT $2
	`, cutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("select_runs_failed: %w", err)
	}
	defer rows.Close()

	out := make([]simulationRunToDelete, 0)
	for rows.Next() {
		var r simulationRunToDelete
		if err := rows.Scan(&r.ID, &r.RunKey, &r.CalcuttaEvaluationRun, &r.FinishedAt); err != nil {
			return nil, fmt.Errorf("scan_run_failed: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select_runs_rows_failed: %w", err)
	}
	return out, nil
}

func evalRunExclusiveToDeletionSet(ctx context.Context, q querier, evalRunID uuid.UUID, deletingRunIDs []uuid.UUID) (bool, error) {
	if evalRunID == uuid.Nil {
		return false, nil
	}
	var nOther int
	err := q.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM derived.simulation_runs
		WHERE deleted_at IS NULL
			AND calcutta_evaluation_run_id = $1::uuid
			AND NOT (id = ANY($2::uuid[]))
	`, evalRunID, deletingRunIDs).Scan(&nOther)
	if err != nil {
		return false, fmt.Errorf("eval_run_refcount_failed: %w", err)
	}
	return nOther == 0, nil
}

func isPathUnderDir(path string, dir string) (bool, error) {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(dir) == "" {
		return false, errors.New("path and dir are required")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false, err
	}
	absPath = filepath.Clean(absPath)
	absDir = filepath.Clean(absDir)

	rel, err := filepath.Rel(absDir, absPath)
	if err != nil {
		return false, err
	}
	if rel == "." {
		return true, nil
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return false, nil
	}
	return true, nil
}
