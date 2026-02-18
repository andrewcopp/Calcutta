package models

import (
	"testing"
)

func TestThatTeamWithByeAndNoGamesPlayedHasZeroWins(t *testing.T) {
	// GIVEN a bracket with a 1-seed team that has not played
	bracket := createTestBracketWithTwoTeams()

	// WHEN calculating wins and byes
	wins, _, _ := CalculateWinsAndByes("team1", bracket)

	// THEN the team has 0 wins
	if wins != 0 {
		t.Errorf("Expected 0 wins, got %d", wins)
	}
}

func TestThatTeamWithByeAndNoGamesPlayedHasOneBye(t *testing.T) {
	// GIVEN a bracket with a 1-seed team that has not played
	bracket := createTestBracketWithTwoTeams()

	// WHEN calculating wins and byes
	_, byes, _ := CalculateWinsAndByes("team1", bracket)

	// THEN the team has 1 bye
	if byes != 1 {
		t.Errorf("Expected 1 bye, got %d", byes)
	}
}

func TestThatTeamWithByeAndNoGamesPlayedIsNotEliminated(t *testing.T) {
	// GIVEN a bracket with a 1-seed team that has not played
	bracket := createTestBracketWithTwoTeams()

	// WHEN calculating wins and byes
	_, _, eliminated := CalculateWinsAndByes("team1", bracket)

	// THEN the team is not eliminated
	if eliminated {
		t.Errorf("Expected team to not be eliminated")
	}
}

func TestThatTeamWinningFirstGameHasOneWin(t *testing.T) {
	// GIVEN a bracket where team1 has won their first game
	bracket := createTestBracketWithTwoTeams()
	bracket.Games["game1"].Winner = bracket.Games["game1"].Team1
	bracket.Games["game2"].Team1 = bracket.Games["game1"].Team1

	// WHEN calculating wins and byes
	wins, _, _ := CalculateWinsAndByes("team1", bracket)

	// THEN the team has 1 win
	if wins != 1 {
		t.Errorf("Expected 1 win, got %d", wins)
	}
}

func TestThatTeamLosingFirstGameIsEliminated(t *testing.T) {
	// GIVEN a bracket where team1 has lost their first game
	bracket := createTestBracketWithTwoTeams()
	bracket.Games["game1"].Winner = bracket.Games["game1"].Team2

	// WHEN calculating wins and byes
	_, _, eliminated := CalculateWinsAndByes("team1", bracket)

	// THEN the team is eliminated
	if !eliminated {
		t.Errorf("Expected team to be eliminated")
	}
}

func createTestBracketWithTwoTeams() *BracketStructure {
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

	return bracket
}

func TestThatValidateWinnerSelectionAcceptsParticipant(t *testing.T) {
	// GIVEN a game with two teams
	game := &BracketGame{
		Team1: &BracketTeam{TeamID: "team1"},
		Team2: &BracketTeam{TeamID: "team2"},
	}

	// WHEN validating team1 as winner
	err := ValidateWinnerSelection(game, "team1")

	// THEN no error is returned
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestThatValidateWinnerSelectionRejectsNonParticipant(t *testing.T) {
	// GIVEN a game with two teams
	game := &BracketGame{
		Team1: &BracketTeam{TeamID: "team1"},
		Team2: &BracketTeam{TeamID: "team2"},
	}

	// WHEN validating a non-participant as winner
	err := ValidateWinnerSelection(game, "team3")

	// THEN an error is returned
	if err == nil {
		t.Error("Expected error for non-participant winner")
	}
}

func TestThatValidateWinnerSelectionRejectsMissingTeam(t *testing.T) {
	// GIVEN a game with only one team
	game := &BracketGame{
		Team1: &BracketTeam{TeamID: "team1"},
	}

	// WHEN validating a winner
	err := ValidateWinnerSelection(game, "team1")

	// THEN an error is returned
	if err == nil {
		t.Error("Expected error when team2 is missing")
	}
}

func TestThatValidateWinnerSelectionRejectsNilGame(t *testing.T) {
	// WHEN validating a nil game
	err := ValidateWinnerSelection(nil, "team1")

	// THEN an error is returned
	if err == nil {
		t.Error("Expected error for nil game")
	}
}
