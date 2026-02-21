package policy

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatUnauthenticatedUserCannotCreateEntry(t *testing.T) {
	// GIVEN no authenticated user
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	tournament := &models.Tournament{ID: "t1"}

	// WHEN checking create entry permission
	decision, err := CanCreateEntry(context.Background(), nil, "", calcutta, tournament, nil, time.Now())

	// THEN access is denied with unauthorized status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, decision.Status)
	}
}

func TestThatCreateEntryDeniesWhenCalcuttaIsNil(t *testing.T) {
	// GIVEN a nil calcutta
	tournament := &models.Tournament{ID: "t1"}

	// WHEN checking create entry permission
	decision, err := CanCreateEntry(context.Background(), nil, "user1", nil, tournament, nil, time.Now())

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatCreateEntryDeniesWhenTournamentIsNil(t *testing.T) {
	// GIVEN a nil tournament
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}

	// WHEN checking create entry permission
	decision, err := CanCreateEntry(context.Background(), nil, "user1", calcutta, nil, nil, time.Now())

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatNonCommissionerCannotCreateEntryForAnotherUser(t *testing.T) {
	// GIVEN a non-commissioner user trying to create an entry for another user
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}
	targetUserID := "other-user"

	// WHEN checking create entry permission for another user
	decision, err := CanCreateEntry(context.Background(), authz, "regular-user", calcutta, tournament, &targetUserID, time.Now())

	// THEN access is denied with forbidden status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, decision.Status)
	}
}

func TestThatNonAdminCannotCreateEntryAfterTournamentStarts(t *testing.T) {
	// GIVEN a tournament that has already started and a non-admin user
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	startingAt := time.Now().Add(-24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking create entry permission
	decision, err := CanCreateEntry(context.Background(), authz, "regular-user", calcutta, tournament, nil, time.Now())

	// THEN access is denied with locked status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusLocked {
		t.Fatalf("expected status %d, got %d", http.StatusLocked, decision.Status)
	}
}

func TestThatUserCanCreateEntryForSelf(t *testing.T) {
	// GIVEN a regular user creating an entry for themselves before the tournament starts
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking create entry permission with no target user
	decision, err := CanCreateEntry(context.Background(), authz, "regular-user", calcutta, tournament, nil, time.Now())

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected user to be able to create entry for self")
	}
}

func TestThatCommissionerCanCreateEntryForAnotherUser(t *testing.T) {
	// GIVEN the calcutta owner creating an entry for another user before the tournament starts
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	targetUserID := "other-user"

	// WHEN checking create entry permission for another user as the commissioner
	decision, err := CanCreateEntry(context.Background(), nil, "owner", calcutta, tournament, &targetUserID, time.Now())

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected commissioner to be able to create entry for another user")
	}
}
