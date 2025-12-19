package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/google/uuid"
)

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

	calcuttas := make([]*models.Calcutta, 0)
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
	log.Printf("Executing GetByID query for calcutta ID: %s", id)

	query := `
		SELECT id, tournament_id, owner_id, name, created_at, updated_at, deleted_at
		FROM calcuttas
		WHERE id = $1 AND deleted_at IS NULL
	`

	calcutta := &models.Calcutta{}
	var createdAt, updatedAt time.Time
	var deletedAt sql.NullTime

	log.Printf("Running query: %s with ID: %s", query, id)
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
			log.Printf("No calcutta found with ID: %s", id)
			return nil, errors.New("calcutta not found")
		}
		log.Printf("Database error while fetching calcutta: %v", err)
		return nil, err
	}

	calcutta.Created = createdAt
	calcutta.Updated = updatedAt
	if deletedAt.Valid {
		calcutta.Deleted = &deletedAt.Time
	}

	log.Printf("Successfully retrieved calcutta from database: %+v", calcutta)
	return calcutta, nil
}

// Create creates a new Calcutta
func (r *CalcuttaRepository) Create(ctx context.Context, calcutta *models.Calcutta) error {
	log.Printf("Creating new calcutta: %+v", calcutta)

	query := `
		INSERT INTO calcuttas (id, tournament_id, owner_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	calcutta.ID = uuid.New().String()
	now := time.Now()
	calcutta.Created = now
	calcutta.Updated = now

	log.Printf("Executing query with values: id=%s, tournamentId=%s, ownerId=%s, name=%s, created=%v, updated=%v",
		calcutta.ID, calcutta.TournamentID, calcutta.OwnerID, calcutta.Name, calcutta.Created, calcutta.Updated)

	result, err := r.db.ExecContext(ctx, query,
		calcutta.ID,
		calcutta.TournamentID,
		calcutta.OwnerID,
		calcutta.Name,
		calcutta.Created,
		calcutta.Updated,
	)
	if err != nil {
		log.Printf("Error executing create calcutta query: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}
	log.Printf("Created calcutta successfully, rows affected: %d", rowsAffected)

	return nil
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

func (r *CalcuttaRepository) GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error) {
	query := `
		SELECT c.id, c.tournament_id, c.created_at, c.updated_at, c.deleted_at
		FROM calcuttas c
		WHERE c.tournament_id = $1 AND c.deleted_at IS NULL
	`
	rows, err := r.db.QueryContext(ctx, query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calcuttas []*models.Calcutta
	for rows.Next() {
		calcutta := &models.Calcutta{}
		err := rows.Scan(&calcutta.ID, &calcutta.TournamentID, &calcutta.Created, &calcutta.Updated, &calcutta.Deleted)
		if err != nil {
			return nil, err
		}
		calcuttas = append(calcuttas, calcutta)
	}

	return calcuttas, nil
}
