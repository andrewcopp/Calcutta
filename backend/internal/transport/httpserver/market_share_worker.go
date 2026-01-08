package httpserver

import (
	"context"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/workers"
)

const (
	defaultMarketShareWorkerPollInterval = 2 * time.Second
	defaultMarketShareWorkerStaleAfter   = 30 * time.Minute
)

func (s *Server) RunMarketShareWorker(ctx context.Context) {
	s.RunMarketShareWorkerWithOptions(ctx, defaultMarketShareWorkerPollInterval, defaultMarketShareWorkerStaleAfter)
}

func (s *Server) RunMarketShareWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	w := workers.NewMarketShareWorker(s.pool, workers.NewDBProgressWriter(s.pool))
	w.RunWithOptions(ctx, pollInterval, staleAfter)
}
