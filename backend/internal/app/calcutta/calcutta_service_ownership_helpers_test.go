package calcutta

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatOwnershipPercentageReturnsZeroWhenTotalBidsIsZero(t *testing.T) {
	// GIVEN
	bid := 10.0
	total := 0.0

	// WHEN
	got := ownershipPercentage(bid, total)

	// THEN
	if got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestThatTotalBidsForTeamSumsBidsAcrossEntries(t *testing.T) {
	// GIVEN
	allTeams := []*models.CalcuttaEntryTeam{
		{EntryID: "e1", TeamID: "t1", Bid: 10},
		{EntryID: "e2", TeamID: "t1", Bid: 15},
		{EntryID: "e3", TeamID: "t2", Bid: 99},
	}

	// WHEN
	got := totalBidsForTeam("t1", allTeams)

	// THEN
	if got != 25 {
		t.Fatalf("expected 25, got %v", got)
	}
}

func TestThatBidForTeamInEntryReturnsMatchingBid(t *testing.T) {
	// GIVEN
	allTeams := []*models.CalcuttaEntryTeam{
		{EntryID: "e1", TeamID: "t1", Bid: 10},
		{EntryID: "e2", TeamID: "t1", Bid: 15},
	}

	// WHEN
	got := bidForTeamInEntry("t1", "e2", allTeams)

	// THEN
	if got != 15 {
		t.Fatalf("expected 15, got %v", got)
	}
}

func TestThatCalculateOwnershipPercentageForPortfolioTeamUsesEntryBid(t *testing.T) {
	// GIVEN
	service := newTestCalcuttaService()
	portfolioTeam := &models.CalcuttaPortfolioTeam{ID: "pt1", PortfolioID: "portfolio1", TeamID: "teamA"}
	allTeams := []*models.CalcuttaEntryTeam{
		{ID: "et1", EntryID: "entry1", TeamID: "teamA", Bid: 20},
		{ID: "et2", EntryID: "entry2", TeamID: "teamA", Bid: 30},
		{ID: "et3", EntryID: "entry3", TeamID: "teamA", Bid: 50},
	}

	// WHEN
	percentage := service.CalculateOwnershipPercentage(context.Background(), portfolioTeam, allTeams)

	// THEN
	if percentage != 0.2 {
		t.Fatalf("expected 0.2, got %v", percentage)
	}
}
