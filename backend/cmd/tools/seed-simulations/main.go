package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"

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

	// Generate tournament simulations for each season
	// Note: Game outcome predictions must already exist (created via lab pipeline).
	simSvc := sim.New(pool)
	var seededCount int
	for _, season := range seasons {
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
		seededCount++
	}

	log.Printf("Done! %d of %d seasons seeded with simulations.", seededCount, len(seasons))
	return nil
}

func getSeasons(ctx context.Context, pool *pgxpool.Pool) ([]int, error) {
	rows, err := pool.Query(ctx, `
		SELECT s.year
		FROM core.tournaments t
		JOIN core.seasons s ON s.id = t.season_id
		WHERE t.deleted_at IS NULL AND s.deleted_at IS NULL
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
