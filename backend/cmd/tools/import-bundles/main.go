package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app/prediction"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/importer"
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
	inDir := flag.String("in", "./exports/bundles", "input bundles directory")
	dryRun := flag.Bool("dry-run", true, "read and validate bundles; rollback DB writes")
	flag.Parse()

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := platform.OpenPGXPool(ctx, cfg, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to database (pgxpool): %w", err)
	}
	defer pool.Close()

	report, err := importer.ImportFromDir(ctx, pool, *inDir, importer.Options{DryRun: *dryRun})
	if err != nil {
		return fmt.Errorf("import bundles failed: %w", err)
	}

	b, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(b))

	if !*dryRun {
		refreshPredictions(ctx, pool, report.TournamentIDs)
	}

	return nil
}

func refreshPredictions(ctx context.Context, pool *pgxpool.Pool, tournamentIDs []string) {
	predSvc := prediction.New(dbadapters.NewPredictionRepository(pool))
	for _, tid := range tournamentIDs {
		results, err := predSvc.RunAllCheckpoints(ctx, prediction.RunParams{
			TournamentID:         tid,
			ProbabilitySourceKey: "kenpom",
		})
		if err != nil {
			slog.Warn("prediction_refresh_failed", "tournament_id", tid, "error", err)
			continue
		}
		for _, result := range results {
			slog.Info("prediction_refresh_succeeded",
				"tournament_id", tid, "batch_id", result.BatchID,
				"team_count", result.TeamCount, "duration_ms", result.Duration.Milliseconds())
		}
	}
}
