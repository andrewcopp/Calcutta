package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"

	sim "github.com/andrewcopp/Calcutta/backend/internal/app/simulate_tournaments"
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
	var workers int
	var batchSize int
	var probabilitySourceKey string

	flag.IntVar(&season, "season", 0, "Tournament season/year (e.g. 2025)")
	flag.IntVar(&nSims, "n-sims", 10000, "Number of Monte Carlo simulations")
	flag.IntVar(&seed, "seed", 42, "Base RNG seed")
	flag.IntVar(&workers, "workers", runtime.GOMAXPROCS(0), "Number of simulation workers")
	flag.IntVar(&batchSize, "batch-size", 1000, "Number of simulations per DB COPY batch")
	flag.StringVar(&probabilitySourceKey, "probability-source-key", "kenpom-v1-go", "Stored in derived.simulated_tournaments.probability_source_key")
	flag.Parse()

	if season <= 0 {
		flag.Usage()
		return fmt.Errorf("--season is required")
	}
	if nSims <= 0 {
		flag.Usage()
		return fmt.Errorf("--n-sims must be positive")
	}
	if batchSize <= 0 {
		flag.Usage()
		return fmt.Errorf("--batch-size must be positive")
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

	svc := sim.New(pool)
	res, err := svc.Run(
		context.Background(),
		sim.RunParams{
			Season:               season,
			NSims:                nSims,
			Seed:                 seed,
			Workers:              workers,
			BatchSize:            batchSize,
			ProbabilitySourceKey: probabilitySourceKey,
		},
	)
	if err != nil {
		return fmt.Errorf("simulate tournaments failed: %w", err)
	}

	log.Printf(
		"Simulated tournaments: season=%d core_tournament_id=%s batch_id=%s n_sims=%d rows=%d",
		season,
		res.CoreTournamentID,
		res.TournamentSimulationBatchID,
		res.NSims,
		res.RowsWritten,
	)
	log.Printf(
		"Timing: load=%s simulate+write=%s overall=%s",
		res.LoadDuration,
		res.SimulateWriteDuration,
		res.OverallDuration,
	)

	return nil
}
