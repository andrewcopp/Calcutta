package services

import (
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// BracketTestHelper provides factory methods for creating test brackets
type BracketTestHelper struct{}

// NewBracketTestHelper creates a new test helper
func NewBracketTestHelper() *BracketTestHelper {
	return &BracketTestHelper{}
}

// CreateTournament68Teams creates a standard 68-team tournament
// Each region gets 17 teams with one duplicate seed
func (h *BracketTestHelper) CreateTournament68Teams() []*models.TournamentTeam {
	regionDuplicates := map[string]int{
		"East":    11,
		"West":    11,
		"South":   16,
		"Midwest": 16,
	}

	teams := make([]*models.TournamentTeam, 0, 68)
	for _, region := range []string{"East", "West", "South", "Midwest"} {
		duplicateSeed := regionDuplicates[region]
		for seed := 1; seed <= 16; seed++ {
			if seed == duplicateSeed {
				teams = append(teams, h.CreateTeam(region+"-seed-"+fmt.Sprintf("%d", seed)+"a", region, seed))
				teams = append(teams, h.CreateTeam(region+"-seed-"+fmt.Sprintf("%d", seed)+"b", region, seed))
			} else {
				teams = append(teams, h.CreateTeam(region+"-seed-"+fmt.Sprintf("%d", seed), region, seed))
			}
		}
	}
	return teams
}

// CreateStandardRegionTeams creates a region with no duplicate seeds
func (h *BracketTestHelper) CreateStandardRegionTeams(region string, count int) []*models.TournamentTeam {
	teams := make([]*models.TournamentTeam, count)
	for i := 0; i < count; i++ {
		seed := i + 1
		teams[i] = h.CreateTeam(fmt.Sprintf("%s-team-%d", region, seed), region, seed)
	}
	return teams
}

// CreateRegionWithDuplicateSeeds creates a region with specified duplicate seeds
func (h *BracketTestHelper) CreateRegionWithDuplicateSeeds(region string, duplicateSeeds ...int) []*models.TournamentTeam {
	teams := make([]*models.TournamentTeam, 0, 18)
	duplicateMap := make(map[int]bool)
	for _, seed := range duplicateSeeds {
		duplicateMap[seed] = true
	}

	for seed := 1; seed <= 16; seed++ {
		if duplicateMap[seed] {
			teams = append(teams, h.CreateTeam(fmt.Sprintf("%s-team-%da", region, seed), region, seed))
			teams = append(teams, h.CreateTeam(fmt.Sprintf("%s-team-%db", region, seed), region, seed))
		} else {
			teams = append(teams, h.CreateTeam(fmt.Sprintf("%s-team-%d", region, seed), region, seed))
		}
	}
	return teams
}

// CreateTeam creates a single tournament team
func (h *BracketTestHelper) CreateTeam(id, region string, seed int) *models.TournamentTeam {
	return &models.TournamentTeam{
		ID:           id,
		TournamentID: "test-tournament",
		SchoolID:     "school-" + id,
		Seed:         seed,
		Region:       region,
		School: &models.School{
			ID:   "school-" + id,
			Name: "School " + id,
		},
	}
}

// CreateFinalFourConfig creates a standard Final Four configuration
func (h *BracketTestHelper) CreateFinalFourConfig() *models.FinalFourConfig {
	return &models.FinalFourConfig{
		TopLeftRegion:     "East",
		BottomLeftRegion:  "West",
		TopRightRegion:    "South",
		BottomRightRegion: "Midwest",
	}
}

// CreateEmptyBracket creates an empty bracket structure
func (h *BracketTestHelper) CreateEmptyBracket() *models.BracketStructure {
	return &models.BracketStructure{
		TournamentID: "test",
		Games:        make(map[string]*models.BracketGame),
		Regions:      []string{"East", "West", "South", "Midwest"},
	}
}

// CreateBracketTeam creates a bracket team with LowestSeedSeen initialized
func (h *BracketTestHelper) CreateBracketTeam(teamID string, seed int) *models.BracketTeam {
	return &models.BracketTeam{
		TeamID:         teamID,
		Seed:           seed,
		LowestSeedSeen: seed,
	}
}

// CreateGame creates a bracket game with specified parameters
func (h *BracketTestHelper) CreateGame(gameID string, round models.BracketRound, team1, team2 *models.BracketTeam) *models.BracketGame {
	return &models.BracketGame{
		GameID: gameID,
		Round:  round,
		Team1:  team1,
		Team2:  team2,
	}
}

// LinkGames links two games together
func (h *BracketTestHelper) LinkGames(fromGame *models.BracketGame, toGameID string, slot int) {
	fromGame.NextGameID = toGameID
	fromGame.NextGameSlot = slot
}

// SelectWinner sets a winner for a game and progresses them to the next game
func (h *BracketTestHelper) SelectWinner(bracket *models.BracketStructure, gameID, teamID string) error {
	game := bracket.Games[gameID]
	if game == nil {
		return fmt.Errorf("game not found: %s", gameID)
	}

	var winner *models.BracketTeam
	if game.Team1 != nil && game.Team1.TeamID == teamID {
		winner = game.Team1
	} else if game.Team2 != nil && game.Team2.TeamID == teamID {
		winner = game.Team2
	}

	if winner == nil {
		return fmt.Errorf("team not found in game: %s", teamID)
	}

	game.Winner = winner

	// Progress to next game if exists
	if game.NextGameID != "" {
		nextGame := bracket.Games[game.NextGameID]
		if nextGame != nil {
			winnerCopy := *winner
			if game.Team1 != nil && game.Team2 != nil {
				minSeed := game.Team1.LowestSeedSeen
				if game.Team2.LowestSeedSeen < minSeed {
					minSeed = game.Team2.LowestSeedSeen
				}
				winnerCopy.LowestSeedSeen = minSeed
			}

			if game.NextGameSlot == 1 {
				nextGame.Team1 = &winnerCopy
			} else {
				nextGame.Team2 = &winnerCopy
			}
		}
	}

	return nil
}

// UnselectWinner clears a winner and all downstream games
func (h *BracketTestHelper) UnselectWinner(bracket *models.BracketStructure, gameID string) error {
	game := bracket.Games[gameID]
	if game == nil {
		return fmt.Errorf("game not found: %s", gameID)
	}

	game.Winner = nil

	if game.NextGameID != "" {
		nextGame := bracket.Games[game.NextGameID]
		if nextGame != nil {
			if game.NextGameSlot == 1 {
				nextGame.Team1 = nil
			} else {
				nextGame.Team2 = nil
			}
			nextGame.Winner = nil

			if nextGame.NextGameID != "" {
				h.UnselectWinner(bracket, nextGame.GameID)
			}
		}
	}

	return nil
}

// Preset bracket scenarios for common test cases

// CreateBracketWithRound64Game creates a bracket with a single Round of 64 game (1 vs 16)
func (h *BracketTestHelper) CreateBracketWithRound64Game() *models.BracketStructure {
	bracket := h.CreateEmptyBracket()
	team1 := h.CreateBracketTeam("team1-seed1", 1)
	team16 := h.CreateBracketTeam("team16-seed16", 16)

	game := h.CreateGame("East-round_of_64-1", models.RoundOf64, team1, team16)
	h.LinkGames(game, "East-round_of_32-1", 1)
	bracket.Games[game.GameID] = game

	return bracket
}

// CreateBracketWithLinkedGames creates a bracket with Round of 64 and Round of 32 games linked
func (h *BracketTestHelper) CreateBracketWithLinkedGames() *models.BracketStructure {
	bracket := h.CreateEmptyBracket()

	team1 := h.CreateBracketTeam("team1-seed1", 1)
	team16 := h.CreateBracketTeam("team16-seed16", 16)
	team8 := h.CreateBracketTeam("team8-seed8", 8)
	team9 := h.CreateBracketTeam("team9-seed9", 9)

	game64_1 := h.CreateGame("East-round_of_64-1", models.RoundOf64, team1, team16)
	h.LinkGames(game64_1, "East-round_of_32-1", 1)

	game64_8 := h.CreateGame("East-round_of_64-8", models.RoundOf64, team8, team9)
	h.LinkGames(game64_8, "East-round_of_32-1", 2)

	game32 := h.CreateGame("East-round_of_32-1", models.RoundOf32, nil, nil)
	h.LinkGames(game32, "East-sweet_16-1", 1)

	bracket.Games[game64_1.GameID] = game64_1
	bracket.Games[game64_8.GameID] = game64_8
	bracket.Games[game32.GameID] = game32

	return bracket
}

// CreateBracketWithFirstFourGame creates a bracket with a First Four game
func (h *BracketTestHelper) CreateBracketWithFirstFourGame() *models.BracketStructure {
	bracket := h.CreateEmptyBracket()

	team11a := h.CreateBracketTeam("team-ff-11a", 11)
	team11b := h.CreateBracketTeam("team-ff-11b", 11)

	gameFF := h.CreateGame("East-first_four-11", models.RoundFirstFour, team11a, team11b)
	h.LinkGames(gameFF, "East-round_of_64-6", 2)

	game64 := h.CreateGame("East-round_of_64-6", models.RoundOf64, nil, nil)

	bracket.Games[gameFF.GameID] = gameFF
	bracket.Games[game64.GameID] = game64

	return bracket
}
