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

func TestThatPublicCalcuttaIsViewableByAnyone(t *testing.T) {
	// GIVEN a public calcutta and no authenticated user
	calcutta := &models.Calcutta{ID: "c1", Visibility: "public", OwnerID: "owner"}

	// WHEN checking view permission
	decision, err := CanViewCalcutta(context.Background(), nil, "", calcutta, nil)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected public calcutta to be viewable by anyone")
	}
}

func TestThatUnlistedCalcuttaIsViewableByAnyone(t *testing.T) {
	// GIVEN an unlisted calcutta and no authenticated user
	calcutta := &models.Calcutta{ID: "c1", Visibility: "unlisted", OwnerID: "owner"}

	// WHEN checking view permission
	decision, err := CanViewCalcutta(context.Background(), nil, "", calcutta, nil)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected unlisted calcutta to be viewable by anyone")
	}
}

func TestThatPrivateCalcuttaIsViewableByOwner(t *testing.T) {
	// GIVEN a private calcutta and the owner
	calcutta := &models.Calcutta{ID: "c1", Visibility: "private", OwnerID: "owner"}

	// WHEN the owner checks view permission
	decision, err := CanViewCalcutta(context.Background(), nil, "owner", calcutta, nil)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected private calcutta to be viewable by owner")
	}
}

func TestThatPrivateCalcuttaIsViewableByParticipant(t *testing.T) {
	// GIVEN a private calcutta and a participant
	calcutta := &models.Calcutta{ID: "c1", Visibility: "private", OwnerID: "owner"}
	participants := []string{"p1", "p2"}

	// WHEN a participant checks view permission
	decision, err := CanViewCalcutta(context.Background(), nil, "p1", calcutta, participants)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected private calcutta to be viewable by participant")
	}
}

func TestThatPrivateCalcuttaIsViewableByAdmin(t *testing.T) {
	// GIVEN a private calcutta and an admin user
	calcutta := &models.Calcutta{ID: "c1", Visibility: "private", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN an admin checks view permission
	decision, err := CanViewCalcutta(context.Background(), authz, "admin-user", calcutta, nil)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected private calcutta to be viewable by admin")
	}
}

func TestThatPrivateCalcuttaDeniesUnrelatedUser(t *testing.T) {
	// GIVEN a private calcutta and an unrelated user
	calcutta := &models.Calcutta{ID: "c1", Visibility: "private", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: false}

	// WHEN an unrelated user checks view permission
	decision, err := CanViewCalcutta(context.Background(), authz, "stranger", calcutta, []string{"p1"})

	// THEN access is denied
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Allowed {
		t.Fatal("expected private calcutta to deny unrelated user")
	}
}

func TestThatPrivateCalcuttaDeniesUnauthenticatedUser(t *testing.T) {
	// GIVEN a private calcutta and no authenticated user
	calcutta := &models.Calcutta{ID: "c1", Visibility: "private", OwnerID: "owner"}

	// WHEN an unauthenticated user checks view permission
	decision, err := CanViewCalcutta(context.Background(), nil, "", calcutta, nil)

	// THEN access is denied with unauthorized status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Allowed {
		t.Fatal("expected private calcutta to deny unauthenticated user")
	}
}
