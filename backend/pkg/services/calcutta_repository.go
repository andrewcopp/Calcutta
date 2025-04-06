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

// CalcuttaRepositoryInterface defines the interface that a Calcutta repository must implement
type CalcuttaRepositoryInterface interface {
	GetAll(ctx context.Context) ([]*models.Calcutta, error)
	GetByID(ctx context.Context, id string) (*models.Calcutta, error)
	Create(ctx context.Context, calcutta *models.Calcutta) error
	Update(ctx context.Context, calcutta *models.Calcutta) error
	Delete(ctx context.Context, id string) error
	GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, error)
	GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error)
	GetPortfolio(ctx context.Context, id string) (*models.CalcuttaPortfolio, error)
	GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error)
	UpdatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error
	GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error)
	UpdatePortfolio(ctx context.Context, portfolio *models.CalcuttaPortfolio) error
	GetRounds(ctx context.Context, calcuttaID string) ([]*models.CalcuttaRound, error)
	GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error)
	CreatePortfolio(ctx context.Context, portfolio *models.CalcuttaPortfolio) error
	CreatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error
	GetPortfolios(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error)
	GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error)
	GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error)
}

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

// GetPortfolio retrieves a portfolio by ID
func (r *CalcuttaRepository) GetPortfolio(ctx context.Context, id string) (*models.CalcuttaPortfolio, error) {
	query := `
		SELECT 
			id, entry_id, maximum_points,
			created_at, updated_at, deleted_at
		FROM calcutta_portfolios
		WHERE id = $1 AND deleted_at IS NULL
	`

	portfolio := &models.CalcuttaPortfolio{}
	var createdAt, updatedAt time.Time
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&portfolio.ID,
		&portfolio.EntryID,
		&portfolio.MaximumPoints,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("portfolio not found")
		}
		return nil, err
	}

	portfolio.Created = createdAt
	portfolio.Updated = updatedAt
	if deletedAt.Valid {
		portfolio.Deleted = &deletedAt.Time
	}

	return portfolio, nil
}

// GetPortfolioTeams retrieves all teams for a portfolio
func (r *CalcuttaRepository) GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error) {
	query := `
		SELECT 
			cpt.id, 
			cpt.portfolio_id, 
			cpt.team_id, 
			cpt.ownership_percentage, 
			cpt.actual_points, 
			cpt.expected_points, 
			cpt.predicted_points, 
			cpt.created_at, 
			cpt.updated_at, 
			cpt.deleted_at
		FROM calcutta_portfolio_teams cpt
		WHERE cpt.portfolio_id = $1 AND cpt.deleted_at IS NULL
	`

	rows, err := r.db.QueryContext(ctx, query, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.CalcuttaPortfolioTeam
	for rows.Next() {
		team := &models.CalcuttaPortfolioTeam{}
		var createdAt, updatedAt time.Time
		var deletedAt sql.NullTime

		err := rows.Scan(
			&team.ID,
			&team.PortfolioID,
			&team.TeamID,
			&team.OwnershipPercentage,
			&team.ActualPoints,
			&team.ExpectedPoints,
			&team.PredictedPoints,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		team.Created = createdAt
		team.Updated = updatedAt
		if deletedAt.Valid {
			team.Deleted = &deletedAt.Time
		}

		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

// UpdatePortfolioTeam updates a portfolio team
func (r *CalcuttaRepository) UpdatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error {
	query := `
		UPDATE calcutta_portfolio_teams
		SET ownership_percentage = $1, actual_points = $2, expected_points = $3, predicted_points = $4, updated_at = $5
		WHERE id = $6 AND deleted_at IS NULL
	`

	log.Printf("Updating portfolio team %s with ownership=%f, actual_points=%f, expected_points=%f, predicted_points=%f",
		team.ID, team.OwnershipPercentage, team.ActualPoints, team.ExpectedPoints, team.PredictedPoints)

	result, err := r.db.ExecContext(ctx, query,
		team.OwnershipPercentage,
		team.ActualPoints,
		team.ExpectedPoints,
		team.PredictedPoints,
		team.Updated,
		team.ID,
	)
	if err != nil {
		log.Printf("Error executing update for portfolio team %s: %v", team.ID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected for portfolio team %s: %v", team.ID, err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No rows updated for portfolio team %s", team.ID)
		return errors.New("portfolio team not found")
	}

	log.Printf("Successfully updated portfolio team %s (rows affected: %d)", team.ID, rowsAffected)
	return nil
}

// GetPortfoliosByEntry retrieves all portfolios for an entry
func (r *CalcuttaRepository) GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	query := `
		SELECT id, entry_id, created_at, updated_at, deleted_at
		FROM calcutta_portfolios
		WHERE entry_id = $1 AND deleted_at IS NULL
	`

	rows, err := r.db.QueryContext(ctx, query, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var portfolios []*models.CalcuttaPortfolio
	for rows.Next() {
		portfolio := &models.CalcuttaPortfolio{}
		var createdAt, updatedAt time.Time
		var deletedAt sql.NullTime

		err := rows.Scan(
			&portfolio.ID,
			&portfolio.EntryID,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		portfolio.Created = createdAt
		portfolio.Updated = updatedAt
		if deletedAt.Valid {
			portfolio.Deleted = &deletedAt.Time
		}

		portfolios = append(portfolios, portfolio)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return portfolios, nil
}

// UpdatePortfolio updates a portfolio
func (r *CalcuttaRepository) UpdatePortfolio(ctx context.Context, portfolio *models.CalcuttaPortfolio) error {
	query := `
		UPDATE calcutta_portfolios
		SET maximum_points = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		portfolio.MaximumPoints,
		portfolio.Updated,
		portfolio.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("portfolio not found")
	}

	return nil
}

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

	var rounds []*models.CalcuttaRound
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
			return nil, errors.New("entry not found")
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

// CreatePortfolio creates a new portfolio
func (r *CalcuttaRepository) CreatePortfolio(ctx context.Context, portfolio *models.CalcuttaPortfolio) error {
	query := `
		INSERT INTO calcutta_portfolios (
			id, entry_id, maximum_points, 
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	portfolio.ID = uuid.New().String()
	now := time.Now()
	portfolio.Created = now
	portfolio.Updated = now

	_, err := r.db.ExecContext(ctx, query,
		portfolio.ID,
		portfolio.EntryID,
		portfolio.MaximumPoints,
		portfolio.Created,
		portfolio.Updated,
	)

	return err
}

// CreatePortfolioTeam creates a new portfolio team
func (r *CalcuttaRepository) CreatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error {
	query := `
		INSERT INTO calcutta_portfolio_teams (
			id, portfolio_id, team_id, ownership_percentage, actual_points, 
			expected_points, predicted_points, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	team.ID = uuid.New().String()
	now := time.Now()
	team.Created = now
	team.Updated = now

	_, err := r.db.ExecContext(ctx, query,
		team.ID,
		team.PortfolioID,
		team.TeamID,
		team.OwnershipPercentage,
		team.ActualPoints,
		team.ExpectedPoints,
		team.PredictedPoints,
		team.Created,
		team.Updated,
	)

	return err
}

// GetPortfolios retrieves all portfolios for a Calcutta entry
func (r *CalcuttaRepository) GetPortfolios(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	query := `
		SELECT 
			id, entry_id, maximum_points,
			created_at, updated_at, deleted_at
		FROM calcutta_portfolios
		WHERE entry_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var portfolios []*models.CalcuttaPortfolio
	for rows.Next() {
		portfolio := &models.CalcuttaPortfolio{}
		var createdAt, updatedAt time.Time
		var deletedAt sql.NullTime

		err := rows.Scan(
			&portfolio.ID,
			&portfolio.EntryID,
			&portfolio.MaximumPoints,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		portfolio.Created = createdAt
		portfolio.Updated = updatedAt
		if deletedAt.Valid {
			portfolio.Deleted = &deletedAt.Time
		}

		portfolios = append(portfolios, portfolio)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return portfolios, nil
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
