package models

import (
	"errors"
	"fmt"
)

// BracketRound represents a round in the tournament bracket
type BracketRound string

const (
	RoundFirstFour    BracketRound = "first_four"
	RoundOf64         BracketRound = "round_of_64"
	RoundOf32         BracketRound = "round_of_32"
	RoundSweet16      BracketRound = "sweet_16"
	RoundElite8       BracketRound = "elite_8"
	RoundFinalFour    BracketRound = "final_four"
	RoundChampionship BracketRound = "championship"
)

// BracketGame represents a single game in the bracket with its participants
type BracketGame struct {
	GameID       string       `json:"gameId"`
	Round        BracketRound `json:"round"`
	Region       string       `json:"region"`
	Team1        *BracketTeam `json:"team1,omitempty"`
	Team2        *BracketTeam `json:"team2,omitempty"`
	Winner       *BracketTeam `json:"winner,omitempty"`
	NextGameID   string       `json:"nextGameId,omitempty"`
	NextGameSlot int          `json:"nextGameSlot,omitempty"` // 1 or 2
	SortOrder    int          `json:"sortOrder"`
}

// BracketTeam represents a team in the bracket
type BracketTeam struct {
	TeamID         string `json:"teamId"`
	SchoolID       string `json:"schoolId"`
	Name           string `json:"name"`
	Seed           int    `json:"seed"`
	Region         string `json:"region"`
	LowestSeedSeen int    `json:"lowestSeedSeen"` // For upset tracking in regional rounds
}

// BracketStructure represents the complete tournament bracket
type BracketStructure struct {
	TournamentID string                  `json:"tournamentId"`
	Regions      []string                `json:"regions"`
	Games        map[string]*BracketGame `json:"games"`
	FinalFour    *FinalFourConfig        `json:"finalFour"`
}

// FinalFourConfig defines which regions play each other in the Final Four
type FinalFourConfig struct {
	TopLeftRegion     string `json:"topLeftRegion"`
	BottomLeftRegion  string `json:"bottomLeftRegion"`
	TopRightRegion    string `json:"topRightRegion"`
	BottomRightRegion string `json:"bottomRightRegion"`
}

// BracketValidator validates bracket operations
type BracketValidator struct{}

// NewBracketValidator creates a new bracket validator
func NewBracketValidator() *BracketValidator {
	return &BracketValidator{}
}

// ValidateWinnerSelection validates that a winner can be selected for a game
func (v *BracketValidator) ValidateWinnerSelection(game *BracketGame, winnerTeamID string) error {
	if game == nil {
		return errors.New("game cannot be nil")
	}

	if game.Team1 == nil || game.Team2 == nil {
		return errors.New("both teams must be present before selecting a winner")
	}

	if game.Team1.TeamID != winnerTeamID && game.Team2.TeamID != winnerTeamID {
		return fmt.Errorf("winner must be one of the participating teams")
	}

	return nil
}

// ValidateBracketProgression validates that the bracket progression is valid
func (v *BracketValidator) ValidateBracketProgression(bracket *BracketStructure) error {
	if bracket == nil {
		return errors.New("bracket cannot be nil")
	}

	// Validate that all games with winners have their winners properly progressed
	for _, game := range bracket.Games {
		if game.Winner != nil && game.NextGameID != "" {
			nextGame, exists := bracket.Games[game.NextGameID]
			if !exists {
				return fmt.Errorf("next game %s not found for game %s", game.NextGameID, game.GameID)
			}

			// Check that the winner is in the correct slot of the next game
			if game.NextGameSlot == 1 && (nextGame.Team1 == nil || nextGame.Team1.TeamID != game.Winner.TeamID) {
				return fmt.Errorf("winner not properly progressed to slot 1 of next game")
			}
			if game.NextGameSlot == 2 && (nextGame.Team2 == nil || nextGame.Team2.TeamID != game.Winner.TeamID) {
				return fmt.Errorf("winner not properly progressed to slot 2 of next game")
			}
		}
	}

	return nil
}

// GetMatchupForSeeds returns the seed that a given seed should play in the first round
// Based on NCAA tournament rules where seeds sum to 17
func GetMatchupForSeeds(seed int) int {
	return 17 - seed
}

// IsFirstFourSeed determines if a seed participates in the First Four
// In a 68-team tournament, there are typically 4 First Four games:
// - Four 16 seeds play each other (2 games)
// - Four 11 seeds play each other (2 games)
func IsFirstFourSeed(seed int, region string, teams []*TournamentTeam) bool {
	// Count how many teams have this seed in this region
	count := 0
	for _, team := range teams {
		if team.Seed == seed && team.Region == region {
			count++
		}
	}

	// If there are 2 teams with the same seed in the same region, they play in First Four
	return count == 2
}

// CalculateWinsAndByes calculates wins and byes for a team based on bracket state
func CalculateWinsAndByes(teamID string, bracket *BracketStructure) (wins int, byes int, eliminated bool) {
	wins = 0
	byes = 0
	eliminated = false

	// Find the team's first game
	var firstGame *BracketGame
	for _, game := range bracket.Games {
		if (game.Team1 != nil && game.Team1.TeamID == teamID) ||
			(game.Team2 != nil && game.Team2.TeamID == teamID) {
			if firstGame == nil || game.Round == RoundFirstFour {
				firstGame = game
			}
		}
	}

	if firstGame == nil {
		return 0, 0, true
	}

	// If first game is not First Four, team has a bye
	if firstGame.Round != RoundFirstFour {
		byes = 1
	}

	// Traverse the bracket to count wins
	currentGame := firstGame
	for currentGame != nil {
		if currentGame.Winner == nil {
			// Game not played yet, team is still alive
			break
		}

		if currentGame.Winner.TeamID == teamID {
			wins++
			// Move to next game
			if currentGame.NextGameID == "" {
				// Won championship
				break
			}
			nextGame := bracket.Games[currentGame.NextGameID]
			currentGame = nextGame
		} else {
			// Team lost, they're eliminated
			eliminated = true
			break
		}
	}

	return wins, byes, eliminated
}
