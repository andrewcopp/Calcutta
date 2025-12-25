package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles/exporter"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
)

func main() {
	outDir := flag.String("out", "./exports/bundles", "output directory")
	flag.Parse()

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := platform.OpenPGXPool(ctx, cfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	generatedAt := time.Now().UTC()
	if err := exporter.ExportToDir(ctx, pool, *outDir, generatedAt); err != nil {
		log.Fatal(err)
	}
}
