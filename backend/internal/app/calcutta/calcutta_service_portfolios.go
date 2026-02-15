package calcutta

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// CalculateTotalBids calculates the total bids for a Calcutta entry
func (s *Service) CalculateTotalBids(ctx context.Context, entryID string) (float64, error) {
	teams, err := s.ports.EntryReader.GetEntryTeams(ctx, entryID)
	if err != nil {
		return 0, err
	}

	var totalBids float64
	for _, team := range teams {
		totalBids += float64(team.Bid)
	}

	return totalBids, nil
}

// GetPortfolioTeams retrieves all teams for a portfolio
func (s *Service) GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error) {
	return s.ports.PortfolioReader.GetPortfolioTeams(ctx, portfolioID)
}
