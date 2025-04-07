package services

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
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
		SELECT tt.id, tt.tournament_id, tt.school_id, tt.seed, tt.byes, tt.wins, tt.eliminated, tt.created_at, tt.updated_at
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
	if err := rows.Scan(&team.ID, &team.TournamentID, &team.SchoolID, &team.Seed, &team.Byes, &team.Wins, &team.Eliminated, &team.Created, &team.Updated); err != nil {
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

// GetTeams returns all teams for a tournament
func (r *TournamentRepository) GetTeams(ctx context.Context, tournamentID string) ([]*models.TournamentTeam, error) {
	query := `
		SELECT id, tournament_id, school_id, seed, byes, wins, eliminated, created_at, updated_at
		FROM tournament_teams
		WHERE tournament_id = $1 AND deleted_at IS NULL
		ORDER BY seed ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.TournamentTeam
	for rows.Next() {
		team := &models.TournamentTeam{}
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&team.ID,
			&team.TournamentID,
			&team.SchoolID,
			&team.Seed,
			&team.Byes,
			&team.Wins,
			&team.Eliminated,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		team.Created = createdAt
		team.Updated = updatedAt

		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

// UpdateTournamentTeam updates a tournament team in the database
func (r *TournamentRepository) UpdateTournamentTeam(ctx context.Context, team *models.TournamentTeam) error {
	query := `
		UPDATE tournament_teams
		SET wins = $1, byes = $2, eliminated = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, team.Wins, team.Byes, team.Eliminated, team.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetGameByID returns a tournament game by ID
func (r *TournamentRepository) GetGameByID(ctx context.Context, id string) (*models.TournamentGame, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tournament_id, team1_id, team2_id, tipoff_time, sort_order, team1_score, team2_score, next_game_id, next_game_slot, is_final, created_at, updated_at
		FROM tournament_games
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // No game found
	}

	var game models.TournamentGame
	if err := rows.Scan(
		&game.ID,
		&game.TournamentID,
		&game.Team1ID,
		&game.Team2ID,
		&game.TipoffTime,
		&game.SortOrder,
		&game.Team1Score,
		&game.Team2Score,
		&game.NextGameID,
		&game.NextGameSlot,
		&game.IsFinal,
		&game.Created,
		&game.Updated,
	); err != nil {
		return nil, err
	}

	return &game, nil
}

// GetGamesByTournamentID returns all games for a tournament
func (r *TournamentRepository) GetGamesByTournamentID(ctx context.Context, tournamentID string) ([]*models.TournamentGame, error) {
	// First, get the game to find its tournament ID
	game, err := r.GetGameByID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	if game == nil {
		return nil, nil // Game not found
	}

	// Get all games for the tournament
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tournament_id, team1_id, team2_id, tipoff_time, sort_order, team1_score, team2_score, next_game_id, next_game_slot, is_final, created_at, updated_at
		FROM tournament_games
		WHERE tournament_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order ASC
	`, game.TournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []*models.TournamentGame
	for rows.Next() {
		var game models.TournamentGame
		if err := rows.Scan(
			&game.ID,
			&game.TournamentID,
			&game.Team1ID,
			&game.Team2ID,
			&game.TipoffTime,
			&game.SortOrder,
			&game.Team1Score,
			&game.Team2Score,
			&game.NextGameID,
			&game.NextGameSlot,
			&game.IsFinal,
			&game.Created,
			&game.Updated,
		); err != nil {
			return nil, err
		}
		games = append(games, &game)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return games, nil
}

// Create creates a new tournament in the database
func (r *TournamentRepository) Create(ctx context.Context, tournament *models.Tournament) error {
	log.Printf("Inserting tournament into database: %+v", tournament)

	query := `
		INSERT INTO tournaments (id, name, rounds, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now()
	tournament.Created = now
	tournament.Updated = now

	result, err := r.db.ExecContext(ctx, query,
		tournament.ID,
		tournament.Name,
		tournament.Rounds,
		tournament.Created,
		tournament.Updated,
	)
	if err != nil {
		log.Printf("Database error creating tournament: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	log.Printf("Successfully inserted tournament. Rows affected: %d", rowsAffected)
	return nil
}
