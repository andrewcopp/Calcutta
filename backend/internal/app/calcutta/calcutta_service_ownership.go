package calcutta

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func totalBidsForTeam(teamID string, allTeams []*models.CalcuttaEntryTeam) float64 {
	totalBids := 0.0
	for _, t := range allTeams {
		if t.TeamID == teamID {
			totalBids += float64(t.Bid)
		}
	}
	return totalBids
}

func bidForTeamInEntry(teamID string, entryID string, allTeams []*models.CalcuttaEntryTeam) float64 {
	for _, t := range allTeams {
		if t.TeamID == teamID && t.EntryID == entryID {
			return float64(t.Bid)
		}
	}
	return 0
}

func ownershipPercentage(bid float64, totalBids float64) float64 {
	if totalBids == 0 {
		return 0
	}
	return bid / totalBids
}

// CalculateOwnershipPercentage calculates the ownership percentage for a team
func (s *Service) CalculateOwnershipPercentage(ctx context.Context, team interface{}, allTeams []*models.CalcuttaEntryTeam) float64 {
	// Rule 8: Ownership percentage = (Player's bid on team) ÷ (Total bids on team)
	var teamID string
	var bid float64

	// Extract team ID and bid based on type
	switch t := team.(type) {
	case *models.CalcuttaEntryTeam:
		teamID = t.TeamID
		bid = float64(t.Bid)
	case *models.CalcuttaPortfolioTeam:
		teamID = t.TeamID
		// For portfolio teams, we need to find the corresponding entry team to get the bid
		// First, get the portfolio to find the entry ID
		if s.ports.PortfolioReader == nil {
			return 0
		}
		portfolio, err := s.ports.PortfolioReader.GetPortfolio(ctx, t.PortfolioID)
		if err != nil {
			return 0
		}
		// Now find the specific entry team for this entry
		bid = bidForTeamInEntry(teamID, portfolio.EntryID, allTeams)
	default:
		return 0
	}

	totalBids := totalBidsForTeam(teamID, allTeams)
	return ownershipPercentage(bid, totalBids)
}
