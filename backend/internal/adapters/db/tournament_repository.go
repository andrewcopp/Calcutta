package db

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Team CRUD operations are in tournament_repository_teams.go

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func computeImportKeySuffix(id string) string {
	hash := md5.Sum([]byte(id))
	return fmt.Sprintf("%x", hash)[:6]
}

type TournamentRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewTournamentRepository(pool *pgxpool.Pool) *TournamentRepository {
	return &TournamentRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *TournamentRepository) GetCompetitions(ctx context.Context) ([]models.Competition, error) {
	rows, err := r.q.ListCompetitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing competitions: %w", err)
	}
	out := make([]models.Competition, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.Competition{ID: row.ID, Name: row.Name})
	}
	return out, nil
}

func (r *TournamentRepository) GetSeasons(ctx context.Context) ([]models.Season, error) {
	rows, err := r.q.ListSeasons(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing seasons: %w", err)
	}
	out := make([]models.Season, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.Season{ID: row.ID, Year: int(row.Year)})
	}
	return out, nil
}

func (r *TournamentRepository) GetAll(ctx context.Context) ([]models.Tournament, error) {
	rows, err := r.q.ListTournaments(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing tournaments: %w", err)
	}

	out := make([]models.Tournament, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.Tournament{
			ID:                   row.ID,
			Name:                 row.Name,
			Rounds:               int(row.Rounds),
			FinalFourTopLeft:     derefString(row.FinalFourTopLeft),
			FinalFourBottomLeft:  derefString(row.FinalFourBottomLeft),
			FinalFourTopRight:    derefString(row.FinalFourTopRight),
			FinalFourBottomRight: derefString(row.FinalFourBottomRight),
			StartingAt:           TimestamptzToPtrTime(row.StartingAt),
			CreatedAt:            row.CreatedAt.Time,
			UpdatedAt:            row.UpdatedAt.Time,
		})
	}
	return out, nil
}

func (r *TournamentRepository) GetByID(ctx context.Context, id string) (*models.Tournament, error) {
	row, err := r.q.GetTournamentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting tournament %s: %w", id, err)
	}
	return &models.Tournament{
		ID:                   row.ID,
		Name:                 row.Name,
		Rounds:               int(row.Rounds),
		FinalFourTopLeft:     derefString(row.FinalFourTopLeft),
		FinalFourBottomLeft:  derefString(row.FinalFourBottomLeft),
		FinalFourTopRight:    derefString(row.FinalFourTopRight),
		FinalFourBottomRight: derefString(row.FinalFourBottomRight),
		StartingAt:           TimestamptzToPtrTime(row.StartingAt),
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
	}, nil
}

func (r *TournamentRepository) Create(ctx context.Context, tournament *models.Tournament, competitionName string, year int) error {
	now := time.Now()
	tournament.CreatedAt = now
	tournament.UpdatedAt = now

	fftl := tournament.FinalFourTopLeft
	ffbl := tournament.FinalFourBottomLeft
	fftr := tournament.FinalFourTopRight
	ffbr := tournament.FinalFourBottomRight

	var startingAt pgtype.Timestamptz
	if tournament.StartingAt != nil {
		startingAt = pgtype.Timestamptz{Time: *tournament.StartingAt, Valid: true}
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	derivedName := fmt.Sprintf("%s (%d)", competitionName, year)
	importKey := slugify(derivedName) + "-" + computeImportKeySuffix(tournament.ID)
	params := sqlc.CreateCoreTournamentParams{
		ID:                   tournament.ID,
		Year:                 int32(year),
		Name:                 competitionName,
		ImportKey:            importKey,
		Rounds:               int32(tournament.Rounds),
		FinalFourTopLeft:     &fftl,
		FinalFourBottomLeft:  &ffbl,
		FinalFourTopRight:    &fftr,
		FinalFourBottomRight: &ffbr,
		StartingAt:           startingAt,
		CreatedAt:            pgtype.Timestamptz{Time: tournament.CreatedAt, Valid: true},
		UpdatedAt:            pgtype.Timestamptz{Time: tournament.UpdatedAt, Valid: true},
	}
	if err = qtx.CreateCoreTournament(ctx, params); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &apperrors.AlreadyExistsError{Resource: "tournament", Field: "competition_season"}
		}
		return fmt.Errorf("create tournament: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing tournament creation: %w", err)
	}
	return nil
}

func (r *TournamentRepository) UpdateStartingAt(ctx context.Context, tournamentID string, startingAt *time.Time) error {
	now := time.Now()

	var start pgtype.Timestamptz
	if startingAt != nil {
		start = pgtype.Timestamptz{Time: *startingAt, Valid: true}
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	params := sqlc.UpdateCoreTournamentStartingAtParams{
		StartingAt: start,
		UpdatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
		ID:         tournamentID,
	}
	affected, err := qtx.UpdateCoreTournamentStartingAt(ctx, params)
	if err != nil {
		return fmt.Errorf("updating tournament starting_at %s: %w", tournamentID, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing tournament starting_at update: %w", err)
	}
	if affected == 0 {
		return nil
	}
	return nil
}

func (r *TournamentRepository) UpdateFinalFour(ctx context.Context, tournamentID, topLeft, bottomLeft, topRight, bottomRight string) error {
	now := time.Now()

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	params := sqlc.UpdateCoreTournamentFinalFourParams{
		FinalFourTopLeft:     &topLeft,
		FinalFourBottomLeft:  &bottomLeft,
		FinalFourTopRight:    &topRight,
		FinalFourBottomRight: &bottomRight,
		UpdatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
		ID:                   tournamentID,
	}
	if _, err = qtx.UpdateCoreTournamentFinalFour(ctx, params); err != nil {
		return fmt.Errorf("updating tournament final four %s: %w", tournamentID, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing tournament final four update: %w", err)
	}
	return nil
}

