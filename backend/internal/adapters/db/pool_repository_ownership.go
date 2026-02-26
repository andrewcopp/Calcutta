package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
)

func (r *PoolRepository) GetOwnershipSummary(ctx context.Context, id string) (*models.OwnershipSummary, error) {
	row, err := r.q.GetOwnershipSummaryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "ownership summary", ID: id}
		}
		return nil, fmt.Errorf("getting ownership summary %s: %w", id, err)
	}

	return &models.OwnershipSummary{
		ID:             row.ID,
		PortfolioID:    row.PortfolioID,
		MaximumReturns: row.MaximumReturns,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		DeletedAt:      TimestamptzToPtrTime(row.DeletedAt),
	}, nil
}

func (r *PoolRepository) GetOwnershipDetails(ctx context.Context, portfolioID string) ([]*models.OwnershipDetail, error) {
	rows, err := r.q.ListOwnershipDetailsByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("listing ownership details for portfolio %s: %w", portfolioID, err)
	}

	out := make([]*models.OwnershipDetail, 0, len(rows))
	for _, row := range rows {
		od := &models.OwnershipDetail{
			ID:                  row.ID,
			PortfolioID:         row.PortfolioID,
			TeamID:              row.TeamID,
			OwnershipPercentage: row.OwnershipPercentage,
			ActualReturns:       row.ActualReturns,
			ExpectedReturns:     row.ExpectedReturns,
			CreatedAt:           row.CreatedAt.Time,
			UpdatedAt:           row.UpdatedAt.Time,
			DeletedAt:           TimestamptzToPtrTime(row.DeletedAt),
		}

		tt := &models.TournamentTeam{
			ID:           row.TournamentTeamID,
			SchoolID:     row.SchoolID,
			TournamentID: row.TournamentID,
			Seed:         int(row.Seed),
			Region:       row.Region,
			Byes:         int(row.Byes),
			Wins:         int(row.Wins),
			IsEliminated: row.IsEliminated,
			CreatedAt:    row.TeamCreatedAt.Time,
			UpdatedAt:    row.TeamUpdatedAt.Time,
		}
		if row.SchoolName != nil {
			tt.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
		}
		od.Team = tt

		out = append(out, od)
	}
	return out, nil
}

func (r *PoolRepository) GetOwnershipSummariesByPortfolio(ctx context.Context, portfolioID string) ([]*models.OwnershipSummary, error) {
	rows, err := r.q.ListOwnershipSummariesByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("listing ownership summaries for portfolio %s: %w", portfolioID, err)
	}

	out := make([]*models.OwnershipSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.OwnershipSummary{
			ID:          row.ID,
			PortfolioID: row.PortfolioID,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
			DeletedAt:   TimestamptzToPtrTime(row.DeletedAt),
		})
	}
	return out, nil
}

func (r *PoolRepository) GetOwnershipSummariesByPortfolioIDs(ctx context.Context, portfolioIDs []string) (map[string][]*models.OwnershipSummary, error) {
	if len(portfolioIDs) == 0 {
		return map[string][]*models.OwnershipSummary{}, nil
	}

	rows, err := r.q.ListOwnershipSummariesByPortfolioIDs(ctx, portfolioIDs)
	if err != nil {
		return nil, fmt.Errorf("listing ownership summaries by portfolio IDs: %w", err)
	}

	out := make(map[string][]*models.OwnershipSummary, len(portfolioIDs))
	for _, row := range rows {
		out[row.PortfolioID] = append(out[row.PortfolioID], &models.OwnershipSummary{
			ID:          row.ID,
			PortfolioID: row.PortfolioID,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
			DeletedAt:   TimestamptzToPtrTime(row.DeletedAt),
		})
	}
	return out, nil
}

func (r *PoolRepository) GetOwnershipDetailsByPortfolioIDs(ctx context.Context, portfolioIDs []string) (map[string][]*models.OwnershipDetail, error) {
	if len(portfolioIDs) == 0 {
		return map[string][]*models.OwnershipDetail{}, nil
	}

	rows, err := r.q.ListOwnershipDetailsByPortfolioIDs(ctx, portfolioIDs)
	if err != nil {
		return nil, fmt.Errorf("listing ownership details by portfolio IDs: %w", err)
	}

	out := make(map[string][]*models.OwnershipDetail, len(portfolioIDs))
	for _, row := range rows {
		od := &models.OwnershipDetail{
			ID:                  row.ID,
			PortfolioID:         row.PortfolioID,
			TeamID:              row.TeamID,
			OwnershipPercentage: row.OwnershipPercentage,
			ActualReturns:       row.ActualReturns,
			ExpectedReturns:     row.ExpectedReturns,
			CreatedAt:           row.CreatedAt.Time,
			UpdatedAt:           row.UpdatedAt.Time,
			DeletedAt:           TimestamptzToPtrTime(row.DeletedAt),
		}

		tt := &models.TournamentTeam{
			ID:           row.TournamentTeamID,
			SchoolID:     row.SchoolID,
			TournamentID: row.TournamentID,
			Seed:         int(row.Seed),
			Region:       row.Region,
			Byes:         int(row.Byes),
			Wins:         int(row.Wins),
			IsEliminated: row.IsEliminated,
			CreatedAt:    row.TeamCreatedAt.Time,
			UpdatedAt:    row.TeamUpdatedAt.Time,
		}
		if row.SchoolName != nil {
			tt.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
		}
		od.Team = tt

		out[row.PortfolioID] = append(out[row.PortfolioID], od)
	}
	return out, nil
}

func (r *PoolRepository) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	row, err := r.q.GetTeamByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "tournament team", ID: id}
		}
		return nil, fmt.Errorf("getting tournament team %s: %w", id, err)
	}

	team := &models.TournamentTeam{
		ID:           row.ID,
		TournamentID: row.TournamentID,
		SchoolID:     row.SchoolID,
		Seed:         int(row.Seed),
		Region:       row.Region,
		Byes:         int(row.Byes),
		Wins:         int(row.Wins),
		IsEliminated: row.IsEliminated,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
	if row.NetRtg != nil || row.ORtg != nil || row.DRtg != nil || row.AdjT != nil {
		team.KenPom = &models.KenPomStats{NetRtg: row.NetRtg, ORtg: row.ORtg, DRtg: row.DRtg, AdjT: row.AdjT}
	}
	if row.SchoolName != nil {
		team.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
	}
	return team, nil
}
