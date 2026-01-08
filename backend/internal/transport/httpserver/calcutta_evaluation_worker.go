package httpserver

import (
	"context"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/workers"
)

const (
	defaultCalcuttaEvalWorkerPollInterval = 2 * time.Second
	defaultCalcuttaEvalWorkerStaleAfter   = 30 * time.Minute
)

func (s *Server) RunCalcuttaEvaluationWorker(ctx context.Context) {
	s.RunCalcuttaEvaluationWorkerWithOptions(ctx, defaultCalcuttaEvalWorkerPollInterval, defaultCalcuttaEvalWorkerStaleAfter)
}

func (s *Server) RunCalcuttaEvaluationWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	w := workers.NewCalcuttaEvaluationWorker(s.pool, workers.NewDBProgressWriter(s.pool))
	w.RunWithOptions(ctx, pollInterval, staleAfter)
}
