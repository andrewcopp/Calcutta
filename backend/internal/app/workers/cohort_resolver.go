package workers

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func resolveCohortNSims(ctx context.Context, pool *pgxpool.Pool, cohortID string, fallback int) int {
	var n int
	if err := pool.QueryRow(ctx, `
		SELECT n_sims
		FROM derived.simulation_cohorts
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, cohortID).Scan(&n); err != nil {
		return fallback
	}
	if n <= 0 {
		return fallback
	}
	return n
}

func resolveCohortSeed(ctx context.Context, pool *pgxpool.Pool, cohortID string, fallback int) int {
	var seed int
	if err := pool.QueryRow(ctx, `
		SELECT seed
		FROM derived.simulation_cohorts
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, cohortID).Scan(&seed); err != nil {
		return fallback
	}
	if seed == 0 {
		return fallback
	}
	return seed
}
