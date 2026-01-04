package main

import (
	"context"
	"flag"
	"log"

	reb "github.com/andrewcopp/Calcutta/backend/internal/features/recommended_entry_bids"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
)

func main() {
	var calcuttaID string
	var runKey string
	var name string
	var optimizerKey string
	var budgetPoints int
	var minTeams int
	var maxTeams int
	var minBid int
	var maxBid int

	flag.StringVar(&calcuttaID, "calcutta-id", "", "Core calcutta UUID")
	flag.StringVar(&runKey, "run-key", "", "Optional run_key (defaults to random UUID)")
	flag.StringVar(&name, "name", "", "Optional human-readable run name")
	flag.StringVar(&optimizerKey, "optimizer", "minlp_v1", "Optimizer key")
	flag.IntVar(&budgetPoints, "budget", 0, "Budget points (default: calcutta budget_points)")
	flag.IntVar(&minTeams, "min-teams", 0, "Min teams (default: calcutta min_teams)")
	flag.IntVar(&maxTeams, "max-teams", 0, "Max teams (default: calcutta max_teams)")
	flag.IntVar(&minBid, "min-bid", 1, "Min bid points")
	flag.IntVar(&maxBid, "max-bid", 0, "Max bid points (default: calcutta max_bid)")
	flag.Parse()

	if calcuttaID == "" {
		log.Fatal("--calcutta-id is required")
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

	svc := reb.New(pool)
	res, err := svc.GenerateAndWrite(context.Background(), reb.GenerateParams{
		CalcuttaID:   calcuttaID,
		RunKey:       runKey,
		Name:         name,
		OptimizerKey: optimizerKey,
		BudgetPoints: budgetPoints,
		MinTeams:     minTeams,
		MaxTeams:     maxTeams,
		MinBidPoints: minBid,
		MaxBidPoints: maxBid,
	})
	if err != nil {
		log.Fatalf("GenerateAndWrite failed: %v", err)
	}

	log.Printf("Generated strategy_generation_run_id=%s run_key=%s n_teams=%d total_bid=%d simulated_tournament_id=%s",
		res.StrategyGenerationRunID,
		res.RunKey,
		res.NTeams,
		res.TotalBidPoints,
		res.SimulatedTournamentID,
	)
}
