package calcutta

import "github.com/andrewcopp/Calcutta/backend/internal/models"

// CalculatePoints calculates the points earned by a team based on its performance
func (s *Service) CalculatePoints(team *models.TournamentTeam, rounds []*models.CalcuttaRound) float64 {
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
	if totalProgress <= 0 {
		return 0
	}

	points := 0
	for _, r := range rounds {
		if r == nil {
			continue
		}
		if r.Round <= totalProgress {
			points += r.Points
		}
	}

	return float64(points)
}

// CalculatePlayerPoints calculates the points earned by a player based on their team ownerships
func (s *Service) CalculatePlayerPoints(portfolio *models.CalcuttaPortfolio, portfolioTeams []*models.CalcuttaPortfolioTeam) float64 {
	totalPoints := 0.0
	for _, portfolioTeam := range portfolioTeams {
		if portfolioTeam.PortfolioID == portfolio.ID {
			totalPoints += portfolioTeam.ActualPoints
		}
	}
	return totalPoints
}
