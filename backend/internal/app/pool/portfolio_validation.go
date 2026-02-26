package pool

import (
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// ValidatePortfolio validates all investments for a portfolio according to the pool's rules.
// This is a pure function that can be tested without mocking repositories.
func ValidatePortfolio(pool *models.Pool, portfolio *models.Portfolio, investments []*models.Investment) error {
	// Rule 1: All investments are sealed until the tournament begins
	// This is handled at the API level, not in the service

	// Rule 2: Investors must invest in a minimum number of teams
	if len(investments) < pool.MinTeams {
		return fmt.Errorf("investors must invest in a minimum of %d teams", pool.MinTeams)
	}

	// Rule 3: Investors may invest in a maximum number of teams
	if len(investments) > pool.MaxTeams {
		return fmt.Errorf("investors may invest in a maximum of %d teams", pool.MaxTeams)
	}

	// Rule 4: Maximum investment in any single team
	for _, inv := range investments {
		if inv.Credits > pool.MaxInvestmentCredits {
			return fmt.Errorf("maximum investment in any single team is %d credits", pool.MaxInvestmentCredits)
		}
	}

	// Rule 5: Total investments cannot exceed budget
	totalCredits := 0
	for _, inv := range investments {
		totalCredits += inv.Credits
	}
	if totalCredits > pool.BudgetCredits {
		return fmt.Errorf("total investments cannot exceed budget of %d credits", pool.BudgetCredits)
	}

	// Rule 6: Minimum investment in any team is 1 credit
	for _, inv := range investments {
		if inv.Credits < 1 {
			return fmt.Errorf("minimum investment in any team is 1 credit")
		}
	}

	// Rule 7: Investors cannot invest in the same team twice
	teamInvestments := make(map[string]bool)
	for _, inv := range investments {
		if teamInvestments[inv.TeamID] {
			return fmt.Errorf("investors cannot invest in the same team twice")
		}
		teamInvestments[inv.TeamID] = true
	}

	return nil
}
