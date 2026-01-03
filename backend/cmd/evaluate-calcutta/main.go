package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/andrewcopp/Calcutta/backend/internal/features/simulated_calcutta"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var calcuttaID string
	var tournamentSimulationBatchID string
	var excludedEntryName string
	var runID string

	flag.StringVar(&calcuttaID, "calcutta-id", "", "Core calcutta UUID")
	flag.StringVar(&tournamentSimulationBatchID, "tournament-simulation-batch-id", "", "Optional: analytics.tournament_simulation_batches.id")
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

	bronzeTournamentID, err := resolveBronzeTournamentID(ctx(), pool, calcuttaID)
	if err != nil {
		log.Fatalf("failed to resolve lab tournament: %v", err)
	}

	var override *string
	if tournamentSimulationBatchID != "" {
		override = &tournamentSimulationBatchID
	}

	svc := simulated_calcutta.New(pool)
	if err := svc.CalculateSimulatedCalcuttaForEvaluationRun(context.Background(), bronzeTournamentID, runID, excludedEntryName, override); err != nil {
		log.Fatalf("evaluation failed: %v", err)
	}

	log.Printf("Calcutta evaluation complete for calcutta_id=%s lab_tournament_id=%s", calcuttaID, bronzeTournamentID)
}

func ctx() context.Context { return context.Background() }

func resolveBronzeTournamentID(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) (string, error) {
	// Resolve the lab_bronze tournament for the core tournament backing this calcutta.
	var bronzeTournamentID string
	err := pool.QueryRow(ctx, `
		SELECT bt.id
		FROM core.calcuttas c
		JOIN lab_bronze.tournaments bt
		  ON bt.core_tournament_id = c.tournament_id
		 AND bt.deleted_at IS NULL
		WHERE c.id = $1::uuid
		  AND c.deleted_at IS NULL
		ORDER BY bt.created_at DESC
		LIMIT 1
	`, calcuttaID).Scan(&bronzeTournamentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("no lab_bronze tournament found for calcutta_id=%s", calcuttaID)
		}
		return "", err
	}
	return bronzeTournamentID, nil
}
