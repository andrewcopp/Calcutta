package calcutta

import (
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// DeriveEntryStatus computes the entry status from its teams.
// Only two states: "incomplete" (no bids) and "accepted" (has bids, validated at save time).
func DeriveEntryStatus(teams []*models.CalcuttaEntryTeam) string {
	if len(teams) == 0 {
		return "incomplete"
	}
	return "accepted"
}

// ValidateEntry validates all bids for an entry according to the calcutta's rules.
// This is a pure function that can be tested without mocking repositories.
func ValidateEntry(calcutta *models.Calcutta, entry *models.CalcuttaEntry, teams []*models.CalcuttaEntryTeam) error {
	// Rule 1: All bids are sealed until the tournament begins
	// This is handled at the API level, not in the service

	// Rule 2: Players must bid on a minimum number of teams
	if len(teams) < calcutta.MinTeams {
		return fmt.Errorf("players must bid on a minimum of %d teams", calcutta.MinTeams)
	}

	// Rule 3: Players may bid on a maximum number of teams
	if len(teams) > calcutta.MaxTeams {
		return fmt.Errorf("players may bid on a maximum of %d teams", calcutta.MaxTeams)
	}

	// Rule 4: Maximum bid on any single team
	for _, team := range teams {
		if team.BidPoints > calcutta.MaxBidPoints {
			return fmt.Errorf("maximum bid on any single team is %d points", calcutta.MaxBidPoints)
		}
	}

	// Rule 5: Total bids cannot exceed starting budget
	totalBids := 0
	for _, team := range teams {
		totalBids += team.BidPoints
	}
	if totalBids > calcutta.BudgetPoints {
		return fmt.Errorf("total bids cannot exceed starting budget of %d points", calcutta.BudgetPoints)
	}

	// Rule 6: Minimum bid on any team is $1
	for _, team := range teams {
		if team.BidPoints < 1 {
			return fmt.Errorf("minimum bid on any team is 1 point")
		}
	}

	// Rule 7: Players cannot bid on the same team multiple times
	teamBids := make(map[string]bool)
	for _, team := range teams {
		if teamBids[team.TeamID] {
			return fmt.Errorf("players cannot bid on the same team multiple times")
		}
		teamBids[team.TeamID] = true
	}

	return nil
}
