package main

import (
	"context"
	"flag"
	"log"
	"runtime"

	pgo "github.com/andrewcopp/Calcutta/backend/internal/features/predicted_game_outcomes"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
)

func main() {
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
		log.Fatal("--season is required")
	}
	if nSims <= 0 {
		log.Fatal("--n-sims must be positive")
	}
	if kenpomScale <= 0 {
		log.Fatal("--kenpom-scale must be positive")
	}

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := platform.OpenPGXPool(context.Background(), cfg, nil)
	if err != nil {
		log.Fatalf("Failed to connect to database (pgxpool): %v", err)
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
	labTournamentID, nRows, err := svc.GenerateAndWrite(
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
		log.Fatalf("GenerateAndWrite failed: %v", err)
	}

	log.Printf("Wrote %d predicted_game_outcomes rows for lab_tournament_id=%s", nRows, labTournamentID)
}
