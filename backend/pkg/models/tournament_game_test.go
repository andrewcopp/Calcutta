package models

import (
	"testing"
	"time"
)

func TestTournamentGame_GetGameStatus(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		game     TournamentGame
		expected string
	}{
		{
			name: "Future game",
			game: TournamentGame{
				TipoffTime: now.Add(24 * time.Hour),
				IsFinal:    false,
			},
			expected: "future",
		},
		{
			name: "Completed game",
			game: TournamentGame{
				TipoffTime: now.Add(-24 * time.Hour),
				IsFinal:    true,
			},
			expected: "completed",
		},
		{
			name: "In progress game",
			game: TournamentGame{
				TipoffTime: now.Add(-1 * time.Hour),
				IsFinal:    false,
			},
			expected: "in_progress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.game.GetGameStatus()
			if result != tt.expected {
				t.Errorf("GetGameStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTournamentGame_GetWinnerID(t *testing.T) {
	team1ID := "team1"
	team2ID := "team2"

	tests := []struct {
		name     string
		game     TournamentGame
		expected *string
	}{
		{
			name: "Team1 wins",
			game: TournamentGame{
				Team1ID:    &team1ID,
				Team2ID:    &team2ID,
				Team1Score: 80,
				Team2Score: 70,
				IsFinal:    true,
			},
			expected: &team1ID,
		},
		{
			name: "Team2 wins",
			game: TournamentGame{
				Team1ID:    &team1ID,
				Team2ID:    &team2ID,
				Team1Score: 70,
				Team2Score: 80,
				IsFinal:    true,
			},
			expected: &team2ID,
		},
		{
			name: "Game not final",
			game: TournamentGame{
				Team1ID:    &team1ID,
				Team2ID:    &team2ID,
				Team1Score: 80,
				Team2Score: 70,
				IsFinal:    false,
			},
			expected: nil,
		},
		{
			name: "Tie game",
			game: TournamentGame{
				Team1ID:    &team1ID,
				Team2ID:    &team2ID,
				Team1Score: 80,
				Team2Score: 80,
				IsFinal:    true,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.game.GetWinnerID()
			if (result == nil && tt.expected != nil) || (result != nil && tt.expected == nil) {
				t.Errorf("GetWinnerID() = %v, want %v", result, tt.expected)
			} else if result != nil && tt.expected != nil && *result != *tt.expected {
				t.Errorf("GetWinnerID() = %v, want %v", *result, *tt.expected)
			}
		})
	}
}

func TestTournamentGame_IsChampionshipGame(t *testing.T) {
	nextGameID := "next-game"
	nextGameSlot := 1

	tests := []struct {
		name     string
		game     TournamentGame
		expected bool
	}{
		{
			name: "Championship game (no next game)",
			game: TournamentGame{
				NextGameID:   nil,
				NextGameSlot: nil,
			},
			expected: true,
		},
		{
			name: "Regular game (has next game)",
			game: TournamentGame{
				NextGameID:   &nextGameID,
				NextGameSlot: &nextGameSlot,
			},
			expected: false,
		},
		{
			name: "Edge case: has next game ID but no slot",
			game: TournamentGame{
				NextGameID:   &nextGameID,
				NextGameSlot: nil,
			},
			expected: false,
		},
		{
			name: "Edge case: has next game slot but no ID",
			game: TournamentGame{
				NextGameID:   nil,
				NextGameSlot: &nextGameSlot,
			},
			expected: true, // We consider it a championship game if there's no next game ID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.game.IsChampionshipGame()
			if result != tt.expected {
				t.Errorf("IsChampionshipGame() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTournamentGame_GetRound(t *testing.T) {
	// Create a map of games for testing
	gamesMap := map[string]*TournamentGame{
		"championship": {
			ID:           "championship",
			NextGameID:   nil,
			NextGameSlot: nil,
			SortOrder:    67,
		},
		"final-four-1": {
			ID:           "final-four-1",
			NextGameID:   stringPtr("championship"),
			NextGameSlot: intPtr(1),
			SortOrder:    65,
		},
		"final-four-2": {
			ID:           "final-four-2",
			NextGameID:   stringPtr("championship"),
			NextGameSlot: intPtr(2),
			SortOrder:    66,
		},
		"elite-eight-1": {
			ID:           "elite-eight-1",
			NextGameID:   stringPtr("final-four-1"),
			NextGameSlot: intPtr(1),
			SortOrder:    61,
		},
		"elite-eight-2": {
			ID:           "elite-eight-2",
			NextGameID:   stringPtr("final-four-1"),
			NextGameSlot: intPtr(2),
			SortOrder:    62,
		},
		"elite-eight-3": {
			ID:           "elite-eight-3",
			NextGameID:   stringPtr("final-four-2"),
			NextGameSlot: intPtr(1),
			SortOrder:    63,
		},
		"elite-eight-4": {
			ID:           "elite-eight-4",
			NextGameID:   stringPtr("final-four-2"),
			NextGameSlot: intPtr(2),
			SortOrder:    64,
		},
		"sweet-sixteen-1": {
			ID:           "sweet-sixteen-1",
			NextGameID:   stringPtr("elite-eight-1"),
			NextGameSlot: intPtr(1),
			SortOrder:    53,
		},
		"round-of-32-1": {
			ID:           "round-of-32-1",
			NextGameID:   stringPtr("sweet-sixteen-1"),
			NextGameSlot: intPtr(1),
			SortOrder:    37,
		},
		"round-of-64-1": {
			ID:           "round-of-64-1",
			NextGameID:   stringPtr("round-of-32-1"),
			NextGameSlot: intPtr(1),
			SortOrder:    5,
		},
		"first-four-1": {
			ID:           "first-four-1",
			NextGameID:   stringPtr("round-of-64-1"),
			NextGameSlot: intPtr(1),
			SortOrder:    1,
		},
	}

	tests := []struct {
		name        string
		gameID      string
		totalRounds int
		expected    int
	}{
		{
			name:        "Championship game",
			gameID:      "championship",
			totalRounds: 7,
			expected:    7,
		},
		{
			name:        "Final Four game",
			gameID:      "final-four-1",
			totalRounds: 7,
			expected:    6,
		},
		{
			name:        "Elite Eight game",
			gameID:      "elite-eight-1",
			totalRounds: 7,
			expected:    5,
		},
		{
			name:        "Sweet Sixteen game",
			gameID:      "sweet-sixteen-1",
			totalRounds: 7,
			expected:    4,
		},
		{
			name:        "Round of 32 game",
			gameID:      "round-of-32-1",
			totalRounds: 7,
			expected:    3,
		},
		{
			name:        "Round of 64 game",
			gameID:      "round-of-64-1",
			totalRounds: 7,
			expected:    2,
		},
		{
			name:        "First Four game",
			gameID:      "first-four-1",
			totalRounds: 7,
			expected:    1,
		},
		{
			name:        "Different tournament structure",
			gameID:      "elite-eight-1",
			totalRounds: 4,
			expected:    2, // In a 4-round tournament, Elite Eight would be round 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := gamesMap[tt.gameID]
			result := game.GetRound(gamesMap, tt.totalRounds)
			if result != tt.expected {
				t.Errorf("GetRound() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func TestGetPotentialOpponents(t *testing.T) {
	// Create a 4-game tournament (2 rounds)
	gamesMap := make(map[string]*TournamentGame)

	// Round 1 games
	game1 := &TournamentGame{
		ID:           "game1",
		Team1ID:      stringPtr("team1"),
		Team2ID:      stringPtr("team2"),
		NextGameID:   stringPtr("championship"),
		NextGameSlot: intPtr(1),
	}
	game2 := &TournamentGame{
		ID:           "game2",
		Team1ID:      stringPtr("team3"),
		Team2ID:      stringPtr("team4"),
		NextGameID:   stringPtr("championship"),
		NextGameSlot: intPtr(2),
	}

	// Championship game
	championship := &TournamentGame{
		ID:         "championship",
		Team1ID:    nil,
		Team2ID:    nil,
		NextGameID: nil,
	}

	gamesMap["game1"] = game1
	gamesMap["game2"] = game2
	gamesMap["championship"] = championship

	// Test cases for 4-game tournament
	tests := []struct {
		name          string
		teamID        string
		round         int
		expectedCount int
		expectedTeams []string
	}{
		{
			name:          "team1_in_current_round",
			teamID:        "team1",
			round:         1,
			expectedCount: 1,
			expectedTeams: []string{"team2"},
		},
		{
			name:          "team1_in_championship_round",
			teamID:        "team1",
			round:         2,
			expectedCount: 2,
			expectedTeams: []string{"team3", "team4"},
		},
		{
			name:          "team1_in_earlier_round",
			teamID:        "team1",
			round:         0,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opponents := GetPotentialOpponents(gamesMap, tt.teamID, tt.round, 2)

			if len(opponents) != tt.expectedCount {
				t.Errorf("GetPotentialOpponents() returned %d opponents, expected %d", len(opponents), tt.expectedCount)
			}

			if tt.expectedTeams != nil {
				for _, expectedTeam := range tt.expectedTeams {
					found := false
					for _, opponent := range opponents {
						if opponent == expectedTeam {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected opponent %s not found in returned opponents", expectedTeam)
					}
				}
			}
		})
	}

	// Create an 8-game tournament (3 rounds)
	gamesMap8 := make(map[string]*TournamentGame)

	// Round 1 games
	game1_8 := &TournamentGame{
		ID:           "game1_8",
		Team1ID:      stringPtr("team1_8"),
		Team2ID:      stringPtr("team2_8"),
		NextGameID:   stringPtr("semifinal1"),
		NextGameSlot: intPtr(1),
	}
	game2_8 := &TournamentGame{
		ID:           "game2_8",
		Team1ID:      stringPtr("team3_8"),
		Team2ID:      stringPtr("team4_8"),
		NextGameID:   stringPtr("semifinal1"),
		NextGameSlot: intPtr(2),
	}
	game3_8 := &TournamentGame{
		ID:           "game3_8",
		Team1ID:      stringPtr("team5_8"),
		Team2ID:      stringPtr("team6_8"),
		NextGameID:   stringPtr("semifinal2"),
		NextGameSlot: intPtr(1),
	}
	game4_8 := &TournamentGame{
		ID:           "game4_8",
		Team1ID:      stringPtr("team7_8"),
		Team2ID:      stringPtr("team8_8"),
		NextGameID:   stringPtr("semifinal2"),
		NextGameSlot: intPtr(2),
	}

	// Semifinal games
	semifinal1 := &TournamentGame{
		ID:           "semifinal1",
		Team1ID:      nil,
		Team2ID:      nil,
		NextGameID:   stringPtr("championship8"),
		NextGameSlot: intPtr(1),
	}
	semifinal2 := &TournamentGame{
		ID:           "semifinal2",
		Team1ID:      nil,
		Team2ID:      nil,
		NextGameID:   stringPtr("championship8"),
		NextGameSlot: intPtr(2),
	}

	// Championship game
	championship8 := &TournamentGame{
		ID:         "championship8",
		Team1ID:    nil,
		Team2ID:    nil,
		NextGameID: nil,
	}

	gamesMap8["game1_8"] = game1_8
	gamesMap8["game2_8"] = game2_8
	gamesMap8["game3_8"] = game3_8
	gamesMap8["game4_8"] = game4_8
	gamesMap8["semifinal1"] = semifinal1
	gamesMap8["semifinal2"] = semifinal2
	gamesMap8["championship8"] = championship8

	// Test cases for 8-game tournament
	tests8 := []struct {
		name          string
		teamID        string
		round         int
		expectedCount int
		expectedTeams []string
	}{
		{
			name:          "team1_8_in_current_round",
			teamID:        "team1_8",
			round:         1,
			expectedCount: 1,
			expectedTeams: []string{"team2_8"},
		},
		{
			name:          "team1_8_in_semifinals",
			teamID:        "team1_8",
			round:         2,
			expectedCount: 2,
			expectedTeams: []string{"team3_8", "team4_8"},
		},
		{
			name:          "team1_8_in_championship",
			teamID:        "team1_8",
			round:         3,
			expectedCount: 4,
			expectedTeams: []string{"team5_8", "team6_8", "team7_8", "team8_8"},
		},
		{
			name:          "team1_8_in_earlier_round",
			teamID:        "team1_8",
			round:         0,
			expectedCount: 0,
		},
	}

	for _, tt := range tests8 {
		t.Run(tt.name, func(t *testing.T) {
			opponents := GetPotentialOpponents(gamesMap8, tt.teamID, tt.round, 3)

			if len(opponents) != tt.expectedCount {
				t.Errorf("GetPotentialOpponents() returned %d opponents, expected %d", len(opponents), tt.expectedCount)
			}

			if tt.expectedTeams != nil {
				for _, expectedTeam := range tt.expectedTeams {
					found := false
					for _, opponent := range opponents {
						if opponent == expectedTeam {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected opponent %s not found in returned opponents", expectedTeam)
					}
				}
			}
		})
	}
}

func TestGetPotentialMatchupRound(t *testing.T) {
	// Create a 4-game tournament (3 rounds)
	// Championship
	//     /    \
	// Game1    Game2
	fourGameMap := map[string]*TournamentGame{
		"championship": {
			ID:           "championship",
			NextGameID:   nil,
			NextGameSlot: nil,
			SortOrder:    4,
		},
		"game1": {
			ID:           "game1",
			NextGameID:   stringPtr("championship"),
			NextGameSlot: intPtr(1),
			SortOrder:    1,
		},
		"game2": {
			ID:           "game2",
			NextGameID:   stringPtr("championship"),
			NextGameSlot: intPtr(2),
			SortOrder:    2,
		},
	}

	team1 := "team1"
	team2 := "team2"
	team3 := "team3"
	team4 := "team4"

	// Assign teams to games
	fourGameMap["game1"].Team1ID = &team1
	fourGameMap["game1"].Team2ID = &team2
	fourGameMap["game2"].Team1ID = &team3
	fourGameMap["game2"].Team2ID = &team4

	// Create an 8-game tournament (4 rounds)
	//           Championship
	//          /           \
	//      Semi1          Semi2
	//     /    \         /    \
	// Game1    Game2  Game3  Game4
	eightGameMap := map[string]*TournamentGame{
		"championship": {
			ID:           "championship",
			NextGameID:   nil,
			NextGameSlot: nil,
			SortOrder:    8,
		},
		"semi1": {
			ID:           "semi1",
			NextGameID:   stringPtr("championship"),
			NextGameSlot: intPtr(1),
			SortOrder:    6,
		},
		"semi2": {
			ID:           "semi2",
			NextGameID:   stringPtr("championship"),
			NextGameSlot: intPtr(2),
			SortOrder:    7,
		},
		"game1": {
			ID:           "game1",
			NextGameID:   stringPtr("semi1"),
			NextGameSlot: intPtr(1),
			SortOrder:    1,
		},
		"game2": {
			ID:           "game2",
			NextGameID:   stringPtr("semi1"),
			NextGameSlot: intPtr(2),
			SortOrder:    2,
		},
		"game3": {
			ID:           "game3",
			NextGameID:   stringPtr("semi2"),
			NextGameSlot: intPtr(1),
			SortOrder:    3,
		},
		"game4": {
			ID:           "game4",
			NextGameID:   stringPtr("semi2"),
			NextGameSlot: intPtr(2),
			SortOrder:    4,
		},
	}

	team5 := "team5"
	team6 := "team6"
	team7 := "team7"
	team8 := "team8"

	// Assign teams to games
	eightGameMap["game1"].Team1ID = &team1
	eightGameMap["game1"].Team2ID = &team2
	eightGameMap["game2"].Team1ID = &team3
	eightGameMap["game2"].Team2ID = &team4
	eightGameMap["game3"].Team1ID = &team5
	eightGameMap["game3"].Team2ID = &team6
	eightGameMap["game4"].Team1ID = &team7
	eightGameMap["game4"].Team2ID = &team8

	tests := []struct {
		name          string
		gamesMap      map[string]*TournamentGame
		team1ID       string
		team2ID       string
		totalRounds   int
		expectedRound int
	}{
		// 4-game tournament tests
		{
			name:          "4-game: teams in same game",
			gamesMap:      fourGameMap,
			team1ID:       team1,
			team2ID:       team2,
			totalRounds:   3,
			expectedRound: 2, // First round
		},
		{
			name:          "4-game: teams meet in championship",
			gamesMap:      fourGameMap,
			team1ID:       team1,
			team2ID:       team3,
			totalRounds:   3,
			expectedRound: 3, // Championship
		},
		{
			name:          "4-game: non-existent team",
			gamesMap:      fourGameMap,
			team1ID:       "non-existent",
			team2ID:       team1,
			totalRounds:   3,
			expectedRound: -1,
		},

		// 8-game tournament tests
		{
			name:          "8-game: teams in same game",
			gamesMap:      eightGameMap,
			team1ID:       team1,
			team2ID:       team2,
			totalRounds:   4,
			expectedRound: 2, // First round
		},
		{
			name:          "8-game: teams meet in semifinals",
			gamesMap:      eightGameMap,
			team1ID:       team1,
			team2ID:       team3,
			totalRounds:   4,
			expectedRound: 3, // Semifinals
		},
		{
			name:          "8-game: teams meet in championship",
			gamesMap:      eightGameMap,
			team1ID:       team1,
			team2ID:       team5,
			totalRounds:   4,
			expectedRound: 4, // Championship
		},
		{
			name:          "8-game: teams from different semifinals",
			gamesMap:      eightGameMap,
			team1ID:       team3,
			team2ID:       team7,
			totalRounds:   4,
			expectedRound: 4, // Championship
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			round := GetPotentialMatchupRound(tt.gamesMap, tt.team1ID, tt.team2ID, tt.totalRounds)
			if round != tt.expectedRound {
				t.Errorf("GetPotentialMatchupRound() = %v, want %v", round, tt.expectedRound)
			}
		})
	}
}
