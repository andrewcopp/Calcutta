package models

import (
	"time"
)

// TournamentGame represents a game in a tournament
type TournamentGame struct {
	ID           string     `json:"id"`
	TournamentID string     `json:"tournamentId"`
	Team1ID      *string    `json:"team1Id,omitempty"` // Nullable reference to TournamentTeam
	Team2ID      *string    `json:"team2Id,omitempty"` // Nullable reference to TournamentTeam
	TipoffTime   time.Time  `json:"tipoffTime"`
	SortOrder    int        `json:"sortOrder"`              // Order in the bracket
	Team1Score   int        `json:"team1Score"`             // Score for team 1 (default 0)
	Team2Score   int        `json:"team2Score"`             // Score for team 2 (default 0)
	NextGameID   *string    `json:"nextGameId,omitempty"`   // Reference to the next game in the bracket
	NextGameSlot *int       `json:"nextGameSlot,omitempty"` // Slot in the next game (1 or 2), null for championship game
	IsFinal      bool       `json:"isFinal"`                // Whether the game is final
	Created      time.Time  `json:"created"`
	Updated      time.Time  `json:"updated"`
	Deleted      *time.Time `json:"deleted,omitempty"`
}

// GetGameStatus returns the current status of the game based on tipoff time and is_final
func (g *TournamentGame) GetGameStatus() string {
	now := time.Now()

	if g.TipoffTime.After(now) {
		return "future"
	} else if g.IsFinal {
		return "completed"
	} else {
		return "in_progress"
	}
}

// GetWinnerID returns the ID of the winning team if the game is final
func (g *TournamentGame) GetWinnerID() *string {
	if !g.IsFinal {
		return nil
	}

	if g.Team1Score > g.Team2Score {
		return g.Team1ID
	} else if g.Team2Score > g.Team1Score {
		return g.Team2ID
	}

	// In case of a tie (shouldn't happen in basketball, but just in case)
	return nil
}

// IsChampionshipGame returns true if this is the championship game (no next game)
func (g *TournamentGame) IsChampionshipGame() bool {
	return g.NextGameID == nil
}

// GetRound returns the round number this game belongs to by traversing the tournament bracket
// This method requires a map of all games in the tournament to properly traverse the bracket
func (g *TournamentGame) GetRound(gamesMap map[string]*TournamentGame, totalRounds int) int {
	// If this is the championship game, it's the last round
	if g.IsChampionshipGame() {
		return totalRounds
	}

	// Count the distance from the championship game by traversing the bracket
	distance := g.countDistanceFromChampionship(gamesMap, 0)

	// If we got a distance of 0, it means we're treating this as a championship game
	// So we should return the total rounds
	if distance == 0 {
		return totalRounds
	}

	// Round = total rounds - distance from championship
	return totalRounds - distance
}

// countDistanceFromChampionship is a helper function that uses tail recursion to count
// the distance from the championship game
func (g *TournamentGame) countDistanceFromChampionship(gamesMap map[string]*TournamentGame, distance int) int {
	// Base case: if this is the championship game, return the current distance
	if g.IsChampionshipGame() {
		return distance
	}

	// If there's no next game, we're done
	if g.NextGameID == nil {
		return distance
	}

	// Get the next game from the map
	nextGame, exists := gamesMap[*g.NextGameID]
	if !exists {
		// If we can't find the next game in the map, treat it as a championship game
		// This handles cases where we're testing with a different tournament structure
		return 0
	}

	// Recursive call with incremented distance
	// This is tail recursion because the recursive call is the last operation
	return nextGame.countDistanceFromChampionship(gamesMap, distance+1)
}

// GetPotentialOpponents returns a list of potential opponents for a team in the tournament.
// It traverses the tournament bracket to find teams that could potentially meet the given team
// in the specified round.
func GetPotentialOpponents(gamesMap map[string]*TournamentGame, teamID string, round int, totalRounds int) []string {
	// Find the game where the team is currently playing
	var currentGame *TournamentGame
	for _, game := range gamesMap {
		if (game.Team1ID != nil && *game.Team1ID == teamID) || (game.Team2ID != nil && *game.Team2ID == teamID) {
			currentGame = game
			break
		}
	}

	if currentGame == nil {
		return []string{} // Team not found in tournament
	}

	// If the game is already played, return empty list
	if currentGame.IsFinal {
		return []string{}
	}

	// Get the current round of the team's game
	currentRound := currentGame.GetRound(gamesMap, totalRounds)

	// If the requested round is earlier than the current round, return empty list
	if round < currentRound {
		return []string{} // Can't meet opponents in earlier rounds
	}

	// If we're looking for opponents in the current round
	if round == currentRound {
		opponents := []string{}
		// Add the opponent from the current game
		if currentGame.Team1ID != nil && *currentGame.Team1ID != teamID {
			opponents = append(opponents, *currentGame.Team1ID)
		}
		if currentGame.Team2ID != nil && *currentGame.Team2ID != teamID {
			opponents = append(opponents, *currentGame.Team2ID)
		}
		return opponents
	}

	// For later rounds, traverse up the bracket until we reach the target round
	targetGame := currentGame
	for targetGame != nil && targetGame.GetRound(gamesMap, totalRounds) < round {
		if targetGame.NextGameID == nil {
			return []string{} // No more games up the bracket
		}
		targetGame = gamesMap[*targetGame.NextGameID]
	}

	if targetGame == nil {
		return []string{}
	}

	// Find the path from current game to target game to determine which slot we'll be in
	var ourSlot int
	game := currentGame
	for game != nil && game.ID != targetGame.ID {
		if game.NextGameID == nil {
			return []string{} // Should not happen
		}
		nextGame := gamesMap[*game.NextGameID]
		if nextGame == nil {
			return []string{} // Should not happen
		}
		ourSlot = *game.NextGameSlot
		game = nextGame
	}

	// Now find all potential opponents by looking at the other slot
	otherSlot := 2
	if ourSlot == 2 {
		otherSlot = 1
	}

	// Find all games that feed into the target game in the other slot
	potentialOpponents := []string{}
	for _, game := range gamesMap {
		if game.NextGameID != nil && *game.NextGameID == targetGame.ID &&
			game.NextGameSlot != nil && *game.NextGameSlot == otherSlot {
			// Traverse down this branch to find all potential teams
			findTeamsInBranch(gamesMap, game, teamID, &potentialOpponents)
		}
	}

	return potentialOpponents
}

// findTeamsInBranch recursively finds all teams in a branch of the tournament tree
func findTeamsInBranch(gamesMap map[string]*TournamentGame, game *TournamentGame, excludeTeamID string, teams *[]string) {
	if game == nil {
		return
	}

	// Add teams from this game if they're not the team we're looking for
	if game.Team1ID != nil && *game.Team1ID != excludeTeamID {
		*teams = append(*teams, *game.Team1ID)
	}
	if game.Team2ID != nil && *game.Team2ID != excludeTeamID {
		*teams = append(*teams, *game.Team2ID)
	}

	// Find games that feed into this game
	for _, g := range gamesMap {
		if g.NextGameID != nil && *g.NextGameID == game.ID {
			findTeamsInBranch(gamesMap, g, excludeTeamID, teams)
		}
	}
}

// GetPotentialMatchupRound determines which round two teams would meet in the tournament.
// Returns -1 if the teams cannot meet (different regions or one/both not in tournament).
func GetPotentialMatchupRound(gamesMap map[string]*TournamentGame, team1ID, team2ID string, totalRounds int) int {
	// Find the games where each team is currently playing
	var team1Game, team2Game *TournamentGame
	for _, game := range gamesMap {
		if team1Game == nil && ((game.Team1ID != nil && *game.Team1ID == team1ID) || (game.Team2ID != nil && *game.Team2ID == team1ID)) {
			team1Game = game
		}
		if team2Game == nil && ((game.Team1ID != nil && *game.Team1ID == team2ID) || (game.Team2ID != nil && *game.Team2ID == team2ID)) {
			team2Game = game
		}
		if team1Game != nil && team2Game != nil {
			break
		}
	}

	// If either team is not in the tournament, return -1
	if team1Game == nil || team2Game == nil {
		return -1
	}

	// If the teams are already playing each other, return the current round
	if team1Game.ID == team2Game.ID {
		return team1Game.GetRound(gamesMap, totalRounds)
	}

	// If both teams feed into the same next game, they'll meet in that game
	if team1Game.NextGameID != nil && team2Game.NextGameID != nil && *team1Game.NextGameID == *team2Game.NextGameID {
		nextGame := gamesMap[*team1Game.NextGameID]
		return nextGame.GetRound(gamesMap, totalRounds)
	}

	// Find the earliest possible meeting point by traversing up the bracket
	meetingPoint := findEarliestMeetingPoint(gamesMap, team1Game, team2Game)
	if meetingPoint == nil {
		return -1 // Teams cannot meet (different regions)
	}

	return meetingPoint.GetRound(gamesMap, totalRounds)
}

// findEarliestMeetingPoint finds the earliest game where two teams could potentially meet
func findEarliestMeetingPoint(gamesMap map[string]*TournamentGame, game1, game2 *TournamentGame) *TournamentGame {
	// Keep track of games we've visited from each path
	visited1 := make(map[string]bool)
	visited2 := make(map[string]bool)

	// Start with both games and traverse up until we find a common ancestor
	current1 := game1
	current2 := game2

	for current1 != nil && current2 != nil {
		// Mark current games as visited
		visited1[current1.ID] = true
		visited2[current2.ID] = true

		// Check if either current game is in the other's path
		if visited2[current1.ID] {
			return current1
		}
		if visited1[current2.ID] {
			return current2
		}

		// Move up the bracket
		if current1.NextGameID != nil {
			next1, exists := gamesMap[*current1.NextGameID]
			if exists {
				current1 = next1
			} else {
				current1 = nil
			}
		} else {
			current1 = nil
		}

		if current2.NextGameID != nil {
			next2, exists := gamesMap[*current2.NextGameID]
			if exists {
				current2 = next2
			} else {
				current2 = nil
			}
		} else {
			current2 = nil
		}
	}

	return nil // No meeting point found
}
