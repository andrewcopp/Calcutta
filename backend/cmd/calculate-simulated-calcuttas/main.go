package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/andrewcopp/Calcutta/backend/internal/app/simulated_calcutta"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {

	// Get tournament ID from command line args
	if len(os.Args) < 2 {
		log.Fatal("Usage: calculate-simulated-calcuttas <tournament_id> [run_id]")
	}

	tournamentID := os.Args[1]

	// Optional: run_id can be provided, otherwise we'll find the latest
	var runID string
	if len(os.Args) >= 3 {
		runID = os.Args[2]
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
		SELECT gor.run_id
		FROM gold_optimization_runs gor
		JOIN bronze_calcuttas bc ON gor.calcutta_id = bc.id
		WHERE bc.tournament_id = $1
		ORDER BY gor.created_at DESC
		LIMIT 1
	`

	var runID string
	err := pool.QueryRow(context.Background(), query, tournamentID).Scan(&runID)
	if err != nil {
		return "", err
	}

	return runID, nil
}
