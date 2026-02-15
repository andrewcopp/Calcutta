package db

import (
	"context"
	"errors"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TournamentRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewTournamentRepository(pool *pgxpool.Pool) *TournamentRepository {
	return &TournamentRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *TournamentRepository) GetAll(ctx context.Context) ([]models.Tournament, error) {
	rows, err := r.q.ListTournaments(ctx)
	if err != nil {
		return nil, err
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
			StartingAt:           timestamptzToTimePtr(row.StartingAt),
			Created:              row.CreatedAt.Time,
			Updated:              row.UpdatedAt.Time,
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
		return nil, err
	}
	return &models.Tournament{
		ID:                   row.ID,
		Name:                 row.Name,
		Rounds:               int(row.Rounds),
		FinalFourTopLeft:     derefString(row.FinalFourTopLeft),
		FinalFourBottomLeft:  derefString(row.FinalFourBottomLeft),
		FinalFourTopRight:    derefString(row.FinalFourTopRight),
		FinalFourBottomRight: derefString(row.FinalFourBottomRight),
		StartingAt:           timestamptzToTimePtr(row.StartingAt),
		Created:              row.CreatedAt.Time,
		Updated:              row.UpdatedAt.Time,
	}, nil
}

func (r *TournamentRepository) Create(ctx context.Context, tournament *models.Tournament) error {
	now := time.Now()
	tournament.Created = now
	tournament.Updated = now

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
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	params := sqlc.CreateCoreTournamentParams{
		ID:                   tournament.ID,
		Name:                 tournament.Name,
		Rounds:               int32(tournament.Rounds),
		FinalFourTopLeft:     &fftl,
		FinalFourBottomLeft:  &ffbl,
		FinalFourTopRight:    &fftr,
		FinalFourBottomRight: &ffbr,
		StartingAt:           startingAt,
		CreatedAt:            pgtype.Timestamptz{Time: tournament.Created, Valid: true},
		UpdatedAt:            pgtype.Timestamptz{Time: tournament.Updated, Valid: true},
	}
	if err = qtx.CreateCoreTournament(ctx, params); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
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
		return err
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
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	if affected == 0 {
		return nil
	}
	return nil
}

func (r *TournamentRepository) GetTeams(ctx context.Context, tournamentID string) ([]*models.TournamentTeam, error) {
	rows, err := r.q.ListTeamsByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	teams := make([]*models.TournamentTeam, 0, len(rows))
	for _, row := range rows {
		teams = append(teams, tournamentTeamFromRow(
			row.ID,
			row.TournamentID,
			row.SchoolID,
			row.Seed,
			row.Region,
			row.Byes,
			row.Wins,
			row.Eliminated,
			row.CreatedAt,
			row.UpdatedAt,
			row.NetRtg,
			row.ORtg,
			row.DRtg,
			row.AdjT,
			row.SchoolName,
		))
	}
	return teams, nil
}

func (r *TournamentRepository) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	row, err := r.q.GetTeamByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return tournamentTeamFromRow(
		row.ID,
		row.TournamentID,
		row.SchoolID,
		row.Seed,
		row.Region,
		row.Byes,
		row.Wins,
		row.Eliminated,
		row.CreatedAt,
		row.UpdatedAt,
		row.NetRtg,
		row.ORtg,
		row.DRtg,
		row.AdjT,
		row.SchoolName,
	), nil
}

func (r *TournamentRepository) UpdateTournamentTeam(ctx context.Context, team *models.TournamentTeam) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	params := sqlc.UpdateTeamParams{
		Wins:       int32(team.Wins),
		Byes:       int32(team.Byes),
		Eliminated: team.Eliminated,
		ID:         team.ID,
	}
	if err = qtx.UpdateTeam(ctx, params); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *TournamentRepository) CreateTeam(ctx context.Context, team *models.TournamentTeam) error {
	now := time.Now()
	team.Created = now
	team.Updated = now

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	params := sqlc.CreateTeamParams{
		ID:           team.ID,
		TournamentID: team.TournamentID,
		SchoolID:     team.SchoolID,
		Seed:         int32(team.Seed),
		Region:       team.Region,
		Byes:         int32(team.Byes),
		Wins:         int32(team.Wins),
		Eliminated:   team.Eliminated,
		CreatedAt:    pgtype.Timestamptz{Time: team.Created, Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: team.Updated, Valid: true},
	}
	if err = qtx.CreateTeam(ctx, params); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *TournamentRepository) GetWinningTeam(ctx context.Context, tournamentID string) (*models.TournamentTeam, error) {
	row, err := r.q.GetWinningTeam(ctx, tournamentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	team := &models.TournamentTeam{
		ID:           row.ID,
		TournamentID: row.TournamentID,
		SchoolID:     row.SchoolID,
		Seed:         int(row.Seed),
		Region:       row.Region,
		Byes:         int(row.Byes),
		Wins:         int(row.Wins),
		Eliminated:   row.Eliminated,
		Created:      row.CreatedAt.Time,
		Updated:      row.UpdatedAt.Time,
	}
	if row.NetRtg != nil || row.ORtg != nil || row.DRtg != nil || row.AdjT != nil {
		team.KenPom = &models.KenPomStats{NetRtg: row.NetRtg, ORtg: row.ORtg, DRtg: row.DRtg, AdjT: row.AdjT}
	}
	return team, nil
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func timestamptzToTimePtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

func tournamentTeamFromRow(id, tournamentID, schoolID string, seed int32, region string, byes, wins int32, eliminated bool, createdAt, updatedAt pgtype.Timestamptz, netRtg, oRtg, dRtg, adjT *float64, schoolName *string) *models.TournamentTeam {
	team := &models.TournamentTeam{
		ID:           id,
		TournamentID: tournamentID,
		SchoolID:     schoolID,
		Seed:         int(seed),
		Region:       region,
		Byes:         int(byes),
		Wins:         int(wins),
		Eliminated:   eliminated,
		Created:      createdAt.Time,
		Updated:      updatedAt.Time,
	}
	if netRtg != nil || oRtg != nil || dRtg != nil || adjT != nil {
		team.KenPom = &models.KenPomStats{NetRtg: netRtg, ORtg: oRtg, DRtg: dRtg, AdjT: adjT}
	}
	if schoolName != nil {
		team.School = &models.School{ID: schoolID, Name: *schoolName}
	}
	return team
}
