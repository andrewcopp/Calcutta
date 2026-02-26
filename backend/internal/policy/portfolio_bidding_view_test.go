package policy

import (
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatPortfolioOwnerCanViewOwnDataDuringBidding(t *testing.T) {
	// GIVEN a tournament that has not started and a user who owns the portfolio
	userID := "u1"
	portfolio := &models.Portfolio{UserID: &userID}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{StartingAt: &startingAt}

	// WHEN checking bidding phase view access
	result := IsBiddingPhaseViewAllowed(userID, portfolio, tournament, time.Now(), false)

	// THEN access is allowed
	if !result {
		t.Fatal("expected portfolio owner to view own data during bidding")
	}
}

func TestThatAdminCanViewAnyPortfolioDuringBidding(t *testing.T) {
	// GIVEN a tournament that has not started and an admin viewing another user's portfolio
	otherUserID := "u2"
	portfolio := &models.Portfolio{UserID: &otherUserID}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{StartingAt: &startingAt}

	// WHEN checking bidding phase view access as admin
	result := IsBiddingPhaseViewAllowed("admin", portfolio, tournament, time.Now(), true)

	// THEN access is allowed
	if !result {
		t.Fatal("expected admin to view any portfolio during bidding")
	}
}

func TestThatParticipantCannotViewOtherPortfolioDuringBidding(t *testing.T) {
	// GIVEN a tournament that has not started and a non-admin viewing another user's portfolio
	otherUserID := "u2"
	portfolio := &models.Portfolio{UserID: &otherUserID}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{StartingAt: &startingAt}

	// WHEN checking bidding phase view access
	result := IsBiddingPhaseViewAllowed("u1", portfolio, tournament, time.Now(), false)

	// THEN access is denied
	if result {
		t.Fatal("expected non-owner participant to be denied during bidding")
	}
}

func TestThatAnyoneCanViewPortfolioDataAfterTournamentStarts(t *testing.T) {
	// GIVEN a tournament that has already started
	otherUserID := "u2"
	portfolio := &models.Portfolio{UserID: &otherUserID}
	startingAt := time.Now().Add(-24 * time.Hour)
	tournament := &models.Tournament{StartingAt: &startingAt}

	// WHEN checking bidding phase view access as a non-owner non-admin
	result := IsBiddingPhaseViewAllowed("u1", portfolio, tournament, time.Now(), false)

	// THEN access is allowed
	if !result {
		t.Fatal("expected anyone to view portfolio data after tournament starts")
	}
}
