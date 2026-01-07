package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type calcuttaRow struct {
	CalcuttaID   string
	TournamentID string
	SeasonYear   int
}

func main() {
	platform.InitLogger()
	if err := run(); err != nil {
		slog.Error("cmd_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var calcuttaID string
	var seasonMin int
	var seasonMax int
	var excludedEntryName string
	var algorithmName string
	var ridgeAlpha float64
	var featureSet string
	var trainYears string
	var skipExisting bool

	flag.StringVar(&calcuttaID, "calcutta-id", "", "Optional core.calcuttas.id (uuid). If empty, process all calcuttas")
	flag.IntVar(&seasonMin, "season-min", 0, "Optional season year lower bound (inclusive)")
	flag.IntVar(&seasonMax, "season-max", 0, "Optional season year upper bound (inclusive)")
	flag.StringVar(&excludedEntryName, "excluded-entry-name", "", "Entry name to exclude from training (defaults to EXCLUDED_ENTRY_NAME env)")
	flag.StringVar(&algorithmName, "algorithm-name", "ridge", "Market share algorithm name (stored on derived.algorithms kind=market_share)")
	flag.Float64Var(&ridgeAlpha, "ridge-alpha", 1.0, "Ridge alpha (stored in derived.market_share_runs.params_json)")
	flag.StringVar(&featureSet, "feature-set", "optimal", "Feature set (stored in derived.market_share_runs.params_json)")
	flag.StringVar(&trainYears, "train-years", "", "Comma-separated training years (optional; stored in derived.market_share_runs.params_json)")
	flag.BoolVar(&skipExisting, "skip-existing", true, "Skip calcuttas that already have a matching market_share_run (same algorithm + excluded_entry_name)")
	flag.Parse()

	if seasonMin > 0 && seasonMax > 0 && seasonMax < seasonMin {
		return fmt.Errorf("invalid season range: season-max < season-min")
	}
	if excludedEntryName == "" {
		excludedEntryName = strings.TrimSpace(os.Getenv("EXCLUDED_ENTRY_NAME"))
	}
	if excludedEntryName == "" {
		return fmt.Errorf("excluded entry name is required (set --excluded-entry-name or EXCLUDED_ENTRY_NAME)")
	}

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := platform.OpenPGXPool(ctx, cfg, nil)
	if err != nil {
		return fmt.Errorf("db_connect_failed: %w", err)
	}
	defer pool.Close()

	calcuttas, err := listCalcuttas(ctx, pool, calcuttaID, seasonMin, seasonMax)
	if err != nil {
		return err
	}
	if len(calcuttas) == 0 {
		return fmt.Errorf("no calcuttas found")
	}

	slog.Info(
		"market_share_batch_start",
		"n_calcuttas", len(calcuttas),
		"algorithm", algorithmName,
		"excluded_entry_name", excludedEntryName,
	)

	for _, c := range calcuttas {
		if skipExisting {
			exists, err := hasMarketShareRun(ctx, pool, c.CalcuttaID, algorithmName, excludedEntryName)
			if err != nil {
				return err
			}
			if exists {
				slog.Info(
					"calcutta_skipped_existing",
					"calcutta_id", c.CalcuttaID,
					"season", c.SeasonYear,
				)
				continue
			}
		}

		runID, err := submitMarketShareRun(ctx, pool, c.CalcuttaID, algorithmName, excludedEntryName, ridgeAlpha, featureSet, trainYears)
		if err != nil {
			return err
		}

		slog.Info(
			"calcutta_enqueued",
			"calcutta_id", c.CalcuttaID,
			"season", c.SeasonYear,
			"market_share_run_id", runID,
		)
	}

	slog.Info("market_share_batch_done")
	return nil
}

func listCalcuttas(ctx context.Context, pool *pgxpool.Pool, calcuttaID string, seasonMin, seasonMax int) ([]calcuttaRow, error) {
	args := []any{}
	where := ""
	idx := 1

	if strings.TrimSpace(calcuttaID) != "" {
		where += fmt.Sprintf(" AND c.id = $%d::uuid ", idx)
		args = append(args, calcuttaID)
		idx++
	}

	if seasonMin > 0 || seasonMax > 0 {
		if seasonMin <= 0 {
			seasonMin = seasonMax
		}
		if seasonMax <= 0 {
			seasonMax = seasonMin
		}
		where += fmt.Sprintf(" AND seas.year BETWEEN $%d::int AND $%d::int ", idx, idx+1)
		args = append(args, seasonMin, seasonMax)
		idx += 2
	}

	rows, err := pool.Query(ctx, fmt.Sprintf(`
		SELECT
			c.id::text,
			c.tournament_id::text,
			seas.year::int
		FROM core.calcuttas c
		JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		JOIN core.seasons seas ON seas.id = t.season_id AND seas.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
		%s
		ORDER BY seas.year ASC, c.created_at ASC
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]calcuttaRow, 0)
	for rows.Next() {
		var c calcuttaRow
		if err := rows.Scan(&c.CalcuttaID, &c.TournamentID, &c.SeasonYear); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func hasMarketShareRun(ctx context.Context, pool *pgxpool.Pool, calcuttaID, algorithmName, excludedEntryName string) (bool, error) {
	var id string
	err := pool.QueryRow(ctx, `
		SELECT msr.id::text
		FROM derived.market_share_runs msr
		JOIN derived.algorithms a ON a.id = msr.algorithm_id AND a.deleted_at IS NULL
		WHERE msr.calcutta_id = $1::uuid
			AND msr.deleted_at IS NULL
			AND a.kind = 'market_share'
			AND a.name = $2::text
			AND msr.params_json->>'excluded_entry_name' = $3::text
		ORDER BY msr.created_at DESC
		LIMIT 1
	`, calcuttaID, algorithmName, excludedEntryName).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	_ = id
	return true, nil
}

func submitMarketShareRun(ctx context.Context, pool *pgxpool.Pool, calcuttaID, algorithmName, excludedEntryName string, ridgeAlpha float64, featureSet string, trainYears string) (string, error) {
	params := map[string]interface{}{
		"excluded_entry_name": excludedEntryName,
		"ridge_alpha":         ridgeAlpha,
		"feature_set":         featureSet,
		"source":              "batch-generate-market-share",
	}
	if strings.TrimSpace(trainYears) != "" {
		parts := strings.Split(trainYears, ",")
		years := make([]int, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			y, err := strconv.Atoi(p)
			if err != nil {
				return "", fmt.Errorf("invalid train year %q", p)
			}
			years = append(years, y)
		}
		if len(years) > 0 {
			params["train_years"] = years
		}
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	var algorithmID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO derived.algorithms (kind, name, params_json)
		VALUES ('market_share', $1::text, '{}'::jsonb)
		ON CONFLICT (kind, name) WHERE deleted_at IS NULL
		DO UPDATE SET updated_at = NOW()
		RETURNING id::text
	`, algorithmName).Scan(&algorithmID); err != nil {
		return "", err
	}

	gitSHA := strings.TrimSpace(os.Getenv("GIT_SHA"))
	var gitSHAParam any
	if gitSHA != "" {
		gitSHAParam = gitSHA
	} else {
		gitSHAParam = nil
	}

	var runID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO derived.market_share_runs (algorithm_id, calcutta_id, params_json, git_sha)
		VALUES ($1::uuid, $2::uuid, $3::jsonb, $4)
		RETURNING id::text
	`, algorithmID, calcuttaID, paramsJSON, gitSHAParam).Scan(&runID); err != nil {
		return "", err
	}

	return runID, nil
}
