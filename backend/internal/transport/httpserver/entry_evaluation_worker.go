package httpserver

import (
	"context"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/workers"
)

const (
	defaultEntryEvalWorkerPollInterval = 2 * time.Second
	defaultEntryEvalWorkerStaleAfter   = 30 * time.Minute
)

func (s *Server) RunEntryEvaluationWorker(ctx context.Context) {
	s.RunEntryEvaluationWorkerWithOptions(ctx, defaultEntryEvalWorkerPollInterval, defaultEntryEvalWorkerStaleAfter)
}

func (s *Server) RunEntryEvaluationWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	w := workers.NewEntryEvaluationWorker(s.pool, workers.NewDBProgressWriter(s.pool))
	w.RunWithOptions(ctx, pollInterval, staleAfter)
}
