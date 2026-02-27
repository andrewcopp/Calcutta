package dtos

import "testing"

func TestThatCreatePoolRequestRequiresScoringRules(t *testing.T) {
	// GIVEN a request with no scoring rules
	req := &CreatePoolRequest{
		Name:         "Test Pool",
		TournamentID: "t1",
		ScoringRules: []ScoringRuleInput{},
	}

	// WHEN validating
	err := req.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for empty scoring rules")
	}
}

func TestThatCreatePoolRequestRejectsDuplicateWinIndex(t *testing.T) {
	// GIVEN a request with duplicate winIndex values
	req := &CreatePoolRequest{
		Name:         "Test Pool",
		TournamentID: "t1",
		ScoringRules: []ScoringRuleInput{
			{WinIndex: 1, PointsAwarded: 50},
			{WinIndex: 1, PointsAwarded: 100},
		},
	}

	// WHEN validating
	err := req.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for duplicate winIndex")
	}
}

func TestThatCreatePoolRequestRejectsNegativePointsAwarded(t *testing.T) {
	// GIVEN a request with negative pointsAwarded
	req := &CreatePoolRequest{
		Name:         "Test Pool",
		TournamentID: "t1",
		ScoringRules: []ScoringRuleInput{
			{WinIndex: 1, PointsAwarded: -10},
		},
	}

	// WHEN validating
	err := req.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for negative pointsAwarded")
	}
}

func TestThatCreatePoolRequestRejectsZeroWinIndex(t *testing.T) {
	// GIVEN a request with winIndex of 0
	req := &CreatePoolRequest{
		Name:         "Test Pool",
		TournamentID: "t1",
		ScoringRules: []ScoringRuleInput{
			{WinIndex: 0, PointsAwarded: 50},
		},
	}

	// WHEN validating
	err := req.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for zero winIndex")
	}
}

func TestThatCreatePoolRequestAcceptsValidScoringRules(t *testing.T) {
	// GIVEN a valid request with scoring rules
	req := &CreatePoolRequest{
		Name:         "Test Pool",
		TournamentID: "t1",
		ScoringRules: []ScoringRuleInput{
			{WinIndex: 1, PointsAwarded: 50},
			{WinIndex: 2, PointsAwarded: 100},
		},
	}

	// WHEN validating
	err := req.Validate()

	// THEN no error is returned
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestThatUpdatePoolRequestRequiresAtLeastOneField(t *testing.T) {
	// GIVEN a request with no fields set
	req := &UpdatePoolRequest{}

	// WHEN validating
	err := req.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for empty update request")
	}
}

func TestThatUpdatePoolRequestRejectsEmptyName(t *testing.T) {
	// GIVEN a request with an empty name
	empty := "   "
	req := &UpdatePoolRequest{Name: &empty}

	// WHEN validating
	err := req.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestThatUpdatePoolRequestAcceptsValidName(t *testing.T) {
	// GIVEN a request with a valid name
	name := "My Pool"
	req := &UpdatePoolRequest{Name: &name}

	// WHEN validating
	err := req.Validate()

	// THEN no error is returned
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
