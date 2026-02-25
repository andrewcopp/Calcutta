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

// expectedEVTotal computes the deterministic sum of all team EVs.
// In the prediction model, all numParticipants teams earn round-1 credit
// (via bye or First Four win), then subsequent rounds follow standard
// single-elimination. The total is:
//
//	numParticipants*inc[1] + (numParticipants/2)*inc[2] + ... + 2*inc[log2(N)]
//
// This is equivalent to TournamentTotal with winners-per-round = [N, N/2, ...].
func expectedEVTotal(rules []scoring.Rule, numParticipants int) float64 {
	winnersPerRound := append([]int{numParticipants}, scoring.GamesPerRoundForBracket(numParticipants)...)
	return float64(scoring.TournamentTotal(rules, winnersPerRound))
}

// expectedFavoritesTotal computes the deterministic Favorites bracket total.
// The Favorites bracket resolves each game by picking the higher-probability
// team, awarding actual game wins only (no participation credit). The total is
// simply the games-based tournament total: games_per_round[i] * inc[i+1].
func expectedFavoritesTotal(rules []scoring.Rule, numTeamsInBracket int) float64 {
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
// Expected EV total: 4*10 + 2*20 = 80
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

	// THEN the sum equals the expected total (4*10 + 2*20 = 80)
	expected := expectedEVTotal(rules, 4)
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
	expected := expectedFavoritesTotal(rules, 4)
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

	// THEN the sum equals the expected total (8*10 + 4*20 + 2*40 = 240)
	expected := expectedEVTotal(rules, 8)
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
	expected := expectedFavoritesTotal(rules, 8)
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

	// THEN the sum equals the expected total for 64 participants
	// (64 teams survive FF, then 6 rounds of single-elimination)
	expected := expectedEVTotal(rules, 64)
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
	expected := expectedFavoritesTotal(rules, 64)
	if math.Abs(favSum-expected) > 0.01 {
		t.Errorf("Favorites sum = %.4f, expected %.1f", favSum, expected)
	}
}

// --- Checkpoint fixtures using 68-team field ---

func TestThatCheckpointOneEVSumsToTournamentTotal(t *testing.T) {
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

	// THEN the sum equals the expected total for 64 participants
	expected := expectedEVTotal(rules, 64)
	if math.Abs(evSum-expected) > 0.01 {
		t.Errorf("EV sum = %.4f, expected %.1f", evSum, expected)
	}
}

func TestThatCheckpointOneFavoritesSumToTournamentTotal(t *testing.T) {
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

	// THEN the sum equals the EV total (3840) — at checkpoint, all alive teams
	// already have base progress from byes/FF wins, so Favorites includes that.
	expected := expectedEVTotal(rules, 64)
	if math.Abs(favSum-expected) > 0.01 {
		t.Errorf("Favorites sum = %.4f, expected %.1f", favSum, expected)
	}
}
