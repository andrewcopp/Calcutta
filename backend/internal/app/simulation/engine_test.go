package simulation

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func toyBracket() *models.BracketStructure {
	b := &models.BracketStructure{
		TournamentID: "t1",
		Regions:      []string{"X"},
		Games:        make(map[string]*models.BracketGame),
		FinalFour:    &models.FinalFourConfig{},
	}

	b.Games["g1"] = &models.BracketGame{
		GameID:       "g1",
		Round:        models.RoundOf64,
		Region:       "X",
		Team1:        &models.BracketTeam{TeamID: "t1"},
		Team2:        &models.BracketTeam{TeamID: "t3"},
		NextGameID:   "g3",
		NextGameSlot: 1,
		SortOrder:    1,
	}
	b.Games["g2"] = &models.BracketGame{
		GameID:       "g2",
		Round:        models.RoundOf64,
		Region:       "X",
		Team1:        &models.BracketTeam{TeamID: "t2"},
		Team2:        &models.BracketTeam{TeamID: "t4"},
		NextGameID:   "g3",
		NextGameSlot: 2,
		SortOrder:    2,
	}
	b.Games["g3"] = &models.BracketGame{
		GameID:    "g3",
		Round:     models.RoundChampionship,
		Region:    "X",
		SortOrder: 1,
	}

	return b
}

func toyProbs() map[MatchupKey]float64 {
	return map[MatchupKey]float64{
		{GameID: "g1", Team1ID: "t1", Team2ID: "t3"}: 0.6,
		{GameID: "g2", Team1ID: "t2", Team2ID: "t4"}: 0.7,
		{GameID: "g3", Team1ID: "t1", Team2ID: "t2"}: 0.5,
		{GameID: "g3", Team1ID: "t1", Team2ID: "t4"}: 0.5,
		{GameID: "g3", Team1ID: "t3", Team2ID: "t2"}: 0.5,
		{GameID: "g3", Team1ID: "t3", Team2ID: "t4"}: 0.5,
	}
}

func TestThatDeterministicSimulationProducesIdenticalResults(t *testing.T) {
	// GIVEN a bracket with explicit probabilities
	b := toyBracket()
	probs := toyProbs()

	// WHEN simulating twice with the same seed
	res1, err1 := Simulate(b, probs, 100, 42, Options{Workers: 1})
	if err1 != nil {
		t.Fatalf("first simulation failed: %v", err1)
	}
	res2, err2 := Simulate(b, probs, 100, 42, Options{Workers: 1})
	if err2 != nil {
		t.Fatalf("second simulation failed: %v", err2)
	}

	// THEN both simulations produce identical team results
	if len(res1) != len(res2) {
		t.Fatalf("expected equal lengths, got %d and %d", len(res1), len(res2))
	}
	for i := range res1 {
		if res1[i] != res2[i] {
			t.Errorf("mismatch at index %d: %+v vs %+v", i, res1[i], res2[i])
		}
	}
}

func TestThatEveryTeamAppearsInSimulationResults(t *testing.T) {
	// GIVEN a bracket with 4 teams
	b := toyBracket()

	// WHEN running a simulation
	res, err := Simulate(b, nil, 1, 1, Options{Workers: 1})
	if err != nil {
		t.Fatalf("simulation failed: %v", err)
	}

	// THEN all 4 teams appear in the results
	seen := make(map[string]bool)
	for _, r := range res {
		seen[r.TeamID] = true
	}
	for _, tid := range []string{"t1", "t2", "t3", "t4"} {
		if !seen[tid] {
			t.Errorf("team %s missing from simulation results", tid)
		}
	}
}

func TestThatWinsAreNonNegative(t *testing.T) {
	// GIVEN a bracket simulated 50 times
	b := toyBracket()

	// WHEN running the simulation
	res, err := Simulate(b, nil, 50, 99, Options{Workers: 1})
	if err != nil {
		t.Fatalf("simulation failed: %v", err)
	}

	// THEN no team result has negative wins
	for _, r := range res {
		if r.Wins < 0 {
			t.Errorf("team %s in sim %d has negative wins: %d", r.TeamID, r.SimID, r.Wins)
		}
	}
}

func TestThatWinsAreBoundedByTotalGames(t *testing.T) {
	// GIVEN a bracket with 3 total games (max possible wins = 3)
	b := toyBracket()

	// WHEN running the simulation
	res, err := Simulate(b, nil, 50, 99, Options{Workers: 1})
	if err != nil {
		t.Fatalf("simulation failed: %v", err)
	}

	// THEN no team wins more than the total number of games
	maxPossibleWins := 3
	for _, r := range res {
		if r.Wins > maxPossibleWins {
			t.Errorf("team %s in sim %d has wins=%d exceeding max=%d", r.TeamID, r.SimID, r.Wins, maxPossibleWins)
		}
	}
}
