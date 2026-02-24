package prediction

import (
	"math"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
)

func favoritesTestRules() []scoring.Rule {
	return []scoring.Rule{
		{WinIndex: 1, PointsAwarded: 10},
		{WinIndex: 2, PointsAwarded: 20},
		{WinIndex: 3, PointsAwarded: 40},
	}
}

func TestThatHigherProbabilityTeamAdvancesInFavoritesBracket(t *testing.T) {
	// GIVEN two alive teams where team-A has 70% win probability
	teams := []TeamInput{
		{ID: "team-A", Wins: 1, Byes: 0},
		{ID: "team-B", Wins: 1, Byes: 0},
	}
	matchups := []PredictedMatchup{
		{
			GameID:                 "R2-East-1",
			RoundOrder:             2,
			Team1ID:                "team-A",
			Team2ID:                "team-B",
			PMatchup:               1.0,
			PTeam1WinsGivenMatchup: 0.7,
			PTeam2WinsGivenMatchup: 0.3,
		},
	}

	// WHEN computing favorites bracket from checkpoint 1
	result := ComputeFavoritesBracket(teams, matchups, 1, favoritesTestRules())

	// THEN team-A gets 2 wins worth of points (10+20=30) and team-B stays at 1 win (10)
	if math.Abs(result["team-A"]-30.0) > 0.001 {
		t.Errorf("expected team-A=30.0, got %.2f", result["team-A"])
	}
	if math.Abs(result["team-B"]-10.0) > 0.001 {
		t.Errorf("expected team-B=10.0, got %.2f", result["team-B"])
	}
}

func TestThatFavoritesBracketPicksMostLikelyMatchupPerGame(t *testing.T) {
	// GIVEN a game with two possible matchups, where only one has both teams alive
	teams := []TeamInput{
		{ID: "team-A", Wins: 1, Byes: 0},
		{ID: "team-B", Wins: 1, Byes: 0},
		{ID: "team-C", Wins: 0, Byes: 0}, // eliminated (progress < throughRound=1)
	}
	matchups := []PredictedMatchup{
		{
			GameID:                 "R2-East-1",
			RoundOrder:             2,
			Team1ID:                "team-A",
			Team2ID:                "team-C",
			PMatchup:               0.3,
			PTeam1WinsGivenMatchup: 0.9,
			PTeam2WinsGivenMatchup: 0.1,
		},
		{
			GameID:                 "R2-East-1",
			RoundOrder:             2,
			Team1ID:                "team-A",
			Team2ID:                "team-B",
			PMatchup:               0.7,
			PTeam1WinsGivenMatchup: 0.6,
			PTeam2WinsGivenMatchup: 0.4,
		},
	}

	// WHEN computing favorites bracket from checkpoint 1
	result := ComputeFavoritesBracket(teams, matchups, 1, favoritesTestRules())

	// THEN team-A wins (the only matchup with both alive is A vs B, and A has 60% > 50%)
	if math.Abs(result["team-A"]-30.0) > 0.001 {
		t.Errorf("expected team-A=30.0 (2 wins), got %.2f", result["team-A"])
	}
	if math.Abs(result["team-B"]-10.0) > 0.001 {
		t.Errorf("expected team-B=10.0 (1 win), got %.2f", result["team-B"])
	}
}

func TestThatFavoritesBracketAppliesScoringRulesCorrectly(t *testing.T) {
	// GIVEN a team that starts with 0 wins and advances through 3 rounds
	teams := []TeamInput{
		{ID: "team-A", Wins: 0, Byes: 0},
		{ID: "team-B", Wins: 0, Byes: 0},
		{ID: "team-C", Wins: 0, Byes: 0},
		{ID: "team-D", Wins: 0, Byes: 0},
	}
	matchups := []PredictedMatchup{
		{GameID: "R1-1", RoundOrder: 1, Team1ID: "team-A", Team2ID: "team-B", PMatchup: 1.0, PTeam1WinsGivenMatchup: 0.8, PTeam2WinsGivenMatchup: 0.2},
		{GameID: "R1-2", RoundOrder: 1, Team1ID: "team-C", Team2ID: "team-D", PMatchup: 1.0, PTeam1WinsGivenMatchup: 0.9, PTeam2WinsGivenMatchup: 0.1},
		{GameID: "R2-1", RoundOrder: 2, Team1ID: "team-A", Team2ID: "team-C", PMatchup: 0.72, PTeam1WinsGivenMatchup: 0.55, PTeam2WinsGivenMatchup: 0.45},
	}

	// WHEN computing favorites bracket from checkpoint 0
	result := ComputeFavoritesBracket(teams, matchups, 0, favoritesTestRules())

	// THEN team-A gets 2 wins (30 pts), team-C gets 1 win (10 pts), team-B and D get 0
	if math.Abs(result["team-A"]-30.0) > 0.001 {
		t.Errorf("expected team-A=30.0, got %.2f", result["team-A"])
	}
	if math.Abs(result["team-C"]-10.0) > 0.001 {
		t.Errorf("expected team-C=10.0, got %.2f", result["team-C"])
	}
	if math.Abs(result["team-B"]-0.0) > 0.001 {
		t.Errorf("expected team-B=0.0, got %.2f", result["team-B"])
	}
}

func TestThatEliminatedTeamsKeepActualPointsInFavoritesBracket(t *testing.T) {
	// GIVEN an eliminated team with 2 wins and an alive team
	teams := []TeamInput{
		{ID: "team-A", Wins: 2, Byes: 0},
		{ID: "team-B", Wins: 1, Byes: 0}, // eliminated (progress < throughRound=2)
	}
	// No matchups since team-A is the only one alive at checkpoint 2
	matchups := []PredictedMatchup{}

	// WHEN computing favorites bracket from checkpoint 2
	result := ComputeFavoritesBracket(teams, matchups, 2, favoritesTestRules())

	// THEN team-A keeps 30 pts, team-B keeps 10 pts
	if math.Abs(result["team-A"]-30.0) > 0.001 {
		t.Errorf("expected team-A=30.0, got %.2f", result["team-A"])
	}
	if math.Abs(result["team-B"]-10.0) > 0.001 {
		t.Errorf("expected team-B=10.0, got %.2f", result["team-B"])
	}
}

func TestThatFavoritesBracketTiebreakIsDeterministic(t *testing.T) {
	// GIVEN a 50/50 matchup (exact tiebreak)
	teams := []TeamInput{
		{ID: "team-A", Wins: 0, Byes: 0},
		{ID: "team-B", Wins: 0, Byes: 0},
	}
	matchups := []PredictedMatchup{
		{
			GameID:                 "R1-1",
			RoundOrder:             1,
			Team1ID:                "team-A",
			Team2ID:                "team-B",
			PMatchup:               1.0,
			PTeam1WinsGivenMatchup: 0.5,
			PTeam2WinsGivenMatchup: 0.5,
		},
	}

	// WHEN computing favorites bracket from checkpoint 0
	result := ComputeFavoritesBracket(teams, matchups, 0, favoritesTestRules())

	// THEN Team1 (team-A) wins deterministically (PTeam1 >= 0.5)
	if math.Abs(result["team-A"]-10.0) > 0.001 {
		t.Errorf("expected team-A=10.0 (winner), got %.2f", result["team-A"])
	}
	if math.Abs(result["team-B"]-0.0) > 0.001 {
		t.Errorf("expected team-B=0.0 (loser), got %.2f", result["team-B"])
	}
}

func TestThatFavoritesBracketOnlyCountsFutureRounds(t *testing.T) {
	// GIVEN teams at checkpoint 1 with matchups for round 2
	teams := []TeamInput{
		{ID: "team-A", Wins: 1, Byes: 0},
		{ID: "team-B", Wins: 1, Byes: 0},
	}
	matchups := []PredictedMatchup{
		{
			GameID:                 "R2-1",
			RoundOrder:             2,
			Team1ID:                "team-A",
			Team2ID:                "team-B",
			PMatchup:               1.0,
			PTeam1WinsGivenMatchup: 0.3,
			PTeam2WinsGivenMatchup: 0.7,
		},
	}

	// WHEN computing favorites bracket from checkpoint 1
	result := ComputeFavoritesBracket(teams, matchups, 1, favoritesTestRules())

	// THEN team-B advances (0.7 > 0.5), gets 2 wins (30 pts); team-A stays at 1 win (10 pts)
	if math.Abs(result["team-B"]-30.0) > 0.001 {
		t.Errorf("expected team-B=30.0, got %.2f", result["team-B"])
	}
	if math.Abs(result["team-A"]-10.0) > 0.001 {
		t.Errorf("expected team-A=10.0, got %.2f", result["team-A"])
	}
}
