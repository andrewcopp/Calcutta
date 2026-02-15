package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"

	"github.com/andrewcopp/Calcutta/backend/internal/app/predicted_game_outcomes"
	sim "github.com/andrewcopp/Calcutta/backend/internal/app/simulate_tournaments"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	platform.InitLogger()
	if err := run(); err != nil {
		log.Fatalf("seed-simulations failed: %v", err)
	}
}

func run() error {
	var nSims int
	var seed int
	flag.IntVar(&nSims, "n-sims", 10000, "Number of Monte Carlo simulations per tournament")
	flag.IntVar(&seed, "seed", 42, "Base RNG seed")
	flag.Parse()

	ctx := context.Background()

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return err
	}

	pool, err := platform.OpenPGXPool(ctx, cfg, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Get all seasons with tournaments
	seasons, err := getSeasons(ctx, pool)
	if err != nil {
		return fmt.Errorf("failed to get seasons: %w", err)
	}

	log.Printf("Found %d seasons to seed", len(seasons))

	// Step 1: Generate game outcome predictions for each season
	goSvc := predicted_game_outcomes.New(pool)
	var readySeasons []int
	for _, season := range seasons {
		log.Printf("Generating game outcome predictions for season %d...", season)
		_, n, err := goSvc.GenerateAndWrite(ctx, predicted_game_outcomes.GenerateParams{
			Season:       season,
			KenPomScale:  11.0,
			NSims:        10000,
			Seed:         seed,
			ModelVersion: "kenpom-v1-sigma11-go",
		})
		if err != nil {
			log.Printf("  SKIP season %d: %v", season, err)
			continue
		}
		if n == 0 {
			log.Printf("  SKIP season %d: no KenPom stats (0 predictions generated)", season)
			continue
		}
		log.Printf("  Generated %d game outcome predictions for season %d", n, season)
		readySeasons = append(readySeasons, season)
	}

	// Step 2: Generate tournament simulations for ready seasons
	simSvc := sim.New(pool)
	for _, season := range readySeasons {
		log.Printf("Simulating tournaments for season %d (%d sims)...", season, nSims)
		res, err := simSvc.Run(ctx, sim.RunParams{
			Season:               season,
			NSims:                nSims,
			Seed:                 seed,
			Workers:              runtime.GOMAXPROCS(0),
			BatchSize:            1000,
			ProbabilitySourceKey: "kenpom-v1-go",
		})
		if err != nil {
			log.Printf("  SKIP season %d simulation: %v", season, err)
			continue
		}
		log.Printf("  Simulated season %d: tournament_id=%s batch_id=%s rows=%d (%s)",
			season, res.CoreTournamentID, res.TournamentSimulationBatchID, res.RowsWritten, res.OverallDuration)
	}

	log.Printf("Done! %d of %d seasons seeded with game outcomes and simulations.", len(readySeasons), len(seasons))
	return nil
}

func getSeasons(ctx context.Context, pool *pgxpool.Pool) ([]int, error) {
	rows, err := pool.Query(ctx, `
		SELECT s.year
		FROM core.tournaments t
		JOIN core.seasons s ON s.id = t.season_id
		ORDER BY s.year
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seasons []int
	for rows.Next() {
		var year int
		if err := rows.Scan(&year); err != nil {
			return nil, err
		}
		seasons = append(seasons, year)
	}
	return seasons, rows.Err()
}
