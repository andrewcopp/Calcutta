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

func TestThatSimulationsAreDeterministic(t *testing.T) {
	b := toyBracket()
	probs := map[MatchupKey]float64{
		{GameID: "g1", Team1ID: "t1", Team2ID: "t3"}: 0.6,
		{GameID: "g2", Team1ID: "t2", Team2ID: "t4"}: 0.7,

		// Championship probabilities depend on who advances.
		{GameID: "g3", Team1ID: "t1", Team2ID: "t2"}: 0.5,
		{GameID: "g3", Team1ID: "t1", Team2ID: "t4"}: 0.5,
		{GameID: "g3", Team1ID: "t3", Team2ID: "t2"}: 0.5,
		{GameID: "g3", Team1ID: "t3", Team2ID: "t4"}: 0.5,
	}

	res1, err := Simulate(b, probs, 100, 42, Options{Workers: 1})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	res2, err := Simulate(b, probs, 100, 42, Options{Workers: 1})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(res1) != len(res2) {
		t.Fatalf("expected equal lengths, got %d and %d", len(res1), len(res2))
	}

	for i := range res1 {
		if res1[i] != res2[i] {
			t.Fatalf("expected deterministic results at index %d: %+v vs %+v", i, res1[i], res2[i])
		}
	}
}

func TestThatAllTeamsIncluded(t *testing.T) {
	b := toyBracket()
	res, err := Simulate(b, nil, 10, 1, Options{Workers: 1})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	teams := map[string]bool{"t1": true, "t2": true, "t3": true, "t4": true}
	for simID := 0; simID < 10; simID++ {
		seen := make(map[string]bool)
		for _, r := range res {
			if r.SimID != simID {
				continue
			}
			seen[r.TeamID] = true
		}
		for tid := range teams {
			if !seen[tid] {
				t.Fatalf("simulation %d missing team %s", simID, tid)
			}
		}
	}
}

func TestThatWinsAreBounded(t *testing.T) {
	b := toyBracket()
	res, err := Simulate(b, nil, 50, 99, Options{Workers: 1})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Total games played in a completed sim is 3.
	maxPossibleWins := 3
	for _, r := range res {
		if r.Wins < 0 {
			t.Fatalf("expected non-negative wins")
		}
		if r.Wins > maxPossibleWins {
			t.Fatalf("expected wins bounded by %d, got %d", maxPossibleWins, r.Wins)
		}
	}
}
