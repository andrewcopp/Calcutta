package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
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

type pythonRunnerResult struct {
	OK           bool    `json:"ok"`
	RunID        *string `json:"run_id"`
	RowsInserted *int    `json:"rows_inserted"`
	Error        *string `json:"error"`
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
	var pythonBin string
	var pythonRunnerPath string
	var excludedEntryName string
	var algorithmName string
	var skipExisting bool

	flag.StringVar(&calcuttaID, "calcutta-id", "", "Optional core.calcuttas.id (uuid). If empty, process all calcuttas")
	flag.IntVar(&seasonMin, "season-min", 0, "Optional season year lower bound (inclusive)")
	flag.IntVar(&seasonMax, "season-max", 0, "Optional season year upper bound (inclusive)")
	flag.StringVar(&pythonBin, "python-bin", "python3", "Python interpreter to run the market-share runner")
	flag.StringVar(&pythonRunnerPath, "python-market-runner", "", "Path to data-science/scripts/run_market_share_runner.py")
	flag.StringVar(&excludedEntryName, "excluded-entry-name", "", "Entry name to exclude from training (defaults to EXCLUDED_ENTRY_NAME env)")
	flag.StringVar(&algorithmName, "algorithm-name", "ridge", "Market share algorithm name (stored on derived.algorithms kind=market_share)")
	flag.BoolVar(&skipExisting, "skip-existing", true, "Skip calcuttas that already have a matching market_share_run (same algorithm + excluded_entry_name)")
	flag.Parse()

	if pythonRunnerPath == "" {
		return fmt.Errorf("--python-market-runner is required")
	}
	if seasonMin > 0 && seasonMax > 0 && seasonMax < seasonMin {
		return fmt.Errorf("invalid season range: season-max < season-min")
	}
	if excludedEntryName == "" {
		excludedEntryName = strings.TrimSpace(os.Getenv("EXCLUDED_ENTRY_NAME"))
	}
	if excludedEntryName == "" {
		return fmt.Errorf("excluded entry name is required (set --excluded-entry-name or EXCLUDED_ENTRY_NAME)")
	}

	absRunner, err := filepath.Abs(pythonRunnerPath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(absRunner); err != nil {
		return fmt.Errorf("python runner not found: %s", absRunner)
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

		runID, inserted, err := runPythonMarketShare(
			ctx,
			pythonBin,
			absRunner,
			c.CalcuttaID,
			algorithmName,
			excludedEntryName,
		)
		if err != nil {
			return err
		}

		slog.Info(
			"calcutta_done",
			"calcutta_id", c.CalcuttaID,
			"season", c.SeasonYear,
			"market_share_run_id", runID,
			"rows", inserted,
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

func runPythonMarketShare(ctx context.Context, pythonBin, runnerPath, calcuttaID, algorithmName, excludedEntryName string) (string, int, error) {
	cmd := exec.CommandContext(
		ctx,
		pythonBin,
		runnerPath,
		"--calcutta-id",
		calcuttaID,
		"--excluded-entry-name",
		excludedEntryName,
		"--algorithm-name",
		algorithmName,
	)
	cmd.Env = os.Environ()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outStr := strings.TrimSpace(stdout.String())
	if outStr == "" {
		outStr = "{}"
	}

	var parsed pythonRunnerResult
	_ = json.Unmarshal([]byte(outStr), &parsed)

	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if parsed.Error != nil && strings.TrimSpace(*parsed.Error) != "" {
			msg = *parsed.Error
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", 0, fmt.Errorf("python runner failed: %s", msg)
	}
	if !parsed.OK {
		msg := "python runner returned ok=false"
		if parsed.Error != nil && strings.TrimSpace(*parsed.Error) != "" {
			msg = *parsed.Error
		}
		return "", 0, errors.New(msg)
	}
	if parsed.RunID == nil || strings.TrimSpace(*parsed.RunID) == "" {
		return "", 0, errors.New("python runner did not return run_id")
	}
	inserted := 0
	if parsed.RowsInserted != nil {
		inserted = *parsed.RowsInserted
	}
	return strings.TrimSpace(*parsed.RunID), inserted, nil
}
