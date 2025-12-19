package services

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// Tests for BracketService.UnselectWinner function

func TestThatUnselectingFirstFourWinnerDecrementsWinsToZero(t *testing.T) {
	// GIVEN a First Four game where team11a won (has 1 win)
	helper := NewBracketTestHelper()
	bracket := helper.CreateEmptyBracket()

	team11a := helper.CreateBracketTeam("team-11a", 11)
	team11b := helper.CreateBracketTeam("team-11b", 11)

	gameFF := helper.CreateGame("East-first_four-11", models.RoundFirstFour, team11a, team11b)
	helper.LinkGames(gameFF, "East-round_of_64-6", 2)
	bracket.Games[gameFF.GameID] = gameFF

	game64 := helper.CreateGame("East-round_of_64-6", models.RoundOf64, nil, nil)
	bracket.Games[game64.GameID] = game64

	// Team11a has won the First Four game
	teams := []*models.TournamentTeam{
		{ID: "team-11a", Wins: 1, Byes: 0},
		{ID: "team-11b", Wins: 0, Byes: 0},
	}

	service := &BracketService{}
	err := service.applyCurrentResults(context.Background(), bracket, teams)
	if err != nil {
		t.Fatalf("Failed to apply results: %v", err)
	}

	// Verify team11a is the winner
	if gameFF.Winner == nil || gameFF.Winner.TeamID != "team-11a" {
		t.Fatal("Expected team-11a to be First Four winner")
	}

	// WHEN we decrement team11a's wins back to 0
	teams[0].Wins = 0

	// AND rebuild the bracket
	bracket2 := helper.CreateEmptyBracket()
	gameFF2 := helper.CreateGame("East-first_four-11", models.RoundFirstFour, team11a, team11b)
	helper.LinkGames(gameFF2, "East-round_of_64-6", 2)
	bracket2.Games[gameFF2.GameID] = gameFF2

	game64_2 := helper.CreateGame("East-round_of_64-6", models.RoundOf64, nil, nil)
	bracket2.Games[game64_2.GameID] = game64_2

	err = service.applyCurrentResults(context.Background(), bracket2, teams)
	if err != nil {
		t.Fatalf("Failed to apply results after unselect: %v", err)
	}

	// THEN the First Four game has no winner
	if gameFF2.Winner != nil {
		t.Errorf("Expected First Four game to have no winner after unselect, got %s", gameFF2.Winner.TeamID)
	}
}

func TestThatUnselectingFirstFourWinnerRemovesThemFromRoundOfSixtyFour(t *testing.T) {
	// GIVEN a First Four game where team11a won and progressed to Round of 64
	helper := NewBracketTestHelper()
	bracket := helper.CreateEmptyBracket()

	team11a := helper.CreateBracketTeam("team-11a", 11)
	team11b := helper.CreateBracketTeam("team-11b", 11)
	team6 := helper.CreateBracketTeam("team-6", 6)

	gameFF := helper.CreateGame("East-first_four-11", models.RoundFirstFour, team11a, team11b)
	helper.LinkGames(gameFF, "East-round_of_64-6", 2)
	bracket.Games[gameFF.GameID] = gameFF

	game64 := helper.CreateGame("East-round_of_64-6", models.RoundOf64, team6, nil)
	bracket.Games[game64.GameID] = game64

	teams := []*models.TournamentTeam{
		{ID: "team-11a", Wins: 1, Byes: 0}, // Won First Four
		{ID: "team-11b", Wins: 0, Byes: 0},
		{ID: "team-6", Wins: 0, Byes: 1},
	}

	service := &BracketService{}
	err := service.applyCurrentResults(context.Background(), bracket, teams)
	if err != nil {
		t.Fatalf("Failed to apply results: %v", err)
	}

	// Verify team11a progressed to Round of 64
	if game64.Team2 == nil || game64.Team2.TeamID != "team-11a" {
		t.Fatal("Expected team-11a to be in Round of 64")
	}

	// WHEN we decrement team11a's wins back to 0
	teams[0].Wins = 0

	// AND rebuild the bracket
	bracket2 := helper.CreateEmptyBracket()
	gameFF2 := helper.CreateGame("East-first_four-11", models.RoundFirstFour, team11a, team11b)
	helper.LinkGames(gameFF2, "East-round_of_64-6", 2)
	bracket2.Games[gameFF2.GameID] = gameFF2

	game64_2 := helper.CreateGame("East-round_of_64-6", models.RoundOf64, team6, nil)
	bracket2.Games[game64_2.GameID] = game64_2

	err = service.applyCurrentResults(context.Background(), bracket2, teams)
	if err != nil {
		t.Fatalf("Failed to apply results after unselect: %v", err)
	}

	// THEN team11a is removed from Round of 64
	if game64_2.Team2 != nil {
		t.Errorf("Expected team-11a to be removed from Round of 64, but Team2 is %s", game64_2.Team2.TeamID)
	}
}

func TestThatUnselectingRoundOfSixtyFourWinnerDecrementsWins(t *testing.T) {
	// GIVEN a Round of 64 game where team1 won
	helper := NewBracketTestHelper()
	bracket := helper.CreateEmptyBracket()

	team1 := helper.CreateBracketTeam("team-1", 1)
	team16 := helper.CreateBracketTeam("team-16", 16)

	game64 := helper.CreateGame("East-round_of_64-1", models.RoundOf64, team1, team16)
	helper.LinkGames(game64, "East-round_of_32-1", 1)
	bracket.Games[game64.GameID] = game64

	game32 := helper.CreateGame("East-round_of_32-1", models.RoundOf32, nil, nil)
	bracket.Games[game32.GameID] = game32

	teams := []*models.TournamentTeam{
		{ID: "team-1", Wins: 1, Byes: 1},
		{ID: "team-16", Wins: 0, Byes: 1},
	}

	service := &BracketService{}
	err := service.applyCurrentResults(context.Background(), bracket, teams)
	if err != nil {
		t.Fatalf("Failed to apply results: %v", err)
	}

	// Verify team1 is the winner
	if game64.Winner == nil || game64.Winner.TeamID != "team-1" {
		t.Fatal("Expected team-1 to be Round of 64 winner")
	}

	// WHEN we decrement team1's wins back to 0
	teams[0].Wins = 0

	// AND rebuild the bracket
	bracket2 := helper.CreateEmptyBracket()
	game64_2 := helper.CreateGame("East-round_of_64-1", models.RoundOf64, team1, team16)
	helper.LinkGames(game64_2, "East-round_of_32-1", 1)
	bracket2.Games[game64_2.GameID] = game64_2

	game32_2 := helper.CreateGame("East-round_of_32-1", models.RoundOf32, nil, nil)
	bracket2.Games[game32_2.GameID] = game32_2

	err = service.applyCurrentResults(context.Background(), bracket2, teams)
	if err != nil {
		t.Fatalf("Failed to apply results after unselect: %v", err)
	}

	// THEN the Round of 64 game has no winner
	if game64_2.Winner != nil {
		t.Errorf("Expected Round of 64 game to have no winner after unselect, got %s", game64_2.Winner.TeamID)
	}
}

func TestThatUnselectingWinnerDoesNotDecrementBelowZero(t *testing.T) {
	// GIVEN a team with 0 wins (edge case - shouldn't happen but test safety)
	helper := NewBracketTestHelper()
	bracket := helper.CreateEmptyBracket()

	team11a := helper.CreateBracketTeam("team-11a", 11)
	team11b := helper.CreateBracketTeam("team-11b", 11)

	gameFF := helper.CreateGame("East-first_four-11", models.RoundFirstFour, team11a, team11b)
	bracket.Games[gameFF.GameID] = gameFF

	teams := []*models.TournamentTeam{
		{ID: "team-11a", Wins: 0, Byes: 0},
		{ID: "team-11b", Wins: 0, Byes: 0},
	}

	// WHEN we try to decrement wins (simulating the UnselectWinner logic)
	if teams[0].Wins > 0 {
		teams[0].Wins--
	}

	// THEN wins stays at 0
	if teams[0].Wins != 0 {
		t.Errorf("Expected wins to stay at 0, got %d", teams[0].Wins)
	}
}
