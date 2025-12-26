package calcutta

import (
	"context"
	"log"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// CalculateOwnershipPercentage calculates the ownership percentage for a team
func (s *Service) CalculateOwnershipPercentage(ctx context.Context, team interface{}, allTeams []*models.CalcuttaEntryTeam) float64 {
	// Rule 8: Ownership percentage = (Player's bid on team) ÷ (Total bids on team)
	totalBids := 0.0
	var teamID string
	var bid float64

	// Extract team ID and bid based on type
	switch t := team.(type) {
	case *models.CalcuttaEntryTeam:
		teamID = t.TeamID
		bid = float64(t.Bid)
		log.Printf("Entry team bid: %f for team %s", bid, teamID)
	case *models.CalcuttaPortfolioTeam:
		teamID = t.TeamID
		// For portfolio teams, we need to find the corresponding entry team to get the bid
		// First, get the portfolio to find the entry ID
		portfolio, err := s.ports.PortfolioReader.GetPortfolio(ctx, t.PortfolioID)
		if err != nil {
			log.Printf("Error getting portfolio: %v", err)
			return 0
		}
		// Now find the specific entry team for this entry
		for _, et := range allTeams {
			if et.TeamID == t.TeamID && et.EntryID == portfolio.EntryID {
				bid = float64(et.Bid)
				log.Printf("Found matching entry team with bid: %f for team %s", bid, teamID)
				break
			}
		}
	default:
		log.Printf("Unknown team type")
		return 0
	}

	// Calculate total bids for this team
	for _, t := range allTeams {
		if t.TeamID == teamID {
			totalBids += float64(t.Bid)
			log.Printf("Adding bid %f to total bids for team %s", float64(t.Bid), teamID)
		}
	}

	if totalBids == 0 {
		log.Printf("Total bids is 0 for team %s", teamID)
		return 0
	}

	ownershipPercentage := bid / totalBids
	log.Printf("Final ownership percentage: %f (bid: %f / total bids: %f) for team %s", ownershipPercentage, bid, totalBids, teamID)
	return ownershipPercentage
}
