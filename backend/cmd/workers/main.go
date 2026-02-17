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

	runBundleImportWorker := flag.Bool("bundle-import-worker", true, "Run the bundle import worker")
	runGameOutcomeWorker := flag.Bool("game-outcome-worker", true, "Run the predicted game outcomes worker")
	runCalcuttaEvalWorker := flag.Bool("calcutta-eval-worker", true, "Run the calcutta evaluation worker")
	runSimulationWorker := flag.Bool("simulation-worker", true, "Run the simulation worker")
	runLabPipelineWorker := flag.Bool("lab-pipeline-worker", true, "Run the lab pipeline worker")
	flag.Parse()

	if !*runBundleImportWorker && !*runGameOutcomeWorker && !*runCalcuttaEvalWorker && !*runSimulationWorker && !*runLabPipelineWorker {
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
	bundleWorker := workers.NewBundleImportWorker(pool)
	gameOutcomeWorker := workers.NewGameOutcomeWorker(pool, progress)
	calcuttaEvalWorker := workers.NewCalcuttaEvaluationWorker(pool, progress)
	simulationWorker := workers.NewSimulationWorker(pool, progress, cfg.ArtifactsDir)
	labPipelineWorker := workers.NewLabPipelineWorker(pool, progress)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup

	if *runBundleImportWorker {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bundleWorker.Run(ctx)
		}()
	}
	if *runGameOutcomeWorker {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gameOutcomeWorker.Run(ctx)
		}()
	}
	if *runCalcuttaEvalWorker {
		wg.Add(1)
		go func() {
			defer wg.Done()
			calcuttaEvalWorker.Run(ctx)
		}()
	}
	if *runSimulationWorker {
		wg.Add(1)
		go func() {
			defer wg.Done()
			simulationWorker.Run(ctx)
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
