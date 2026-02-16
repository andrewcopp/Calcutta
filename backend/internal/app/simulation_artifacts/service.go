package simulation_artifacts

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service handles simulation artifact export and batch status updates.
type Service struct {
	pool         *pgxpool.Pool
	artifactsDir string
}

// New creates a new Service instance.
func New(pool *pgxpool.Pool, artifactsDir string) *Service {
	return &Service{
		pool:         pool,
		artifactsDir: strings.TrimSpace(artifactsDir),
	}
}

// ExportArtifacts exports simulation artifacts (entry performance and simulation outcomes)
// to JSONL files and records them in run_artifacts.
func (s *Service) ExportArtifacts(ctx context.Context, simulationRunID, runKey, calcuttaEvaluationRunID string) error {
	return s.exportArtifacts(ctx, simulationRunID, runKey, calcuttaEvaluationRunID)
}
