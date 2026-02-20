package dtos

import "testing"

func TestThatCreateTournamentRequestRejectsEmptyCompetition(t *testing.T) {
	// GIVEN a request with an empty competition
	req := &CreateTournamentRequest{
		Competition: "",
		Year:        2026,
		Rounds:      6,
	}

	// WHEN validating
	err := req.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for empty competition")
	}
}

func TestThatCreateTournamentRequestRejectsWhitespaceCompetition(t *testing.T) {
	// GIVEN a request with a whitespace-only competition
	req := &CreateTournamentRequest{
		Competition: "   ",
		Year:        2026,
		Rounds:      6,
	}

	// WHEN validating
	err := req.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for whitespace competition")
	}
}

func TestThatCreateTournamentRequestRejectsYearAtOrBelow2000(t *testing.T) {
	// GIVEN a request with year=2000
	req := &CreateTournamentRequest{
		Competition: "NCAA Men's",
		Year:        2000,
		Rounds:      6,
	}

	// WHEN validating
	err := req.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for year 2000")
	}
}

func TestThatCreateTournamentRequestRejectsZeroRounds(t *testing.T) {
	// GIVEN a request with rounds=0
	req := &CreateTournamentRequest{
		Competition: "NCAA Men's",
		Year:        2026,
		Rounds:      0,
	}

	// WHEN validating
	err := req.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for zero rounds")
	}
}

func TestThatCreateTournamentRequestRejectsNegativeRounds(t *testing.T) {
	// GIVEN a request with rounds=-1
	req := &CreateTournamentRequest{
		Competition: "NCAA Men's",
		Year:        2026,
		Rounds:      -1,
	}

	// WHEN validating
	err := req.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for negative rounds")
	}
}

func TestThatCreateTournamentRequestAcceptsValidInput(t *testing.T) {
	// GIVEN a valid request
	req := &CreateTournamentRequest{
		Competition: "NCAA Men's",
		Year:        2026,
		Rounds:      6,
	}

	// WHEN validating
	err := req.Validate()

	// THEN no error is returned
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestThatDerivedNameFormatsCompetitionAndYear(t *testing.T) {
	// GIVEN a request with competition="NCAA Men's" and year=2026
	req := &CreateTournamentRequest{
		Competition: "NCAA Men's",
		Year:        2026,
		Rounds:      6,
	}

	// WHEN deriving the name
	name := req.DerivedName()

	// THEN the name is "NCAA Men's 2026"
	if name != "NCAA Men's 2026" {
		t.Errorf("expected \"NCAA Men's 2026\", got %q", name)
	}
}
