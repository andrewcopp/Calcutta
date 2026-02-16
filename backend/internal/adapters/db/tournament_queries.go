package db

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ResolveCoreTournamentID finds the core tournament ID for a given season year.
func ResolveCoreTournamentID(ctx context.Context, pool *pgxpool.Pool, season int) (string, error) {
	var id string
	if err := pool.QueryRow(ctx, `
		SELECT t.id
		FROM core.tournaments t
		JOIN core.seasons s
			ON s.id = t.season_id
			AND s.deleted_at IS NULL
		WHERE s.year = $1::int
			AND t.deleted_at IS NULL
		ORDER BY t.created_at DESC
		LIMIT 1
	`, season).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

// LoadFinalFourConfig loads the Final Four region configuration for a tournament.
func LoadFinalFourConfig(ctx context.Context, pool *pgxpool.Pool, coreTournamentID string) (*models.FinalFourConfig, error) {
	var tl, bl, tr, br *string
	err := pool.QueryRow(ctx, `
		SELECT
			final_four_top_left,
			final_four_bottom_left,
			final_four_top_right,
			final_four_bottom_right
		FROM core.tournaments
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, coreTournamentID).Scan(&tl, &bl, &tr, &br)
	if err != nil {
		return nil, err
	}

	cfg := &models.FinalFourConfig{}
	if tl != nil {
		cfg.TopLeftRegion = *tl
	}
	if bl != nil {
		cfg.BottomLeftRegion = *bl
	}
	if tr != nil {
		cfg.TopRightRegion = *tr
	}
	if br != nil {
		cfg.BottomRightRegion = *br
	}

	if cfg.TopLeftRegion == "" {
		cfg.TopLeftRegion = "East"
	}
	if cfg.BottomLeftRegion == "" {
		cfg.BottomLeftRegion = "West"
	}
	if cfg.TopRightRegion == "" {
		cfg.TopRightRegion = "South"
	}
	if cfg.BottomRightRegion == "" {
		cfg.BottomRightRegion = "Midwest"
	}

	return cfg, nil
}
