package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/google/uuid"
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

	runKeyText := strings.TrimSpace(runKey)
	runKeyUUID := ""
	if runKeyText != "" {
		if _, err := uuid.Parse(runKeyText); err == nil {
			runKeyUUID = runKeyText
		} else {
			runKeyUUID = uuid.NewString()
		}
	} else {
		runKeyUUID = uuid.NewString()
		runKeyText = runKeyUUID
	}

	if name == "" {
		name = optimizerKey
	}
	if optimizerKey == "" {
		optimizerKey = "minlp_v1"
	}

	params := map[string]any{
		"budget_points":  budgetPoints,
		"min_teams":      minTeams,
		"max_teams":      maxTeams,
		"min_bid_points": minBid,
		"max_bid_points": maxBid,
		"source":         "generate-recommended-entry-bids",
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}

	gitSHA := strings.TrimSpace(os.Getenv("GIT_SHA"))
	var gitSHAParam any
	if gitSHA != "" {
		gitSHAParam = gitSHA
	} else {
		gitSHAParam = nil
	}

	marketShareRunID := (any)(nil)
	if v, ok := params["market_share_run_id"]; ok {
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				marketShareRunID = s
			}
		}
	}

	var runID string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO derived.strategy_generation_runs (
			run_key,
			run_key_uuid,
			name,
			simulated_tournament_id,
			calcutta_id,
			purpose,
			returns_model_key,
			investment_model_key,
			optimizer_key,
			market_share_run_id,
			params_json,
			git_sha
		)
		VALUES ($1, $2::uuid, $3, NULL, $4::uuid, 'go_recommended_entry_bids', 'legacy', 'predicted_market_share', $5, $6::uuid, $7::jsonb, $8)
		RETURNING id::text
	`, runKeyText, runKeyUUID, name, calcuttaID, optimizerKey, marketShareRunID, string(paramsJSON), gitSHAParam).Scan(&runID); err != nil {
		return err
	}

	log.Printf("Enqueued strategy_generation_run_id=%s run_key=%s", runID, runKeyText)
	return nil
}
