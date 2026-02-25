package prediction

import (
	"testing"
)

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

	// WHEN computing future wins from checkpoint 1
	result := computeFavoritesFutureWins(teams, matchups, 1)

	// THEN team-A gets 1 future win, team-B gets 0
	if result["team-A"] != 1 {
		t.Errorf("expected team-A=1 future win, got %d", result["team-A"])
	}
	if result["team-B"] != 0 {
		t.Errorf("expected team-B=0 future wins, got %d", result["team-B"])
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

	// WHEN computing future wins from checkpoint 1
	result := computeFavoritesFutureWins(teams, matchups, 1)

	// THEN team-A wins (the only matchup with both alive is A vs B, and A has 60% > 50%)
	if result["team-A"] != 1 {
		t.Errorf("expected team-A=1 future win, got %d", result["team-A"])
	}
	if result["team-B"] != 0 {
		t.Errorf("expected team-B=0 future wins, got %d", result["team-B"])
	}
}

func TestThatFavoritesBracketResolvesMultipleRoundsCorrectly(t *testing.T) {
	// GIVEN a team that advances through 2 rounds in the favorites bracket
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

	// WHEN computing future wins from checkpoint 0
	result := computeFavoritesFutureWins(teams, matchups, 0)

	// THEN team-A gets 2 future wins, team-C gets 1, team-B and D get 0
	if result["team-A"] != 2 {
		t.Errorf("expected team-A=2 future wins, got %d", result["team-A"])
	}
	if result["team-C"] != 1 {
		t.Errorf("expected team-C=1 future win, got %d", result["team-C"])
	}
	if result["team-B"] != 0 {
		t.Errorf("expected team-B=0 future wins, got %d", result["team-B"])
	}
}

func TestThatEliminatedTeamsGetZeroFutureWins(t *testing.T) {
	// GIVEN an eliminated team with 2 wins and an alive team
	teams := []TeamInput{
		{ID: "team-A", Wins: 2, Byes: 0},
		{ID: "team-B", Wins: 1, Byes: 0}, // eliminated (progress < throughRound=2)
	}
	// No matchups since team-A is the only one alive at checkpoint 2
	matchups := []PredictedMatchup{}

	// WHEN computing future wins from checkpoint 2
	result := computeFavoritesFutureWins(teams, matchups, 2)

	// THEN both teams get 0 future wins (no future games to play)
	if result["team-A"] != 0 {
		t.Errorf("expected team-A=0 future wins, got %d", result["team-A"])
	}
	if result["team-B"] != 0 {
		t.Errorf("expected team-B=0 future wins, got %d", result["team-B"])
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

	// WHEN computing future wins from checkpoint 0
	result := computeFavoritesFutureWins(teams, matchups, 0)

	// THEN Team1 (team-A) wins deterministically (PTeam1 >= 0.5)
	if result["team-A"] != 1 {
		t.Errorf("expected team-A=1 future win (winner), got %d", result["team-A"])
	}
	if result["team-B"] != 0 {
		t.Errorf("expected team-B=0 future wins (loser), got %d", result["team-B"])
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

	// WHEN computing future wins from checkpoint 1
	result := computeFavoritesFutureWins(teams, matchups, 1)

	// THEN team-B advances (0.7 > 0.5), gets 1 future win; team-A gets 0
	if result["team-B"] != 1 {
		t.Errorf("expected team-B=1 future win, got %d", result["team-B"])
	}
	if result["team-A"] != 0 {
		t.Errorf("expected team-A=0 future wins, got %d", result["team-A"])
	}
}
