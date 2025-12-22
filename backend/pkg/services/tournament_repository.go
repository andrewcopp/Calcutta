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
		SELECT t.id, t.name, t.rounds,
			t.final_four_top_left, t.final_four_bottom_left, t.final_four_top_right, t.final_four_bottom_right,
			t.created_at, t.updated_at
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
		if err := rows.Scan(
			&tournament.ID,
			&tournament.Name,
			&tournament.Rounds,
			&tournament.FinalFourTopLeft,
			&tournament.FinalFourBottomLeft,
			&tournament.FinalFourTopRight,
			&tournament.FinalFourBottomRight,
			&tournament.Created,
			&tournament.Updated,
		); err != nil {
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
		SELECT
			tt.id, tt.tournament_id, tt.school_id, tt.seed, tt.byes, tt.wins, tt.eliminated, tt.created_at, tt.updated_at,
			kps.net_rtg,
			kps.o_rtg,
			kps.d_rtg,
			kps.adj_t
		FROM tournament_teams tt
		LEFT JOIN tournament_team_kenpom_stats kps ON kps.tournament_team_id = tt.id AND kps.deleted_at IS NULL
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
	var netRtg sql.NullFloat64
	var oRtg sql.NullFloat64
	var dRtg sql.NullFloat64
	var adjT sql.NullFloat64
	if err := rows.Scan(
		&team.ID,
		&team.TournamentID,
		&team.SchoolID,
		&team.Seed,
		&team.Byes,
		&team.Wins,
		&team.Eliminated,
		&team.Created,
		&team.Updated,
		&netRtg,
		&oRtg,
		&dRtg,
		&adjT,
	); err != nil {
		return nil, err
	}

	hasKenPom := netRtg.Valid || oRtg.Valid || dRtg.Valid || adjT.Valid
	if hasKenPom {
		team.KenPom = &models.KenPomStats{
			NetRtg: nullFloat64Ptr(netRtg),
			ORtg:   nullFloat64Ptr(oRtg),
			DRtg:   nullFloat64Ptr(dRtg),
			AdjT:   nullFloat64Ptr(adjT),
		}
	}

	return &team, nil
}

// GetByID returns a tournament by ID
func (r *TournamentRepository) GetByID(ctx context.Context, id string) (*models.Tournament, error) {
	var tournament models.Tournament
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, rounds,
			final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right,
			created_at, updated_at
		FROM tournaments
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&tournament.ID,
		&tournament.Name,
		&tournament.Rounds,
		&tournament.FinalFourTopLeft,
		&tournament.FinalFourBottomLeft,
		&tournament.FinalFourTopRight,
		&tournament.FinalFourBottomRight,
		&tournament.Created,
		&tournament.Updated,
	)

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
		SELECT 
			tt.id, 
			tt.tournament_id, 
			tt.school_id, 
			tt.seed, 
			tt.region, 
			tt.byes, 
			tt.wins, 
			tt.eliminated, 
			tt.created_at, 
			tt.updated_at,
			kps.net_rtg,
			kps.o_rtg,
			kps.d_rtg,
			kps.adj_t,
			s.id as school_id,
			s.name as school_name
		FROM tournament_teams tt
		LEFT JOIN tournament_team_kenpom_stats kps ON kps.tournament_team_id = tt.id AND kps.deleted_at IS NULL
		LEFT JOIN schools s ON tt.school_id = s.id
		WHERE tt.tournament_id = $1 AND tt.deleted_at IS NULL
		ORDER BY tt.seed ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tournamentID)
	if err != nil {
		return []*models.TournamentTeam{}, err
	}
	defer rows.Close()

	teams := make([]*models.TournamentTeam, 0)
	for rows.Next() {
		team := &models.TournamentTeam{}
		var createdAt, updatedAt time.Time
		var schoolIDFromJoin sql.NullString
		var schoolName sql.NullString
		var netRtg sql.NullFloat64
		var oRtg sql.NullFloat64
		var dRtg sql.NullFloat64
		var adjT sql.NullFloat64

		err := rows.Scan(
			&team.ID,
			&team.TournamentID,
			&team.SchoolID,
			&team.Seed,
			&team.Region,
			&team.Byes,
			&team.Wins,
			&team.Eliminated,
			&createdAt,
			&updatedAt,
			&netRtg,
			&oRtg,
			&dRtg,
			&adjT,
			&schoolIDFromJoin,
			&schoolName,
		)
		if err != nil {
			return []*models.TournamentTeam{}, err
		}

		team.Created = createdAt
		team.Updated = updatedAt

		hasKenPom := netRtg.Valid || oRtg.Valid || dRtg.Valid || adjT.Valid
		if hasKenPom {
			team.KenPom = &models.KenPomStats{
				NetRtg: nullFloat64Ptr(netRtg),
				ORtg:   nullFloat64Ptr(oRtg),
				DRtg:   nullFloat64Ptr(dRtg),
				AdjT:   nullFloat64Ptr(adjT),
			}
		}

		// Add school information if available
		if schoolIDFromJoin.Valid && schoolName.Valid {
			team.School = &models.School{
				ID:   schoolIDFromJoin.String,
				Name: schoolName.String,
			}
		}

		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return []*models.TournamentTeam{}, err
	}

	return teams, nil
}

// GetTournamentTeam retrieves a single tournament team by ID
func (r *TournamentRepository) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	query := `
		SELECT 
			tt.id, tt.tournament_id, tt.school_id, tt.seed, tt.region, tt.byes, tt.wins, tt.eliminated,
			tt.created_at, tt.updated_at,
			kps.net_rtg,
			kps.o_rtg,
			kps.d_rtg,
			kps.adj_t,
			s.id as school_id_from_join, s.name as school_name
		FROM tournament_teams tt
		LEFT JOIN tournament_team_kenpom_stats kps ON kps.tournament_team_id = tt.id AND kps.deleted_at IS NULL
		LEFT JOIN schools s ON tt.school_id = s.id
		WHERE tt.id = $1 AND tt.deleted_at IS NULL
	`

	var team models.TournamentTeam
	var schoolIDFromJoin sql.NullString
	var schoolName sql.NullString
	var createdAt, updatedAt time.Time
	var netRtg sql.NullFloat64
	var oRtg sql.NullFloat64
	var dRtg sql.NullFloat64
	var adjT sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&team.ID,
		&team.TournamentID,
		&team.SchoolID,
		&team.Seed,
		&team.Region,
		&team.Byes,
		&team.Wins,
		&team.Eliminated,
		&createdAt,
		&updatedAt,
		&netRtg,
		&oRtg,
		&dRtg,
		&adjT,
		&schoolIDFromJoin,
		&schoolName,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	team.Created = createdAt
	team.Updated = updatedAt

	hasKenPom := netRtg.Valid || oRtg.Valid || dRtg.Valid || adjT.Valid
	if hasKenPom {
		team.KenPom = &models.KenPomStats{
			NetRtg: nullFloat64Ptr(netRtg),
			ORtg:   nullFloat64Ptr(oRtg),
			DRtg:   nullFloat64Ptr(dRtg),
			AdjT:   nullFloat64Ptr(adjT),
		}
	}

	if schoolIDFromJoin.Valid && schoolName.Valid {
		team.School = &models.School{
			ID:   schoolIDFromJoin.String,
			Name: schoolName.String,
		}
	}

	return &team, nil
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
		INSERT INTO tournaments (
			id,
			name,
			rounds,
			final_four_top_left,
			final_four_bottom_left,
			final_four_top_right,
			final_four_bottom_right,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	tournament.Created = now
	tournament.Updated = now

	result, err := r.db.ExecContext(ctx, query,
		tournament.ID,
		tournament.Name,
		tournament.Rounds,
		tournament.FinalFourTopLeft,
		tournament.FinalFourBottomLeft,
		tournament.FinalFourTopRight,
		tournament.FinalFourBottomRight,
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

// CreateTeam creates a new tournament team in the database
func (r *TournamentRepository) CreateTeam(ctx context.Context, team *models.TournamentTeam) error {
	log.Printf("CreateTeam: Starting to create team with ID: %s", team.ID)
	log.Printf("CreateTeam: Team data: %+v", team)

	query := `
		INSERT INTO tournament_teams (
			id, tournament_id, school_id, seed, region, byes, wins, eliminated,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	team.Created = now
	team.Updated = now

	log.Printf("CreateTeam: Executing SQL query with values: id=%s, tournament_id=%s, school_id=%s, seed=%d, region=%s, byes=%d, wins=%d, eliminated=%v, created_at=%v, updated_at=%v",
		team.ID, team.TournamentID, team.SchoolID, team.Seed, team.Region, team.Byes, team.Wins, team.Eliminated, team.Created, team.Updated)

	result, err := r.db.ExecContext(ctx, query,
		team.ID,
		team.TournamentID,
		team.SchoolID,
		team.Seed,
		team.Region,
		team.Byes,
		team.Wins,
		team.Eliminated,
		team.Created,
		team.Updated,
	)
	if err != nil {
		log.Printf("CreateTeam: Error executing query: %v", err)
		return err
	}

	if team.KenPom != nil {
		if err := r.UpsertTournamentTeamKenPomStats(ctx, team.ID, team.KenPom); err != nil {
			return err
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("CreateTeam: Error getting rows affected: %v", err)
		return err
	}

	log.Printf("CreateTeam: Successfully inserted team. Rows affected: %d", rowsAffected)
	return nil
}

func (r *TournamentRepository) UpsertTournamentTeamKenPomStats(ctx context.Context, tournamentTeamID string, stats *models.KenPomStats) error {
	if stats == nil {
		return nil
	}

	query := `
		INSERT INTO tournament_team_kenpom_stats (
			tournament_team_id,
			net_rtg, o_rtg, d_rtg, adj_t,
			updated_at
		) VALUES (
			$1,
			$2, $3, $4, $5,
			NOW()
		)
		ON CONFLICT (tournament_team_id)
		DO UPDATE SET
			net_rtg = EXCLUDED.net_rtg,
			o_rtg = EXCLUDED.o_rtg,
			d_rtg = EXCLUDED.d_rtg,
			adj_t = EXCLUDED.adj_t,
			updated_at = NOW(),
			deleted_at = NULL
		WHERE tournament_team_kenpom_stats.deleted_at IS NULL OR tournament_team_kenpom_stats.deleted_at IS NOT NULL
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		tournamentTeamID,
		float64OrNil(stats.NetRtg),
		float64OrNil(stats.ORtg),
		float64OrNil(stats.DRtg),
		float64OrNil(stats.AdjT),
	)
	return err
}

func nullFloat64Ptr(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Float64
	return &v
}

func float64OrNil(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}
