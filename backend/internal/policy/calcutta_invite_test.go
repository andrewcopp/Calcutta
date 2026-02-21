package policy

import (
	"context"
	"net/http"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatUnauthenticatedUserCannotInviteToCalcutta(t *testing.T) {
	// GIVEN no authenticated user
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}

	// WHEN checking invite permission
	decision, err := CanInviteToCalcutta(context.Background(), nil, "", calcutta)

	// THEN access is denied with unauthorized status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, decision.Status)
	}
}

func TestThatInviteToCalcuttaDeniesWhenCalcuttaIsNil(t *testing.T) {
	// GIVEN a nil calcutta
	// WHEN checking invite permission
	decision, err := CanInviteToCalcutta(context.Background(), nil, "user1", nil)

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatNonAdminUserCannotInviteToCalcutta(t *testing.T) {
	// GIVEN a regular user who is not the owner and has no admin grant
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking invite permission
	decision, err := CanInviteToCalcutta(context.Background(), authz, "stranger", calcutta)

	// THEN access is denied with forbidden status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, decision.Status)
	}
}

func TestThatCalcuttaOwnerCanInviteToCalcutta(t *testing.T) {
	// GIVEN the calcutta owner
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}

	// WHEN checking invite permission
	decision, err := CanInviteToCalcutta(context.Background(), nil, "owner", calcutta)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected calcutta owner to be able to invite")
	}
}

func TestThatAdminCanInviteToCalcutta(t *testing.T) {
	// GIVEN a user with admin grant via authz
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN checking invite permission
	decision, err := CanInviteToCalcutta(context.Background(), authz, "admin-user", calcutta)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected admin to be able to invite")
	}
}
