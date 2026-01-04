package main

import (
	"context"
	"flag"
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
	slog.Info("workers_starting")

	runBundleImportWorker := flag.Bool("bundle-import-worker", true, "Run the bundle import worker")
	runEntryEvaluationWorker := flag.Bool("entry-eval-worker", true, "Run the entry evaluation worker")
	flag.Parse()

	if !*runBundleImportWorker && !*runEntryEvaluationWorker {
		slog.Error("no_workers_selected")
		os.Exit(1)
	}

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		slog.Error("config_load_failed", "error", err)
		os.Exit(1)
	}

	pool, err := platform.OpenPGXPool(context.Background(), cfg, nil)
	if err != nil {
		slog.Error("db_connect_failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	s, err := httpserver.NewServer(pool, cfg)
	if err != nil {
		slog.Error("server_init_failed", "error", err)
		os.Exit(1)
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
}
