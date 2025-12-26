package calcutta

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

func TestThatOwnershipPercentageIsCalculatedCorrectlyWithMultipleBids(t *testing.T) {
	service := newTestCalcuttaService()
	team := &models.CalcuttaEntryTeam{ID: "team1", EntryID: "entry1", TeamID: "teamA", Bid: 20}
	allTeams := []*models.CalcuttaEntryTeam{
		{ID: "team1", EntryID: "entry1", TeamID: "teamA", Bid: 20},
		{ID: "team2", EntryID: "entry2", TeamID: "teamA", Bid: 30},
		{ID: "team3", EntryID: "entry3", TeamID: "teamA", Bid: 50},
	}

	percentage := service.CalculateOwnershipPercentage(context.Background(), team, allTeams)
	if percentage != 0.2 {
		t.Fatalf("expected 0.2, got %v", percentage)
	}
}

func TestThatOwnershipPercentageIs100PercentWithSingleBid(t *testing.T) {
	service := newTestCalcuttaService()
	team := &models.CalcuttaEntryTeam{ID: "team1", EntryID: "entry1", TeamID: "teamB", Bid: 20}
	allTeams := []*models.CalcuttaEntryTeam{{ID: "team1", EntryID: "entry1", TeamID: "teamB", Bid: 20}}

	percentage := service.CalculateOwnershipPercentage(context.Background(), team, allTeams)
	if percentage != 1.0 {
		t.Fatalf("expected 1.0, got %v", percentage)
	}
}
