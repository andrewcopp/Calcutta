package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles/importer"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
)

func main() {
	inDir := flag.String("in", "./exports/bundles", "input bundles directory")
	dryRun := flag.Bool("dry-run", true, "read and validate bundles; rollback DB writes")
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

	report, err := importer.ImportFromDir(ctx, pool, *inDir, importer.Options{DryRun: *dryRun})
	if err != nil {
		log.Fatal(err)
	}

	b, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(b))
}
