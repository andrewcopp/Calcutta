package simulation_game_outcomes

import (
	"math"
	"testing"
)

func TestThatValidateReturnsErrorForNilSpec(t *testing.T) {
	// GIVEN a nil spec
	var s *Spec

	// WHEN Validate is called
	err := s.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for nil spec")
	}
}

func TestThatValidateReturnsErrorForUnsupportedKind(t *testing.T) {
	// GIVEN a spec with an unsupported kind
	s := &Spec{Kind: "unknown", Sigma: 10.0}

	// WHEN Validate is called
	err := s.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for unsupported kind")
	}
}

func TestThatValidateReturnsErrorForNonPositiveSigma(t *testing.T) {
	// GIVEN a spec with a non-positive sigma
	s := &Spec{Kind: "kenpom", Sigma: 0}

	// WHEN Validate is called
	err := s.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for non-positive sigma")
	}
}

func TestThatValidateSucceedsForValidSpec(t *testing.T) {
	// GIVEN a valid spec
	s := &Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN Validate is called
	err := s.Validate()

	// THEN no error is returned
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestThatNormalizeDefaultsKindToKenpom(t *testing.T) {
	// GIVEN a spec with an empty kind
	s := &Spec{Kind: "", Sigma: 10.0}

	// WHEN Normalize is called
	s.Normalize()

	// THEN kind is set to "kenpom"
	if s.Kind != "kenpom" {
		t.Errorf("expected kind 'kenpom', got %q", s.Kind)
	}
}

func TestThatNormalizeDefaultsSigmaTo10(t *testing.T) {
	// GIVEN a spec with a non-positive sigma
	s := &Spec{Kind: "kenpom", Sigma: 0}

	// WHEN Normalize is called
	s.Normalize()

	// THEN sigma is set to 10.0
	if s.Sigma != 10.0 {
		t.Errorf("expected sigma 10.0, got %f", s.Sigma)
	}
}

func TestThatWinProbReturns0Point5ForEqualTeams(t *testing.T) {
	// GIVEN a spec and two teams with equal net ratings
	s := &Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN WinProb is called with equal ratings
	prob := s.WinProb(5.0, 5.0)

	// THEN the probability is 0.5
	if math.Abs(prob-0.5) > 1e-9 {
		t.Errorf("expected 0.5, got %f", prob)
	}
}

func TestThatWinProbFavorsBetterTeam(t *testing.T) {
	// GIVEN a spec and two teams where team1 is better
	s := &Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN WinProb is called with team1 having higher rating
	prob := s.WinProb(20.0, 5.0)

	// THEN team1 has greater than 50% probability
	if prob <= 0.5 {
		t.Errorf("expected probability > 0.5, got %f", prob)
	}
}
