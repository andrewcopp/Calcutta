package models

import (
	"testing"
)

func TestGetMatchupForSeeds(t *testing.T) {
	tests := []struct {
		seed     int
		expected int
	}{
		{1, 16},
		{2, 15},
		{3, 14},
		{4, 13},
		{5, 12},
		{6, 11},
		{7, 10},
		{8, 9},
		{9, 8},
		{16, 1},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.seed)), func(t *testing.T) {
			result := GetMatchupForSeeds(tt.seed)
			if result != tt.expected {
				t.Errorf("GetMatchupForSeeds(%d) = %d; want %d", tt.seed, result, tt.expected)
			}
		})
	}
}

func TestCalculateWinsAndByes(t *testing.T) {
	bracket := &BracketStructure{
		TournamentID: "test-tournament",
		Games:        make(map[string]*BracketGame),
	}

	team1 := &BracketTeam{
		TeamID: "team1",
		Seed:   1,
		Region: "East",
	}

	team16 := &BracketTeam{
		TeamID: "team16",
		Seed:   16,
		Region: "East",
	}

	game1 := &BracketGame{
		GameID:       "game1",
		Round:        RoundOf64,
		Region:       "East",
		Team1:        team1,
		Team2:        team16,
		NextGameID:   "game2",
		NextGameSlot: 1,
	}

	game2 := &BracketGame{
		GameID: "game2",
		Round:  RoundOf32,
		Region: "East",
	}

	bracket.Games["game1"] = game1
	bracket.Games["game2"] = game2

	t.Run("Team with bye, no games played", func(t *testing.T) {
		wins, byes, eliminated := CalculateWinsAndByes("team1", bracket)
		if wins != 0 {
			t.Errorf("Expected 0 wins, got %d", wins)
		}
		if byes != 1 {
			t.Errorf("Expected 1 bye, got %d", byes)
		}
		if eliminated {
			t.Errorf("Expected team to not be eliminated")
		}
	})

	t.Run("Team wins first game", func(t *testing.T) {
		game1.Winner = team1
		game2.Team1 = team1

		wins, byes, eliminated := CalculateWinsAndByes("team1", bracket)
		if wins != 1 {
			t.Errorf("Expected 1 win, got %d", wins)
		}
		if byes != 1 {
			t.Errorf("Expected 1 bye, got %d", byes)
		}
		if eliminated {
			t.Errorf("Expected team to not be eliminated")
		}
	})

	t.Run("Team loses first game", func(t *testing.T) {
		game1.Winner = team16

		wins, _, eliminated := CalculateWinsAndByes("team1", bracket)
		if wins != 0 {
			t.Errorf("Expected 0 wins, got %d", wins)
		}
		if eliminated != true {
			t.Errorf("Expected team to be eliminated")
		}
	})
}

func TestBracketValidator_ValidateWinnerSelection(t *testing.T) {
	validator := NewBracketValidator()

	team1 := &BracketTeam{TeamID: "team1"}
	team2 := &BracketTeam{TeamID: "team2"}

	t.Run("Valid winner selection", func(t *testing.T) {
		game := &BracketGame{
			Team1: team1,
			Team2: team2,
		}

		err := validator.ValidateWinnerSelection(game, "team1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Invalid winner - not a participant", func(t *testing.T) {
		game := &BracketGame{
			Team1: team1,
			Team2: team2,
		}

		err := validator.ValidateWinnerSelection(game, "team3")
		if err == nil {
			t.Error("Expected error for non-participant winner")
		}
	})

	t.Run("Invalid - missing teams", func(t *testing.T) {
		game := &BracketGame{
			Team1: team1,
		}

		err := validator.ValidateWinnerSelection(game, "team1")
		if err == nil {
			t.Error("Expected error when team2 is missing")
		}
	})

	t.Run("Invalid - nil game", func(t *testing.T) {
		err := validator.ValidateWinnerSelection(nil, "team1")
		if err == nil {
			t.Error("Expected error for nil game")
		}
	})
}

func TestBracketValidator_ValidateBracketProgression(t *testing.T) {
	validator := NewBracketValidator()

	team1 := &BracketTeam{TeamID: "team1"}
	team2 := &BracketTeam{TeamID: "team2"}

	t.Run("Valid bracket progression", func(t *testing.T) {
		bracket := &BracketStructure{
			Games: map[string]*BracketGame{
				"game1": {
					GameID:       "game1",
					Team1:        team1,
					Team2:        team2,
					Winner:       team1,
					NextGameID:   "game2",
					NextGameSlot: 1,
				},
				"game2": {
					GameID: "game2",
					Team1:  team1,
				},
			},
		}

		err := validator.ValidateBracketProgression(bracket)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Invalid - winner not progressed", func(t *testing.T) {
		bracket := &BracketStructure{
			Games: map[string]*BracketGame{
				"game1": {
					GameID:       "game1",
					Team1:        team1,
					Team2:        team2,
					Winner:       team1,
					NextGameID:   "game2",
					NextGameSlot: 1,
				},
				"game2": {
					GameID: "game2",
					Team1:  team2,
				},
			},
		}

		err := validator.ValidateBracketProgression(bracket)
		if err == nil {
			t.Error("Expected error when winner not properly progressed")
		}
	})

	t.Run("Invalid - next game not found", func(t *testing.T) {
		bracket := &BracketStructure{
			Games: map[string]*BracketGame{
				"game1": {
					GameID:       "game1",
					Team1:        team1,
					Team2:        team2,
					Winner:       team1,
					NextGameID:   "game2",
					NextGameSlot: 1,
				},
			},
		}

		err := validator.ValidateBracketProgression(bracket)
		if err == nil {
			t.Error("Expected error when next game not found")
		}
	})
}
