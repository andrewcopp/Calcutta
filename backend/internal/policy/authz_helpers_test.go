package policy

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatCalcuttaOwnerIsCalcuttaAdmin(t *testing.T) {
	// GIVEN a user who owns the calcutta
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}

	// WHEN checking admin status
	ok, err := isCalcuttaAdminOrOwner(context.Background(), nil, "owner", calcutta)

	// THEN they are recognized as admin
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected owner to be calcutta admin")
	}
}

func TestThatGlobalAdminIsCalcuttaAdmin(t *testing.T) {
	// GIVEN a global admin (HasPermission returns true for any scope)
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN checking admin status
	ok, err := isCalcuttaAdminOrOwner(context.Background(), authz, "admin-user", calcutta)

	// THEN they are recognized as admin
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected global admin to be calcutta admin")
	}
}

func TestThatCalcuttaScopedAdminIsCalcuttaAdmin(t *testing.T) {
	// GIVEN a user with calcutta-scoped admin grant (HasPermission returns true)
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN checking admin status
	ok, err := isCalcuttaAdminOrOwner(context.Background(), authz, "co-manager", calcutta)

	// THEN they are recognized as admin
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected scoped admin to be calcutta admin")
	}
}

func TestThatRegularUserIsNotCalcuttaAdmin(t *testing.T) {
	// GIVEN a regular user with no grants
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking admin status
	ok, err := isCalcuttaAdminOrOwner(context.Background(), authz, "stranger", calcutta)

	// THEN they are not recognized as admin
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected regular user to not be calcutta admin")
	}
}

func TestThatNilAuthzFallsBackToOwnerCheck(t *testing.T) {
	// GIVEN nil authz and a non-owner user
	calcutta := &models.Calcutta{ID: "c1", OwnerID: "owner"}

	// WHEN checking admin status with nil authz
	ok, err := isCalcuttaAdminOrOwner(context.Background(), nil, "stranger", calcutta)

	// THEN they are not recognized as admin
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected non-owner with nil authz to not be calcutta admin")
	}
}
