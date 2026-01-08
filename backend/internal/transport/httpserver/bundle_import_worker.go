package httpserver

import (
	"context"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/workers"
)

const (
	defaultBundleWorkerPollInterval = 2 * time.Second
	defaultBundleWorkerStaleAfter   = 30 * time.Minute
)

func (s *Server) RunBundleImportWorker(ctx context.Context) {
	s.RunBundleImportWorkerWithOptions(ctx, defaultBundleWorkerPollInterval, defaultBundleWorkerStaleAfter)
}

func (s *Server) RunBundleImportWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	w := workers.NewBundleImportWorker(s.pool)
	w.RunWithOptions(ctx, pollInterval, staleAfter)
}
