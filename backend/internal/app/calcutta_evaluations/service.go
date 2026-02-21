package calcutta_evaluations

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TournamentResolver resolves tournament metadata without importing adapters.
// Includes all methods needed by both this service and the simulation service
// it creates internally.
type TournamentResolver interface {
	ResolveCoreTournamentID(ctx context.Context, season int) (string, error)
	ResolveSeasonFromTournamentID(ctx context.Context, tournamentID string) (int, error)
	LoadFinalFourConfig(ctx context.Context, coreTournamentID string) (*models.FinalFourConfig, error)
}

// Service handles simulated calcutta analysis
type Service struct {
	pool               *pgxpool.Pool
	tournamentResolver TournamentResolver
}

// New creates a new simulated calcutta service
func New(pool *pgxpool.Pool, opts ...Option) *Service {
	s := &Service{pool: pool}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Option configures the Service.
type Option func(*Service)

// WithTournamentResolver sets the TournamentResolver.
func WithTournamentResolver(r TournamentResolver) Option {
	return func(s *Service) { s.tournamentResolver = r }
}
