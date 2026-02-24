package main

import (
	"context"
	"fmt"
	"log"
	"os"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app/prediction"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	tournamentID := ""
	if len(os.Args) > 1 {
		tournamentID = os.Args[1]
	}
	if tournamentID == "" {
		log.Fatal("Usage: refresh-predictions <tournament-id>")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	repo := dbadapters.NewPredictionRepository(pool)
	svc := prediction.New(repo)
	results, err := svc.RunAllCheckpoints(ctx, prediction.RunParams{
		TournamentID:         tournamentID,
		ProbabilitySourceKey: "kenpom",
	})
	if err != nil {
		log.Fatalf("prediction run failed: %v", err)
	}

	for _, result := range results {
		fmt.Printf("  batch=%s teams=%d duration=%s\n",
			result.BatchID, result.TeamCount, result.Duration)
	}

	fmt.Printf("Done: %d batches generated\n", len(results))
}
