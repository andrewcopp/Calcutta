package policy

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatIsEntryOwnerOrCalcuttaOwnerReturnsFalseWhenEntryNotOwned(t *testing.T) {
	// GIVEN a user who does not own the entry or calcutta
	userID := "u1"
	otherUserID := "u2"
	entry := &models.CalcuttaEntry{UserID: &otherUserID}
	calcutta := &models.Calcutta{OwnerID: "owner"}

	// WHEN checking ownership
	result := IsEntryOwnerOrCalcuttaOwner(userID, entry, calcutta)

	// THEN the result is false
	if result {
		t.Fatalf("expected false")
	}
}

func TestThatIsEntryOwnerOrCalcuttaOwnerReturnsTrueWhenEntryOwned(t *testing.T) {
	// GIVEN a user who owns the entry
	userID := "u1"
	entry := &models.CalcuttaEntry{UserID: &userID}
	calcutta := &models.Calcutta{OwnerID: "owner"}

	// WHEN checking ownership
	result := IsEntryOwnerOrCalcuttaOwner(userID, entry, calcutta)

	// THEN the result is true
	if !result {
		t.Fatalf("expected true")
	}
}

func TestThatIsEntryOwnerOrCalcuttaOwnerReturnsTrueWhenCalcuttaOwner(t *testing.T) {
	// GIVEN a user who owns the calcutta
	userID := "u1"
	entry := &models.CalcuttaEntry{UserID: nil}
	calcutta := &models.Calcutta{OwnerID: userID}

	// WHEN checking ownership
	result := IsEntryOwnerOrCalcuttaOwner(userID, entry, calcutta)

	// THEN the result is true
	if !result {
		t.Fatalf("expected true")
	}
}

func TestThatIsEntryOwnerOrCalcuttaOwnerReturnsFalseForNilEntry(t *testing.T) {
	// GIVEN a nil entry
	userID := "u1"
	calcutta := &models.Calcutta{OwnerID: "owner"}

	// WHEN checking ownership
	result := IsEntryOwnerOrCalcuttaOwner(userID, nil, calcutta)

	// THEN the result is false
	if result {
		t.Fatalf("expected false")
	}
}

func TestThatIsEntryOwnerOrCalcuttaOwnerReturnsFalseForNilCalcutta(t *testing.T) {
	// GIVEN a nil calcutta and entry with different owner
	userID := "u1"
	otherUserID := "u2"
	entry := &models.CalcuttaEntry{UserID: &otherUserID}

	// WHEN checking ownership
	result := IsEntryOwnerOrCalcuttaOwner(userID, entry, nil)

	// THEN the result is false
	if result {
		t.Fatalf("expected false")
	}
}
