package winprob

import (
	"math"
	"testing"
)

func TestThatValidateReturnsErrorForNilModel(t *testing.T) {
	// GIVEN a nil model
	var m *Model

	// WHEN Validate is called
	err := m.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for nil model")
	}
}

func TestThatValidateReturnsErrorForUnsupportedKind(t *testing.T) {
	// GIVEN a model with an unsupported kind
	m := &Model{Kind: "unknown", Sigma: 10.0}

	// WHEN Validate is called
	err := m.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for unsupported kind")
	}
}

func TestThatValidateReturnsErrorForNonPositiveSigma(t *testing.T) {
	// GIVEN a model with a non-positive sigma
	m := &Model{Kind: "kenpom", Sigma: 0}

	// WHEN Validate is called
	err := m.Validate()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for non-positive sigma")
	}
}

func TestThatValidateSucceedsForValidModel(t *testing.T) {
	// GIVEN a valid model
	m := &Model{Kind: "kenpom", Sigma: 10.0}

	// WHEN Validate is called
	err := m.Validate()

	// THEN no error is returned
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestThatNormalizeDefaultsKindToKenpom(t *testing.T) {
	// GIVEN a model with an empty kind
	m := &Model{Kind: "", Sigma: 10.0}

	// WHEN Normalize is called
	m.Normalize()

	// THEN kind is set to "kenpom"
	if m.Kind != "kenpom" {
		t.Errorf("expected kind 'kenpom', got %q", m.Kind)
	}
}

func TestThatNormalizeDefaultsSigmaTo10(t *testing.T) {
	// GIVEN a model with a non-positive sigma
	m := &Model{Kind: "kenpom", Sigma: 0}

	// WHEN Normalize is called
	m.Normalize()

	// THEN sigma is set to 10.0
	if m.Sigma != 10.0 {
		t.Errorf("expected sigma 10.0, got %f", m.Sigma)
	}
}

func TestThatWinProbReturns0Point5ForEqualTeams(t *testing.T) {
	// GIVEN a model and two teams with equal net ratings
	m := &Model{Kind: "kenpom", Sigma: 10.0}

	// WHEN WinProb is called with equal ratings
	prob := m.WinProb(5.0, 5.0)

	// THEN the probability is 0.5
	if math.Abs(prob-0.5) > 1e-9 {
		t.Errorf("expected 0.5, got %f", prob)
	}
}

func TestThatWinProbFavorsBetterTeam(t *testing.T) {
	// GIVEN a model and two teams where team1 is better
	m := &Model{Kind: "kenpom", Sigma: 10.0}

	// WHEN WinProb is called with team1 having higher rating
	prob := m.WinProb(20.0, 5.0)

	// THEN team1 has greater than 50% probability
	if prob <= 0.5 {
		t.Errorf("expected probability > 0.5, got %f", prob)
	}
}
