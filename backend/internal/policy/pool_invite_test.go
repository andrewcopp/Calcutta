package policy

import (
	"context"
	"net/http"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatUnauthenticatedUserCannotInviteToPool(t *testing.T) {
	// GIVEN no authenticated user
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}

	// WHEN checking invite permission
	decision, err := CanInviteToPool(context.Background(), nil, "", pool)

	// THEN access is denied with unauthorized status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, decision.Status)
	}
}

func TestThatInviteToPoolDeniesWhenPoolIsNil(t *testing.T) {
	// GIVEN a nil pool
	// WHEN checking invite permission
	decision, err := CanInviteToPool(context.Background(), nil, "user1", nil)

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatNonAdminUserCannotInviteToPool(t *testing.T) {
	// GIVEN a regular user who is not the owner and has no admin grant
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking invite permission
	decision, err := CanInviteToPool(context.Background(), authz, "stranger", pool)

	// THEN access is denied with forbidden status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, decision.Status)
	}
}

func TestThatPoolOwnerCanInviteToPool(t *testing.T) {
	// GIVEN the pool owner
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}

	// WHEN checking invite permission
	decision, err := CanInviteToPool(context.Background(), nil, "owner", pool)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected pool owner to be able to invite")
	}
}

func TestThatAdminCanInviteToPool(t *testing.T) {
	// GIVEN a user with admin grant via authz
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN checking invite permission
	decision, err := CanInviteToPool(context.Background(), authz, "admin-user", pool)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected admin to be able to invite")
	}
}
