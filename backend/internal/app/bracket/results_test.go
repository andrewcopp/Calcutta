package bracket

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatApplyCurrentResultsSetsWinnerWhenTeamHasMoreProgress(t *testing.T) {
	// GIVEN a Round of 64 game where team1 has 2 wins+byes (> minRequired=1) and team2 has 1 bye (>= minRequired=1)
	bracket := &models.BracketStructure{
		Games: map[string]*models.BracketGame{
			"game1": {
				GameID: "game1",
				Round:  models.RoundOf64,
				Team1:  &models.BracketTeam{TeamID: "team-a", Seed: 1, LowestSeedSeen: 1},
				Team2:  &models.BracketTeam{TeamID: "team-b", Seed: 16, LowestSeedSeen: 16},
			},
		},
	}
	teams := []*models.TournamentTeam{
		{ID: "team-a", Wins: 1, Byes: 1}, // progress = 2, minRequired = 1, 2 > 1 = true
		{ID: "team-b", Wins: 0, Byes: 1}, // progress = 1, minRequired = 1, 1 >= 1 = true
	}
	svc := &Service{}

	// WHEN applying current results
	_ = svc.applyCurrentResults(context.Background(), bracket, teams)

	// THEN team1 is the winner
	game := bracket.Games["game1"]
	if game.Winner == nil || game.Winner.TeamID != "team-a" {
		winnerID := ""
		if game.Winner != nil {
			winnerID = game.Winner.TeamID
		}
		t.Errorf("expected winner to be 'team-a', got '%s'", winnerID)
	}
}

func TestThatApplyCurrentResultsPropagatesWinnerToNextGame(t *testing.T) {
	// GIVEN a Round of 64 game with a next game where the winner should slot into Team1
	bracket := &models.BracketStructure{
		Games: map[string]*models.BracketGame{
			"game1": {
				GameID:       "game1",
				Round:        models.RoundOf64,
				Team1:        &models.BracketTeam{TeamID: "team-a", Seed: 1, LowestSeedSeen: 1},
				Team2:        &models.BracketTeam{TeamID: "team-b", Seed: 16, LowestSeedSeen: 16},
				NextGameID:   "game2",
				NextGameSlot: 1,
			},
			"game2": {
				GameID: "game2",
				Round:  models.RoundOf32,
			},
		},
	}
	teams := []*models.TournamentTeam{
		{ID: "team-a", Wins: 1, Byes: 1}, // progress = 2 > minRequired = 1
		{ID: "team-b", Wins: 0, Byes: 1}, // progress = 1 >= minRequired = 1
	}
	svc := &Service{}

	// WHEN applying current results
	_ = svc.applyCurrentResults(context.Background(), bracket, teams)

	// THEN the next game's Team1 is set to the winner
	nextGame := bracket.Games["game2"]
	if nextGame.Team1 == nil || nextGame.Team1.TeamID != "team-a" {
		slotID := ""
		if nextGame.Team1 != nil {
			slotID = nextGame.Team1.TeamID
		}
		t.Errorf("expected next game Team1 to be 'team-a', got '%s'", slotID)
	}
}

func TestThatApplyCurrentResultsDoesNotSetWinnerWhenBothTeamsLackProgress(t *testing.T) {
	// GIVEN a Round of 64 game where neither team has progress > minRequired (1)
	bracket := &models.BracketStructure{
		Games: map[string]*models.BracketGame{
			"game1": {
				GameID: "game1",
				Round:  models.RoundOf64,
				Team1:  &models.BracketTeam{TeamID: "team-a", Seed: 1, LowestSeedSeen: 1},
				Team2:  &models.BracketTeam{TeamID: "team-b", Seed: 16, LowestSeedSeen: 16},
			},
		},
	}
	teams := []*models.TournamentTeam{
		{ID: "team-a", Wins: 0, Byes: 1}, // progress = 1, not > minRequired = 1
		{ID: "team-b", Wins: 0, Byes: 1}, // progress = 1, not > minRequired = 1
	}
	svc := &Service{}

	// WHEN applying current results
	_ = svc.applyCurrentResults(context.Background(), bracket, teams)

	// THEN no winner is set
	game := bracket.Games["game1"]
	if game.Winner != nil {
		t.Errorf("expected no winner, got '%s'", game.Winner.TeamID)
	}
}

func TestThatApplyCurrentResultsNilsNextSlotWhenNoWinner(t *testing.T) {
	// GIVEN a Round of 64 game with no winner and a next game that previously had a Team1
	bracket := &models.BracketStructure{
		Games: map[string]*models.BracketGame{
			"game1": {
				GameID:       "game1",
				Round:        models.RoundOf64,
				Team1:        &models.BracketTeam{TeamID: "team-a", Seed: 1, LowestSeedSeen: 1},
				Team2:        &models.BracketTeam{TeamID: "team-b", Seed: 16, LowestSeedSeen: 16},
				NextGameID:   "game2",
				NextGameSlot: 1,
			},
			"game2": {
				GameID: "game2",
				Round:  models.RoundOf32,
				Team1:  &models.BracketTeam{TeamID: "stale-team", Seed: 1, LowestSeedSeen: 1},
			},
		},
	}
	teams := []*models.TournamentTeam{
		{ID: "team-a", Wins: 0, Byes: 1}, // progress = 1, not > minRequired = 1
		{ID: "team-b", Wins: 0, Byes: 1}, // progress = 1, not > minRequired = 1
	}
	svc := &Service{}

	// WHEN applying current results
	_ = svc.applyCurrentResults(context.Background(), bracket, teams)

	// THEN the next game's Team1 slot is nil
	nextGame := bracket.Games["game2"]
	if nextGame.Team1 != nil {
		t.Errorf("expected next game Team1 to be nil, got '%s'", nextGame.Team1.TeamID)
	}
}
