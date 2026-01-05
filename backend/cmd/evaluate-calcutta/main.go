package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/andrewcopp/Calcutta/backend/internal/features/simulated_calcutta"
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
	var calcuttaID string
	var tournamentSimulationBatchID string
	var excludedEntryName string
	var runID string

	flag.StringVar(&calcuttaID, "calcutta-id", "", "Core calcutta UUID")
	flag.StringVar(&tournamentSimulationBatchID, "tournament-simulation-batch-id", "", "Optional: derived.simulated_tournaments.id")
	flag.StringVar(&excludedEntryName, "excluded-entry-name", "", "Optional: entry name to exclude")
	flag.StringVar(&runID, "run-id", "", "Optional: run_id tag for legacy compatibility")
	flag.Parse()

	if calcuttaID == "" {
		flag.Usage()
		return fmt.Errorf("--calcutta-id is required")
	}
	if excludedEntryName == "" {
		excludedEntryName = os.Getenv("EXCLUDED_ENTRY_NAME")
	}
	if runID == "" {
		runID = "go_eval"
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

	var override *string
	if tournamentSimulationBatchID != "" {
		override = &tournamentSimulationBatchID
	}

	svc := simulated_calcutta.New(pool)
	_, err = svc.CalculateSimulatedCalcuttaForEvaluationRun(context.Background(), calcuttaID, runID, excludedEntryName, override)
	if err != nil {
		return fmt.Errorf("evaluation failed: %w", err)
	}

	log.Printf("Calcutta evaluation complete for calcutta_id=%s", calcuttaID)
	return nil
}

func ctx() context.Context { return context.Background() }
