package httpserver

import (
	"context"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/workers"
)

const (
	defaultSimulationWorkerPollInterval = 2 * time.Second
	defaultSimulationWorkerStaleAfter   = 30 * time.Minute
)

func (s *Server) RunSimulationWorker(ctx context.Context) {
	s.RunSimulationWorkerWithOptions(ctx, defaultSimulationWorkerPollInterval, defaultSimulationWorkerStaleAfter)
}

func (s *Server) RunSimulationWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	w := workers.NewSimulationWorker(s.pool, workers.NewDBProgressWriter(s.pool), s.cfg.ArtifactsDir)
	w.RunWithOptions(ctx, pollInterval, staleAfter)
}
