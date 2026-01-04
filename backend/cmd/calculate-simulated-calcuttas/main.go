package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/andrewcopp/Calcutta/backend/internal/features/simulated_calcutta"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {

	// Get tournament ID from command line args
	if len(os.Args) < 2 {
		log.Fatal("Usage: calculate-simulated-calcuttas <tournament_id> [run_id] [excluded_entry_id]")
	}

	tournamentID := os.Args[1]

	// Optional: run_id can be provided, otherwise we'll find the latest
	var runID string
	if len(os.Args) >= 3 {
		runID = os.Args[2]
	}

	// Optional: excluded_entry_name can be provided via command line or env var
	excludedEntryName := ""
	if len(os.Args) >= 4 {
		excludedEntryName = os.Args[3]
	} else {
		excludedEntryName = os.Getenv("EXCLUDED_ENTRY_NAME")
	}

	// Connect to database
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Get run_id if not provided
	if runID == "" {
		runID, err = getLatestRunID(pool, tournamentID)
		if err != nil {
			log.Fatalf("Failed to get latest run ID: %v", err)
		}
		log.Printf("Using latest run ID: %s", runID)
	}

	// Log excluded entry name if provided
	if excludedEntryName != "" {
		log.Printf("Excluding entry name: %s", excludedEntryName)
	}

	// Create service and run calculation
	service := simulated_calcutta.New(pool)

	log.Printf("Starting simulated calcutta calculation for tournament %s, run %s", tournamentID, runID)

	if err := service.CalculateSimulatedCalcutta(context.Background(), tournamentID, runID); err != nil {
		log.Fatalf("Failed to calculate simulated calcutta: %v", err)
	}

	log.Printf("Successfully completed simulated calcutta calculation!")
}

func getLatestRunID(pool *pgxpool.Pool, tournamentID string) (string, error) {
	query := `
		SELECT sgr.run_key
		FROM derived.tournaments bt
		JOIN core.tournaments t
			ON t.id = bt.core_tournament_id
			AND t.deleted_at IS NULL
		JOIN core.calcuttas c
			ON c.tournament_id = t.id
			AND c.deleted_at IS NULL
		JOIN derived.strategy_generation_runs sgr
			ON sgr.calcutta_id = c.id
			AND sgr.deleted_at IS NULL
			AND sgr.run_key IS NOT NULL
		WHERE bt.id = $1
			AND bt.deleted_at IS NULL
		ORDER BY sgr.created_at DESC
		LIMIT 1
	`

	var runID string
	err := pool.QueryRow(context.Background(), query, tournamentID).Scan(&runID)
	if err != nil {
		return "", err
	}

	return runID, nil
}
