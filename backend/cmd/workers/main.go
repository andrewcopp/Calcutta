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

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver"
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
	runEntryEvaluationWorker := flag.Bool("entry-eval-worker", true, "Run the entry evaluation worker")
	flag.Parse()

	if !*runBundleImportWorker && !*runEntryEvaluationWorker {
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

	s, err := httpserver.NewServer(pool, cfg)
	if err != nil {
		return fmt.Errorf("server_init_failed: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup

	if *runBundleImportWorker {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.RunBundleImportWorker(ctx)
		}()
	}
	if *runEntryEvaluationWorker {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.RunEntryEvaluationWorker(ctx)
		}()
	}

	<-ctx.Done()
	wg.Wait()

	slog.Info("workers_stopping")
	return nil
}
