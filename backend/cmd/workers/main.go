package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/andrewcopp/Calcutta/backend/internal/app/workers"
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
	slog.Info("workers_starting")

	runTournamentImportWorker := flag.Bool("tournament-import-worker", true, "Run the tournament import worker")
	runLabPipelineWorker := flag.Bool("lab-pipeline-worker", true, "Run the lab pipeline worker")
	flag.Parse()

	if !*runTournamentImportWorker && !*runLabPipelineWorker {
		flag.Usage()
		return fmt.Errorf("no workers selected")
	}

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("config_load_failed: %w", err)
	}

	pool, err := platform.OpenPGXPool(context.Background(), cfg, nil)
	if err != nil {
		return fmt.Errorf("db_connect_failed: %w", err)
	}
	defer pool.Close()

	progress := workers.NewDBProgressWriter(pool)
	tournamentImportWorker := workers.NewTournamentImportWorker(pool)
	labPipelineWorker := workers.NewLabPipelineWorker(pool, progress, workers.LabPipelineWorkerConfig{
		PythonBin:          cfg.PythonBin,
		RunJobsMaxAttempts: cfg.RunJobsMaxAttempts,
		WorkerID:           cfg.WorkerID,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup

	if *runTournamentImportWorker {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tournamentImportWorker.Run(ctx)
		}()
	}
	if *runLabPipelineWorker {
		wg.Add(1)
		go func() {
			defer wg.Done()
			labPipelineWorker.Run(ctx)
		}()
	}

	<-ctx.Done()
	wg.Wait()

	slog.Info("workers_stopping")
	return nil
}
