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
			return nil, &NotFoundError{Resource: "portfolio", ID: id}
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

	teams := make([]*models.CalcuttaPortfolioTeam, 0)
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
		return &NotFoundError{Resource: "portfolio team", ID: team.ID}
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

	portfolios := make([]*models.CalcuttaPortfolio, 0)
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
		return &NotFoundError{Resource: "portfolio", ID: portfolio.ID}
	}

	return nil
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
