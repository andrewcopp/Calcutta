package calcutta

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

func TestThatValidEntryPassesValidation(t *testing.T) {
	service := newTestCalcuttaService()
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

	if err := service.ValidateEntry(entry, teams); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestThatFewerThanThreeTeamsFailsValidation(t *testing.T) {
	service := newTestCalcuttaService()
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

	if err := service.ValidateEntry(entry, teams); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatMoreThanTenTeamsFailsValidation(t *testing.T) {
	service := newTestCalcuttaService()
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

	if err := service.ValidateEntry(entry, teams); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatBidOver50FailsValidation(t *testing.T) {
	service := newTestCalcuttaService()
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

	if err := service.ValidateEntry(entry, teams); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatTotalBidsOver100FailsValidation(t *testing.T) {
	service := newTestCalcuttaService()
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

	if err := service.ValidateEntry(entry, teams); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatBidUnder1FailsValidation(t *testing.T) {
	service := newTestCalcuttaService()
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

	if err := service.ValidateEntry(entry, teams); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestThatDuplicateTeamBidsFailsValidation(t *testing.T) {
	service := newTestCalcuttaService()
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

	if err := service.ValidateEntry(entry, teams); err == nil {
		t.Fatalf("expected error, got nil")
	}
}
