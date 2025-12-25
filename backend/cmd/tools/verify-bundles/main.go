package main

import (
	"context"
	"flag"
	"log"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles/verifier"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
)

func main() {
	inDir := flag.String("in", "./exports/bundles", "input bundles directory")
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

	report, err := verifier.VerifyDirAgainstDB(ctx, pool, *inDir)
	if err != nil {
		log.Fatal(err)
	}
	if !report.OK {
		for _, m := range report.Mismatches {
			log.Printf("%s: %s", m.Where, m.What)
		}
		log.Fatalf("verify failed: %d mismatches", report.MismatchCount)
	}
}
