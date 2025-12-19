package services

import "github.com/andrewcopp/Calcutta/backend/pkg/models"

// CalculatePoints calculates the points earned by a team based on its performance
func (s *CalcuttaService) CalculatePoints(team *models.TournamentTeam, rounds []*models.CalcuttaRound) float64 {
	// Calculate total progress (wins + byes)
	totalProgress := team.Wins + team.Byes

	// Map total progress to points
	// 0 wins/byes = First Four Loss (0 points)
	// 1 win/bye = First Four Win or Bye (0 points)
	// 2 wins/byes = First Round Win (50 points)
	// 3 wins/byes = Round of 32 Win (150 points)
	// 4 wins/byes = Sweet 16 Win (300 points)
	// 5 wins/byes = Elite 8 Win (500 points)
	// 6 wins/byes = Final Four Win (750 points)
	// 7 wins/byes = Championship Game Win (1050 points)
	switch totalProgress {
	case 0:
		return 0
	case 1:
		return 0
	case 2:
		return 50 // First Round Winner
	case 3:
		return 150 // Round of 32 Winner
	case 4:
		return 300 // Sweet 16 Winner
	case 5:
		return 500 // Elite 8 Winner
	case 6:
		return 750 // Final Four Winner
	case 7:
		return 1050 // Tournament Winner
	default:
		// For any other value (shouldn't happen in a normal tournament)
		return 0
	}
}

// CalculatePlayerPoints calculates the points earned by a player based on their team ownerships
func (s *CalcuttaService) CalculatePlayerPoints(portfolio *models.CalcuttaPortfolio, portfolioTeams []*models.CalcuttaPortfolioTeam) float64 {
	totalPoints := 0.0
	for _, portfolioTeam := range portfolioTeams {
		if portfolioTeam.PortfolioID == portfolio.ID {
			totalPoints += portfolioTeam.ActualPoints
		}
	}
	return totalPoints
}
