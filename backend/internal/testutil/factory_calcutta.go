package testutil

import "github.com/andrewcopp/Calcutta/backend/internal/models"

// NewPool returns a fully-populated Pool with sensible defaults.
func NewPool() *models.Pool {
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
		CreatedAt:            DefaultTime,
		UpdatedAt:            DefaultTime,
	}
}

// NewPortfolio returns a fully-populated Portfolio with sensible defaults.
func NewPortfolio() *models.Portfolio {
	return &models.Portfolio{
		ID:        "portfolio-1",
		Name:      "Test Portfolio",
		UserID:    StringPtr("user-1"),
		PoolID:    "pool-1",
		CreatedAt: DefaultTime,
		UpdatedAt: DefaultTime,
	}
}

// NewInvestment returns a fully-populated Investment with sensible defaults.
func NewInvestment() *models.Investment {
	return &models.Investment{
		ID:          "investment-1",
		PortfolioID: "portfolio-1",
		TeamID:      "team-1",
		Credits:     10,
		CreatedAt:   DefaultTime,
		UpdatedAt:   DefaultTime,
	}
}

// NewPayout returns a fully-populated PoolPayout with sensible defaults.
func NewPayout() *models.PoolPayout {
	return &models.PoolPayout{
		ID:          "payout-1",
		PoolID:      "pool-1",
		Position:    1,
		AmountCents: 100,
		CreatedAt:   DefaultTime,
		UpdatedAt:   DefaultTime,
	}
}

// NewScoringRule returns a fully-populated ScoringRule with sensible defaults.
func NewScoringRule() *models.ScoringRule {
	return &models.ScoringRule{
		ID:            "scoring-rule-1",
		PoolID:        "pool-1",
		WinIndex:      1,
		PointsAwarded: 1,
		CreatedAt:     DefaultTime,
		UpdatedAt:     DefaultTime,
	}
}

// NewInvitation returns a fully-populated PoolInvitation with sensible defaults.
func NewInvitation() *models.PoolInvitation {
	return &models.PoolInvitation{
		ID:        "invitation-1",
		PoolID:    "pool-1",
		UserID:    "user-1",
		InvitedBy: "owner-1",
		Status:    "pending",
		CreatedAt: DefaultTime,
		UpdatedAt: DefaultTime,
	}
}
