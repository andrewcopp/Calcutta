package policy

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// --- CanEditEntryBids tests ---

func TestThatUnauthenticatedUserCannotEditEntryBids(t *testing.T) {
	// GIVEN no authenticated user
	userID := "u1"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &userID}
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	tournament := &models.Tournament{ID: "t1"}

	// WHEN checking edit bids permission
	decision, err := CanEditEntryBids(context.Background(), nil, "", entry, calcutta, tournament, time.Now())

	// THEN access is denied with unauthorized status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, decision.Status)
	}
}

func TestThatEditEntryBidsDeniesWhenEntryIsNil(t *testing.T) {
	// GIVEN a nil entry
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	tournament := &models.Tournament{ID: "t1"}

	// WHEN checking edit bids permission
	decision, err := CanEditEntryBids(context.Background(), nil, "user1", nil, calcutta, tournament, time.Now())

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatEditEntryBidsDeniesWhenCalcuttaIsNil(t *testing.T) {
	// GIVEN a nil calcutta
	userID := "u1"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &userID}
	tournament := &models.Tournament{ID: "t1"}

	// WHEN checking edit bids permission
	decision, err := CanEditEntryBids(context.Background(), nil, "u1", entry, nil, tournament, time.Now())

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatEditEntryBidsDeniesWhenTournamentIsNil(t *testing.T) {
	// GIVEN a nil tournament
	userID := "u1"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &userID}
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}

	// WHEN checking edit bids permission
	decision, err := CanEditEntryBids(context.Background(), nil, "u1", entry, calcutta, nil, time.Now())

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatNonOwnerNonAdminCannotEditEntryBids(t *testing.T) {
	// GIVEN a user who is neither the entry owner nor an admin
	otherUserID := "other-user"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &otherUserID}
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking edit bids permission
	decision, err := CanEditEntryBids(context.Background(), authz, "stranger", entry, calcutta, tournament, time.Now())

	// THEN access is denied with forbidden status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, decision.Status)
	}
}

func TestThatEntryOwnerCannotEditBidsAfterTournamentStarts(t *testing.T) {
	// GIVEN the entry owner and a tournament that has already started
	userID := "u1"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &userID}
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	startingAt := time.Now().Add(-24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking edit bids permission
	decision, err := CanEditEntryBids(context.Background(), authz, "u1", entry, calcutta, tournament, time.Now())

	// THEN access is denied with locked status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusLocked {
		t.Fatalf("expected status %d, got %d", http.StatusLocked, decision.Status)
	}
}

func TestThatEntryOwnerCanEditBidsBeforeTournamentStarts(t *testing.T) {
	// GIVEN the entry owner and a tournament that has not started
	userID := "u1"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &userID}
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking edit bids permission
	decision, err := CanEditEntryBids(context.Background(), authz, "u1", entry, calcutta, tournament, time.Now())

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected entry owner to be able to edit bids before tournament starts")
	}
}

func TestThatAdminCanEditBidsAfterTournamentStarts(t *testing.T) {
	// GIVEN an admin user and a tournament that has already started
	otherUserID := "other-user"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &otherUserID}
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	startingAt := time.Now().Add(-24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}

	// WHEN the calcutta owner checks edit bids permission
	decision, err := CanEditEntryBids(context.Background(), nil, "owner", entry, calcutta, tournament, time.Now())

	// THEN access is allowed because admins bypass the tournament lock
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected admin to be able to edit bids even after tournament starts")
	}
}

// --- CanViewEntryData tests ---

func TestThatUnauthenticatedUserCannotViewEntryData(t *testing.T) {
	// GIVEN no authenticated user
	userID := "u1"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &userID}
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}

	// WHEN checking view entry data permission
	decision, err := CanViewEntryData(context.Background(), nil, "", entry, calcutta)

	// THEN access is denied with unauthorized status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, decision.Status)
	}
}

func TestThatViewEntryDataDeniesWhenEntryIsNil(t *testing.T) {
	// GIVEN a nil entry
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}

	// WHEN checking view entry data permission
	decision, err := CanViewEntryData(context.Background(), nil, "user1", nil, calcutta)

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatViewEntryDataDeniesWhenCalcuttaIsNil(t *testing.T) {
	// GIVEN a nil calcutta
	userID := "u1"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &userID}

	// WHEN checking view entry data permission
	decision, err := CanViewEntryData(context.Background(), nil, "u1", entry, nil)

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatNonOwnerNonAdminCannotViewEntryData(t *testing.T) {
	// GIVEN a user who is neither the entry owner, calcutta owner, nor an admin
	otherUserID := "other-user"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &otherUserID}
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking view entry data permission
	decision, err := CanViewEntryData(context.Background(), authz, "stranger", entry, calcutta)

	// THEN access is denied with forbidden status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, decision.Status)
	}
}

func TestThatEntryOwnerCanViewEntryData(t *testing.T) {
	// GIVEN the entry owner
	userID := "u1"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &userID}
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking view entry data permission
	decision, err := CanViewEntryData(context.Background(), authz, "u1", entry, calcutta)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected entry owner to be able to view entry data")
	}
}

func TestThatCalcuttaOwnerCanViewEntryData(t *testing.T) {
	// GIVEN the calcutta owner viewing another user's entry
	otherUserID := "other-user"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &otherUserID}
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}

	// WHEN checking view entry data permission as the calcutta owner
	decision, err := CanViewEntryData(context.Background(), nil, "owner", entry, calcutta)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected calcutta owner to be able to view any entry data")
	}
}

func TestThatAdminCanViewEntryData(t *testing.T) {
	// GIVEN an admin user viewing another user's entry
	otherUserID := "other-user"
	entry := &models.CalcuttaEntry{ID: "e1", UserID: &otherUserID}
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN checking view entry data permission as an admin
	decision, err := CanViewEntryData(context.Background(), authz, "admin-user", entry, calcutta)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected admin to be able to view any entry data")
	}
}
