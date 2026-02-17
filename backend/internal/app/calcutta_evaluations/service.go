package calcutta_evaluations

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service handles simulated calcutta analysis
type Service struct {
	pool *pgxpool.Pool
}

// New creates a new simulated calcutta service
func New(pool *pgxpool.Pool) *Service {
	return &Service{
		pool: pool,
	}
}
