package policy

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatUnauthenticatedUserCannotAcceptInvitation(t *testing.T) {
	// GIVEN no authenticated user
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	tournament := &models.Tournament{ID: "t1"}

	// WHEN checking accept invitation permission
	decision, err := CanAcceptInvitation(context.Background(), nil, "", calcutta, tournament, time.Now())

	// THEN access is denied with unauthorized status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, decision.Status)
	}
}

func TestThatNilCalcuttaRejectsInvitationAcceptance(t *testing.T) {
	// GIVEN a nil calcutta
	tournament := &models.Tournament{ID: "t1"}

	// WHEN checking accept invitation permission
	decision, err := CanAcceptInvitation(context.Background(), nil, "user1", nil, tournament, time.Now())

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatNilTournamentRejectsInvitationAcceptance(t *testing.T) {
	// GIVEN a nil tournament
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}

	// WHEN checking accept invitation permission
	decision, err := CanAcceptInvitation(context.Background(), nil, "user1", calcutta, nil, time.Now())

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatNonAdminCannotAcceptInvitationAfterTournamentStarts(t *testing.T) {
	// GIVEN a tournament that has already started and a non-admin user
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	startingAt := time.Now().Add(-24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking accept invitation permission
	decision, err := CanAcceptInvitation(context.Background(), authz, "regular-user", calcutta, tournament, time.Now())

	// THEN access is denied with locked status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusLocked {
		t.Fatalf("expected status %d, got %d", http.StatusLocked, decision.Status)
	}
}

func TestThatNonAdminCanAcceptInvitationBeforeTournamentStarts(t *testing.T) {
	// GIVEN a tournament that has not started and a non-admin user
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking accept invitation permission
	decision, err := CanAcceptInvitation(context.Background(), authz, "regular-user", calcutta, tournament, time.Now())

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected non-admin to be able to accept invitation before tournament starts")
	}
}

func TestThatAdminCanAcceptInvitationAfterTournamentStarts(t *testing.T) {
	// GIVEN a tournament that has already started and an admin user
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	startingAt := time.Now().Add(-24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}

	// WHEN checking accept invitation permission as the calcutta owner
	decision, err := CanAcceptInvitation(context.Background(), nil, "owner", calcutta, tournament, time.Now())

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected admin to be able to accept invitation after tournament starts")
	}
}
