package workers

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func nullUUIDParam(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func nullUUIDParamPtr(v *string) any {
	if v == nil {
		return nil
	}
	if *v == "" {
		return nil
	}
	return *v
}

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
