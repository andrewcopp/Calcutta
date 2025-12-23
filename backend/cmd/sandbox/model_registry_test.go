package main

import (
	"testing"
)

func TestCanonicalModelName(t *testing.T) {
	if got := canonicalModelName("  Ridge-Returns "); got != "ridge-returns" {
		t.Fatalf("canonicalModelName mismatch: got %q", got)
	}
}

func TestClamp01(t *testing.T) {
	if clamp01(-0.1) != 0 {
		t.Fatalf("expected clamp01(-0.1)=0")
	}
	if clamp01(1.1) != 1 {
		t.Fatalf("expected clamp01(1.1)=1")
	}
	if clamp01(0.3) != 0.3 {
		t.Fatalf("expected clamp01(0.3)=0.3")
	}
}

func TestNormalizeNonNegativeScoresForTeams(t *testing.T) {
	ids := []string{"a", "b", "c"}
	scores := map[string]float64{"a": 1, "b": 1, "c": -5}
	shares, sum, err := normalizeNonNegativeScoresForTeams(ids, scores)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if sum != 2 {
		t.Fatalf("expected sum=2 got %v", sum)
	}
	if shares["a"] != 0.5 || shares["b"] != 0.5 || shares["c"] != 0 {
		t.Fatalf("unexpected shares: %#v", shares)
	}
}

func TestRegistryHasBuiltins(t *testing.T) {
	inv, err := GetInvestmentModel("seed")
	if err != nil {
		t.Fatalf("expected seed investment model to exist: %v", err)
	}
	if inv == nil {
		t.Fatalf("expected non-nil seed investment model")
	}
	pts, err := GetPointsModel("seed")
	if err != nil {
		t.Fatalf("expected seed points model to exist: %v", err)
	}
	if pts == nil {
		t.Fatalf("expected non-nil seed points model")
	}
}
