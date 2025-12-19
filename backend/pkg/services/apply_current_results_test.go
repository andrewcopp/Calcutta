package services

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// Tests for BracketService.applyCurrentResults function

func TestThatPlayInWinnerAppearsInRoundOfSixtyFourButDoesNotAutoWin(t *testing.T) {
	// GIVEN a bracket with a play-in game where team11a won (1 win, 0 byes)
	// and their Round of 64 opponent has 0 wins, 1 bye
	helper := NewBracketTestHelper()
	bracket := helper.CreateEmptyBracket()

	team11a := helper.CreateBracketTeam("team-11a", 11)
	team11b := helper.CreateBracketTeam("team-11b", 11)
	team6 := helper.CreateBracketTeam("team-6", 6)

	// First Four game
	gameFF := helper.CreateGame("East-first_four-11", models.RoundFirstFour, team11a, team11b)
	helper.LinkGames(gameFF, "East-round_of_64-6", 2)
	bracket.Games[gameFF.GameID] = gameFF

	// Round of 64 game
	game64 := helper.CreateGame("East-round_of_64-6", models.RoundOf64, team6, nil)
	bracket.Games[game64.GameID] = game64

	// Create team data with wins
	teams := []*models.TournamentTeam{
		{ID: "team-11a", Wins: 1, Byes: 0}, // Won play-in
		{ID: "team-11b", Wins: 0, Byes: 0}, // Lost play-in
		{ID: "team-6", Wins: 0, Byes: 1},   // Bye to Round of 64
	}

	service := &BracketService{}

	// WHEN applying current results
	err := service.applyCurrentResults(context.Background(), bracket, teams)
	if err != nil {
		t.Fatalf("Failed to apply results: %v", err)
	}

	// THEN the play-in winner appears in Round of 64
	if game64.Team2 == nil {
		t.Error("Expected play-in winner to appear in Round of 64 game")
		return
	}
	if game64.Team2.TeamID != "team-11a" {
		t.Errorf("Expected team-11a in Round of 64, got %s", game64.Team2.TeamID)
	}

	// AND the Round of 64 game has no winner yet
	if game64.Winner != nil {
		t.Errorf("Expected Round of 64 game to have no winner, but winner is %s", game64.Winner.TeamID)
	}
}

func TestThatTeamsWithEqualWinsHaveNoWinnerDetermined(t *testing.T) {
	// GIVEN a Round of 64 game with both teams having 0 wins
	helper := NewBracketTestHelper()
	bracket := helper.CreateEmptyBracket()

	team1 := helper.CreateBracketTeam("team-1", 1)
	team16 := helper.CreateBracketTeam("team-16", 16)

	game64 := helper.CreateGame("East-round_of_64-1", models.RoundOf64, team1, team16)
	bracket.Games[game64.GameID] = game64

	teams := []*models.TournamentTeam{
		{ID: "team-1", Wins: 0, Byes: 1},
		{ID: "team-16", Wins: 0, Byes: 1},
	}

	service := &BracketService{}

	// WHEN applying current results
	err := service.applyCurrentResults(context.Background(), bracket, teams)
	if err != nil {
		t.Fatalf("Failed to apply results: %v", err)
	}

	// THEN the game has no winner
	if game64.Winner != nil {
		t.Errorf("Expected no winner when both teams have equal wins, got %s", game64.Winner.TeamID)
	}
}

func TestThatTeamWithMoreWinsIsMarkedAsWinner(t *testing.T) {
	// GIVEN a Round of 64 game where team1 has 1 win and team16 has 0 wins
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

	// WHEN applying current results
	err := service.applyCurrentResults(context.Background(), bracket, teams)
	if err != nil {
		t.Fatalf("Failed to apply results: %v", err)
	}

	// THEN team1 is marked as the winner
	if game64.Winner == nil {
		t.Error("Expected game to have a winner")
		return
	}
	if game64.Winner.TeamID != "team-1" {
		t.Errorf("Expected team-1 to be winner, got %s", game64.Winner.TeamID)
	}
}

func TestThatWinnerProgressesToNextGame(t *testing.T) {
	// GIVEN a Round of 64 game with a winner
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

	// WHEN applying current results
	err := service.applyCurrentResults(context.Background(), bracket, teams)
	if err != nil {
		t.Fatalf("Failed to apply results: %v", err)
	}

	// THEN the winner appears in Round of 32
	if game32.Team1 == nil {
		t.Error("Expected winner to progress to Round of 32")
		return
	}
	if game32.Team1.TeamID != "team-1" {
		t.Errorf("Expected team-1 in Round of 32, got %s", game32.Team1.TeamID)
	}
}

func TestThatLowestSeedSeenIsUpdatedWhenWinnerProgresses(t *testing.T) {
	// GIVEN a Round of 64 game where 9-seed beats 8-seed (upset)
	helper := NewBracketTestHelper()
	bracket := helper.CreateEmptyBracket()

	team8 := helper.CreateBracketTeam("team-8", 8)
	team9 := helper.CreateBracketTeam("team-9", 9)

	game64 := helper.CreateGame("East-round_of_64-8", models.RoundOf64, team8, team9)
	helper.LinkGames(game64, "East-round_of_32-1", 2)
	bracket.Games[game64.GameID] = game64

	game32 := helper.CreateGame("East-round_of_32-1", models.RoundOf32, nil, nil)
	bracket.Games[game32.GameID] = game32

	teams := []*models.TournamentTeam{
		{ID: "team-8", Wins: 0, Byes: 1},
		{ID: "team-9", Wins: 1, Byes: 1}, // 9-seed won (upset)
	}

	service := &BracketService{}

	// WHEN applying current results
	err := service.applyCurrentResults(context.Background(), bracket, teams)
	if err != nil {
		t.Fatalf("Failed to apply results: %v", err)
	}

	// THEN the winner in Round of 32 has LowestSeedSeen = 8 (the upset victim)
	if game32.Team2 == nil {
		t.Error("Expected winner to progress to Round of 32")
		return
	}
	if game32.Team2.LowestSeedSeen != 8 {
		t.Errorf("Expected LowestSeedSeen = 8 (upset), got %d", game32.Team2.LowestSeedSeen)
	}
}

func TestThatFirstFourWinnerHasLowestSeedSeenEqualToSeed(t *testing.T) {
	// GIVEN a First Four game where both teams are 11-seeds
	helper := NewBracketTestHelper()
	bracket := helper.CreateEmptyBracket()

	team11a := helper.CreateBracketTeam("team-11a", 11)
	team11b := helper.CreateBracketTeam("team-11b", 11)

	gameFF := helper.CreateGame("East-first_four-11", models.RoundFirstFour, team11a, team11b)
	helper.LinkGames(gameFF, "East-round_of_64-6", 2)
	bracket.Games[gameFF.GameID] = gameFF

	game64 := helper.CreateGame("East-round_of_64-6", models.RoundOf64, nil, nil)
	bracket.Games[game64.GameID] = game64

	teams := []*models.TournamentTeam{
		{ID: "team-11a", Wins: 1, Byes: 0}, // Won First Four
		{ID: "team-11b", Wins: 0, Byes: 0},
	}

	service := &BracketService{}

	// WHEN applying current results
	err := service.applyCurrentResults(context.Background(), bracket, teams)
	if err != nil {
		t.Fatalf("Failed to apply results: %v", err)
	}

	// THEN the winner in Round of 64 has LowestSeedSeen = 11 (their own seed)
	if game64.Team2 == nil {
		t.Error("Expected First Four winner to progress to Round of 64")
		return
	}
	if game64.Team2.LowestSeedSeen != 11 {
		t.Errorf("Expected LowestSeedSeen = 11, got %d", game64.Team2.LowestSeedSeen)
	}
}

func TestThatMultipleRoundsAreProcessedInOrder(t *testing.T) {
	// GIVEN a bracket with winners through multiple rounds
	helper := NewBracketTestHelper()
	bracket := helper.CreateEmptyBracket()

	team1 := helper.CreateBracketTeam("team-1", 1)
	team16 := helper.CreateBracketTeam("team-16", 16)
	team8 := helper.CreateBracketTeam("team-8", 8)
	team9 := helper.CreateBracketTeam("team-9", 9)

	game64_1 := helper.CreateGame("East-round_of_64-1", models.RoundOf64, team1, team16)
	helper.LinkGames(game64_1, "East-round_of_32-1", 1)
	bracket.Games[game64_1.GameID] = game64_1

	game64_8 := helper.CreateGame("East-round_of_64-8", models.RoundOf64, team8, team9)
	helper.LinkGames(game64_8, "East-round_of_32-1", 2)
	bracket.Games[game64_8.GameID] = game64_8

	game32 := helper.CreateGame("East-round_of_32-1", models.RoundOf32, nil, nil)
	helper.LinkGames(game32, "East-sweet_16-1", 1)
	bracket.Games[game32.GameID] = game32

	sweet16 := helper.CreateGame("East-sweet_16-1", models.RoundSweet16, nil, nil)
	bracket.Games[sweet16.GameID] = sweet16

	teams := []*models.TournamentTeam{
		{ID: "team-1", Wins: 2, Byes: 1}, // Won Round of 64 and Round of 32
		{ID: "team-16", Wins: 0, Byes: 1},
		{ID: "team-8", Wins: 0, Byes: 1},
		{ID: "team-9", Wins: 1, Byes: 1}, // Won Round of 64
	}

	service := &BracketService{}

	// WHEN applying current results
	err := service.applyCurrentResults(context.Background(), bracket, teams)
	if err != nil {
		t.Fatalf("Failed to apply results: %v", err)
	}

	// THEN team-1 appears in Sweet 16
	if sweet16.Team1 == nil {
		t.Error("Expected team-1 to progress to Sweet 16")
		return
	}
	if sweet16.Team1.TeamID != "team-1" {
		t.Errorf("Expected team-1 in Sweet 16, got %s", sweet16.Team1.TeamID)
	}
}
