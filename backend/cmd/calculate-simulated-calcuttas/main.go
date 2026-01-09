package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/andrewcopp/Calcutta/backend/internal/app/calcutta_evaluations"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
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
	var calcuttaID string
	var runID string
	var excludedEntryName string

	flag.StringVar(&calcuttaID, "calcutta-id", "", "Calcutta ID (uuid)")
	flag.StringVar(&runID, "run-id", "", "Strategy generation run key (optional; defaults to latest)")
	flag.StringVar(&excludedEntryName, "excluded-entry-name", "", "Entry name to exclude (optional; defaults to EXCLUDED_ENTRY_NAME env var)")
	flag.Parse()

	// Backward-compatible positional args:
	//   calculate-simulated-calcuttas <calcutta_id> [run_id] [excluded_entry_name]
	if calcuttaID == "" && flag.NArg() >= 1 {
		calcuttaID = flag.Arg(0)
	}
	if runID == "" && flag.NArg() >= 2 {
		runID = flag.Arg(1)
	}
	if excludedEntryName == "" {
		if flag.NArg() >= 3 {
			excludedEntryName = flag.Arg(2)
		} else {
			excludedEntryName = os.Getenv("EXCLUDED_ENTRY_NAME")
		}
	}

	if calcuttaID == "" {
		flag.Usage()
		return fmt.Errorf("--calcutta-id is required")
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

	// Get run_id if not provided
	if runID == "" {
		runID, err = getLatestRunID(pool, calcuttaID)
		if err != nil {
			return fmt.Errorf("failed to get latest run ID: %w", err)
		}
		log.Printf("Using latest run ID: %s", runID)
	}

	// Log excluded entry name if provided
	if excludedEntryName != "" {
		log.Printf("Excluding entry name: %s", excludedEntryName)
	}

	// Create service and run calculation
	svc := calcutta_evaluations.New(pool)

	log.Printf("Starting simulated calcutta calculation for calcutta %s, run %s", calcuttaID, runID)
	evalRunID, err := svc.CalculateSimulatedCalcuttaForEvaluationRun(context.Background(), calcuttaID, runID, excludedEntryName, nil)
	if err != nil {
		return fmt.Errorf("failed to calculate simulated calcutta: %w", err)
	}

	log.Printf("Successfully completed simulated calcutta calculation eval_run_id=%s", evalRunID)
	return nil
}

func getLatestRunID(pool *pgxpool.Pool, calcuttaID string) (string, error) {
	query := `
		SELECT sgr.run_key
		FROM derived.strategy_generation_runs sgr
		WHERE sgr.calcutta_id = $1::uuid
			AND sgr.deleted_at IS NULL
			AND sgr.run_key IS NOT NULL
		ORDER BY sgr.created_at DESC
		LIMIT 1
	`

	var runID string
	err := pool.QueryRow(context.Background(), query, calcuttaID).Scan(&runID)
	if err != nil {
		return "", err
	}

	return runID, nil
}
