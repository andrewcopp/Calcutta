package calcutta

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatDeriveEntryStatusReturnsIncompleteWhenNoTeams(t *testing.T) {
	// GIVEN an empty slice of teams
	teams := []*models.CalcuttaEntryTeam{}

	// WHEN deriving the entry status
	status := DeriveEntryStatus(teams)

	// THEN the status is "incomplete"
	if status != "incomplete" {
		t.Errorf("expected 'incomplete', got %q", status)
	}
}

func TestThatDeriveEntryStatusReturnsAcceptedWhenTeamsExist(t *testing.T) {
	// GIVEN a non-empty slice of teams
	teams := []*models.CalcuttaEntryTeam{
		{ID: "team1", EntryID: "entry1", TeamID: "team1", BidPoints: 20},
	}

	// WHEN deriving the entry status
	status := DeriveEntryStatus(teams)

	// THEN the status is "accepted"
	if status != "accepted" {
		t.Errorf("expected 'accepted', got %q", status)
	}
}

func TestThatDeriveEntryStatusReturnsIncompleteWhenTeamsIsNil(t *testing.T) {
	// GIVEN a nil slice of teams
	var teams []*models.CalcuttaEntryTeam

	// WHEN deriving the entry status
	status := DeriveEntryStatus(teams)

	// THEN the status is "incomplete"
	if status != "incomplete" {
		t.Errorf("expected 'incomplete', got %q", status)
	}
}
