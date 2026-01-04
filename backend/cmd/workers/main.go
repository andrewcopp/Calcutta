package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver"
)

func main() {
	platform.InitLogger()
	slog.Info("workers_starting")

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		slog.Error("config_load_failed", "error", err)
		os.Exit(1)
	}

	env := os.Getenv("NODE_ENV")
	if env == "" {
		env = "development"
	}
	if env != "development" && cfg.JWTSecret == "" {
		slog.Error("jwt_secret_missing")
		os.Exit(1)
	}

	pool, err := platform.OpenPGXPool(context.Background(), cfg, nil)
	if err != nil {
		slog.Error("db_connect_failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	s := httpserver.NewServer(pool, cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s.RunBundleImportWorker(ctx)
	s.RunEntryEvaluationWorker(ctx)

	slog.Info("workers_stopping")
}
