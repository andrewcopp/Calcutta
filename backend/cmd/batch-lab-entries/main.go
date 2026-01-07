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

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app/analytics"
	pgo "github.com/andrewcopp/Calcutta/backend/internal/app/predicted_game_outcomes"
	reb "github.com/andrewcopp/Calcutta/backend/internal/app/recommended_entry_bids"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type suiteRow struct {
	ID                  string
	Name                string
	GameOutcomesAlgID   string
	GameOutcomesAlgName string
	MarketShareAlgID    string
	MarketShareAlgName  string
	OptimizerKey        string
	StartingStateKey    string
	ExcludedEntryName   *string
}

type calcuttaRow struct {
	CalcuttaID                   string
	TournamentID                 string
	SeasonYear                   int
	FocusStrategyGenerationRunID *string
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
	var suiteID string
	var calcuttaID string
	var seasonMin int
	var seasonMax int
	var pythonBin string
	var pythonRunnerPath string
	var excludedEntryName string
	var nSims int
	var seed int
	var kenpomScale float64
	var skipExistingFocus bool
	var useExistingMarketShare bool
	var dryRun bool

	flag.StringVar(&suiteID, "suite-id", "", "Optional derived.synthetic_calcutta_cohorts.id (uuid). If empty, process all cohorts")
	flag.StringVar(&calcuttaID, "calcutta-id", "", "Optional core.calcuttas.id (uuid). If empty, process all calcuttas in range")
	flag.IntVar(&seasonMin, "season-min", 0, "Optional season year lower bound (inclusive)")
	flag.IntVar(&seasonMax, "season-max", 0, "Optional season year upper bound (inclusive)")
	flag.StringVar(&pythonBin, "python-bin", "python3", "Python interpreter to run the market-share runner")
	flag.StringVar(&pythonRunnerPath, "python-market-runner", "", "Optional path to data-science/scripts/run_market_share_runner.py (used only as fallback if no matching market_share_run exists)")
	flag.StringVar(&excludedEntryName, "excluded-entry-name", "", "Override excluded entry name (defaults to cohort.excluded_entry_name or EXCLUDED_ENTRY_NAME)")
	flag.IntVar(&nSims, "pgo-n-sims", 5000, "predicted_game_outcomes n_sims")
	flag.IntVar(&seed, "pgo-seed", 42, "predicted_game_outcomes seed")
	flag.Float64Var(&kenpomScale, "pgo-kenpom-scale", 10.0, "predicted_game_outcomes kenpom scale")
	flag.BoolVar(&skipExistingFocus, "skip-existing-focus", true, "Skip synthetic_calcuttas that already have focus_strategy_generation_run_id")
	flag.BoolVar(&useExistingMarketShare, "use-existing-market-share", true, "Reuse existing market_share_runs (matching algorithm name + excluded_entry_name) and only call Python runner if missing")
	flag.BoolVar(&dryRun, "dry-run", false, "If set, compute selections but do not write synthetic_calcuttas focus run")
	flag.Parse()

	if seasonMin > 0 && seasonMax > 0 && seasonMax < seasonMin {
		return fmt.Errorf("invalid season range: season-max < season-min")
	}

	absRunner := ""
	if strings.TrimSpace(pythonRunnerPath) != "" {
		v, err := filepath.Abs(pythonRunnerPath)
		if err != nil {
			return err
		}
		if _, err := os.Stat(v); err != nil {
			return fmt.Errorf("python runner not found: %s", v)
		}
		absRunner = v
	}

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := platform.OpenPGXPool(ctx, cfg, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to database (pgxpool): %w", err)
	}
	defer pool.Close()

	suites, err := loadSuites(ctx, pool, suiteID)
	if err != nil {
		return err
	}
	if len(suites) == 0 {
		return fmt.Errorf("no suites found")
	}

	arepo := db.NewAnalyticsRepository(pool)
	asvc := analytics.New(arepo)
	pgoSvc := pgo.New(pool)
	rebSvc := reb.New(pool)

	for _, s := range suites {
		slog.Info("processing_suite", "suite_id", s.ID, "suite_name", s.Name)

		// Ensure suite_scenarios exist for the requested season range.
		if err := ensureSuiteScenarios(ctx, pool, s.ID, seasonMin, seasonMax, calcuttaID); err != nil {
			return err
		}

		calcuttas, err := listSuiteScenarioCalcuttas(ctx, pool, s.ID, calcuttaID)
		if err != nil {
			return err
		}
		if len(calcuttas) == 0 {
			slog.Info("no_calcuttas_for_suite", "suite_id", s.ID)
			continue
		}

		for _, c := range calcuttas {
			if skipExistingFocus && c.FocusStrategyGenerationRunID != nil && strings.TrimSpace(*c.FocusStrategyGenerationRunID) != "" {
				slog.Info(
					"scenario_skipped_existing_focus",
					"suite_id", s.ID,
					"calcutta_id", c.CalcuttaID,
					"focus_strategy_generation_run_id", *c.FocusStrategyGenerationRunID,
				)
				continue
			}

			// Ensure PGO run exists for this tournament+algorithm.
			goRunID, err := ensureGameOutcomeRun(ctx, pool, pgoSvc, s, c, nSims, seed, kenpomScale)
			if err != nil {
				return err
			}

			// Resolve excluded_entry_name for combo identity.
			excl := excludedEntryName
			if excl == "" {
				if s.ExcludedEntryName != nil {
					excl = strings.TrimSpace(*s.ExcludedEntryName)
				}
			}
			if excl == "" {
				excl = strings.TrimSpace(os.Getenv("EXCLUDED_ENTRY_NAME"))
			}
			if excl == "" {
				return fmt.Errorf("excluded entry name is required (set suite.excluded_entry_name, --excluded-entry-name, or EXCLUDED_ENTRY_NAME)")
			}

			msRunID := ""
			if useExistingMarketShare {
				id, ok, err := resolveMarketShareRunID(ctx, pool, c.CalcuttaID, s.MarketShareAlgName, excl)
				if err != nil {
					return err
				}
				if ok {
					msRunID = id
				}
			}
			if msRunID == "" {
				if absRunner == "" {
					return fmt.Errorf("no matching market_share_run found for calcutta_id=%s algorithm=%s excluded_entry_name=%s (provide --python-market-runner to compute)", c.CalcuttaID, s.MarketShareAlgName, excl)
				}
				id, err := runPythonMarketShare(ctx, pythonBin, absRunner, c.CalcuttaID, s.MarketShareAlgName, excl)
				if err != nil {
					return err
				}
				msRunID = id
			}

			// Compute expected points from analytics service (PGO DP).
			_, _, returns, err := asvc.GetCalcuttaPredictedReturns(ctx, c.CalcuttaID, nil, &goRunID)
			if err != nil {
				return err
			}
			pts := make([]reb.ExpectedTeam, 0, len(returns))
			for _, r := range returns {
				pts = append(pts, reb.ExpectedTeam{TeamID: r.TeamID, ExpectedPoints: r.ExpectedValue})
			}

			shares, err := loadMarketSharesByRunID(ctx, pool, msRunID)
			if err != nil {
				return err
			}

			res, err := rebSvc.GenerateFromPredictionsAndWrite(ctx, reb.GenerateFromPredictionsParams{
				CalcuttaID:           c.CalcuttaID,
				RunKey:               "",
				Name:                 fmt.Sprintf("lab_entries/%s", s.Name),
				OptimizerKey:         s.OptimizerKey,
				GameOutcomeRunID:     goRunID,
				MarketShareRunID:     msRunID,
				ExcludedEntryName:    excl,
				ExpectedPointsByTeam: pts,
				PredictedShareByTeam: shares,
				MinBidPoints:         1,
				MaxBidPoints:         0,
				MinTeams:             0,
				MaxTeams:             0,
				BudgetPoints:         0,
			})
			if err != nil {
				return err
			}

			slog.Info(
				"generated_entry",
				"suite_id", s.ID,
				"calcutta_id", c.CalcuttaID,
				"strategy_generation_run_id", res.StrategyGenerationRunID,
				"total_bid", res.TotalBidPoints,
			)

			if !dryRun {
				if err := attachFocusRunToSuiteScenario(ctx, pool, s.ID, c.CalcuttaID, res.StrategyGenerationRunID, s.StartingStateKey, excl); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func loadSuites(ctx context.Context, pool *pgxpool.Pool, suiteID string) ([]suiteRow, error) {
	args := []any{}
	where := ""
	if strings.TrimSpace(suiteID) != "" {
		where = " AND s.id = $1::uuid "
		args = append(args, suiteID)
	}

	rows, err := pool.Query(ctx, fmt.Sprintf(`
		SELECT
			s.id::text,
			s.name,
			s.game_outcomes_algorithm_id::text,
			goa.name,
			s.market_share_algorithm_id::text,
			msa.name,
			s.optimizer_key,
			COALESCE(NULLIF(s.starting_state_key, ''), 'post_first_four') AS starting_state_key,
			s.excluded_entry_name
		FROM derived.synthetic_calcutta_cohorts s
		JOIN derived.algorithms goa ON goa.id = s.game_outcomes_algorithm_id AND goa.deleted_at IS NULL
		JOIN derived.algorithms msa ON msa.id = s.market_share_algorithm_id AND msa.deleted_at IS NULL
		WHERE s.deleted_at IS NULL
		%s
		ORDER BY s.created_at DESC
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]suiteRow, 0)
	for rows.Next() {
		var s suiteRow
		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.GameOutcomesAlgID,
			&s.GameOutcomesAlgName,
			&s.MarketShareAlgID,
			&s.MarketShareAlgName,
			&s.OptimizerKey,
			&s.StartingStateKey,
			&s.ExcludedEntryName,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func resolveMarketShareRunID(ctx context.Context, pool *pgxpool.Pool, calcuttaID, algorithmName, excludedEntryName string) (string, bool, error) {
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
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}

func ensureSuiteScenarios(ctx context.Context, pool *pgxpool.Pool, suiteID string, seasonMin, seasonMax int, calcuttaID string) error {
	if strings.TrimSpace(calcuttaID) != "" {
		_, err := pool.Exec(ctx, `
			INSERT INTO derived.synthetic_calcuttas (cohort_id, calcutta_id, starting_state_key, excluded_entry_name)
			VALUES ($1::uuid, $2::uuid, NULL, NULL)
			ON CONFLICT (cohort_id, calcutta_id) WHERE deleted_at IS NULL
			DO NOTHING
		`, suiteID, calcuttaID)
		return err
	}

	if seasonMin <= 0 && seasonMax <= 0 {
		return nil
	}
	if seasonMin <= 0 {
		seasonMin = seasonMax
	}
	if seasonMax <= 0 {
		seasonMax = seasonMin
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO derived.synthetic_calcuttas (cohort_id, calcutta_id, starting_state_key, excluded_entry_name)
		SELECT $1::uuid, c.id, NULL, NULL
		FROM core.calcuttas c
		JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		JOIN core.seasons s ON s.id = t.season_id AND s.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
		  AND s.year BETWEEN $2::int AND $3::int
		ON CONFLICT (cohort_id, calcutta_id) WHERE deleted_at IS NULL
		DO NOTHING
	`, suiteID, seasonMin, seasonMax)
	return err
}

func listSuiteScenarioCalcuttas(ctx context.Context, pool *pgxpool.Pool, suiteID string, calcuttaID string) ([]calcuttaRow, error) {
	args := []any{suiteID}
	where := ""
	if strings.TrimSpace(calcuttaID) != "" {
		where = " AND sc.calcutta_id = $2::uuid "
		args = append(args, calcuttaID)
	}

	rows, err := pool.Query(ctx, fmt.Sprintf(`
		SELECT
			sc.calcutta_id::text,
			c.tournament_id::text,
			seas.year::int,
			sc.focus_strategy_generation_run_id::text
		FROM derived.synthetic_calcuttas sc
		JOIN core.calcuttas c ON c.id = sc.calcutta_id AND c.deleted_at IS NULL
		JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		JOIN core.seasons seas ON seas.id = t.season_id AND seas.deleted_at IS NULL
		WHERE sc.cohort_id = $1::uuid
		  AND sc.deleted_at IS NULL
		  %s
		ORDER BY seas.year ASC
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]calcuttaRow, 0)
	for rows.Next() {
		var c calcuttaRow
		if err := rows.Scan(&c.CalcuttaID, &c.TournamentID, &c.SeasonYear, &c.FocusStrategyGenerationRunID); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func ensureGameOutcomeRun(
	ctx context.Context,
	pool *pgxpool.Pool,
	svc *pgo.Service,
	s suiteRow,
	c calcuttaRow,
	nSims int,
	seed int,
	kenpomScale float64,
) (string, error) {
	var runID string
	err := pool.QueryRow(ctx, `
		SELECT gor.id::text
		FROM derived.game_outcome_runs gor
		WHERE gor.tournament_id = $1::uuid
		  AND gor.algorithm_id = $2::uuid
		  AND gor.deleted_at IS NULL
		ORDER BY gor.created_at DESC
		LIMIT 1
	`, c.TournamentID, s.GameOutcomesAlgID).Scan(&runID)
	if err == nil {
		return runID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	slog.Info("generating_pgo", "season", c.SeasonYear, "algorithm", s.GameOutcomesAlgName)
	_, _, err = svc.GenerateAndWrite(ctx, pgo.GenerateParams{
		Season:       c.SeasonYear,
		KenPomScale:  kenpomScale,
		NSims:        nSims,
		Seed:         seed,
		ModelVersion: s.GameOutcomesAlgName,
	})
	if err != nil {
		return "", err
	}

	err = pool.QueryRow(ctx, `
		SELECT gor.id::text
		FROM derived.game_outcome_runs gor
		WHERE gor.tournament_id = $1::uuid
		  AND gor.algorithm_id = $2::uuid
		  AND gor.deleted_at IS NULL
		ORDER BY gor.created_at DESC
		LIMIT 1
	`, c.TournamentID, s.GameOutcomesAlgID).Scan(&runID)
	if err != nil {
		return "", err
	}
	return runID, nil
}

func runPythonMarketShare(ctx context.Context, pythonBin, runnerPath, calcuttaID, algorithmName, excludedEntryName string) (string, error) {
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
		return "", fmt.Errorf("python runner failed: %s", msg)
	}
	if !parsed.OK {
		msg := "python runner returned ok=false"
		if parsed.Error != nil && strings.TrimSpace(*parsed.Error) != "" {
			msg = *parsed.Error
		}
		return "", errors.New(msg)
	}
	if parsed.RunID == nil || strings.TrimSpace(*parsed.RunID) == "" {
		return "", errors.New("python runner did not return run_id")
	}
	return strings.TrimSpace(*parsed.RunID), nil
}

func loadMarketSharesByRunID(ctx context.Context, pool *pgxpool.Pool, runID string) (map[string]float64, error) {
	rows, err := pool.Query(ctx, `
		SELECT team_id::text, predicted_share
		FROM derived.predicted_market_share
		WHERE run_id = $1::uuid
		  AND deleted_at IS NULL
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]float64)
	for rows.Next() {
		var teamID string
		var share float64
		if err := rows.Scan(&teamID, &share); err != nil {
			return nil, err
		}
		out[teamID] = share
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no predicted_market_share rows for run_id=%s", runID)
	}
	return out, nil
}

func attachFocusRunToSuiteScenario(ctx context.Context, pool *pgxpool.Pool, suiteID, calcuttaID, strategyRunID, startingStateKey, excludedEntryName string) error {
	_, err := pool.Exec(ctx, `
		UPDATE derived.synthetic_calcuttas
		SET focus_strategy_generation_run_id = $3::uuid,
		    starting_state_key = COALESCE(NULLIF($4::text, ''), starting_state_key),
		    excluded_entry_name = $5::text,
		    updated_at = NOW()
		WHERE cohort_id = $1::uuid
		  AND calcutta_id = $2::uuid
		  AND deleted_at IS NULL
	`, suiteID, calcuttaID, strategyRunID, startingStateKey, excludedEntryName)
	return err
}
