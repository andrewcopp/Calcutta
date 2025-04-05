package services

import (
	"calcutta/internal/models"
	"context"
	"database/sql"
)

// TournamentRepository handles database operations for tournaments
type TournamentRepository struct {
	db *sql.DB
}

// NewTournamentRepository creates a new TournamentRepository
func NewTournamentRepository(db *sql.DB) *TournamentRepository {
	return &TournamentRepository{db: db}
}

// GetAll returns all tournaments with their winning teams
func (r *TournamentRepository) GetAll(ctx context.Context) ([]models.Tournament, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.name, t.rounds, t.created_at, t.updated_at
		FROM tournaments t
		WHERE t.deleted_at IS NULL
		ORDER BY t.name DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tournaments []models.Tournament
	for rows.Next() {
		var tournament models.Tournament
		if err := rows.Scan(&tournament.ID, &tournament.Name, &tournament.Rounds, &tournament.Created, &tournament.Updated); err != nil {
			return nil, err
		}
		tournaments = append(tournaments, tournament)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tournaments, nil
}

// GetWinningTeam returns the winning team for a tournament
func (r *TournamentRepository) GetWinningTeam(ctx context.Context, tournamentID string) (*models.TournamentTeam, error) {
	// Find the team with the most wins in the tournament
	// In a typical tournament, the winner will have the maximum number of wins
	rows, err := r.db.QueryContext(ctx, `
		SELECT tt.id, tt.tournament_id, tt.school_id, tt.seed, tt.byes, tt.wins, tt.created_at, tt.updated_at
		FROM tournament_teams tt
		WHERE tt.tournament_id = $1 AND tt.deleted_at IS NULL
		ORDER BY tt.wins DESC
		LIMIT 1
	`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // No winning team found
	}

	var team models.TournamentTeam
	if err := rows.Scan(&team.ID, &team.TournamentID, &team.SchoolID, &team.Seed, &team.Byes, &team.Wins, &team.Created, &team.Updated); err != nil {
		return nil, err
	}

	return &team, nil
}

// GetTournamentWithWinner returns a tournament with its winning team and school
func (r *TournamentRepository) GetTournamentWithWinner(ctx context.Context, tournamentID string) (*models.Tournament, *models.TournamentTeam, *models.School, error) {
	// Get the tournament
	tournament, err := r.GetByID(ctx, tournamentID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get the winning team
	team, err := r.GetWinningTeam(ctx, tournamentID)
	if err != nil {
		return tournament, nil, nil, err
	}

	if team == nil {
		return tournament, nil, nil, nil
	}

	// Get the school
	schoolRepo := NewSchoolRepository(r.db)
	school, err := schoolRepo.GetByID(ctx, team.SchoolID)
	if err != nil {
		return tournament, team, nil, err
	}

	return tournament, team, &school, nil
}

// GetByID returns a tournament by ID
func (r *TournamentRepository) GetByID(ctx context.Context, id string) (*models.Tournament, error) {
	var tournament models.Tournament
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, rounds, created_at, updated_at
		FROM tournaments
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&tournament.ID, &tournament.Name, &tournament.Rounds, &tournament.Created, &tournament.Updated)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &tournament, nil
}
