package calcutta

import (
	"errors"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// ValidateEntry validates all bids for an entry according to the rules
func (s *Service) ValidateEntry(entry *models.CalcuttaEntry, teams []*models.CalcuttaEntryTeam) error {
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
	totalBids := 0.0
	for _, team := range teams {
		totalBids += float64(team.Bid)
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
