package db

import (
	"context"
	"errors"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
)

func (r *CalcuttaRepository) GetPortfolio(ctx context.Context, id string) (*models.CalcuttaPortfolio, error) {
	row, err := r.q.GetPortfolioByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "portfolio", ID: id}
		}
		return nil, err
	}

	return &models.CalcuttaPortfolio{
		ID:            row.ID,
		EntryID:       row.EntryID,
		MaximumPoints: row.MaximumPoints,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		DeletedAt:     TimestamptzToPtrTime(row.DeletedAt),
	}, nil
}

func (r *CalcuttaRepository) GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error) {
	rows, err := r.q.ListPortfolioTeamsByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.CalcuttaPortfolioTeam, 0, len(rows))
	for _, row := range rows {
		pt := &models.CalcuttaPortfolioTeam{
			ID:                  row.ID,
			PortfolioID:         row.PortfolioID,
			TeamID:              row.TeamID,
			OwnershipPercentage: row.OwnershipPercentage,
			ActualPoints:        row.ActualPoints,
			ExpectedPoints:      row.ExpectedPoints,
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
			Eliminated:   row.Eliminated,
			CreatedAt:    row.TeamCreatedAt.Time,
			UpdatedAt:    row.TeamUpdatedAt.Time,
		}
		if row.SchoolName != nil {
			tt.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
		}
		pt.Team = tt

		out = append(out, pt)
	}
	return out, nil
}

func (r *CalcuttaRepository) GetPortfoliosByEntryIDs(ctx context.Context, entryIDs []string) (map[string][]*models.CalcuttaPortfolio, error) {
	if len(entryIDs) == 0 {
		return map[string][]*models.CalcuttaPortfolio{}, nil
	}

	rows, err := r.q.ListPortfoliosByEntryIDs(ctx, entryIDs)
	if err != nil {
		return nil, err
	}

	out := make(map[string][]*models.CalcuttaPortfolio, len(entryIDs))
	for _, row := range rows {
		out[row.EntryID] = append(out[row.EntryID], &models.CalcuttaPortfolio{
			ID:        row.ID,
			EntryID:   row.EntryID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
			DeletedAt: TimestamptzToPtrTime(row.DeletedAt),
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) GetPortfolioTeamsByPortfolioIDs(ctx context.Context, portfolioIDs []string) (map[string][]*models.CalcuttaPortfolioTeam, error) {
	if len(portfolioIDs) == 0 {
		return map[string][]*models.CalcuttaPortfolioTeam{}, nil
	}

	rows, err := r.q.ListPortfolioTeamsByPortfolioIDs(ctx, portfolioIDs)
	if err != nil {
		return nil, err
	}

	out := make(map[string][]*models.CalcuttaPortfolioTeam, len(portfolioIDs))
	for _, row := range rows {
		pt := &models.CalcuttaPortfolioTeam{
			ID:                  row.ID,
			PortfolioID:         row.PortfolioID,
			TeamID:              row.TeamID,
			OwnershipPercentage: row.OwnershipPercentage,
			ActualPoints:        row.ActualPoints,
			ExpectedPoints:      row.ExpectedPoints,
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
			Eliminated:   row.Eliminated,
			CreatedAt:    row.TeamCreatedAt.Time,
			UpdatedAt:    row.TeamUpdatedAt.Time,
		}
		if row.SchoolName != nil {
			tt.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
		}
		pt.Team = tt

		out[row.PortfolioID] = append(out[row.PortfolioID], pt)
	}
	return out, nil
}

func (r *CalcuttaRepository) GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	rows, err := r.q.ListPortfoliosByEntryID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.CalcuttaPortfolio, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.CalcuttaPortfolio{
			ID:        row.ID,
			EntryID:   row.EntryID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
			DeletedAt: TimestamptzToPtrTime(row.DeletedAt),
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	row, err := r.q.GetTeamByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "tournament team", ID: id}
		}
		return nil, err
	}

	team := &models.TournamentTeam{
		ID:           row.ID,
		TournamentID: row.TournamentID,
		SchoolID:     row.SchoolID,
		Seed:         int(row.Seed),
		Region:       row.Region,
		Byes:         int(row.Byes),
		Wins:         int(row.Wins),
		Eliminated:   row.Eliminated,
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
