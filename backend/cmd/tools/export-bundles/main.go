package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles/exporter"
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
	outDir := flag.String("out", "./exports/bundles", "output directory")
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

	generatedAt := time.Now().UTC()
	if err := exporter.ExportToDir(ctx, pool, *outDir, generatedAt); err != nil {
		return fmt.Errorf("export bundles failed: %w", err)
	}
	return nil
}
