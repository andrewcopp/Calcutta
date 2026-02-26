package policy

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type mockAuthzChecker struct {
	result bool
	err    error
}

func (m *mockAuthzChecker) HasPermission(_ context.Context, _, _, _, _ string) (bool, error) {
	return m.result, m.err
}

func TestThatPublicPoolIsViewableByAnyone(t *testing.T) {
	// GIVEN a public pool and no authenticated user
	pool := &models.Pool{ID: "p1", Visibility: "public", OwnerID: "owner"}

	// WHEN checking view permission
	decision, err := CanViewPool(context.Background(), nil, "", pool, nil)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected public pool to be viewable by anyone")
	}
}

func TestThatUnlistedPoolIsViewableByAnyone(t *testing.T) {
	// GIVEN an unlisted pool and no authenticated user
	pool := &models.Pool{ID: "p1", Visibility: "unlisted", OwnerID: "owner"}

	// WHEN checking view permission
	decision, err := CanViewPool(context.Background(), nil, "", pool, nil)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected unlisted pool to be viewable by anyone")
	}
}

func TestThatPrivatePoolIsViewableByOwner(t *testing.T) {
	// GIVEN a private pool and the owner
	pool := &models.Pool{ID: "p1", Visibility: "private", OwnerID: "owner"}

	// WHEN the owner checks view permission
	decision, err := CanViewPool(context.Background(), nil, "owner", pool, nil)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected private pool to be viewable by owner")
	}
}

func TestThatPrivatePoolIsViewableByParticipant(t *testing.T) {
	// GIVEN a private pool and a participant
	pool := &models.Pool{ID: "p1", Visibility: "private", OwnerID: "owner"}
	participants := []string{"p1", "p2"}

	// WHEN a participant checks view permission
	decision, err := CanViewPool(context.Background(), nil, "p1", pool, participants)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected private pool to be viewable by participant")
	}
}

func TestThatPrivatePoolIsViewableByAdmin(t *testing.T) {
	// GIVEN a private pool and an admin user
	pool := &models.Pool{ID: "p1", Visibility: "private", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN an admin checks view permission
	decision, err := CanViewPool(context.Background(), authz, "admin-user", pool, nil)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected private pool to be viewable by admin")
	}
}

func TestThatPrivatePoolDeniesUnrelatedUser(t *testing.T) {
	// GIVEN a private pool and an unrelated user
	pool := &models.Pool{ID: "p1", Visibility: "private", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: false}

	// WHEN an unrelated user checks view permission
	decision, err := CanViewPool(context.Background(), authz, "stranger", pool, []string{"p1"})

	// THEN access is denied
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Allowed {
		t.Fatal("expected private pool to deny unrelated user")
	}
}

func TestThatPrivatePoolDeniesUnauthenticatedUser(t *testing.T) {
	// GIVEN a private pool and no authenticated user
	pool := &models.Pool{ID: "p1", Visibility: "private", OwnerID: "owner"}

	// WHEN an unauthenticated user checks view permission
	decision, err := CanViewPool(context.Background(), nil, "", pool, nil)

	// THEN access is denied with unauthorized status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Allowed {
		t.Fatal("expected private pool to deny unauthenticated user")
	}
}
