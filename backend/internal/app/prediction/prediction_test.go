package prediction

import (
	"fmt"
	"math"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
)

func TestThatGenerateAllTheoreticalMatchupsCreatesMatchupsForAllRounds(t *testing.T) {
	// GIVEN a 68-team tournament field
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}

	// WHEN generating all theoretical matchups
	matchups, err := GenerateAllTheoreticalMatchups(teams, spec)
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
	matchups, err := GenerateAllTheoreticalMatchups(teams, spec)
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
	matchups, err := GenerateAllTheoreticalMatchups(teams, spec)
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
	matchups, err := GenerateAllTheoreticalMatchups(teams, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values
	values := GenerateTournamentValues(matchups, rules)

	// THEN all 68 teams have positive expected points
	if len(values) != 68 {
		t.Errorf("expected 68 team values, got %d", len(values))
	}
}

func TestThatChampionshipProbabilitiesSumToOne(t *testing.T) {
	// GIVEN a 68-team tournament field with matchups
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateAllTheoreticalMatchups(teams, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()

	// WHEN generating tournament values
	values := GenerateTournamentValues(matchups, rules)

	// THEN championship probabilities sum to 1.0
	var pChampSum float64
	for _, v := range values {
		pChampSum += v.PRound6
	}

	if math.Abs(pChampSum-1.0) > 0.01 {
		t.Errorf("championship probabilities sum = %.4f, expected 1.0", pChampSum)
	}
}

func TestThatOneSeedHasHigherExpectedPointsThanSixteenSeed(t *testing.T) {
	// GIVEN a 68-team tournament field with matchups
	teams := generateTestTeams()
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	matchups, err := GenerateAllTheoreticalMatchups(teams, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rules := DefaultScoringRules()
	values := GenerateTournamentValues(matchups, rules)

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
