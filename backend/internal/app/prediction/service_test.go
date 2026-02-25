package prediction

import "testing"

func TestThatDetectThroughRoundReturnsZeroWhenNoTeamsHaveWins(t *testing.T) {
	// GIVEN a set of teams with no wins
	teams := []TeamInput{
		{ID: "t1", Seed: 1, Wins: 0, Byes: 1},
		{ID: "t2", Seed: 16, Wins: 0, Byes: 0},
		{ID: "t3", Seed: 8, Wins: 0, Byes: 1},
	}

	// WHEN detecting the through round
	result := detectThroughRoundFromTeams(teams)

	// THEN through round is 0 (tournament hasn't started)
	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}
}

func TestThatDetectThroughRoundReturnsMaxProgressWhenWinsExist(t *testing.T) {
	// GIVEN teams where one has 2 wins and another has 1 win + 1 bye (progress=2)
	teams := []TeamInput{
		{ID: "t1", Seed: 1, Wins: 2, Byes: 1},  // progress = 3
		{ID: "t2", Seed: 16, Wins: 0, Byes: 0},  // progress = 0
		{ID: "t3", Seed: 8, Wins: 1, Byes: 1},   // progress = 2
		{ID: "t4", Seed: 9, Wins: 1, Byes: 0},   // progress = 1
	}

	// WHEN detecting the through round
	result := detectThroughRoundFromTeams(teams)

	// THEN through round equals the maximum progress (3)
	if result != 3 {
		t.Errorf("expected 3, got %d", result)
	}
}

func TestThatDetectThroughRoundReturnsZeroForEmptySlice(t *testing.T) {
	// GIVEN no teams
	teams := []TeamInput{}

	// WHEN detecting the through round
	result := detectThroughRoundFromTeams(teams)

	// THEN through round is 0
	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}
}

func TestThatDetectThroughRoundIncludesByesInProgressWhenWinsExist(t *testing.T) {
	// GIVEN a team with only byes and another with 1 win
	// The byes-only team should count toward progress since maxWins > 0
	teams := []TeamInput{
		{ID: "t1", Seed: 1, Wins: 0, Byes: 2},  // progress = 2, but 0 wins
		{ID: "t2", Seed: 16, Wins: 1, Byes: 0},  // progress = 1, triggers maxWins > 0
	}

	// WHEN detecting the through round
	result := detectThroughRoundFromTeams(teams)

	// THEN through round is 2 (t1's byes count because tournament has started)
	if result != 2 {
		t.Errorf("expected 2, got %d", result)
	}
}
