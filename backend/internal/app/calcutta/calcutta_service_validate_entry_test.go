package calcutta

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatValidEntryPassesValidation(t *testing.T) {
	// GIVEN a valid entry with 3 teams and total bids of 90
	userID := "user1"
	entry := &models.CalcuttaEntry{
		ID:         "entry1",
		Name:       "Test User",
		UserID:     &userID,
		CalcuttaID: "calcutta1",
	}
	teams := []*models.CalcuttaEntryTeam{
		{ID: "team1", EntryID: "entry1", TeamID: "team1", Bid: 20},
		{ID: "team2", EntryID: "entry1", TeamID: "team2", Bid: 30},
		{ID: "team3", EntryID: "entry1", TeamID: "team3", Bid: 40},
	}

	// WHEN validating the entry
	err := ValidateEntry(entry, teams)

	// THEN no error is returned
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestThatFewerThanThreeTeamsFailsValidation(t *testing.T) {
	// GIVEN an entry with only 2 teams
	userID := "user1"
	entry := &models.CalcuttaEntry{
		ID:         "entry1",
		Name:       "Test User",
		UserID:     &userID,
		CalcuttaID: "calcutta1",
	}
	teams := []*models.CalcuttaEntryTeam{
		{ID: "team1", EntryID: "entry1", TeamID: "team1", Bid: 20},
		{ID: "team2", EntryID: "entry1", TeamID: "team2", Bid: 30},
	}

	// WHEN validating the entry
	err := ValidateEntry(entry, teams)

	// THEN an error is returned
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatMoreThanTenTeamsFailsValidation(t *testing.T) {
	// GIVEN an entry with 11 teams
	userID := "user1"
	entry := &models.CalcuttaEntry{
		ID:         "entry1",
		Name:       "Test User",
		UserID:     &userID,
		CalcuttaID: "calcutta1",
	}
	teams := make([]*models.CalcuttaEntryTeam, 11)
	for i := 0; i < 11; i++ {
		teams[i] = &models.CalcuttaEntryTeam{ID: "team" + string(rune('1'+i)), EntryID: "entry1", TeamID: "team" + string(rune('1'+i)), Bid: 10}
	}

	// WHEN validating the entry
	err := ValidateEntry(entry, teams)

	// THEN an error is returned
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatBidOver50FailsValidation(t *testing.T) {
	// GIVEN an entry with a bid of 51 on one team
	userID := "user1"
	entry := &models.CalcuttaEntry{
		ID:         "entry1",
		Name:       "Test User",
		UserID:     &userID,
		CalcuttaID: "calcutta1",
	}
	teams := []*models.CalcuttaEntryTeam{
		{ID: "team1", EntryID: "entry1", TeamID: "team1", Bid: 51},
		{ID: "team2", EntryID: "entry1", TeamID: "team2", Bid: 20},
		{ID: "team3", EntryID: "entry1", TeamID: "team3", Bid: 20},
	}

	// WHEN validating the entry
	err := ValidateEntry(entry, teams)

	// THEN an error is returned
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatTotalBidsOver100FailsValidation(t *testing.T) {
	// GIVEN an entry with total bids of 102
	userID := "user1"
	entry := &models.CalcuttaEntry{
		ID:         "entry1",
		Name:       "Test User",
		UserID:     &userID,
		CalcuttaID: "calcutta1",
	}
	teams := []*models.CalcuttaEntryTeam{
		{ID: "team1", EntryID: "entry1", TeamID: "team1", Bid: 50},
		{ID: "team2", EntryID: "entry1", TeamID: "team2", Bid: 51},
		{ID: "team3", EntryID: "entry1", TeamID: "team3", Bid: 1},
	}

	// WHEN validating the entry
	err := ValidateEntry(entry, teams)

	// THEN an error is returned
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatBidUnder1FailsValidation(t *testing.T) {
	// GIVEN an entry with a bid of 0 on one team
	userID := "user1"
	entry := &models.CalcuttaEntry{
		ID:         "entry1",
		Name:       "Test User",
		UserID:     &userID,
		CalcuttaID: "calcutta1",
	}
	teams := []*models.CalcuttaEntryTeam{
		{ID: "team1", EntryID: "entry1", TeamID: "team1", Bid: 20},
		{ID: "team2", EntryID: "entry1", TeamID: "team2", Bid: 30},
		{ID: "team3", EntryID: "entry1", TeamID: "team3", Bid: 0},
	}

	// WHEN validating the entry
	err := ValidateEntry(entry, teams)

	// THEN an error is returned
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatDuplicateTeamBidsFailsValidation(t *testing.T) {
	// GIVEN an entry with duplicate team IDs
	userID := "user1"
	entry := &models.CalcuttaEntry{
		ID:         "entry1",
		Name:       "Test User",
		UserID:     &userID,
		CalcuttaID: "calcutta1",
	}
	teams := []*models.CalcuttaEntryTeam{
		{ID: "team1", EntryID: "entry1", TeamID: "team1", Bid: 20},
		{ID: "team2", EntryID: "entry1", TeamID: "team2", Bid: 30},
		{ID: "team3", EntryID: "entry1", TeamID: "team1", Bid: 10},
	}

	// WHEN validating the entry
	err := ValidateEntry(entry, teams)

	// THEN an error is returned
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
