package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles/verifier"
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

	report, err := verifier.VerifyDirAgainstDB(ctx, pool, *inDir)
	if err != nil {
		return fmt.Errorf("verify bundles failed: %w", err)
	}
	if !report.OK {
		for _, m := range report.Mismatches {
			log.Printf("%s: %s", m.Where, m.What)
		}
		return fmt.Errorf("verify failed: %d mismatches", report.MismatchCount)
	}

	return nil
}
