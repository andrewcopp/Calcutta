package policy

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatPoolOwnerIsPoolAdmin(t *testing.T) {
	// GIVEN a user who owns the pool
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}

	// WHEN checking admin status
	ok, err := isPoolAdminOrOwner(context.Background(), nil, "owner", pool)

	// THEN they are recognized as admin
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected owner to be pool admin")
	}
}

func TestThatGlobalAdminIsPoolAdmin(t *testing.T) {
	// GIVEN a global admin (HasPermission returns true for any scope)
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN checking admin status
	ok, err := isPoolAdminOrOwner(context.Background(), authz, "admin-user", pool)

	// THEN they are recognized as admin
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected global admin to be pool admin")
	}
}

func TestThatPoolScopedAdminIsPoolAdmin(t *testing.T) {
	// GIVEN a user with pool-scoped admin grant (HasPermission returns true)
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN checking admin status
	ok, err := isPoolAdminOrOwner(context.Background(), authz, "co-manager", pool)

	// THEN they are recognized as admin
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected scoped admin to be pool admin")
	}
}

func TestThatRegularUserIsNotPoolAdmin(t *testing.T) {
	// GIVEN a regular user with no grants
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking admin status
	ok, err := isPoolAdminOrOwner(context.Background(), authz, "stranger", pool)

	// THEN they are not recognized as admin
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected regular user to not be pool admin")
	}
}

func TestThatNilAuthzFallsBackToOwnerCheck(t *testing.T) {
	// GIVEN nil authz and a non-owner user
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}

	// WHEN checking admin status with nil authz
	ok, err := isPoolAdminOrOwner(context.Background(), nil, "stranger", pool)

	// THEN they are not recognized as admin
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected non-owner with nil authz to not be pool admin")
	}
}
