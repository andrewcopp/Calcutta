package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/google/uuid"
)

// CalcuttaRepository handles data access for Calcutta entities
type CalcuttaRepository struct {
	db *sql.DB
}

// NewCalcuttaRepository creates a new CalcuttaRepository
func NewCalcuttaRepository(db *sql.DB) *CalcuttaRepository {
	return &CalcuttaRepository{db: db}
}

// GetAll retrieves all Calcuttas
func (r *CalcuttaRepository) GetAll(ctx context.Context) ([]*models.Calcutta, error) {
	query := `
		SELECT id, tournament_id, owner_id, name, created_at, updated_at, deleted_at
		FROM calcuttas
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calcuttas []*models.Calcutta
	for rows.Next() {
		calcutta := &models.Calcutta{}
		var createdAt, updatedAt time.Time
		var deletedAt sql.NullTime

		err := rows.Scan(
			&calcutta.ID,
			&calcutta.TournamentID,
			&calcutta.OwnerID,
			&calcutta.Name,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		calcutta.Created = createdAt
		calcutta.Updated = updatedAt
		if deletedAt.Valid {
			calcutta.Deleted = &deletedAt.Time
		}

		calcuttas = append(calcuttas, calcutta)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return calcuttas, nil
}

// GetByID retrieves a Calcutta by ID
func (r *CalcuttaRepository) GetByID(ctx context.Context, id string) (*models.Calcutta, error) {
	query := `
		SELECT id, tournament_id, owner_id, name, created_at, updated_at, deleted_at
		FROM calcuttas
		WHERE id = $1 AND deleted_at IS NULL
	`

	calcutta := &models.Calcutta{}
	var createdAt, updatedAt time.Time
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&calcutta.ID,
		&calcutta.TournamentID,
		&calcutta.OwnerID,
		&calcutta.Name,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("calcutta not found")
		}
		return nil, err
	}

	calcutta.Created = createdAt
	calcutta.Updated = updatedAt
	if deletedAt.Valid {
		calcutta.Deleted = &deletedAt.Time
	}

	return calcutta, nil
}

// Create creates a new Calcutta
func (r *CalcuttaRepository) Create(ctx context.Context, calcutta *models.Calcutta) error {
	query := `
		INSERT INTO calcuttas (id, tournament_id, owner_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	calcutta.ID = uuid.New().String()
	now := time.Now()
	calcutta.Created = now
	calcutta.Updated = now

	_, err := r.db.ExecContext(ctx, query,
		calcutta.ID,
		calcutta.TournamentID,
		calcutta.OwnerID,
		calcutta.Name,
		calcutta.Created,
		calcutta.Updated,
	)

	return err
}

// Update updates an existing Calcutta
func (r *CalcuttaRepository) Update(ctx context.Context, calcutta *models.Calcutta) error {
	query := `
		UPDATE calcuttas
		SET tournament_id = $1, owner_id = $2, name = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	calcutta.Updated = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		calcutta.TournamentID,
		calcutta.OwnerID,
		calcutta.Name,
		calcutta.Updated,
		calcutta.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("calcutta not found")
	}

	return nil
}

// Delete soft-deletes a Calcutta
func (r *CalcuttaRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE calcuttas
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("calcutta not found")
	}

	return nil
}

// GetEntries retrieves all entries for a Calcutta
func (r *CalcuttaRepository) GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, error) {
	query := `
		SELECT id, name, user_id, calcutta_id, created_at, updated_at, deleted_at
		FROM calcutta_entries
		WHERE calcutta_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, calcuttaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.CalcuttaEntry
	for rows.Next() {
		entry := &models.CalcuttaEntry{}
		var userID sql.NullString
		var createdAt, updatedAt time.Time
		var deletedAt sql.NullTime

		err := rows.Scan(
			&entry.ID,
			&entry.Name,
			&userID,
			&entry.CalcuttaID,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		if userID.Valid {
			entry.UserID = &userID.String
		}
		entry.Created = createdAt
		entry.Updated = updatedAt
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

	var teams []*models.CalcuttaEntryTeam
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
