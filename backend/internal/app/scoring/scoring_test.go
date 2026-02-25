package scoring

import "testing"

func TestThatPointsForProgressReturnsZeroWhenRulesEmpty(t *testing.T) {
	// GIVEN
	GIVENRules := []Rule(nil)
	GIVENWins := 3
	GIVENByes := 0

	// WHEN
	WHENPoints := PointsForProgress(GIVENRules, GIVENWins, GIVENByes)

	// THEN
	if WHENPoints != 0 {
		t.Fatalf("expected 0, got %d", WHENPoints)
	}
}

func TestThatPointsForProgressReturnsZeroWhenProgressIsZero(t *testing.T) {
	// GIVEN
	GIVENRules := []Rule{{WinIndex: 1, PointsAwarded: 10}}

	// WHEN
	WHENPoints := PointsForProgress(GIVENRules, 0, 0)

	// THEN
	if WHENPoints != 0 {
		t.Fatalf("expected 0, got %d", WHENPoints)
	}
}

func TestThatPointsForProgressReturnsZeroWhenProgressIsNegative(t *testing.T) {
	// GIVEN
	GIVENRules := []Rule{{WinIndex: 1, PointsAwarded: 10}}

	// WHEN
	WHENPoints := PointsForProgress(GIVENRules, -1, 0)

	// THEN
	if WHENPoints != 0 {
		t.Fatalf("expected 0, got %d", WHENPoints)
	}
}

func TestThatPointsForProgressSumsRulesUpToProgress(t *testing.T) {
	// GIVEN
	GIVENRules := []Rule{
		{WinIndex: 1, PointsAwarded: 0},
		{WinIndex: 2, PointsAwarded: 50},
		{WinIndex: 3, PointsAwarded: 100},
		{WinIndex: 4, PointsAwarded: 150},
	}
	GIVENWins := 3
	GIVENByes := 0

	// WHEN
	WHENPoints := PointsForProgress(GIVENRules, GIVENWins, GIVENByes)

	// THEN
	if WHENPoints != 150 {
		t.Fatalf("expected 150, got %d", WHENPoints)
	}
}

func TestThatPointsForProgressCountsByesAsProgress(t *testing.T) {
	// GIVEN
	GIVENRules := []Rule{
		{WinIndex: 1, PointsAwarded: 0},
		{WinIndex: 2, PointsAwarded: 50},
		{WinIndex: 3, PointsAwarded: 100},
	}
	GIVENWins := 2
	GIVENByes := 1

	// WHEN
	WHENPoints := PointsForProgress(GIVENRules, GIVENWins, GIVENByes)

	// THEN
	if WHENPoints != 150 {
		t.Fatalf("expected 150, got %d", WHENPoints)
	}
}

func TestThatPointsForProgressIsOrderIndependent(t *testing.T) {
	// GIVEN
	GIVENRules := []Rule{
		{WinIndex: 3, PointsAwarded: 100},
		{WinIndex: 1, PointsAwarded: 0},
		{WinIndex: 2, PointsAwarded: 50},
	}

	// WHEN
	WHENPoints := PointsForProgress(GIVENRules, 3, 0)

	// THEN
	if WHENPoints != 150 {
		t.Fatalf("expected 150, got %d", WHENPoints)
	}
}

// --- TournamentTotal tests ---

func TestThatTournamentTotalForFourTeamBracketIsForty(t *testing.T) {
	// GIVEN a 4-team bracket [2,1] with rules {10,20}
	rules := []Rule{{WinIndex: 1, PointsAwarded: 10}, {WinIndex: 2, PointsAwarded: 20}}
	gamesPerRound := []int{2, 1}

	// WHEN computing tournament total
	total := TournamentTotal(rules, gamesPerRound)

	// THEN total is 2*10 + 1*20 = 40
	if total != 40 {
		t.Fatalf("expected 40, got %d", total)
	}
}

func TestThatTournamentTotalForEightTeamBracketIsOneHundredTwenty(t *testing.T) {
	// GIVEN an 8-team bracket [4,2,1] with rules {10,20,40}
	rules := []Rule{
		{WinIndex: 1, PointsAwarded: 10},
		{WinIndex: 2, PointsAwarded: 20},
		{WinIndex: 3, PointsAwarded: 40},
	}
	gamesPerRound := []int{4, 2, 1}

	// WHEN computing tournament total
	total := TournamentTotal(rules, gamesPerRound)

	// THEN total is 4*10 + 2*20 + 1*40 = 120
	if total != 120 {
		t.Fatalf("expected 120, got %d", total)
	}
}

func TestThatTournamentTotalForSixtyFourTeamBracketIsNineteenTwenty(t *testing.T) {
	// GIVEN a 64-team bracket with default scoring rules
	rules := []Rule{
		{WinIndex: 1, PointsAwarded: 10},
		{WinIndex: 2, PointsAwarded: 20},
		{WinIndex: 3, PointsAwarded: 40},
		{WinIndex: 4, PointsAwarded: 80},
		{WinIndex: 5, PointsAwarded: 160},
		{WinIndex: 6, PointsAwarded: 320},
	}
	gamesPerRound := []int{32, 16, 8, 4, 2, 1}

	// WHEN computing tournament total
	total := TournamentTotal(rules, gamesPerRound)

	// THEN total is 32*10 + 16*20 + 8*40 + 4*80 + 2*160 + 1*320 = 1920
	if total != 1920 {
		t.Fatalf("expected 1920, got %d", total)
	}
}

func TestThatTournamentTotalIsZeroForEmptyGames(t *testing.T) {
	// GIVEN rules but no games
	rules := []Rule{{WinIndex: 1, PointsAwarded: 10}}
	gamesPerRound := []int{}

	// WHEN computing tournament total
	total := TournamentTotal(rules, gamesPerRound)

	// THEN total is 0
	if total != 0 {
		t.Fatalf("expected 0, got %d", total)
	}
}

func TestThatTournamentTotalIgnoresExtraRulesBeyondGamesRounds(t *testing.T) {
	// GIVEN 2-round bracket but 4 scoring rules
	rules := []Rule{
		{WinIndex: 1, PointsAwarded: 10},
		{WinIndex: 2, PointsAwarded: 20},
		{WinIndex: 3, PointsAwarded: 40},
		{WinIndex: 4, PointsAwarded: 80},
	}
	gamesPerRound := []int{2, 1}

	// WHEN computing tournament total
	total := TournamentTotal(rules, gamesPerRound)

	// THEN only first 2 rounds count: 2*10 + 1*20 = 40
	if total != 40 {
		t.Fatalf("expected 40, got %d", total)
	}
}

// --- GamesPerRoundForBracket tests ---

func TestThatGamesPerRoundForFourTeamBracketIsTwoOne(t *testing.T) {
	// GIVEN a 4-team bracket
	// WHEN computing games per round
	result := GamesPerRoundForBracket(4)

	// THEN result is [2, 1]
	expected := []int{2, 1}
	if len(result) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	}
}

func TestThatGamesPerRoundForSixtyFourTeamBracketIsCorrect(t *testing.T) {
	// GIVEN a 64-team bracket
	// WHEN computing games per round
	result := GamesPerRoundForBracket(64)

	// THEN result is [32, 16, 8, 4, 2, 1]
	expected := []int{32, 16, 8, 4, 2, 1}
	if len(result) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	}
}

func TestThatGamesPerRoundForOneTeamReturnsNil(t *testing.T) {
	// GIVEN a single team (no bracket possible)
	// WHEN computing games per round
	result := GamesPerRoundForBracket(1)

	// THEN result is nil
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}
