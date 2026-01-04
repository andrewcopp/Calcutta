package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver"
)

func main() {
	log.Printf("starting workers")

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	env := os.Getenv("NODE_ENV")
	if env == "" {
		env = "development"
	}
	if env != "development" && cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	pool, err := platform.OpenPGXPool(context.Background(), cfg, nil)
	if err != nil {
		log.Fatalf("Failed to connect to database (pgxpool): %v", err)
	}
	defer pool.Close()

	s := httpserver.NewServer(pool, cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s.RunBundleImportWorker(ctx)
	s.RunEntryEvaluationWorker(ctx)

	log.Printf("stopping workers")
}
