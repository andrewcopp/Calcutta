package policy

import (
	"context"
	"net/http"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatUnauthenticatedUserCannotManageCalcutta(t *testing.T) {
	// GIVEN no authenticated user
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}

	// WHEN checking manage permission
	decision, err := CanManageCalcutta(context.Background(), nil, "", calcutta)

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

func TestThatCalcuttaOwnerCanManageCalcutta(t *testing.T) {
	// GIVEN the calcutta owner
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}

	// WHEN checking manage permission
	decision, err := CanManageCalcutta(context.Background(), nil, "owner", calcutta)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected owner to be able to manage calcutta")
	}
}

func TestThatGlobalAdminCanManageCalcutta(t *testing.T) {
	// GIVEN a global admin
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN checking manage permission
	decision, err := CanManageCalcutta(context.Background(), authz, "admin-user", calcutta)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected global admin to be able to manage calcutta")
	}
}

func TestThatCalcuttaScopedAdminCanManageCalcutta(t *testing.T) {
	// GIVEN a user with calcutta-scoped admin grant
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN checking manage permission
	decision, err := CanManageCalcutta(context.Background(), authz, "co-manager", calcutta)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected scoped admin to be able to manage calcutta")
	}
}

func TestThatRegularUserCannotManageCalcutta(t *testing.T) {
	// GIVEN a regular user
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking manage permission
	decision, err := CanManageCalcutta(context.Background(), authz, "stranger", calcutta)

	// THEN access is denied
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Allowed {
		t.Fatal("expected regular user to be denied")
	}
}
