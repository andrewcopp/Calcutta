package policy

import (
	"context"
	"errors"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type fakeAuthz struct {
	ok  bool
	err error
}

func (a *fakeAuthz) HasPermission(ctx context.Context, userID string, scope string, scopeID string, permission string) (bool, error) {
	return a.ok, a.err
}

func TestThatCanViewEntryDataRejectsWhenEntryNotOwned(t *testing.T) {
	// GIVEN
	userID := "u1"
	otherUserID := "u2"
	entry := &models.CalcuttaEntry{UserID: &otherUserID}
	calcutta := &models.Calcutta{OwnerID: "owner"}

	// WHEN
	decision, err := CanViewEntryData(context.Background(), &fakeAuthz{ok: false}, userID, entry, calcutta)

	// THEN
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Allowed {
		t.Fatalf("expected not allowed")
	}
}

func TestThatCanViewEntryDataAllowsWhenEntryOwned(t *testing.T) {
	// GIVEN
	userID := "u1"
	entry := &models.CalcuttaEntry{UserID: &userID}
	calcutta := &models.Calcutta{OwnerID: "owner"}

	// WHEN
	decision, err := CanViewEntryData(context.Background(), &fakeAuthz{ok: false}, userID, entry, calcutta)

	// THEN
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("expected allowed")
	}
}

func TestThatCanViewEntryDataAllowsWhenCalcuttaOwner(t *testing.T) {
	// GIVEN
	userID := "u1"
	entry := &models.CalcuttaEntry{UserID: nil}
	calcutta := &models.Calcutta{OwnerID: userID}

	// WHEN
	decision, err := CanViewEntryData(context.Background(), &fakeAuthz{ok: false}, userID, entry, calcutta)

	// THEN
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("expected allowed")
	}
}

func TestThatCanViewEntryDataReturnsErrorWhenAuthzErrors(t *testing.T) {
	// GIVEN
	userID := "u1"
	errBoom := errors.New("boom")
	entry := &models.CalcuttaEntry{UserID: nil}
	calcutta := &models.Calcutta{OwnerID: "owner"}

	// WHEN
	_, err := CanViewEntryData(context.Background(), &fakeAuthz{ok: false, err: errBoom}, userID, entry, calcutta)

	// THEN
	if err == nil {
		t.Fatalf("expected error")
	}
}
