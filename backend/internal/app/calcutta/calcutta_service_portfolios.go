package calcutta

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// GetPortfolioTeams retrieves all teams for a portfolio
func (s *Service) GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error) {
	return s.ports.PortfolioReader.GetPortfolioTeams(ctx, portfolioID)
}
