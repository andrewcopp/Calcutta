package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// BracketService manages tournament bracket operations
type BracketService struct {
	tournamentRepo *TournamentRepository
	builder        *BracketBuilder
	validator      *models.BracketValidator
}

// NewBracketService creates a new bracket service
func NewBracketService(tournamentRepo *TournamentRepository) *BracketService {
	return &BracketService{
		tournamentRepo: tournamentRepo,
		builder:        NewBracketBuilder(),
		validator:      models.NewBracketValidator(),
	}
}

// GetBracket retrieves the current bracket state for a tournament
func (s *BracketService) GetBracket(ctx context.Context, tournamentID string) (*models.BracketStructure, error) {
	// Get tournament
	tournament, err := s.tournamentRepo.GetByID(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	if tournament == nil {
		return nil, fmt.Errorf("tournament not found")
	}

	// Get all teams
	teams, err := s.tournamentRepo.GetTeams(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}

	if len(teams) == 0 {
		return nil, fmt.Errorf("tournament has no teams")
	}

	// Build initial bracket structure with default Final Four config if not set
	finalFour := &models.FinalFourConfig{
		TopLeftRegion:     tournament.FinalFourTopLeft,
		BottomLeftRegion:  tournament.FinalFourBottomLeft,
		TopRightRegion:    tournament.FinalFourTopRight,
		BottomRightRegion: tournament.FinalFourBottomRight,
	}

	// Set defaults if empty
	if finalFour.TopLeftRegion == "" {
		finalFour.TopLeftRegion = "East"
	}
	if finalFour.BottomLeftRegion == "" {
		finalFour.BottomLeftRegion = "West"
	}
	if finalFour.TopRightRegion == "" {
		finalFour.TopRightRegion = "South"
	}
	if finalFour.BottomRightRegion == "" {
		finalFour.BottomRightRegion = "Midwest"
	}

	bracket, err := s.builder.BuildBracket(tournamentID, teams, finalFour)
	if err != nil {
		return nil, fmt.Errorf("failed to build bracket: %w", err)
	}

	// Apply current game results to bracket
	if err := s.applyCurrentResults(ctx, bracket, teams); err != nil {
		return nil, fmt.Errorf("failed to apply results: %w", err)
	}

	return bracket, nil
}

// SelectWinner selects a winner for a game and progresses the bracket
func (s *BracketService) SelectWinner(ctx context.Context, tournamentID, gameID, winnerTeamID string) (*models.BracketStructure, error) {
	// Get current bracket
	bracket, err := s.GetBracket(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	// Find the game
	game, exists := bracket.Games[gameID]
	if !exists {
		return nil, errors.New("game not found")
	}

	// Validate winner selection
	if err := s.validator.ValidateWinnerSelection(game, winnerTeamID); err != nil {
		return nil, err
	}

	// Set the winner
	if game.Team1 != nil && game.Team1.TeamID == winnerTeamID {
		game.Winner = game.Team1
	} else if game.Team2 != nil && game.Team2.TeamID == winnerTeamID {
		game.Winner = game.Team2
	}

	// Progress winner to next game
	if game.NextGameID != "" {
		nextGame, exists := bracket.Games[game.NextGameID]
		if !exists {
			return nil, fmt.Errorf("next game %s not found", game.NextGameID)
		}

		// Update LowestSeedSeen: winner inherits the minimum of both teams' LowestSeedSeen
		winnerCopy := *game.Winner
		if game.Team1 != nil && game.Team2 != nil {
			lowestInGame := game.Team1.LowestSeedSeen
			if game.Team2.LowestSeedSeen < lowestInGame {
				lowestInGame = game.Team2.LowestSeedSeen
			}
			winnerCopy.LowestSeedSeen = lowestInGame
		}

		if game.NextGameSlot == 1 {
			nextGame.Team1 = &winnerCopy
		} else if game.NextGameSlot == 2 {
			nextGame.Team2 = &winnerCopy
		}
	}

	// Validate bracket progression
	if err := s.validator.ValidateBracketProgression(bracket); err != nil {
		return nil, fmt.Errorf("bracket validation failed: %w", err)
	}

	// Calculate and update wins/byes for all teams
	if err := s.updateTeamStats(ctx, bracket); err != nil {
		return nil, fmt.Errorf("failed to update team stats: %w", err)
	}

	return bracket, nil
}

// UnselectWinner removes a winner selection and clears downstream games
func (s *BracketService) UnselectWinner(ctx context.Context, tournamentID, gameID string) (*models.BracketStructure, error) {
	// Get current bracket
	bracket, err := s.GetBracket(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	// Find the game
	game, exists := bracket.Games[gameID]
	if !exists {
		return nil, errors.New("game not found")
	}

	if game.Winner == nil {
		return bracket, nil // Nothing to unselect
	}

	// Clear winner and downstream games recursively
	s.clearDownstreamGames(bracket, gameID)

	// Calculate and update wins/byes for all teams
	if err := s.updateTeamStats(ctx, bracket); err != nil {
		return nil, fmt.Errorf("failed to update team stats: %w", err)
	}

	return bracket, nil
}

// clearDownstreamGames recursively clears winners and team assignments downstream
func (s *BracketService) clearDownstreamGames(bracket *models.BracketStructure, gameID string) {
	game, exists := bracket.Games[gameID]
	if !exists {
		return
	}

	// Clear the winner
	game.Winner = nil

	// If there's a next game, clear the team slot and recurse
	if game.NextGameID != "" {
		nextGame, exists := bracket.Games[game.NextGameID]
		if exists {
			if game.NextGameSlot == 1 {
				nextGame.Team1 = nil
			} else if game.NextGameSlot == 2 {
				nextGame.Team2 = nil
			}

			// If both teams are now cleared, clear this game too
			if nextGame.Team1 == nil && nextGame.Team2 == nil {
				s.clearDownstreamGames(bracket, game.NextGameID)
			}
		}
	}
}

// applyCurrentResults applies existing team results to the bracket
func (s *BracketService) applyCurrentResults(ctx context.Context, bracket *models.BracketStructure, teams []*models.TournamentTeam) error {
	// For each team, use their wins to determine which games they won
	// This reconstructs the bracket state from the team stats

	// This is a simplified version - in practice, you might want to store
	// game results separately or derive them from a more detailed state

	// For now, we'll just ensure the bracket structure is valid
	return s.validator.ValidateBracketProgression(bracket)
}

// updateTeamStats calculates and persists wins/byes/eliminated for all teams
func (s *BracketService) updateTeamStats(ctx context.Context, bracket *models.BracketStructure) error {
	// Get all teams
	teams, err := s.tournamentRepo.GetTeams(ctx, bracket.TournamentID)
	if err != nil {
		return fmt.Errorf("failed to get teams: %w", err)
	}

	// Calculate stats for each team
	for i, team := range teams {
		wins, byes, eliminated := models.CalculateWinsAndByes(team.ID, bracket)

		// Update team
		team.Wins = wins
		team.Byes = byes
		team.Eliminated = eliminated

		if err := s.tournamentRepo.UpdateTournamentTeam(ctx, team); err != nil {
			return fmt.Errorf("failed to update team %d/%d (ID: %s, Name: %s): %w", i+1, len(teams), team.ID, team.SchoolID, err)
		}
	}

	return nil
}

// ValidateBracketSetup validates that a tournament has the correct setup for bracket generation
func (s *BracketService) ValidateBracketSetup(ctx context.Context, tournamentID string) error {
	teams, err := s.tournamentRepo.GetTeams(ctx, tournamentID)
	if err != nil {
		return err
	}

	if len(teams) != 68 {
		return fmt.Errorf("tournament must have exactly 68 teams, has %d", len(teams))
	}

	// Validate regions - each region needs at least 16 teams
	regionCounts := make(map[string]int)
	for _, team := range teams {
		regionCounts[team.Region]++
	}

	expectedRegions := []string{"East", "West", "South", "Midwest"}
	for _, region := range expectedRegions {
		count := regionCounts[region]
		if count < 16 {
			return fmt.Errorf("region %s must have at least 16 teams, has %d", region, count)
		}
	}

	// Validate that each region has seeds 1-16 represented
	// Group teams by region and seed
	type regionSeed struct {
		region string
		seed   int
	}
	regionSeedCounts := make(map[regionSeed]int)

	for _, team := range teams {
		key := regionSeed{region: team.Region, seed: team.Seed}
		regionSeedCounts[key]++
	}

	// Check each region has all seeds 1-16
	for _, region := range expectedRegions {
		for seed := 1; seed <= 16; seed++ {
			key := regionSeed{region: region, seed: seed}
			count := regionSeedCounts[key]

			if count == 0 {
				return fmt.Errorf("region %s is missing seed %d", region, seed)
			}

			// If there are 2 teams with the same seed in the same region, that's a play-in game
			// This is valid. If there are more than 2, that's an error.
			if count > 2 {
				return fmt.Errorf("region %s has %d teams with seed %d (max 2 for play-in game)", region, count, seed)
			}
		}
	}

	return nil
}
