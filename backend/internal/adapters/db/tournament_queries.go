package db

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TournamentQueryRepository implements ports.TournamentResolver.
type TournamentQueryRepository struct {
	pool *pgxpool.Pool
}

func NewTournamentQueryRepository(pool *pgxpool.Pool) *TournamentQueryRepository {
	return &TournamentQueryRepository{pool: pool}
}

func (r *TournamentQueryRepository) ResolveCoreTournamentID(ctx context.Context, season int) (string, error) {
	return ResolveCoreTournamentID(ctx, r.pool, season)
}

func (r *TournamentQueryRepository) ResolveSeasonFromTournamentID(ctx context.Context, tournamentID string) (int, error) {
	return ResolveSeasonFromTournamentID(ctx, r.pool, tournamentID)
}

func (r *TournamentQueryRepository) LoadFinalFourConfig(ctx context.Context, coreTournamentID string) (*models.FinalFourConfig, error) {
	return LoadFinalFourConfig(ctx, r.pool, coreTournamentID)
}

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

// ResolveSeasonFromTournamentID finds the season year for a given tournament ID.
func ResolveSeasonFromTournamentID(ctx context.Context, pool *pgxpool.Pool, tournamentID string) (int, error) {
	var year int
	if err := pool.QueryRow(ctx, `
		SELECT s.year
		FROM core.tournaments t
		JOIN core.seasons s
			ON s.id = t.season_id
			AND s.deleted_at IS NULL
		WHERE t.id = $1::uuid
			AND t.deleted_at IS NULL
		LIMIT 1
	`, tournamentID).Scan(&year); err != nil {
		return 0, err
	}
	return year, nil
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

	if err := cfg.ApplyDefaults(); err != nil {
		return nil, err
	}

	return cfg, nil
}
