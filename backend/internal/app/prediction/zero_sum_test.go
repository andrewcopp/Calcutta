package prediction

import (
	"math"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
)

// --- helpers ---

func sumExpectedPoints(values []PredictedTeamValue) float64 {
	var total float64
	for _, v := range values {
		total += v.ExpectedPoints
	}
	return total
}

func sumFavoritesPoints(values []PredictedTeamValue) float64 {
	var total float64
	for _, v := range values {
		total += v.FavoritesTotalPoints
	}
	return total
}

// expectedTournamentTotal computes the deterministic tournament total:
// games_per_round[i] * inc[i+1] for a standard single-elimination bracket.
// Both EV and Favorites should sum to this value for pre-tournament predictions.
func expectedTournamentTotal(rules []scoring.Rule, numTeamsInBracket int) float64 {
	return float64(scoring.TournamentTotal(rules, scoring.GamesPerRoundForBracket(numTeamsInBracket)))
}

// --- 4-team fixture ---

// fourTeamValues builds a 4-team bracket with hand-chosen win probabilities.
//
// Teams: A, B, C, D
// Semi 1: A vs B  (roundOrder=1)
// Semi 2: C vs D  (roundOrder=1)
// Final: {A,B} vs {C,D} (roundOrder=2)
//
// Rules: {WinIndex:1, 10}, {WinIndex:2, 20}
// Expected EV total: 2*10 + 1*20 = 40
func fourTeamValues() ([]PredictedTeamValue, []scoring.Rule) {
	teams := []TeamInput{
		{ID: "A", Seed: 1, Region: "R1"},
		{ID: "B", Seed: 2, Region: "R1"},
		{ID: "C", Seed: 1, Region: "R2"},
		{ID: "D", Seed: 2, Region: "R2"},
	}
	rules := []scoring.Rule{
		{WinIndex: 1, PointsAwarded: 10},
		{WinIndex: 2, PointsAwarded: 20},
	}

	strength := map[string]float64{"A": 20, "B": 10, "C": 15, "D": 5}
	calcWinProb := func(id1, id2 string) float64 {
		return strength[id1] / (strength[id1] + strength[id2])
	}

	pAdvance := map[string]float64{"A": 1.0, "B": 1.0, "C": 1.0, "D": 1.0}
	semiGames := []gameSetup{
		{gameID: "S1", side1: teams[0:1], side2: teams[1:2]},
		{gameID: "S2", side1: teams[2:3], side2: teams[3:4]},
	}
	r1Matchups, pAdv := computeRound(semiGames, pAdvance, calcWinProb, 1)

	finalGames := []gameSetup{
		{gameID: "F1", side1: teams[0:2], side2: teams[2:4]},
	}
	r2Matchups, _ := computeRound(finalGames, pAdv, calcWinProb, 2)

	allMatchups := append(r1Matchups, r2Matchups...)
	values := GenerateTournamentValues(teams, allMatchups, 0, rules)
	return values, rules
}

func TestThatFourTeamEVSumsToTournamentTotal(t *testing.T) {
	// GIVEN a 4-team bracket with known win probabilities
	values, rules := fourTeamValues()

	// WHEN summing all team EVs
	evSum := sumExpectedPoints(values)

	// THEN the sum equals the expected total (2*10 + 1*20 = 40)
	expected := expectedTournamentTotal(rules, 4)
	if math.Abs(evSum-expected) > 1e-9 {
		t.Errorf("EV sum = %.10f, expected %.1f", evSum, expected)
	}
}

func TestThatFourTeamFavoritesSumToTournamentTotal(t *testing.T) {
	// GIVEN a 4-team bracket with known win probabilities
	values, rules := fourTeamValues()

	// WHEN summing all team Favorites points
	favSum := sumFavoritesPoints(values)

	// THEN the sum equals the Favorites total (2*10 + 1*20 = 40)
	expected := expectedTournamentTotal(rules, 4)
	if math.Abs(favSum-expected) > 1e-9 {
		t.Errorf("Favorites sum = %.10f, expected %.1f", favSum, expected)
	}
}

// --- 8-team fixture ---

func eightTeamValues() ([]PredictedTeamValue, []scoring.Rule) {
	teams := []TeamInput{
		{ID: "A", Seed: 1, Region: "R1"},
		{ID: "B", Seed: 2, Region: "R1"},
		{ID: "C", Seed: 3, Region: "R1"},
		{ID: "D", Seed: 4, Region: "R1"},
		{ID: "E", Seed: 1, Region: "R2"},
		{ID: "F", Seed: 2, Region: "R2"},
		{ID: "G", Seed: 3, Region: "R2"},
		{ID: "H", Seed: 4, Region: "R2"},
	}
	rules := []scoring.Rule{
		{WinIndex: 1, PointsAwarded: 10},
		{WinIndex: 2, PointsAwarded: 20},
		{WinIndex: 3, PointsAwarded: 40},
	}

	strength := map[string]float64{
		"A": 20, "B": 15, "C": 10, "D": 5,
		"E": 18, "F": 12, "G": 8, "H": 3,
	}
	calcWinProb := func(id1, id2 string) float64 {
		return strength[id1] / (strength[id1] + strength[id2])
	}

	pAdvance := make(map[string]float64)
	for _, t := range teams {
		pAdvance[t.ID] = 1.0
	}

	r1Games := []gameSetup{
		{gameID: "QF1", side1: teams[0:1], side2: teams[1:2]},
		{gameID: "QF2", side1: teams[2:3], side2: teams[3:4]},
		{gameID: "QF3", side1: teams[4:5], side2: teams[5:6]},
		{gameID: "QF4", side1: teams[6:7], side2: teams[7:8]},
	}
	r1Matchups, pAdv := computeRound(r1Games, pAdvance, calcWinProb, 1)

	r2Games := []gameSetup{
		{gameID: "SF1", side1: teams[0:2], side2: teams[2:4]},
		{gameID: "SF2", side1: teams[4:6], side2: teams[6:8]},
	}
	r2Matchups, pAdv := computeRound(r2Games, pAdv, calcWinProb, 2)

	r3Games := []gameSetup{
		{gameID: "F1", side1: teams[0:4], side2: teams[4:8]},
	}
	r3Matchups, _ := computeRound(r3Games, pAdv, calcWinProb, 3)

	allMatchups := append(r1Matchups, r2Matchups...)
	allMatchups = append(allMatchups, r3Matchups...)
	values := GenerateTournamentValues(teams, allMatchups, 0, rules)
	return values, rules
}

func TestThatEightTeamEVSumsToTournamentTotal(t *testing.T) {
	// GIVEN an 8-team bracket with strength-based win probabilities
	values, rules := eightTeamValues()

	// WHEN summing all team EVs
	evSum := sumExpectedPoints(values)

	// THEN the sum equals the expected total (4*10 + 2*20 + 1*40 = 120)
	expected := expectedTournamentTotal(rules, 8)
	if math.Abs(evSum-expected) > 1e-9 {
		t.Errorf("EV sum = %.10f, expected %.1f", evSum, expected)
	}
}

func TestThatEightTeamFavoritesSumToTournamentTotal(t *testing.T) {
	// GIVEN an 8-team bracket with strength-based win probabilities
	values, rules := eightTeamValues()

	// WHEN summing all team Favorites points
	favSum := sumFavoritesPoints(values)

	// THEN the sum equals the Favorites total (4*10 + 2*20 + 1*40 = 120)
	expected := expectedTournamentTotal(rules, 8)
	if math.Abs(favSum-expected) > 1e-9 {
		t.Errorf("Favorites sum = %.10f, expected %.1f", favSum, expected)
	}
}

// --- Full 68-team tournament fixture ---

func TestThatFullTournamentEVSumsToTournamentTotal(t *testing.T) {
	// GIVEN a 68-team tournament field with real matchup generation
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values and summing EVs
	values := GenerateTournamentValues(teams, matchups, 0, rules)
	evSum := sumExpectedPoints(values)

	// THEN the sum equals the expected NCAA total (1920)
	expected := float64(scoring.TournamentTotal(rules, scoring.NCAAgamesPerRound()))
	if math.Abs(evSum-expected) > 0.01 {
		t.Errorf("EV sum = %.4f, expected %.1f", evSum, expected)
	}
}

func TestThatFullTournamentFavoritesSumToTournamentTotal(t *testing.T) {
	// GIVEN a 68-team tournament field with real matchup generation
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN computing tournament values and summing Favorites
	values := GenerateTournamentValues(teams, matchups, 0, rules)
	favSum := sumFavoritesPoints(values)

	// THEN the sum equals the Favorites total (1920)
	expected := float64(scoring.TournamentTotal(rules, scoring.NCAAgamesPerRound()))
	if math.Abs(favSum-expected) > 0.01 {
		t.Errorf("Favorites sum = %.4f, expected %.1f", favSum, expected)
	}
}

// --- Checkpoint fixtures using 68-team field ---

func TestThatCheckpointOneEVSumsToExpectedTotal(t *testing.T) {
	// GIVEN a 68-team field at checkpoint 1 (First Four resolved)
	teams := generateCheckpoint1Teams()
	survivors := filterSurvivors(teams, 1)
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(survivors, 1, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating values for all teams and summing EVs
	values := GenerateTournamentValues(teams, matchups, 1, rules)
	evSum := sumExpectedPoints(values)

	// THEN the sum equals 1920 (same as pre-tournament).
	// With WinIndex 1=0, R128 survival awards no points, so the pool total
	// is conserved across the First Four checkpoint. All 64 alive teams have
	// actualPoints=0 and the future rounds sum to 1920.
	expected := 1920.0
	if math.Abs(evSum-expected) > 0.01 {
		t.Errorf("EV sum = %.4f, expected %.1f", evSum, expected)
	}
}

func TestThatCompletedTournamentAtCheckpointZeroSumsToExpectedTotal(t *testing.T) {
	// GIVEN a 68-team field with final wins/byes (as if tournament is complete)
	teams := generateTestTeams()
	for i := range teams {
		teams[i].Byes = 1
		if i == 0 {
			teams[i].Wins = 6 // champion
		} else if i < 4 {
			teams[i].Wins = 4 // F4 losers
		} else if i < 8 {
			teams[i].Wins = 3 // E8 losers
		} else {
			teams[i].Wins = 1 // early losers
		}
	}
	rules := DefaultScoringRules()
	data := &TournamentData{Teams: teams, Rules: rules}
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating predictions at throughRound=0 (pre-tournament checkpoint)
	state := NewTournamentState(data, 0)
	values, err := generatePredictions(state, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	evSum := sumExpectedPoints(values)

	// THEN EV sum equals the tournament total (1920), not inflated by final progress
	expected := expectedTournamentTotal(rules, 128)
	if math.Abs(evSum-expected) > 0.01 {
		t.Errorf("EV sum = %.4f, expected %.1f", evSum, expected)
	}
}

func TestThatCheckpointOneFavoritesSumToExpectedTotal(t *testing.T) {
	// GIVEN a 68-team field at checkpoint 1 (First Four resolved)
	teams := generateCheckpoint1Teams()
	survivors := filterSurvivors(teams, 1)
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(survivors, 1, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN computing values and summing Favorites
	values := GenerateTournamentValues(teams, matchups, 1, rules)
	favSum := sumFavoritesPoints(values)

	// THEN the sum equals 1920 (same as pre-tournament).
	// With WinIndex 1=0, R128 survival awards no points, so the Favorites
	// total is conserved across the First Four checkpoint.
	expected := 1920.0
	if math.Abs(favSum-expected) > 0.01 {
		t.Errorf("Favorites sum = %.4f, expected %.1f", favSum, expected)
	}
}
