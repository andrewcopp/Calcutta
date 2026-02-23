package prediction

import (
	"math"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
)

func testRules() []scoring.Rule {
	return []scoring.Rule{
		{WinIndex: 1, PointsAwarded: 10},
		{WinIndex: 2, PointsAwarded: 20},
		{WinIndex: 3, PointsAwarded: 40},
		{WinIndex: 4, PointsAwarded: 80},
		{WinIndex: 5, PointsAwarded: 160},
		{WinIndex: 6, PointsAwarded: 320},
	}
}

func TestThatProjectedTeamEVReturnsActualPointsForEliminatedTeam(t *testing.T) {
	// GIVEN a team that has been eliminated after 2 wins
	ptv := PredictedTeamValue{
		TeamID:         "team-1",
		ExpectedPoints: 200.0,
		PRound1:        1.0,
		PRound2:        0.8,
		PRound3:        0.5,
		PRound4:        0.3,
		PRound5:        0.1,
		PRound6:        0.05,
	}
	rules := testRules()
	tp := TeamProgress{Wins: 2, Byes: 0, IsEliminated: true}

	// WHEN computing projected EV
	result := ProjectedTeamEV(ptv, rules, tp, 0)

	// THEN result equals actual points for 2 wins (10 + 20 = 30)
	expected := 30.0
	if math.Abs(result-expected) > 0.001 {
		t.Errorf("expected %.2f, got %.2f", expected, result)
	}
}

func TestThatProjectedTeamEVReturnsExpectedPointsPreTournament(t *testing.T) {
	// GIVEN a team with no wins or byes (pre-tournament)
	ptv := PredictedTeamValue{
		TeamID:         "team-1",
		ExpectedPoints: 150.0,
		PRound1:        1.0,
		PRound2:        0.9,
		PRound3:        0.6,
		PRound4:        0.3,
		PRound5:        0.1,
		PRound6:        0.05,
	}
	rules := testRules()
	tp := TeamProgress{Wins: 0, Byes: 0, IsEliminated: false}

	// WHEN computing projected EV
	result := ProjectedTeamEV(ptv, rules, tp, 0)

	// THEN result equals the pre-computed expected points
	if math.Abs(result-150.0) > 0.001 {
		t.Errorf("expected 150.00, got %.2f", result)
	}
}

func TestThatProjectedTeamEVComputesConditionalRemainingForAliveTeam(t *testing.T) {
	// GIVEN a team alive after 1 win with known probabilities
	ptv := PredictedTeamValue{
		TeamID:         "team-1",
		ExpectedPoints: 100.0,
		PRound1:        1.0,  // P(survive round 1) = 1.0
		PRound2:        0.50, // P(survive round 2) = 0.50
		PRound3:        0.25, // P(survive round 3) = 0.25
		PRound4:        0.0,
		PRound5:        0.0,
		PRound6:        0.0,
	}
	rules := []scoring.Rule{
		{WinIndex: 1, PointsAwarded: 10},
		{WinIndex: 2, PointsAwarded: 20},
		{WinIndex: 3, PointsAwarded: 40},
	}
	tp := TeamProgress{Wins: 1, Byes: 0, IsEliminated: false}

	// WHEN computing projected EV
	result := ProjectedTeamEV(ptv, rules, tp, 0)

	// THEN result = actual(10) + conditional remaining
	// progress = 1, pAlive = PRound1 = 1.0
	// round 2: (0.50/1.0) * 20 = 10.0
	// round 3: (0.25/1.0) * 40 = 10.0
	// total = 10 + 10 + 10 = 30.0
	expected := 30.0
	if math.Abs(result-expected) > 0.001 {
		t.Errorf("expected %.2f, got %.2f", expected, result)
	}
}

func TestThatProjectedTeamEVFallsBackToActualWhenPAliveIsZero(t *testing.T) {
	// GIVEN a team alive after 2 wins but model predicted 0% chance of being here
	ptv := PredictedTeamValue{
		TeamID:         "team-1",
		ExpectedPoints: 50.0,
		PRound1:        1.0,
		PRound2:        0.0, // model predicted 0% chance of surviving round 2
		PRound3:        0.0,
		PRound4:        0.0,
		PRound5:        0.0,
		PRound6:        0.0,
	}
	rules := testRules()
	tp := TeamProgress{Wins: 2, Byes: 0, IsEliminated: false}

	// WHEN computing projected EV
	result := ProjectedTeamEV(ptv, rules, tp, 0)

	// THEN falls back to actual points (10 + 20 = 30)
	expected := 30.0
	if math.Abs(result-expected) > 0.001 {
		t.Errorf("expected %.2f, got %.2f", expected, result)
	}
}

func TestThatProjectedTeamEVHandlesTeamWithBye(t *testing.T) {
	// GIVEN a team with 1 bye and 1 win (progress = 2), still alive
	ptv := PredictedTeamValue{
		TeamID:         "team-1",
		ExpectedPoints: 100.0,
		PRound1:        1.0,
		PRound2:        1.0, // guaranteed to survive round 2 (had bye)
		PRound3:        0.6,
		PRound4:        0.3,
		PRound5:        0.1,
		PRound6:        0.05,
	}
	rules := testRules()
	tp := TeamProgress{Wins: 1, Byes: 1, IsEliminated: false}

	// WHEN computing projected EV
	result := ProjectedTeamEV(ptv, rules, tp, 0)

	// THEN progress = 2, pAlive = PRound2 = 1.0, actual = PointsForProgress(1 win, 1 bye) = 30
	// conditional remaining:
	// round 3: (0.6/1.0) * 40 = 24.0
	// round 4: (0.3/1.0) * 80 = 24.0
	// round 5: (0.1/1.0) * 160 = 16.0
	// round 6: (0.05/1.0) * 320 = 16.0
	// total = 30 + 24 + 24 + 16 + 16 = 110.0
	expected := 110.0
	if math.Abs(result-expected) > 0.001 {
		t.Errorf("expected %.2f, got %.2f", expected, result)
	}
}
