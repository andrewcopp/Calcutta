package httpserver

import (
	"context"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/workers"
)

const (
	defaultStrategyGenWorkerPollInterval = 2 * time.Second
	defaultStrategyGenWorkerStaleAfter   = 30 * time.Minute
)

func (s *Server) RunStrategyGenerationWorker(ctx context.Context) {
	s.RunStrategyGenerationWorkerWithOptions(ctx, defaultStrategyGenWorkerPollInterval, defaultStrategyGenWorkerStaleAfter)
}

func (s *Server) RunStrategyGenerationWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	w := workers.NewStrategyGenerationWorker(s.pool, workers.NewDBProgressWriter(s.pool))
	w.RunWithOptions(ctx, pollInterval, staleAfter)
}
