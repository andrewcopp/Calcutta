package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
)

func main() {
	platform.InitLogger()
	if err := run(); err != nil {
		slog.Error("cmd_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var season int
	var nSims int
	var seed int
	var kenpomScale float64
	var modelVersion string

	flag.IntVar(&season, "season", 0, "Tournament season/year (e.g. 2025)")
	flag.IntVar(&nSims, "n-sims", 5000, "Number of Monte Carlo simulations used to estimate matchup probabilities")
	flag.IntVar(&seed, "seed", 42, "Base RNG seed")
	flag.Float64Var(&kenpomScale, "kenpom-scale", 10.0, "KenPom scale parameter")
	flag.StringVar(&modelVersion, "model-version", "kenpom-v1-go", "Model version stored on derived.predicted_game_outcomes")
	flag.Parse()

	if season <= 0 {
		flag.Usage()
		return fmt.Errorf("--season is required")
	}
	if nSims <= 0 {
		flag.Usage()
		return fmt.Errorf("--n-sims must be positive")
	}
	if kenpomScale <= 0 {
		flag.Usage()
		return fmt.Errorf("--kenpom-scale must be positive")
	}

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return err
	}

	pool, err := platform.OpenPGXPool(context.Background(), cfg, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to database (pgxpool): %w", err)
	}
	defer pool.Close()

	log.Printf("Enqueuing predicted_game_outcomes season=%d n_sims=%d seed=%d kenpom_scale=%.3f",
		season,
		nSims,
		seed,
		kenpomScale,
	)

	var coreTournamentID string
	if err := pool.QueryRow(context.Background(), `
		SELECT t.id::text
		FROM core.tournaments t
		JOIN core.seasons s ON s.id = t.season_id AND s.deleted_at IS NULL
		WHERE s.year = $1::int
			AND t.deleted_at IS NULL
		ORDER BY t.created_at DESC
		LIMIT 1
	`, season).Scan(&coreTournamentID); err != nil {
		return err
	}

	params := map[string]any{
		"season":        season,
		"kenpom_scale":  kenpomScale,
		"n_sims":        nSims,
		"seed":          seed,
		"model_version": modelVersion,
		"source":        "generate-predicted-game-outcomes",
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}

	algorithmName := modelVersion
	if algorithmName == "" {
		algorithmName = "kenpom"
	}

	var algorithmID string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO derived.algorithms (kind, name, params_json)
		VALUES ('game_outcomes', $1, $2::jsonb)
		ON CONFLICT (kind, name) WHERE deleted_at IS NULL
		DO UPDATE SET
			params_json = EXCLUDED.params_json,
			updated_at = NOW()
		RETURNING id::text
	`, algorithmName, string(paramsJSON)).Scan(&algorithmID); err != nil {
		return err
	}

	gitSHA := os.Getenv("GIT_SHA")
	var gitSHAParam any
	if gitSHA != "" {
		gitSHAParam = gitSHA
	} else {
		gitSHAParam = nil
	}

	var runID string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO derived.game_outcome_runs (algorithm_id, tournament_id, params_json, git_sha)
		VALUES ($1::uuid, $2::uuid, $3::jsonb, $4)
		RETURNING id::text
	`, algorithmID, coreTournamentID, string(paramsJSON), gitSHAParam).Scan(&runID); err != nil {
		return err
	}

	log.Printf("Enqueued game_outcome_run_id=%s tournament_id=%s", runID, coreTournamentID)
	return nil
}
