package db

import (
	"context"
	"errors"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

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
			row.IsEliminated,
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
		row.IsEliminated,
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
		IsEliminated: team.IsEliminated,
		ID:           team.ID,
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
	team.CreatedAt = now
	team.UpdatedAt = now

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
		IsEliminated: team.IsEliminated,
		CreatedAt:    pgtype.Timestamptz{Time: team.CreatedAt, Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: team.UpdatedAt, Valid: true},
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
		IsEliminated: row.IsEliminated,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
	if row.NetRtg != nil || row.ORtg != nil || row.DRtg != nil || row.AdjT != nil {
		team.KenPom = &models.KenPomStats{NetRtg: row.NetRtg, ORtg: row.ORtg, DRtg: row.DRtg, AdjT: row.AdjT}
	}
	return team, nil
}

func (r *TournamentRepository) ListWinningTeams(ctx context.Context) (map[string]*models.TournamentTeam, error) {
	rows, err := r.q.ListWinningTeams(ctx)
	if err != nil {
		return nil, err
	}

	out := make(map[string]*models.TournamentTeam, len(rows))
	for _, row := range rows {
		team := &models.TournamentTeam{
			ID:           row.ID,
			TournamentID: row.TournamentID,
			SchoolID:     row.SchoolID,
			Seed:         int(row.Seed),
			Region:       row.Region,
			Byes:         int(row.Byes),
			Wins:         int(row.Wins),
			IsEliminated: row.IsEliminated,
			CreatedAt:    row.CreatedAt.Time,
			UpdatedAt:    row.UpdatedAt.Time,
		}
		if row.NetRtg != nil || row.ORtg != nil || row.DRtg != nil || row.AdjT != nil {
			team.KenPom = &models.KenPomStats{NetRtg: row.NetRtg, ORtg: row.ORtg, DRtg: row.DRtg, AdjT: row.AdjT}
		}
		out[row.TournamentID] = team
	}
	return out, nil
}

func (r *TournamentRepository) ReplaceTeams(ctx context.Context, tournamentID string, teams []*models.TournamentTeam) error {
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
	now := time.Now()

	if _, err = qtx.SoftDeleteTeamsByTournamentID(ctx, sqlc.SoftDeleteTeamsByTournamentIDParams{
		DeletedAt:    pgtype.Timestamptz{Time: now, Valid: true},
		TournamentID: tournamentID,
	}); err != nil {
		return err
	}

	for _, team := range teams {
		if team == nil {
			continue
		}
		team.CreatedAt = now
		team.UpdatedAt = now
		params := sqlc.CreateTeamParams{
			ID:           team.ID,
			TournamentID: team.TournamentID,
			SchoolID:     team.SchoolID,
			Seed:         int32(team.Seed),
			Region:       team.Region,
			Byes:         int32(team.Byes),
			Wins:         int32(team.Wins),
			IsEliminated: team.IsEliminated,
			CreatedAt:    pgtype.Timestamptz{Time: team.CreatedAt, Valid: true},
			UpdatedAt:    pgtype.Timestamptz{Time: team.UpdatedAt, Valid: true},
		}
		if err = qtx.CreateTeam(ctx, params); err != nil {
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *TournamentRepository) BulkUpsertKenPomStats(ctx context.Context, updates []models.TeamKenPomUpdate) error {
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
	for _, u := range updates {
		params := sqlc.UpsertTeamKenPomStatsParams{
			TeamID: u.TeamID,
			NetRtg: &u.NetRtg,
			ORtg:   &u.ORtg,
			DRtg:   &u.DRtg,
			AdjT:   &u.AdjT,
		}
		if err = qtx.UpsertTeamKenPomStats(ctx, params); err != nil {
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func tournamentTeamFromRow(id, tournamentID, schoolID string, seed int32, region string, byes, wins int32, isEliminated bool, createdAt, updatedAt pgtype.Timestamptz, netRtg, oRtg, dRtg, adjT *float64, schoolName *string) *models.TournamentTeam {
	team := &models.TournamentTeam{
		ID:           id,
		TournamentID: tournamentID,
		SchoolID:     schoolID,
		Seed:         int(seed),
		Region:       region,
		Byes:         int(byes),
		Wins:         int(wins),
		IsEliminated: isEliminated,
		CreatedAt:    createdAt.Time,
		UpdatedAt:    updatedAt.Time,
	}
	if netRtg != nil || oRtg != nil || dRtg != nil || adjT != nil {
		team.KenPom = &models.KenPomStats{NetRtg: netRtg, ORtg: oRtg, DRtg: dRtg, AdjT: adjT}
	}
	if schoolName != nil {
		team.School = &models.School{ID: schoolID, Name: *schoolName}
	}
	return team
}
