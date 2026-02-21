package policy

import (
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatEntryOwnerCanViewOwnDataDuringBidding(t *testing.T) {
	// GIVEN a tournament that has not started and a user who owns the entry
	userID := "u1"
	entry := &models.CalcuttaEntry{UserID: &userID}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{StartingAt: &startingAt}

	// WHEN checking bidding phase view access
	result := IsBiddingPhaseViewAllowed(userID, entry, tournament, time.Now(), false)

	// THEN access is allowed
	if !result {
		t.Fatal("expected entry owner to view own data during bidding")
	}
}

func TestThatAdminCanViewAnyEntryDuringBidding(t *testing.T) {
	// GIVEN a tournament that has not started and an admin viewing another user's entry
	otherUserID := "u2"
	entry := &models.CalcuttaEntry{UserID: &otherUserID}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{StartingAt: &startingAt}

	// WHEN checking bidding phase view access as admin
	result := IsBiddingPhaseViewAllowed("admin", entry, tournament, time.Now(), true)

	// THEN access is allowed
	if !result {
		t.Fatal("expected admin to view any entry during bidding")
	}
}

func TestThatParticipantCannotViewOtherEntryDuringBidding(t *testing.T) {
	// GIVEN a tournament that has not started and a non-admin viewing another user's entry
	otherUserID := "u2"
	entry := &models.CalcuttaEntry{UserID: &otherUserID}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{StartingAt: &startingAt}

	// WHEN checking bidding phase view access
	result := IsBiddingPhaseViewAllowed("u1", entry, tournament, time.Now(), false)

	// THEN access is denied
	if result {
		t.Fatal("expected non-owner participant to be denied during bidding")
	}
}

func TestThatAnyoneCanViewEntryDataAfterTournamentStarts(t *testing.T) {
	// GIVEN a tournament that has already started
	otherUserID := "u2"
	entry := &models.CalcuttaEntry{UserID: &otherUserID}
	startingAt := time.Now().Add(-24 * time.Hour)
	tournament := &models.Tournament{StartingAt: &startingAt}

	// WHEN checking bidding phase view access as a non-owner non-admin
	result := IsBiddingPhaseViewAllowed("u1", entry, tournament, time.Now(), false)

	// THEN access is allowed
	if !result {
		t.Fatal("expected anyone to view entry data after tournament starts")
	}
}
