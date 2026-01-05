package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"

	pgo "github.com/andrewcopp/Calcutta/backend/internal/features/predicted_game_outcomes"
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

	log.Printf("Generating predicted_game_outcomes season=%d n_sims=%d seed=%d kenpom_scale=%.3f gomaxprocs=%d",
		season,
		nSims,
		seed,
		kenpomScale,
		runtime.GOMAXPROCS(0),
	)

	svc := pgo.New(pool)
	coreTournamentID, nRows, err := svc.GenerateAndWrite(
		context.Background(),
		pgo.GenerateParams{
			Season:       season,
			KenPomScale:  kenpomScale,
			NSims:        nSims,
			Seed:         seed,
			ModelVersion: modelVersion,
		},
	)
	if err != nil {
		return fmt.Errorf("GenerateAndWrite failed: %w", err)
	}

	log.Printf("Wrote %d predicted_game_outcomes rows for core_tournament_id=%s", nRows, coreTournamentID)
	return nil
}
