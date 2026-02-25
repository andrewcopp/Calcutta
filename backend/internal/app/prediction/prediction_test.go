package prediction

import (
	"fmt"
	"math"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
)

func TestThatGenerateMatchupsCreatesMatchupsForAllRounds(t *testing.T) {
	// GIVEN a 68-team tournament field
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating all theoretical matchups
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN matchups exist for all 6 rounds including championship
	roundCounts := make(map[int]int)
	for _, m := range matchups {
		roundCounts[m.RoundOrder]++
	}

	if roundCounts[6] == 0 {
		t.Error("expected round 6 (championship) matchups")
	}
}

func TestThatPMatchupSumsToOneForEachGame(t *testing.T) {
	// GIVEN a 68-team tournament field
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating all theoretical matchups
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN p_matchup sums to 1.0 for each game
	pSumByGame := make(map[string]float64)
	for _, m := range matchups {
		pSumByGame[m.GameID] += m.PMatchup
	}

	for gameID, pSum := range pSumByGame {
		if math.Abs(pSum-1.0) > 0.001 {
			t.Errorf("game %s: p_matchup sum = %.4f, expected 1.0", gameID, pSum)
		}
	}
}

func TestThatWinProbabilitiesSumToOneForEachMatchup(t *testing.T) {
	// GIVEN a 68-team tournament field
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating all theoretical matchups
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN win probabilities sum to 1.0 for each matchup
	for _, m := range matchups {
		pSum := m.PTeam1WinsGivenMatchup + m.PTeam2WinsGivenMatchup
		if math.Abs(pSum-1.0) > 0.0001 {
			t.Errorf("matchup %s: win probs sum = %.4f, expected 1.0", m.GameID, pSum)
		}
	}
}

func TestThatAllTeamsHavePositiveExpectedPoints(t *testing.T) {
	// GIVEN a 68-team tournament field with matchups
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values
	values := GenerateTournamentValues(teams, matchups, 0, rules)

	// THEN all 68 teams have positive expected points
	if len(values) != 68 {
		t.Errorf("expected 68 team values, got %d", len(values))
	}
}

func TestThatChampionshipProbabilitiesSumToOne(t *testing.T) {
	// GIVEN a 68-team tournament field with matchups
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values
	values := GenerateTournamentValues(teams, matchups, 0, rules)

	// THEN championship probabilities sum to 1.0
	var pChampSum float64
	for _, v := range values {
		pChampSum += v.PRound7
	}

	if math.Abs(pChampSum-1.0) > 0.01 {
		t.Errorf("championship probabilities sum = %.4f, expected 1.0", pChampSum)
	}
}

func TestThatOneSeedHasHigherExpectedPointsThanSixteenSeed(t *testing.T) {
	// GIVEN a 68-team tournament field with matchups
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()
	values := GenerateTournamentValues(teams, matchups, 0, rules)

	// Build map of team_id -> expected_points
	expectedByTeam := make(map[string]float64)
	for _, v := range values {
		expectedByTeam[v.TeamID] = v.ExpectedPoints
	}

	// Find a 1-seed and a 16-seed from the same region
	var oneSeed, sixteenSeed TeamInput
	for _, team := range teams {
		if team.Region == "East" {
			if team.Seed == 1 {
				oneSeed = team
			} else if team.Seed == 16 {
				sixteenSeed = team
			}
		}
	}

	if oneSeed.ID == "" || sixteenSeed.ID == "" {
		t.Skip("could not find 1-seed and 16-seed in East region")
	}

	// WHEN comparing expected points
	oneExpected := expectedByTeam[oneSeed.ID]
	sixteenExpected := expectedByTeam[sixteenSeed.ID]

	// THEN 1-seed has higher expected points than 16-seed
	if oneExpected <= sixteenExpected {
		t.Errorf("1-seed expected points (%.2f) should be > 16-seed (%.2f)", oneExpected, sixteenExpected)
	}
}

func TestThatDefaultScoringRulesHasSixRounds(t *testing.T) {
	// GIVEN the default scoring rules
	// WHEN getting the rules
	rules := DefaultScoringRules()

	// THEN there are 6 rounds
	if len(rules) != 6 {
		t.Errorf("expected 6 scoring rules, got %d", len(rules))
	}
}

func TestThatDefaultScoringRulesDoublePointsEachRound(t *testing.T) {
	// GIVEN the default scoring rules
	rules := DefaultScoringRules()

	// WHEN checking the points progression
	expectedPoints := []int{10, 20, 40, 80, 160, 320}

	// THEN points double each round
	for i, r := range rules {
		if r.PointsAwarded != expectedPoints[i] {
			t.Errorf("rule %d: points = %d, expected %d", i, r.PointsAwarded, expectedPoints[i])
		}
	}
}

func TestThatChampionEarnsSixHundredThirtyPoints(t *testing.T) {
	// GIVEN the default scoring rules
	rules := DefaultScoringRules()

	// WHEN calculating total points for a champion (6 wins, 0 byes)
	totalPoints := scoring.PointsForProgress(rules, 6, 0)

	// THEN the champion earns 630 points
	expectedTotal := 10 + 20 + 40 + 80 + 160 + 320 // 630
	if totalPoints != expectedTotal {
		t.Errorf("total points for champion = %d, expected %d", totalPoints, expectedTotal)
	}
}

func TestThatNonFirstFourTeamsHaveFFProbabilityOfOne(t *testing.T) {
	// GIVEN a 68-team tournament field with matchups
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values
	values := GenerateTournamentValues(teams, matchups, 0, rules)

	// THEN non-First Four teams have PRound1 (FF survival) = 1.0
	ffTeamIDs := make(map[string]bool)
	for _, team := range teams {
		// FF teams are the extra 11-seeds and 16-seeds (seeds with >1 team per region)
		regionSeedCount := make(map[string]int)
		for _, t := range teams {
			if t.Region == team.Region && t.Seed == team.Seed {
				regionSeedCount[fmt.Sprintf("%s-%d", t.Region, t.Seed)]++
			}
		}
		if regionSeedCount[fmt.Sprintf("%s-%d", team.Region, team.Seed)] > 1 {
			ffTeamIDs[team.ID] = true
		}
	}

	for _, v := range values {
		if !ffTeamIDs[v.TeamID] && v.PRound1 != 1.0 {
			t.Errorf("non-FF team %s: PRound1 = %.4f, expected 1.0", v.TeamID, v.PRound1)
		}
	}
}

func TestThatFirstFourTeamsHaveFFProbabilityLessThanOne(t *testing.T) {
	// GIVEN a 68-team tournament field with matchups
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values
	values := GenerateTournamentValues(teams, matchups, 0, rules)

	// THEN First Four teams have PRound1 (FF survival) < 1.0
	ffTeamIDs := make(map[string]bool)
	for _, team := range teams {
		regionSeedCount := 0
		for _, t := range teams {
			if t.Region == team.Region && t.Seed == team.Seed {
				regionSeedCount++
			}
		}
		if regionSeedCount > 1 {
			ffTeamIDs[team.ID] = true
		}
	}

	valueByID := make(map[string]PredictedTeamValue)
	for _, v := range values {
		valueByID[v.TeamID] = v
	}

	for id := range ffTeamIDs {
		v := valueByID[id]
		if v.PRound1 >= 1.0 {
			t.Errorf("FF team %s: PRound1 = %.4f, expected < 1.0", id, v.PRound1)
		}
	}
}

func TestThatProbabilitiesAreMonotonicallyDecreasing(t *testing.T) {
	// GIVEN a 68-team tournament field with matchups
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values
	values := GenerateTournamentValues(teams, matchups, 0, rules)

	// THEN all teams have monotonically decreasing probabilities
	for _, v := range values {
		probs := []float64{v.PRound1, v.PRound2, v.PRound3, v.PRound4, v.PRound5, v.PRound6, v.PRound7}
		for i := 1; i < len(probs); i++ {
			if probs[i] > probs[i-1] {
				t.Errorf("team %s: PRound%d (%.4f) > PRound%d (%.4f)", v.TeamID, i+1, probs[i], i, probs[i-1])
			}
		}
	}
}

func TestThatOneSeedReachesSweetSixteenMoreThanFiftyPercent(t *testing.T) {
	// GIVEN a 68-team tournament field with matchups
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values for the East 1-seed
	values := GenerateTournamentValues(teams, matchups, 0, rules)
	valueByID := make(map[string]PredictedTeamValue)
	for _, v := range values {
		valueByID[v.TeamID] = v
	}

	var oneSeedID string
	for _, team := range teams {
		if team.Region == "East" && team.Seed == 1 {
			oneSeedID = team.ID
		}
	}

	// THEN the 1-seed has >50% probability of reaching the Sweet 16
	v := valueByID[oneSeedID]
	if v.PRound3 <= 0.50 {
		t.Errorf("1-seed PRound3 (reach S16) = %.4f, expected > 0.50", v.PRound3)
	}
}

func TestThatOneSeedChampionshipProbabilityExceedsFivePercent(t *testing.T) {
	// GIVEN a 68-team tournament field with matchups
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values for the East 1-seed
	values := GenerateTournamentValues(teams, matchups, 0, rules)
	valueByID := make(map[string]PredictedTeamValue)
	for _, v := range values {
		valueByID[v.TeamID] = v
	}

	var oneSeedID string
	for _, team := range teams {
		if team.Region == "East" && team.Seed == 1 {
			oneSeedID = team.ID
		}
	}

	// THEN the 1-seed has >5% championship probability
	v := valueByID[oneSeedID]
	if v.PRound7 <= 0.05 {
		t.Errorf("1-seed PRound7 (win championship) = %.4f, expected > 0.05", v.PRound7)
	}
}

func TestThatSixteenSeedReachesSweetSixteenLessThanFivePercent(t *testing.T) {
	// GIVEN a 68-team tournament field with matchups
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values for non-FF 16-seeds
	values := GenerateTournamentValues(teams, matchups, 0, rules)
	valueByID := make(map[string]PredictedTeamValue)
	for _, v := range values {
		valueByID[v.TeamID] = v
	}

	// Find a non-FF 16-seed (South region has exactly one 16-seed)
	var sixteenSeedID string
	for _, team := range teams {
		if team.Region == "South" && team.Seed == 16 {
			sixteenSeedID = team.ID
		}
	}

	// THEN the 16-seed has <5% probability of reaching the Sweet 16
	v := valueByID[sixteenSeedID]
	if v.PRound3 >= 0.05 {
		t.Errorf("16-seed PRound3 (reach S16) = %.4f, expected < 0.05", v.PRound3)
	}
}

func TestThatStrongerFirstFourTeamHasHigherFFProbability(t *testing.T) {
	// GIVEN a 68-team field where East has two 16-seeds: team-16 (KenPom -10.0) and team-65 (KenPom -12.0)
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values
	values := GenerateTournamentValues(teams, matchups, 0, rules)
	valueByID := make(map[string]PredictedTeamValue)
	for _, v := range values {
		valueByID[v.TeamID] = v
	}

	// THEN the stronger East 16-seed (team-16, KenPom -10.0) has higher PRound1 than the weaker one (team-65, KenPom -12.0)
	stronger := valueByID["team-16"]
	weaker := valueByID["team-65"]
	if stronger.PRound1 <= weaker.PRound1 {
		t.Errorf("stronger FF team PRound1 (%.4f) should be > weaker FF team PRound1 (%.4f)", stronger.PRound1, weaker.PRound1)
	}
}

func TestThatFirstFourPairProbabilitiesSumToOne(t *testing.T) {
	// GIVEN a 68-team field where East has two 16-seeds
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values
	values := GenerateTournamentValues(teams, matchups, 0, rules)
	valueByID := make(map[string]PredictedTeamValue)
	for _, v := range values {
		valueByID[v.TeamID] = v
	}

	// THEN the East 16-seed pair PRound1 values sum to 1.0
	pSum := valueByID["team-16"].PRound1 + valueByID["team-65"].PRound1
	if math.Abs(pSum-1.0) > 0.001 {
		t.Errorf("East 16-seed FF pair PRound1 sum = %.4f, expected 1.0", pSum)
	}
}

// --- Checkpoint tests ---

func TestThatCheckpointMatchupsHavePMatchupOneForKnownGames(t *testing.T) {
	// GIVEN survivors after R64 (throughRound=2): one team per bracket slot per region
	allTeams := generateTestTeams()
	// Simulate R64 results: only the higher seed survives each matchup
	// For each region, 1-seed beats 16, 8 beats 9, 5 beats 12, etc.
	winnerSeeds := []int{1, 8, 5, 4, 6, 3, 7, 2}
	var survivors []TeamInput
	for _, team := range allTeams {
		for _, ws := range winnerSeeds {
			if team.Seed == ws {
				survivors = append(survivors, TeamInput{
					ID:        team.ID,
					Seed:      team.Seed,
					Region:    team.Region,
					KenPomNet: team.KenPomNet,
					Wins:      1,
					Byes:      1,
				})
				break
			}
		}
	}
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating checkpoint matchups
	matchups, err := GenerateMatchups(survivors, 2, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the first remaining round (R32) has pMatchup = 1.0 for each game
	for _, m := range matchups {
		if m.RoundOrder == 2 && math.Abs(m.PMatchup-1.0) > 0.001 {
			t.Errorf("R32 game %s: pMatchup = %.4f, expected 1.0", m.GameID, m.PMatchup)
		}
	}
}

func TestThatCheckpointChampionshipProbsSumToOne(t *testing.T) {
	// GIVEN 4 Final Four survivors (throughRound=5)
	survivors := []TeamInput{
		{ID: "t-east", Seed: 1, Region: "East", KenPomNet: 25.0, Wins: 4, Byes: 1},
		{ID: "t-west", Seed: 2, Region: "West", KenPomNet: 20.0, Wins: 4, Byes: 1},
		{ID: "t-south", Seed: 3, Region: "South", KenPomNet: 18.0, Wins: 4, Byes: 1},
		{ID: "t-midwest", Seed: 1, Region: "Midwest", KenPomNet: 22.0, Wins: 4, Byes: 1},
	}
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating checkpoint matchups and values
	matchups, err := GenerateMatchups(survivors, 5, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()
	values := GenerateTournamentValues(survivors, matchups, 5, rules)

	// THEN championship probabilities (PRound7) sum to 1.0
	var pChampSum float64
	for _, v := range values {
		pChampSum += v.PRound7
	}
	if math.Abs(pChampSum-1.0) > 0.01 {
		t.Errorf("championship probabilities sum = %.4f, expected 1.0", pChampSum)
	}
}

func TestThatCheckpointValuesSetResolvedRoundsCorrectly(t *testing.T) {
	// GIVEN a mix of alive and eliminated teams at throughRound=3
	allTeams := []TeamInput{
		{ID: "alive-1", Seed: 1, Region: "East", KenPomNet: 25.0, Wins: 2, Byes: 1},     // progress=3, alive
		{ID: "alive-2", Seed: 2, Region: "West", KenPomNet: 20.0, Wins: 2, Byes: 1},     // progress=3, alive
		{ID: "alive-3", Seed: 1, Region: "South", KenPomNet: 22.0, Wins: 2, Byes: 1},    // progress=3, alive
		{ID: "alive-4", Seed: 1, Region: "Midwest", KenPomNet: 18.0, Wins: 2, Byes: 1},  // progress=3, alive
		{ID: "elim-1", Seed: 16, Region: "East", KenPomNet: -10.0, Wins: 0, Byes: 1},    // progress=1, eliminated
		{ID: "elim-2", Seed: 8, Region: "West", KenPomNet: 5.0, Wins: 1, Byes: 1},       // progress=2, eliminated
	}
	survivors := []TeamInput{allTeams[0], allTeams[1], allTeams[2], allTeams[3]}
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	matchups, err := GenerateMatchups(survivors, 3, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating checkpoint values for all teams
	values := GenerateTournamentValues(allTeams, matchups, 3, rules)
	valueByID := make(map[string]PredictedTeamValue)
	for _, v := range values {
		valueByID[v.TeamID] = v
	}

	// THEN eliminated team has 0.0 for rounds beyond its progress
	elim1 := valueByID["elim-1"]
	if elim1.PRound1 != 1.0 {
		t.Errorf("eliminated team PRound1 = %.4f, expected 1.0", elim1.PRound1)
	}
	if elim1.PRound2 != 0.0 {
		t.Errorf("eliminated team PRound2 = %.4f, expected 0.0", elim1.PRound2)
	}
}

func TestThatCheckpointExpectedPointsIncludesActualPlusRemaining(t *testing.T) {
	// GIVEN 4 F4 survivors at throughRound=5
	survivors := []TeamInput{
		{ID: "t-east", Seed: 1, Region: "East", KenPomNet: 25.0, Wins: 4, Byes: 1},
		{ID: "t-west", Seed: 2, Region: "West", KenPomNet: 20.0, Wins: 4, Byes: 1},
		{ID: "t-south", Seed: 3, Region: "South", KenPomNet: 18.0, Wins: 4, Byes: 1},
		{ID: "t-midwest", Seed: 1, Region: "Midwest", KenPomNet: 22.0, Wins: 4, Byes: 1},
	}
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(survivors, 5, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating checkpoint values
	values := GenerateTournamentValues(survivors, matchups, 5, rules)

	// THEN each survivor's expectedPoints > actualPoints (they have future value)
	actualPoints := float64(scoring.PointsForProgress(rules, 4, 1)) // 4 wins + 1 bye = progress 5 = 310
	for _, v := range values {
		if v.ExpectedPoints <= actualPoints {
			t.Errorf("team %s: expectedPoints (%.2f) should be > actualPoints (%.2f)", v.TeamID, v.ExpectedPoints, actualPoints)
		}
	}
}

func TestThatCompletedTournamentProducesValuesForAllTeams(t *testing.T) {
	// GIVEN a completed tournament (throughRound=7) with all teams having final results
	allTeams := generateTestTeams()
	// Set wins to simulate a completed tournament: champion has 6 wins, others fewer
	for i := range allTeams {
		allTeams[i].Byes = 1
		// Give the first team (East 1-seed) the championship
		if i == 0 {
			allTeams[i].Wins = 6
		} else if i < 4 {
			allTeams[i].Wins = 4 // F4 losers
		} else if i < 8 {
			allTeams[i].Wins = 3 // E8 losers
		} else {
			allTeams[i].Wins = 1 // Early losers
		}
	}
	rules := DefaultScoringRules()

	// WHEN generating values with throughRound=7 and no remaining matchups
	values := GenerateTournamentValues(allTeams, nil, 7, rules)

	// THEN all 68 teams have values
	if len(values) != 68 {
		t.Errorf("expected 68 team values, got %d", len(values))
	}
}

func TestThatThroughRoundZeroProducesSixRoundsOfMatchups(t *testing.T) {
	// GIVEN a 68-team tournament field
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating matchups at throughRound=0
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN rounds 1-6 are all present
	roundPresent := make(map[int]bool)
	for _, m := range matchups {
		roundPresent[m.RoundOrder] = true
	}
	for r := 1; r <= 6; r++ {
		if !roundPresent[r] {
			t.Errorf("expected round %d matchups to be present", r)
		}
	}
}

func TestThatThroughRoundZeroFirstFourPMatchupLessThanOne(t *testing.T) {
	// GIVEN a 68-team tournament field
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating matchups at throughRound=0
	matchups, err := GenerateMatchups(teams, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN R1 games involving First Four seeds have pMatchup < 1.0
	ffSeeds := make(map[string]bool)
	for _, region := range []string{"East", "West", "South", "Midwest"} {
		seedCounts := make(map[int]int)
		for _, t := range teams {
			if t.Region == region {
				seedCounts[t.Seed]++
			}
		}
		for _, t := range teams {
			if t.Region == region && seedCounts[t.Seed] > 1 {
				ffSeeds[t.ID] = true
			}
		}
	}

	foundFFMatchup := false
	for _, m := range matchups {
		if m.RoundOrder == 1 && (ffSeeds[m.Team1ID] || ffSeeds[m.Team2ID]) {
			foundFFMatchup = true
			if m.PMatchup >= 1.0 {
				t.Errorf("R1 game %s involving FF team: pMatchup = %.4f, expected < 1.0", m.GameID, m.PMatchup)
			}
		}
	}
	if !foundFFMatchup {
		t.Error("expected to find R1 matchups involving First Four teams")
	}
}

func TestThatChampionshipCheckpointGeneratesOneMatchup(t *testing.T) {
	// GIVEN 2 finalists from opposite brackets at throughRound=6
	survivors := []TeamInput{
		{ID: "t-east", Seed: 1, Region: "East", KenPomNet: 25.0, Wins: 5, Byes: 1},
		{ID: "t-south", Seed: 2, Region: "South", KenPomNet: 20.0, Wins: 5, Byes: 1},
	}
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating checkpoint matchups
	matchups, err := GenerateMatchups(survivors, 6, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN exactly 1 matchup at roundOrder=6
	var r6Count int
	for _, m := range matchups {
		if m.RoundOrder == 6 {
			r6Count++
		}
	}
	if r6Count != 1 {
		t.Errorf("expected 1 championship matchup, got %d", r6Count)
	}
}

func TestThatChampionshipCheckpointProbsSumToOne(t *testing.T) {
	// GIVEN 2 finalists from opposite brackets at throughRound=6
	survivors := []TeamInput{
		{ID: "t-east", Seed: 1, Region: "East", KenPomNet: 25.0, Wins: 5, Byes: 1},
		{ID: "t-south", Seed: 2, Region: "South", KenPomNet: 20.0, Wins: 5, Byes: 1},
	}
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating checkpoint matchups and values
	matchups, err := GenerateMatchups(survivors, 6, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()
	values := GenerateTournamentValues(survivors, matchups, 6, rules)

	// THEN PRound7 values sum to 1.0
	var pChampSum float64
	for _, v := range values {
		pChampSum += v.PRound7
	}
	if math.Abs(pChampSum-1.0) > 0.01 {
		t.Errorf("championship probabilities sum = %.4f, expected 1.0", pChampSum)
	}
}

func TestThatChampionshipCheckpointFinalistsHaveNonZeroPRound7(t *testing.T) {
	// GIVEN 2 finalists from opposite brackets at throughRound=6
	survivors := []TeamInput{
		{ID: "t-east", Seed: 1, Region: "East", KenPomNet: 25.0, Wins: 5, Byes: 1},
		{ID: "t-south", Seed: 2, Region: "South", KenPomNet: 20.0, Wins: 5, Byes: 1},
	}
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating checkpoint matchups and values
	matchups, err := GenerateMatchups(survivors, 6, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()
	values := GenerateTournamentValues(survivors, matchups, 6, rules)

	// THEN both finalists have non-zero championship probability
	for _, v := range values {
		if v.PRound7 <= 0.0 {
			t.Errorf("team %s: PRound7 = %.4f, expected > 0.0", v.TeamID, v.PRound7)
		}
	}
}

// --- computeRound unit tests ---

func TestThatComputeRoundProducesPMatchupFromPAdvance(t *testing.T) {
	// GIVEN two teams with different advance probabilities
	games := []gameSetup{{
		gameID: "test-1",
		side1:  []TeamInput{{ID: "a", KenPomNet: 20.0}},
		side2:  []TeamInput{{ID: "b", KenPomNet: 15.0}},
	}}
	pAdvance := map[string]float64{"a": 0.8, "b": 0.6}
	calcWinProb := func(id1, id2 string) float64 { return 0.7 }

	// WHEN computing the round
	matchups, newPAdv := computeRound(games, pAdvance, calcWinProb, 1)

	// THEN pMatchup = pAdvance[a] * pAdvance[b] = 0.48
	if len(matchups) != 1 {
		t.Fatalf("expected 1 matchup, got %d", len(matchups))
	}
	if math.Abs(matchups[0].PMatchup-0.48) > 0.001 {
		t.Errorf("pMatchup = %.4f, expected 0.48", matchups[0].PMatchup)
	}

	// AND new advance probs are computed correctly
	expectedA := 0.48 * 0.7  // 0.336
	expectedB := 0.48 * 0.3  // 0.144
	if math.Abs(newPAdv["a"]-expectedA) > 0.001 {
		t.Errorf("new pAdvance[a] = %.4f, expected %.4f", newPAdv["a"], expectedA)
	}
	if math.Abs(newPAdv["b"]-expectedB) > 0.001 {
		t.Errorf("new pAdvance[b] = %.4f, expected %.4f", newPAdv["b"], expectedB)
	}
}

func TestThatComputeRoundSetsRoundOrder(t *testing.T) {
	// GIVEN a simple game setup
	games := []gameSetup{{
		gameID: "g1",
		side1:  []TeamInput{{ID: "a"}},
		side2:  []TeamInput{{ID: "b"}},
	}}
	pAdvance := map[string]float64{"a": 1.0, "b": 1.0}
	calcWinProb := func(_, _ string) float64 { return 0.5 }

	// WHEN computing with roundOrder=4
	matchups, _ := computeRound(games, pAdvance, calcWinProb, 4)

	// THEN matchup has correct round order
	if matchups[0].RoundOrder != 4 {
		t.Errorf("roundOrder = %d, expected 4", matchups[0].RoundOrder)
	}
}

// --- TournamentState tests ---

func TestThatNewTournamentStatePartitionsSurvivors(t *testing.T) {
	// GIVEN tournament data with some eliminated teams (progress < throughRound)
	data := &TournamentData{
		Teams: []TeamInput{
			{ID: "alive-1", Wins: 2, Byes: 1}, // progress=3, survives
			{ID: "alive-2", Wins: 3, Byes: 0}, // progress=3, survives
			{ID: "elim-1", Wins: 1, Byes: 1},  // progress=2, eliminated
			{ID: "elim-2", Wins: 0, Byes: 1},  // progress=1, eliminated
		},
		Rules:    DefaultScoringRules(),
		FFConfig: nil,
	}

	// WHEN creating state at throughRound=3
	state := NewTournamentState(data, 3)

	// THEN survivors excludes eliminated teams
	if len(state.Survivors) != 2 {
		t.Errorf("expected 2 survivors, got %d", len(state.Survivors))
	}
}

func TestThatNewTournamentStateRetainsAllTeams(t *testing.T) {
	// GIVEN tournament data with 4 teams, some eliminated
	data := &TournamentData{
		Teams: []TeamInput{
			{ID: "alive-1", Wins: 2, Byes: 1},
			{ID: "alive-2", Wins: 3, Byes: 0},
			{ID: "elim-1", Wins: 1, Byes: 1},
			{ID: "elim-2", Wins: 0, Byes: 1},
		},
		Rules:    DefaultScoringRules(),
		FFConfig: nil,
	}

	// WHEN creating state at throughRound=3
	state := NewTournamentState(data, 3)

	// THEN AllTeams includes all 4 teams
	if len(state.AllTeams) != 4 {
		t.Errorf("expected 4 AllTeams, got %d", len(state.AllTeams))
	}
}

func TestThatNewTournamentStateAtZeroMakesAllTeamsSurvivors(t *testing.T) {
	// GIVEN a full 68-team tournament field at pre-tournament
	teams := generateTestTeams()
	data := &TournamentData{
		Teams:    teams,
		Rules:    DefaultScoringRules(),
		FFConfig: nil,
	}

	// WHEN creating state at throughRound=0
	state := NewTournamentState(data, 0)

	// THEN all 68 teams are survivors
	if len(state.Survivors) != 68 {
		t.Errorf("expected 68 survivors, got %d", len(state.Survivors))
	}
}

func TestThatGeneratePredictionsProducesValuesForAllTeams(t *testing.T) {
	// GIVEN a full 68-team tournament state at throughRound=0
	teams := generateTestTeams()
	data := &TournamentData{
		Teams:    teams,
		Rules:    DefaultScoringRules(),
		FFConfig: nil,
	}
	state := NewTournamentState(data, 0)
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating predictions
	values, err := generatePredictions(state, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN values are produced for all 68 teams
	if len(values) != 68 {
		t.Errorf("expected 68 team values, got %d", len(values))
	}
}

// --- Double-counting and Favorites regression tests ---

func TestThatCheckpointZeroEVDoesNotDoubleCountByePoints(t *testing.T) {
	// GIVEN two identical 68-team fields — one with Byes=0, one with Byes=1
	teamsNoByes := generateTestTeams()
	teamsWithByes := generateTestTeams()
	for i := range teamsWithByes {
		teamsWithByes[i].Byes = 1
	}
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	rules := DefaultScoringRules()

	// Matchups are identical since they depend on seeds/KenPom, not Byes
	matchups, err := GenerateMatchups(teamsNoByes, 0, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valuesNoByes := GenerateTournamentValues(teamsNoByes, matchups, 0, rules)
	valuesWithByes := GenerateTournamentValues(teamsWithByes, matchups, 0, rules)

	evByTeam := func(values []PredictedTeamValue) map[string]float64 {
		m := make(map[string]float64)
		for _, v := range values {
			m[v.TeamID] = v.ExpectedPoints
		}
		return m
	}
	noByes := evByTeam(valuesNoByes)
	withByes := evByTeam(valuesWithByes)

	// THEN a non-FF 1-seed's EV should be the same regardless of Byes.
	// The bye gives already-earned progress that shifts points from "future" to "actual"
	// but shouldn't change the total EV.
	// With the double-counting bug, withByes would be ~10 points higher.
	teamID := "team-1" // East 1-seed (non-FF)
	diff := withByes[teamID] - noByes[teamID]
	if math.Abs(diff) > 1.0 {
		t.Errorf("East 1-seed EV: noByes=%.2f, withByes=%.2f, diff=%.2f (expected ~0)",
			noByes[teamID], withByes[teamID], diff)
	}
}

func TestThatEVIsConsistentAcrossCheckpoints(t *testing.T) {
	// GIVEN predictions at checkpoint 0 (Byes=1) and checkpoint 1 (after FF resolved)
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	rules := DefaultScoringRules()

	// Checkpoint 0: non-FF teams have Byes=1
	teams0 := generateTestTeamsWithByes()
	matchups0, err := GenerateMatchups(teams0, 0, spec, nil)
	if err != nil {
		t.Fatalf("checkpoint 0 matchups: %v", err)
	}
	values0 := GenerateTournamentValues(teams0, matchups0, 0, rules)

	// Checkpoint 1: FF resolved, survivors have progress >= 1
	teams1 := generateCheckpoint1Teams()
	survivors1 := filterSurvivors(teams1, 1)
	matchups1, err := GenerateMatchups(survivors1, 1, spec, nil)
	if err != nil {
		t.Fatalf("checkpoint 1 matchups: %v", err)
	}
	values1 := GenerateTournamentValues(teams1, matchups1, 1, rules)

	findEV := func(values []PredictedTeamValue, teamID string) float64 {
		for _, v := range values {
			if v.TeamID == teamID {
				return v.ExpectedPoints
			}
		}
		return 0
	}

	// THEN the East 1-seed's EV should not change by more than 10% between checkpoints
	teamID := "team-1" // East 1-seed
	ev0 := findEV(values0, teamID)
	ev1 := findEV(values1, teamID)

	if ev1 < ev0*0.90 || ev1 > ev0*1.10 {
		t.Errorf("East 1-seed EV: checkpoint0=%.2f, checkpoint1=%.2f (>10%% change)", ev0, ev1)
	}
}

func TestThatFavoritesTotalPointsAreSimilarAcrossCheckpoints(t *testing.T) {
	// GIVEN Favorites computed at checkpoint 0 (Byes=1) and checkpoint 1 (FF resolved)
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	rules := DefaultScoringRules()

	// Checkpoint 0
	teams0 := generateTestTeamsWithByes()
	matchups0, err := GenerateMatchups(teams0, 0, spec, nil)
	if err != nil {
		t.Fatalf("checkpoint 0 matchups: %v", err)
	}
	fav0 := ComputeFavoritesBracket(teams0, matchups0, 0, rules)

	// Checkpoint 1
	teams1 := generateCheckpoint1Teams()
	survivors1 := filterSurvivors(teams1, 1)
	matchups1, err := GenerateMatchups(survivors1, 1, spec, nil)
	if err != nil {
		t.Fatalf("checkpoint 1 matchups: %v", err)
	}
	fav1 := ComputeFavoritesBracket(teams1, matchups1, 1, rules)

	// THEN the East 8-seed's Favorites total should be similar across checkpoints.
	// Without the R64 fix, checkpoint 1 would miss R64 and produce much lower values.
	teamID := "team-8" // East 8-seed
	diff := math.Abs(fav0[teamID] - fav1[teamID])
	if diff > 5.0 {
		t.Errorf("East 8-seed Favorites: checkpoint0=%.2f, checkpoint1=%.2f, diff=%.2f (expected similar)",
			fav0[teamID], fav1[teamID], diff)
	}
}

func TestThatFavoritesBracketProcessesR64AtCheckpointOne(t *testing.T) {
	// GIVEN survivors at checkpoint 1 with R64 matchups
	teams := generateCheckpoint1Teams()
	survivors := filterSurvivors(teams, 1)
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateMatchups(survivors, 1, spec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN computing the Favorites bracket
	favMap := ComputeFavoritesBracket(teams, matchups, 1, rules)

	// THEN the East 8-seed (KenPom 10.0) should beat the 9-seed (KenPom 7.5) in R64,
	// giving it a Favorites total reflecting the R64 win.
	// With R64 win: totalWins=0+1=1, progress=1+1=2, points>=30
	// Without R64 win (bug): totalWins=0, progress=0+1=1, points=10
	team8Points := favMap["team-8"]
	if team8Points < 30.0 {
		t.Errorf("East 8-seed Favorites total = %.2f, expected >= 30 (R64 win should be included)", team8Points)
	}
}

// generateTestTeamsWithByes creates a 68-team field where non-FF teams have Byes=1.
func generateTestTeamsWithByes() []TeamInput {
	teams := generateTestTeams()
	ffTeamIDs := identifyFFTeamIDs(teams)
	for i := range teams {
		if !ffTeamIDs[teams[i].ID] {
			teams[i].Byes = 1
		}
	}
	return teams
}

// generateCheckpoint1Teams creates a 68-team field reflecting First Four results.
// The stronger KenPom team wins each FF matchup.
func generateCheckpoint1Teams() []TeamInput {
	teams := generateTestTeams()
	ffTeamIDs := identifyFFTeamIDs(teams)

	// Determine FF winners by KenPom within each FF pair.
	ffPairs := identifyFFPairs(teams)
	ffWinners := make(map[string]bool)
	for _, pair := range ffPairs {
		if pair[0].KenPomNet >= pair[1].KenPomNet {
			ffWinners[pair[0].ID] = true
		} else {
			ffWinners[pair[1].ID] = true
		}
	}

	for i := range teams {
		if ffTeamIDs[teams[i].ID] {
			if ffWinners[teams[i].ID] {
				teams[i].Wins = 1
			}
			// FF losers: Wins=0, Byes=0 → progress=0 (eliminated at throughRound=1)
		} else {
			teams[i].Byes = 1
		}
	}
	return teams
}

// identifyFFTeamIDs returns IDs of First Four teams (seeds with >1 team in a region).
func identifyFFTeamIDs(teams []TeamInput) map[string]bool {
	type regionSeed struct {
		region string
		seed   int
	}
	counts := make(map[regionSeed]int)
	for _, t := range teams {
		counts[regionSeed{t.Region, t.Seed}]++
	}
	result := make(map[string]bool)
	for _, t := range teams {
		if counts[regionSeed{t.Region, t.Seed}] > 1 {
			result[t.ID] = true
		}
	}
	return result
}

// identifyFFPairs returns pairs of First Four teams grouped by region+seed.
func identifyFFPairs(teams []TeamInput) [][2]TeamInput {
	type regionSeed struct {
		region string
		seed   int
	}
	groups := make(map[regionSeed][]TeamInput)
	for _, t := range teams {
		key := regionSeed{t.Region, t.Seed}
		groups[key] = append(groups[key], t)
	}
	var pairs [][2]TeamInput
	for _, group := range groups {
		if len(group) == 2 {
			pairs = append(pairs, [2]TeamInput{group[0], group[1]})
		}
	}
	return pairs
}

// filterSurvivors returns teams with progress >= throughRound.
func filterSurvivors(teams []TeamInput, throughRound int) []TeamInput {
	var result []TeamInput
	for _, t := range teams {
		if t.Wins+t.Byes >= throughRound {
			result = append(result, t)
		}
	}
	return result
}

// generateTestTeams creates a realistic 68-team tournament field for testing.
func generateTestTeams() []TeamInput {
	regions := []string{"East", "West", "South", "Midwest"}
	var teams []TeamInput

	teamID := 1
	for _, region := range regions {
		// Standard 16 seeds per region
		for seed := 1; seed <= 16; seed++ {
			// KenPom net rating: higher seeds have higher ratings
			// 1-seeds around +25, 16-seeds around -10
			kenpomNet := 30.0 - float64(seed)*2.5

			teams = append(teams, TeamInput{
				ID:        fmt.Sprintf("team-%d", teamID),
				Seed:      seed,
				Region:    region,
				KenPomNet: kenpomNet,
			})
			teamID++
		}
	}

	// Add 4 First Four teams (extra 11 and 16 seeds)
	// Two extra 16-seeds
	teams = append(teams, TeamInput{
		ID:        fmt.Sprintf("team-%d", teamID),
		Seed:      16,
		Region:    "East",
		KenPomNet: -12.0,
	})
	teamID++
	teams = append(teams, TeamInput{
		ID:        fmt.Sprintf("team-%d", teamID),
		Seed:      16,
		Region:    "West",
		KenPomNet: -11.0,
	})
	teamID++

	// Two extra 11-seeds
	teams = append(teams, TeamInput{
		ID:        fmt.Sprintf("team-%d", teamID),
		Seed:      11,
		Region:    "South",
		KenPomNet: 3.0,
	})
	teamID++
	teams = append(teams, TeamInput{
		ID:        fmt.Sprintf("team-%d", teamID),
		Seed:      11,
		Region:    "Midwest",
		KenPomNet: 2.5,
	})

	return teams
}
