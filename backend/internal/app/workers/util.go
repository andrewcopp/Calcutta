package workers

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func resolveSeasonYearByCalcuttaID(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) (int, error) {
	var year int
	q := `
		SELECT seas.year
		FROM core.calcuttas c
		JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		JOIN core.seasons seas ON seas.id = t.season_id AND seas.deleted_at IS NULL
		WHERE c.id = $1::uuid
			AND c.deleted_at IS NULL
		LIMIT 1
	`
	if err := pool.QueryRow(ctx, q, calcuttaID).Scan(&year); err != nil {
		return 0, err
	}
	return year, nil
}

func resolveSeasonYearByTournamentID(ctx context.Context, pool *pgxpool.Pool, tournamentID string) (int, error) {
	var year int
	q := `
		SELECT seas.year
		FROM core.tournaments t
		JOIN core.seasons seas ON seas.id = t.season_id AND seas.deleted_at IS NULL
		WHERE t.id = $1::uuid
			AND t.deleted_at IS NULL
		LIMIT 1
	`
	if err := pool.QueryRow(ctx, q, tournamentID).Scan(&year); err != nil {
		return 0, err
	}
	return year, nil
}

func resolveRunJobsMaxAttempts(defaultValue int) int {
	v := strings.TrimSpace(os.Getenv("RUN_JOBS_MAX_ATTEMPTS"))
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	if n <= 0 {
		return defaultValue
	}
	return n
}
