package services

import (
	"errors"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// CalcuttaService handles business logic for Calcutta auctions
type CalcuttaService struct {
	// In a real implementation, this would have repositories for data access
}

// NewCalcuttaService creates a new CalcuttaService
func NewCalcuttaService() *CalcuttaService {
	return &CalcuttaService{}
}

// ValidateEntry validates all bids for an entry according to the rules
func (s *CalcuttaService) ValidateEntry(entry *models.CalcuttaEntry, teams []*models.CalcuttaEntryTeam) error {
	// Rule 1: All bids are sealed until the tournament begins
	// This is handled at the API level, not in the service

	// Rule 2: Players must bid on a minimum of 3 teams
	if len(teams) < 3 {
		return errors.New("players must bid on a minimum of 3 teams")
	}

	// Rule 3: Players may bid on a maximum of 10 teams
	if len(teams) > 10 {
		return errors.New("players may bid on a maximum of 10 teams")
	}

	// Rule 4: Maximum bid on any single team is $50
	for _, team := range teams {
		if team.Bid > 50 {
			return errors.New("maximum bid on any single team is $50")
		}
	}

	// Rule 5: Total bids cannot exceed starting budget of $100
	totalBids := 0
	for _, team := range teams {
		totalBids += team.Bid
	}
	if totalBids > 100 {
		return errors.New("total bids cannot exceed starting budget of $100")
	}

	// Rule 6: Minimum bid on any team is $1
	for _, team := range teams {
		if team.Bid < 1 {
			return errors.New("minimum bid on any team is $1")
		}
	}

	// Rule 7: Players cannot bid on the same team multiple times
	teamBids := make(map[string]bool)
	for _, team := range teams {
		if teamBids[team.TeamID] {
			return errors.New("players cannot bid on the same team multiple times")
		}
		teamBids[team.TeamID] = true
	}

	return nil
}

// CalculateOwnershipPercentage calculates the ownership percentage for a team
func (s *CalcuttaService) CalculateOwnershipPercentage(team *models.CalcuttaEntryTeam, allTeams []*models.CalcuttaEntryTeam) float64 {
	// Rule 8: Ownership percentage = (Player's bid on team) ÷ (Total bids on team)
	totalBids := 0
	for _, t := range allTeams {
		if t.TeamID == team.TeamID {
			totalBids += t.Bid
		}
	}

	if totalBids == 0 {
		return 0
	}

	return float64(team.Bid) / float64(totalBids)
}

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
			totalPoints += portfolioTeam.PointsEarned
		}
	}
	return totalPoints
}
