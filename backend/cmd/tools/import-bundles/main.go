package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles/importer"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
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
	return nil
}
