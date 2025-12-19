package services

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

func TestThatPlayerPointsAreSummedCorrectlyForPortfolio(t *testing.T) {
	service := newTestCalcuttaService()
	portfolio := &models.CalcuttaPortfolio{ID: "portfolio1", EntryID: "entry1"}
	portfolioTeams := []*models.CalcuttaPortfolioTeam{
		{ID: "portfolioTeam1", PortfolioID: "portfolio1", TeamID: "team1", OwnershipPercentage: 0.25, ActualPoints: 125.0},
		{ID: "portfolioTeam2", PortfolioID: "portfolio1", TeamID: "team2", OwnershipPercentage: 0.15, ActualPoints: 0.0},
		{ID: "portfolioTeam3", PortfolioID: "portfolio1", TeamID: "team3", OwnershipPercentage: 0.1, ActualPoints: 15.0},
		{ID: "portfolioTeam4", PortfolioID: "portfolio2", TeamID: "team4", OwnershipPercentage: 0.5, ActualPoints: 500.0},
	}

	points := service.CalculatePlayerPoints(portfolio, portfolioTeams)
	if points != 140.0 {
		t.Fatalf("expected 140.0, got %v", points)
	}
}
