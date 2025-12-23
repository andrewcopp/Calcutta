package services

import (
	"math"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

func floatPtr(v float64) *float64 { return &v }

func TestKenPomWinProbFromMargin_Symmetry(t *testing.T) {
	sigma := 11.0
	margins := []float64{-30, -20, -10, -5, 0, 5, 10, 20, 30}
	for _, m := range margins {
		p := kenPomWinProbFromMargin(m, sigma)
		q := kenPomWinProbFromMargin(-m, sigma)
		if math.Abs((p+q)-1.0) > 1e-9 {
			t.Fatalf("expected symmetry p(m)+p(-m)=1, got p(%g)=%g q=%g sum=%g", m, p, q, p+q)
		}
	}
}

func TestKenPomWinProbFromMargin_Monotonicity(t *testing.T) {
	sigma := 11.0
	margins := []float64{-25, -10, -1, 0, 1, 10, 25}
	prev := kenPomWinProbFromMargin(margins[0], sigma)
	for i := 1; i < len(margins); i++ {
		cur := kenPomWinProbFromMargin(margins[i], sigma)
		if cur < prev {
			t.Fatalf("expected monotonic non-decreasing; margin %g p=%g < margin %g p=%g", margins[i], cur, margins[i-1], prev)
		}
		prev = cur
	}
}

func TestKenPomExpectedMargin_UsesAdjEMDifference(t *testing.T) {
	teamA := &models.TournamentTeam{ID: "A", KenPom: &models.KenPomStats{ORtg: floatPtr(120), DRtg: floatPtr(90), AdjT: floatPtr(70)}}
	teamB := &models.TournamentTeam{ID: "B", KenPom: &models.KenPomStats{ORtg: floatPtr(110), DRtg: floatPtr(100), AdjT: floatPtr(70)}}

	m, ok := kenPomExpectedMargin(teamA, teamB)
	if !ok {
		t.Fatalf("expected kenPomExpectedMargin ok")
	}

	// AdjEM_A = 120-90=30, AdjEM_B = 110-100=10, diff=20 points per 100 poss.
	// Possessions = (70+70)/2 = 70 => margin = 20*(70/100)=14.
	want := 14.0
	if math.Abs(m-want) > 1e-9 {
		t.Fatalf("expected margin %g, got %g", want, m)
	}

	p, err := kenPomWinProb(teamA, teamB, 11.0)
	if err != nil {
		t.Fatalf("expected kenPomWinProb success, got err: %v", err)
	}
	if p <= 0.5 {
		t.Fatalf("expected teamA favored (p>0.5), got %g", p)
	}
}

func TestKenPomExpectedMargin_MissingStats(t *testing.T) {
	teamA := &models.TournamentTeam{ID: "A", KenPom: &models.KenPomStats{ORtg: floatPtr(120), DRtg: floatPtr(90), AdjT: floatPtr(70)}}
	teamB := &models.TournamentTeam{ID: "B"}

	_, ok := kenPomExpectedMargin(teamA, teamB)
	if ok {
		t.Fatalf("expected kenPomExpectedMargin not ok when stats missing")
	}

	_, err := kenPomWinProb(teamA, teamB, 11.0)
	if err == nil {
		t.Fatalf("expected kenPomWinProb error when stats missing")
	}
}
