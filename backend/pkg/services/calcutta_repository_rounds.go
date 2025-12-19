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

// GetRounds retrieves all rounds for a calcutta
func (r *CalcuttaRepository) GetRounds(ctx context.Context, calcuttaID string) ([]*models.CalcuttaRound, error) {
	query := `
		SELECT id, calcutta_id, round, points, created_at, updated_at, deleted_at
		FROM calcutta_rounds
		WHERE calcutta_id = $1 AND deleted_at IS NULL
		ORDER BY round ASC
	`

	rows, err := r.db.QueryContext(ctx, query, calcuttaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rounds := make([]*models.CalcuttaRound, 0)
	for rows.Next() {
		round := &models.CalcuttaRound{}
		var createdAt, updatedAt time.Time
		var deletedAt sql.NullTime

		err := rows.Scan(
			&round.ID,
			&round.CalcuttaID,
			&round.Round,
			&round.Points,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		round.Created = createdAt
		round.Updated = updatedAt
		if deletedAt.Valid {
			round.Deleted = &deletedAt.Time
		}

		rounds = append(rounds, round)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rounds, nil
}

// CreateRound creates a new round for a Calcutta
func (r *CalcuttaRepository) CreateRound(ctx context.Context, round *models.CalcuttaRound) error {
	log.Printf("Creating new round: %+v", round)

	query := `
		INSERT INTO calcutta_rounds (id, calcutta_id, round, points, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	round.ID = uuid.New().String()
	now := time.Now()
	round.Created = now
	round.Updated = now

	log.Printf("Executing query with values: id=%s, calcuttaId=%s, round=%d, points=%d, created=%v, updated=%v",
		round.ID, round.CalcuttaID, round.Round, round.Points, round.Created, round.Updated)

	result, err := r.db.ExecContext(ctx, query,
		round.ID,
		round.CalcuttaID,
		round.Round,
		round.Points,
		round.Created,
		round.Updated,
	)
	if err != nil {
		log.Printf("Error executing create round query: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}
	log.Printf("Created round successfully, rows affected: %d", rowsAffected)

	return nil
}

// GetTournamentTeam retrieves a tournament team by ID
func (r *CalcuttaRepository) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	query := `
		SELECT 
			id, school_id, tournament_id, seed, byes, wins, eliminated,
			created_at, updated_at, deleted_at
		FROM tournament_teams
		WHERE id = $1 AND deleted_at IS NULL
	`

	team := &models.TournamentTeam{}
	var createdAt, updatedAt time.Time
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&team.ID,
		&team.SchoolID,
		&team.TournamentID,
		&team.Seed,
		&team.Byes,
		&team.Wins,
		&team.Eliminated,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("tournament team not found")
		}
		return nil, err
	}

	team.Created = createdAt
	team.Updated = updatedAt
	if deletedAt.Valid {
		team.Deleted = &deletedAt.Time
	}

	return team, nil
}
