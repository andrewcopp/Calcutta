package policy

import (
	"context"
	"net/http"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatUnauthenticatedUserCannotManagePool(t *testing.T) {
	// GIVEN no authenticated user
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}

	// WHEN checking manage permission
	decision, err := CanManagePool(context.Background(), nil, "", pool)

	// THEN access is denied with unauthorized status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Allowed {
		t.Fatal("expected unauthenticated user to be denied")
	}
	if decision.Status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, decision.Status)
	}
}

func TestThatPoolOwnerCanManagePool(t *testing.T) {
	// GIVEN the pool owner
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}

	// WHEN checking manage permission
	decision, err := CanManagePool(context.Background(), nil, "owner", pool)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected owner to be able to manage pool")
	}
}

func TestThatGlobalAdminCanManagePool(t *testing.T) {
	// GIVEN a global admin
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN checking manage permission
	decision, err := CanManagePool(context.Background(), authz, "admin-user", pool)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected global admin to be able to manage pool")
	}
}

func TestThatPoolScopedAdminCanManagePool(t *testing.T) {
	// GIVEN a user with pool-scoped admin grant
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN checking manage permission
	decision, err := CanManagePool(context.Background(), authz, "co-manager", pool)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected scoped admin to be able to manage pool")
	}
}

func TestThatRegularUserCannotManagePool(t *testing.T) {
	// GIVEN a regular user
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking manage permission
	decision, err := CanManagePool(context.Background(), authz, "stranger", pool)

	// THEN access is denied
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Allowed {
		t.Fatal("expected regular user to be denied")
	}
}
