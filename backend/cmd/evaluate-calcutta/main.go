package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/andrewcopp/Calcutta/backend/internal/features/simulated_calcutta"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
)

func main() {
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
		log.Fatal("--calcutta-id is required")
	}
	if excludedEntryName == "" {
		excludedEntryName = os.Getenv("EXCLUDED_ENTRY_NAME")
	}
	if runID == "" {
		runID = "go_eval"
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

	var override *string
	if tournamentSimulationBatchID != "" {
		override = &tournamentSimulationBatchID
	}

	svc := simulated_calcutta.New(pool)
	if err := svc.CalculateSimulatedCalcuttaForEvaluationRun(context.Background(), calcuttaID, runID, excludedEntryName, override); err != nil {
		log.Fatalf("evaluation failed: %v", err)
	}

	log.Printf("Calcutta evaluation complete for calcutta_id=%s", calcuttaID)
}

func ctx() context.Context { return context.Background() }
