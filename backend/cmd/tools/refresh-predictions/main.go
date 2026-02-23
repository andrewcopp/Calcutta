package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	svc := prediction.New(pool)

	currentRound, err := svc.DetectThroughRound(ctx, tournamentID)
	if err != nil {
		log.Fatalf("failed to detect tournament checkpoint: %v", err)
	}
	fmt.Printf("Tournament checkpoint: through_round=%d\n", currentRound)

	for checkpoint := 0; checkpoint <= currentRound; checkpoint++ {
		cp := checkpoint
		result, err := svc.Run(ctx, prediction.RunParams{
			TournamentID:         tournamentID,
			ProbabilitySourceKey: "kenpom",
			ThroughRound:         &cp,
		})
		if err != nil {
			log.Fatalf("prediction run failed for checkpoint %d: %v", checkpoint, err)
		}
		fmt.Printf("  checkpoint=%d batch=%s teams=%d duration=%s\n",
			checkpoint, result.BatchID, result.TeamCount, result.Duration)
	}

	fmt.Printf("Done: %d batches generated (rounds 0-%d)\n", currentRound+1, currentRound)
}
