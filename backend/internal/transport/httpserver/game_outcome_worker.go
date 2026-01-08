package httpserver

import (
	"context"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/workers"
)

const (
	defaultGameOutcomeWorkerPollInterval = 2 * time.Second
	defaultGameOutcomeWorkerStaleAfter   = 30 * time.Minute
)

func (s *Server) RunGameOutcomeWorker(ctx context.Context) {
	s.RunGameOutcomeWorkerWithOptions(ctx, defaultGameOutcomeWorkerPollInterval, defaultGameOutcomeWorkerStaleAfter)
}

func (s *Server) RunGameOutcomeWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	w := workers.NewGameOutcomeWorker(s.pool, workers.NewDBProgressWriter(s.pool))
	w.RunWithOptions(ctx, pollInterval, staleAfter)
}
