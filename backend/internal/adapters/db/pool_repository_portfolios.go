package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PoolRepository) CreatePortfolio(ctx context.Context, portfolio *models.Portfolio) error {
	portfolio.ID = uuid.New().String()
	now := time.Now()
	portfolio.CreatedAt = now
	portfolio.UpdatedAt = now

	var userID pgtype.UUID
	if portfolio.UserID != nil {
		parsed, err := uuid.Parse(*portfolio.UserID)
		if err != nil {
			return fmt.Errorf("parsing user ID for portfolio: %w", err)
		}
		userID = pgtype.UUID{Bytes: parsed, Valid: true}
	}

	params := sqlc.CreatePortfolioParams{
		ID:     portfolio.ID,
		Name:   portfolio.Name,
		UserID: userID,
		PoolID: portfolio.PoolID,
	}
	if err := r.q.CreatePortfolio(ctx, params); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &apperrors.AlreadyExistsError{Resource: "portfolio", Field: "user_id"}
		}
		return fmt.Errorf("creating portfolio: %w", err)
	}
	return nil
}

func (r *PoolRepository) GetPortfolios(ctx context.Context, poolID string) ([]*models.Portfolio, map[string]float64, error) {
	rows, err := r.q.ListPortfoliosByPoolID(ctx, poolID)
	if err != nil {
		return nil, nil, fmt.Errorf("listing portfolios for pool %s: %w", poolID, err)
	}

	portfolios := make([]*models.Portfolio, 0, len(rows))
	returnsByPortfolio := make(map[string]float64, len(rows))
	for _, row := range rows {
		portfolios = append(portfolios, &models.Portfolio{
			ID:        row.ID,
			Name:      row.Name,
			UserID:    uuidToPtrString(row.UserID),
			PoolID:    row.PoolID,
			Status:    row.Status,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
			DeletedAt: TimestamptzToPtrTime(row.DeletedAt),
		})
		returnsByPortfolio[row.ID] = row.TotalReturns
	}
	return portfolios, returnsByPortfolio, nil
}

func (r *PoolRepository) GetPortfolio(ctx context.Context, id string) (*models.Portfolio, error) {
	row, err := r.q.GetPortfolioByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "portfolio", ID: id}
		}
		return nil, fmt.Errorf("getting portfolio %s: %w", id, err)
	}

	return &models.Portfolio{
		ID:        row.ID,
		Name:      row.Name,
		UserID:    uuidToPtrString(row.UserID),
		PoolID:    row.PoolID,
		Status:    row.Status,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		DeletedAt: TimestamptzToPtrTime(row.DeletedAt),
	}, nil
}

func (r *PoolRepository) GetInvestments(ctx context.Context, portfolioID string) ([]*models.Investment, error) {
	rows, err := r.q.ListInvestmentsByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("listing investments for portfolio %s: %w", portfolioID, err)
	}

	out := make([]*models.Investment, 0, len(rows))
	for _, row := range rows {
		inv := &models.Investment{
			ID:          row.ID,
			PortfolioID: row.PortfolioID,
			TeamID:      row.TeamID,
			Credits:     int(row.Credits),
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
			DeletedAt:   TimestamptzToPtrTime(row.DeletedAt),
		}

		tt := &models.TournamentTeam{
			ID:           row.TournamentTeamID,
			SchoolID:     row.SchoolID,
			TournamentID: row.TournamentID,
			Seed:         int(row.Seed),
			Region:       row.Region,
			Byes:         int(row.Byes),
			Wins:         int(row.Wins),
			CreatedAt:    row.TeamCreatedAt.Time,
			UpdatedAt:    row.TeamUpdatedAt.Time,
			DeletedAt:    TimestamptzToPtrTime(row.TeamDeletedAt),
		}
		if row.SchoolName != nil {
			tt.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
		}
		inv.Team = tt

		out = append(out, inv)
	}
	return out, nil
}

func (r *PoolRepository) GetInvestmentsByPortfolioIDs(ctx context.Context, portfolioIDs []string) (map[string][]*models.Investment, error) {
	if len(portfolioIDs) == 0 {
		return map[string][]*models.Investment{}, nil
	}

	rows, err := r.q.ListInvestmentsByPortfolioIDs(ctx, portfolioIDs)
	if err != nil {
		return nil, fmt.Errorf("listing investments by portfolio IDs: %w", err)
	}

	out := make(map[string][]*models.Investment, len(portfolioIDs))
	for _, row := range rows {
		inv := &models.Investment{
			ID:          row.ID,
			PortfolioID: row.PortfolioID,
			TeamID:      row.TeamID,
			Credits:     int(row.Credits),
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
			DeletedAt:   TimestamptzToPtrTime(row.DeletedAt),
		}

		tt := &models.TournamentTeam{
			ID:           row.TournamentTeamID,
			SchoolID:     row.SchoolID,
			TournamentID: row.TournamentID,
			Seed:         int(row.Seed),
			Region:       row.Region,
			Byes:         int(row.Byes),
			Wins:         int(row.Wins),
			CreatedAt:    row.TeamCreatedAt.Time,
			UpdatedAt:    row.TeamUpdatedAt.Time,
			DeletedAt:    TimestamptzToPtrTime(row.TeamDeletedAt),
		}
		if row.SchoolName != nil {
			tt.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
		}
		inv.Team = tt

		out[row.PortfolioID] = append(out[row.PortfolioID], inv)
	}
	return out, nil
}

func (r *PoolRepository) UpdatePortfolioStatus(ctx context.Context, id string, status string) error {
	if err := r.q.UpdatePortfolioStatus(ctx, sqlc.UpdatePortfolioStatusParams{ID: id, Status: status}); err != nil {
		return fmt.Errorf("updating portfolio status for %s: %w", id, err)
	}
	return nil
}

func (r *PoolRepository) ReplaceInvestments(ctx context.Context, portfolioID string, investments []*models.Investment) error {
	// Validate that portfolio exists (access control handled at higher layers)
	if _, err := r.GetPortfolio(ctx, portfolioID); err != nil {
		return err
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction to replace investments for portfolio %s: %w", portfolioID, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	now := time.Now()

	if _, err = qtx.SoftDeleteInvestmentsByPortfolioID(ctx, sqlc.SoftDeleteInvestmentsByPortfolioIDParams{
		DeletedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		PortfolioID: portfolioID,
	}); err != nil {
		return fmt.Errorf("soft-deleting investments for portfolio %s: %w", portfolioID, err)
	}

	for _, inv := range investments {
		if inv == nil {
			continue
		}
		id := uuid.New().String()
		params := sqlc.CreateInvestmentParams{
			ID:          id,
			PortfolioID: portfolioID,
			TeamID:      inv.TeamID,
			Credits:     int32(inv.Credits),
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		}
		if err = qtx.CreateInvestment(ctx, params); err != nil {
			return fmt.Errorf("creating investment for portfolio %s: %w", portfolioID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction to replace investments for portfolio %s: %w", portfolioID, err)
	}
	return nil
}
