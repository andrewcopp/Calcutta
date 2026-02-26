package testutil

import "github.com/andrewcopp/Calcutta/backend/internal/models"

// NewCalcutta returns a fully-populated Calcutta with sensible defaults.
func NewCalcutta() *models.Calcutta {
	return &models.Calcutta{
		ID:           "calcutta-1",
		TournamentID: "tournament-1",
		OwnerID:      "owner-1",
		CreatedBy:    "owner-1",
		Name:         "Test Calcutta",
		MinTeams:     3,
		MaxTeams:     10,
		MaxBidPoints: 50,
		BudgetPoints: 100,
		Visibility:   "private",
		CreatedAt:    DefaultTime,
		UpdatedAt:    DefaultTime,
	}
}

// NewEntry returns a fully-populated CalcuttaEntry with sensible defaults.
func NewEntry() *models.CalcuttaEntry {
	return &models.CalcuttaEntry{
		ID:         "entry-1",
		Name:       "Test Entry",
		UserID:     StringPtr("user-1"),
		CalcuttaID: "calcutta-1",
		Status:     "submitted",
		CreatedAt:  DefaultTime,
		UpdatedAt:  DefaultTime,
	}
}

// NewEntryTeam returns a fully-populated CalcuttaEntryTeam with sensible defaults.
func NewEntryTeam() *models.CalcuttaEntryTeam {
	return &models.CalcuttaEntryTeam{
		ID:        "entry-team-1",
		EntryID:   "entry-1",
		TeamID:    "team-1",
		BidPoints: 10,
		CreatedAt: DefaultTime,
		UpdatedAt: DefaultTime,
	}
}

// NewPayout returns a fully-populated CalcuttaPayout with sensible defaults.
func NewPayout() *models.CalcuttaPayout {
	return &models.CalcuttaPayout{
		ID:          "payout-1",
		CalcuttaID:  "calcutta-1",
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
		CalcuttaID:    "calcutta-1",
		WinIndex:      1,
		PointsAwarded: 1,
		CreatedAt:     DefaultTime,
		UpdatedAt:     DefaultTime,
	}
}

// NewInvitation returns a fully-populated CalcuttaInvitation with sensible defaults.
func NewInvitation() *models.CalcuttaInvitation {
	return &models.CalcuttaInvitation{
		ID:         "invitation-1",
		CalcuttaID: "calcutta-1",
		UserID:     "user-1",
		InvitedBy:  "owner-1",
		Status:     "pending",
		CreatedAt:  DefaultTime,
		UpdatedAt:  DefaultTime,
	}
}
