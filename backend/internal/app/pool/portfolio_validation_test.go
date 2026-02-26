package pool

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func newTestPool() *models.Pool {
	return &models.Pool{
		ID:                   "pool-1",
		TournamentID:         "tournament-1",
		OwnerID:              "owner-1",
		CreatedBy:            "owner-1",
		Name:                 "Test Pool",
		MinTeams:             3,
		MaxTeams:             10,
		MaxInvestmentCredits: 50,
		BudgetCredits:        100,
		Visibility:           "private",
	}
}

func newTestInvestment(id, portfolioID, teamID string, credits int) *models.Investment {
	return &models.Investment{
		ID:          id,
		PortfolioID: portfolioID,
		TeamID:      teamID,
		Credits:     credits,
	}
}

func TestThatValidPortfolioPassesValidation(t *testing.T) {
	// GIVEN a valid portfolio with 3 investments and total credits of 90
	pool := newTestPool()
	portfolio := &models.Portfolio{ID: "p1", PoolID: pool.ID}
	investments := []*models.Investment{
		newTestInvestment("i1", "p1", "team1", 20),
		newTestInvestment("i2", "p1", "team2", 30),
		newTestInvestment("i3", "p1", "team3", 40),
	}

	// WHEN validating the portfolio
	err := ValidatePortfolio(pool, portfolio, investments)

	// THEN no error is returned
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestThatFewerThanThreeTeamsFailsValidation(t *testing.T) {
	// GIVEN a portfolio with only 2 investments
	pool := newTestPool()
	portfolio := &models.Portfolio{ID: "p1", PoolID: pool.ID}
	investments := []*models.Investment{
		newTestInvestment("i1", "p1", "team1", 20),
		newTestInvestment("i2", "p1", "team2", 30),
	}

	// WHEN validating the portfolio
	err := ValidatePortfolio(pool, portfolio, investments)

	// THEN an error is returned
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatMoreThanTenTeamsFailsValidation(t *testing.T) {
	// GIVEN a portfolio with 11 investments
	pool := newTestPool()
	portfolio := &models.Portfolio{ID: "p1", PoolID: pool.ID}
	investments := make([]*models.Investment, 11)
	for i := 0; i < 11; i++ {
		investments[i] = newTestInvestment("i"+string(rune('1'+i)), "p1", "team"+string(rune('1'+i)), 10)
	}

	// WHEN validating the portfolio
	err := ValidatePortfolio(pool, portfolio, investments)

	// THEN an error is returned
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatInvestmentOver50CreditsFailsValidation(t *testing.T) {
	// GIVEN a portfolio with an investment of 51 credits on one team
	pool := newTestPool()
	portfolio := &models.Portfolio{ID: "p1", PoolID: pool.ID}
	investments := []*models.Investment{
		newTestInvestment("i1", "p1", "team1", 51),
		newTestInvestment("i2", "p1", "team2", 20),
		newTestInvestment("i3", "p1", "team3", 20),
	}

	// WHEN validating the portfolio
	err := ValidatePortfolio(pool, portfolio, investments)

	// THEN an error is returned
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatTotalCreditsOver100FailsValidation(t *testing.T) {
	// GIVEN a portfolio with total credits of 102
	pool := newTestPool()
	portfolio := &models.Portfolio{ID: "p1", PoolID: pool.ID}
	investments := []*models.Investment{
		newTestInvestment("i1", "p1", "team1", 50),
		newTestInvestment("i2", "p1", "team2", 51),
		newTestInvestment("i3", "p1", "team3", 1),
	}

	// WHEN validating the portfolio
	err := ValidatePortfolio(pool, portfolio, investments)

	// THEN an error is returned
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatInvestmentUnder1CreditFailsValidation(t *testing.T) {
	// GIVEN a portfolio with an investment of 0 credits on one team
	pool := newTestPool()
	portfolio := &models.Portfolio{ID: "p1", PoolID: pool.ID}
	investments := []*models.Investment{
		newTestInvestment("i1", "p1", "team1", 20),
		newTestInvestment("i2", "p1", "team2", 30),
		newTestInvestment("i3", "p1", "team3", 0),
	}

	// WHEN validating the portfolio
	err := ValidatePortfolio(pool, portfolio, investments)

	// THEN an error is returned
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatDuplicateTeamInvestmentsFailsValidation(t *testing.T) {
	// GIVEN a portfolio with duplicate team IDs
	pool := newTestPool()
	portfolio := &models.Portfolio{ID: "p1", PoolID: pool.ID}
	investments := []*models.Investment{
		newTestInvestment("i1", "p1", "team1", 20),
		newTestInvestment("i2", "p1", "team2", 30),
		newTestInvestment("i3", "p1", "team1", 10),
	}

	// WHEN validating the portfolio
	err := ValidatePortfolio(pool, portfolio, investments)

	// THEN an error is returned
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
