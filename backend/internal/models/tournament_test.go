package models

import (
	"testing"
	"time"
)

func TestThatNilTournamentIsNotStarted(t *testing.T) {
	GIVENNilTournament := (*Tournament)(nil)
	WHENNow := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	THENStarted := GIVENNilTournament.HasStarted(WHENNow)
	if THENStarted != false {
		t.Fatalf("expected false")
	}
}

func TestThatTournamentWithNilStartingAtIsNotStarted(t *testing.T) {
	GIVENTournamentWithNilStartingAt := &Tournament{StartingAt: nil}
	WHENNow := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	THENStarted := GIVENTournamentWithNilStartingAt.HasStarted(WHENNow)
	if THENStarted != false {
		t.Fatalf("expected false")
	}
}

func TestThatTournamentIsNotStartedBeforeStartingAt(t *testing.T) {
	GIVENStartingAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	GIVENTournament := &Tournament{StartingAt: &GIVENStartingAt}
	WHENNow := GIVENStartingAt.Add(-1 * time.Second)
	THENStarted := GIVENTournament.HasStarted(WHENNow)
	if THENStarted != false {
		t.Fatalf("expected false")
	}
}

func TestThatTournamentIsStartedAtStartingAt(t *testing.T) {
	GIVENStartingAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	GIVENTournament := &Tournament{StartingAt: &GIVENStartingAt}
	WHENNow := GIVENStartingAt
	THENStarted := GIVENTournament.HasStarted(WHENNow)
	if THENStarted != true {
		t.Fatalf("expected true")
	}
}

func TestThatTournamentIsStartedAfterStartingAt(t *testing.T) {
	GIVENStartingAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	GIVENTournament := &Tournament{StartingAt: &GIVENStartingAt}
	WHENNow := GIVENStartingAt.Add(1 * time.Second)
	THENStarted := GIVENTournament.HasStarted(WHENNow)
	if THENStarted != true {
		t.Fatalf("expected true")
	}
}

func TestThatNilTournamentCannotEditBids(t *testing.T) {
	GIVENNilTournament := (*Tournament)(nil)
	WHENNow := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	WHENAllowed, _ := GIVENNilTournament.CanEditBids(WHENNow, false)
	if WHENAllowed != false {
		t.Fatalf("expected false")
	}
}

func TestThatNilTournamentCannotEditBidsWithReasonTournamentMissing(t *testing.T) {
	GIVENNilTournament := (*Tournament)(nil)
	WHENNow := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	_, WHENReason := GIVENNilTournament.CanEditBids(WHENNow, false)
	if WHENReason != TournamentEditDeniedReasonTournamentMissing {
		t.Fatalf("expected %s", TournamentEditDeniedReasonTournamentMissing)
	}
}

func TestThatAdminCanEditBidsAfterTournamentStarts(t *testing.T) {
	GIVENStartingAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	GIVENTournament := &Tournament{StartingAt: &GIVENStartingAt}
	WHENNow := GIVENStartingAt.Add(1 * time.Second)
	WHENAllowed, _ := GIVENTournament.CanEditBids(WHENNow, true)
	if WHENAllowed != true {
		t.Fatalf("expected true")
	}
}

func TestThatAdminEditBidsHasEmptyReason(t *testing.T) {
	GIVENStartingAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	GIVENTournament := &Tournament{StartingAt: &GIVENStartingAt}
	WHENNow := GIVENStartingAt.Add(1 * time.Second)
	_, WHENReason := GIVENTournament.CanEditBids(WHENNow, true)
	if WHENReason != "" {
		t.Fatalf("expected empty string")
	}
}

func TestThatNonAdminCanEditBidsBeforeTournamentStarts(t *testing.T) {
	GIVENStartingAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	GIVENTournament := &Tournament{StartingAt: &GIVENStartingAt}
	WHENNow := GIVENStartingAt.Add(-1 * time.Second)
	WHENAllowed, _ := GIVENTournament.CanEditBids(WHENNow, false)
	if WHENAllowed != true {
		t.Fatalf("expected true")
	}
}

func TestThatNonAdminEditBidsBeforeTournamentStartsHasEmptyReason(t *testing.T) {
	GIVENStartingAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	GIVENTournament := &Tournament{StartingAt: &GIVENStartingAt}
	WHENNow := GIVENStartingAt.Add(-1 * time.Second)
	_, WHENReason := GIVENTournament.CanEditBids(WHENNow, false)
	if WHENReason != "" {
		t.Fatalf("expected empty string")
	}
}

func TestThatNonAdminCannotEditBidsAfterTournamentStarts(t *testing.T) {
	GIVENStartingAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	GIVENTournament := &Tournament{StartingAt: &GIVENStartingAt}
	WHENNow := GIVENStartingAt
	WHENAllowed, _ := GIVENTournament.CanEditBids(WHENNow, false)
	if WHENAllowed != false {
		t.Fatalf("expected false")
	}
}

func TestThatNonAdminCannotEditBidsAfterTournamentStartsWithReasonTournamentStarted(t *testing.T) {
	GIVENStartingAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	GIVENTournament := &Tournament{StartingAt: &GIVENStartingAt}
	WHENNow := GIVENStartingAt
	_, WHENReason := GIVENTournament.CanEditBids(WHENNow, false)
	if WHENReason != TournamentEditDeniedReasonTournamentStarted {
		t.Fatalf("expected %s", TournamentEditDeniedReasonTournamentStarted)
	}
}

