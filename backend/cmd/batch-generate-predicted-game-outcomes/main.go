package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	pgo "github.com/andrewcopp/Calcutta/backend/internal/features/predicted_game_outcomes"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	platform.InitLogger()
	if err := run(); err != nil {
		slog.Error("cmd_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var seasonMin int
	var seasonMax int
	var nSims int
	var seed int
	var kenpomScale float64
	var modelVersion string
	var skipExisting bool

	flag.IntVar(&seasonMin, "season-min", 0, "Season year lower bound (inclusive)")
	flag.IntVar(&seasonMax, "season-max", 0, "Season year upper bound (inclusive)")
	flag.IntVar(&nSims, "n-sims", 5000, "Number of Monte Carlo simulations used to estimate matchup probabilities")
	flag.IntVar(&seed, "seed", 42, "Base RNG seed")
	flag.Float64Var(&kenpomScale, "kenpom-scale", 10.0, "KenPom scale parameter")
	flag.StringVar(&modelVersion, "model-version", "kenpom-v1-go", "Algorithm name stored on derived.algorithms(kind=game_outcomes)")
	flag.BoolVar(&skipExisting, "skip-existing", true, "Skip seasons that already have a game_outcome_run for this algorithm")
	flag.Parse()

	if seasonMin <= 0 || seasonMax <= 0 {
		flag.Usage()
		return fmt.Errorf("--season-min and --season-max are required")
	}
	if seasonMax < seasonMin {
		return fmt.Errorf("invalid season range: season-max < season-min")
	}
	if nSims <= 0 {
		return fmt.Errorf("--n-sims must be positive")
	}
	if kenpomScale <= 0 {
		return fmt.Errorf("--kenpom-scale must be positive")
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

	pgoSvc := pgo.New(pool)

	algID, _ := resolveAlgorithmID(ctx, pool, modelVersion)

	for season := seasonMin; season <= seasonMax; season++ {
		tid, err := resolveCoreTournamentID(ctx, pool, season)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				slog.Info("season_skipped_no_tournament", "season", season)
				continue
			}
			return err
		}

		if skipExisting && algID != "" {
			ok, err := hasGameOutcomeRun(ctx, pool, tid, algID)
			if err != nil {
				return err
			}
			if ok {
				slog.Info("season_skipped_existing", "season", season, "tournament_id", tid)
				continue
			}
		}

		slog.Info("season_generating", "season", season, "tournament_id", tid, "model_version", modelVersion)
		_, nRows, err := pgoSvc.GenerateAndWrite(ctx, pgo.GenerateParams{
			Season:       season,
			KenPomScale:  kenpomScale,
			NSims:        nSims,
			Seed:         seed,
			ModelVersion: modelVersion,
		})
		if err != nil {
			return fmt.Errorf("GenerateAndWrite failed for season=%d: %w", season, err)
		}
		slog.Info("season_done", "season", season, "rows", nRows)

		if algID == "" {
			algID, _ = resolveAlgorithmID(ctx, pool, modelVersion)
		}
	}

	return nil
}

func resolveCoreTournamentID(ctx context.Context, pool *pgxpool.Pool, season int) (string, error) {
	var id string
	err := pool.QueryRow(ctx, `
		SELECT t.id::text
		FROM core.tournaments t
		JOIN core.seasons s
			ON s.id = t.season_id
			AND s.deleted_at IS NULL
		WHERE s.year = $1::int
			AND t.deleted_at IS NULL
		ORDER BY t.created_at DESC
		LIMIT 1
	`, season).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func resolveAlgorithmID(ctx context.Context, pool *pgxpool.Pool, modelVersion string) (string, error) {
	var id string
	err := pool.QueryRow(ctx, `
		SELECT a.id::text
		FROM derived.algorithms a
		WHERE a.kind = 'game_outcomes'
			AND a.name = $1::text
			AND a.deleted_at IS NULL
		LIMIT 1
	`, modelVersion).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func hasGameOutcomeRun(ctx context.Context, pool *pgxpool.Pool, tournamentID string, algorithmID string) (bool, error) {
	var id string
	err := pool.QueryRow(ctx, `
		SELECT gor.id::text
		FROM derived.game_outcome_runs gor
		WHERE gor.tournament_id = $1::uuid
			AND gor.algorithm_id = $2::uuid
			AND gor.deleted_at IS NULL
		LIMIT 1
	`, tournamentID, algorithmID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	_ = id
	return true, nil
}
