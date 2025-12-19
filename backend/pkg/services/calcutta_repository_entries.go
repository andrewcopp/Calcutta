package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// GetEntries retrieves all entries for a Calcutta
func (r *CalcuttaRepository) GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, error) {
	query := `
		WITH entry_points AS (
			SELECT 
				cp.entry_id,
				COALESCE(SUM(cpt.actual_points), 0) as total_points
			FROM calcutta_portfolios cp
			LEFT JOIN calcutta_portfolio_teams cpt ON cp.id = cpt.portfolio_id
			WHERE cp.deleted_at IS NULL AND cpt.deleted_at IS NULL
			GROUP BY cp.entry_id
		)
		SELECT 
			ce.id, 
			ce.name, 
			ce.user_id, 
			ce.calcutta_id, 
			ce.created_at, 
			ce.updated_at, 
			ce.deleted_at,
			COALESCE(ep.total_points, 0) as total_points
		FROM calcutta_entries ce
		LEFT JOIN entry_points ep ON ce.id = ep.entry_id
		WHERE ce.calcutta_id = $1 AND ce.deleted_at IS NULL
		ORDER BY ep.total_points DESC NULLS LAST, ce.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, calcuttaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]*models.CalcuttaEntry, 0)
	for rows.Next() {
		entry := &models.CalcuttaEntry{}
		var userID sql.NullString
		var createdAt, updatedAt time.Time
		var deletedAt sql.NullTime
		var totalPoints float64

		err := rows.Scan(
			&entry.ID,
			&entry.Name,
			&userID,
			&entry.CalcuttaID,
			&createdAt,
			&updatedAt,
			&deletedAt,
			&totalPoints,
		)
		if err != nil {
			return nil, err
		}

		if userID.Valid {
			entry.UserID = &userID.String
		}
		entry.Created = createdAt
		entry.Updated = updatedAt
		entry.TotalPoints = totalPoints
		if deletedAt.Valid {
			entry.Deleted = &deletedAt.Time
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// GetEntryTeams retrieves all teams for a Calcutta entry
func (r *CalcuttaRepository) GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error) {
	query := `
		SELECT 
			cet.id, 
			cet.entry_id, 
			cet.team_id, 
			cet.bid, 
			cet.created_at, 
			cet.updated_at, 
			cet.deleted_at,
			tt.id as team_id,
			tt.school_id,
			tt.tournament_id,
			tt.seed,
			tt.byes,
			tt.wins,
			tt.created_at as team_created_at,
			tt.updated_at as team_updated_at,
			tt.deleted_at as team_deleted_at
		FROM calcutta_entry_teams cet
		JOIN tournament_teams tt ON cet.team_id = tt.id
		WHERE cet.entry_id = $1 AND cet.deleted_at IS NULL
		ORDER BY cet.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]*models.CalcuttaEntryTeam, 0)
	for rows.Next() {
		team := &models.CalcuttaEntryTeam{}
		var createdAt, updatedAt time.Time
		var deletedAt sql.NullTime

		// Team fields
		var teamID, schoolID, tournamentID string
		var seed, byes, wins int
		var teamCreatedAt, teamUpdatedAt time.Time
		var teamDeletedAt sql.NullTime

		err := rows.Scan(
			&team.ID,
			&team.EntryID,
			&team.TeamID,
			&team.Bid,
			&createdAt,
			&updatedAt,
			&deletedAt,
			&teamID,
			&schoolID,
			&tournamentID,
			&seed,
			&byes,
			&wins,
			&teamCreatedAt,
			&teamUpdatedAt,
			&teamDeletedAt,
		)
		if err != nil {
			return nil, err
		}

		team.Created = createdAt
		team.Updated = updatedAt
		if deletedAt.Valid {
			team.Deleted = &deletedAt.Time
		}

		// Create the nested team object
		team.Team = &models.TournamentTeam{
			ID:           teamID,
			SchoolID:     schoolID,
			TournamentID: tournamentID,
			Seed:         seed,
			Byes:         byes,
			Wins:         wins,
			Created:      teamCreatedAt,
			Updated:      teamUpdatedAt,
		}
		if teamDeletedAt.Valid {
			team.Team.Deleted = &teamDeletedAt.Time
		}

		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

// GetEntry retrieves an entry by ID
func (r *CalcuttaRepository) GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error) {
	query := `
		SELECT id, name, user_id, calcutta_id, created_at, updated_at, deleted_at
		FROM calcutta_entries
		WHERE id = $1 AND deleted_at IS NULL
	`

	entry := &models.CalcuttaEntry{}
	var createdAt, updatedAt time.Time
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.Name,
		&entry.UserID,
		&entry.CalcuttaID,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{Resource: "entry", ID: id}
		}
		return nil, err
	}

	entry.Created = createdAt
	entry.Updated = updatedAt
	if deletedAt.Valid {
		entry.Deleted = &deletedAt.Time
	}

	return entry, nil
}
