package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// TournamentTeamRepository defines the interface for tournament team data access
type TournamentTeamRepository interface {
	GetByID(ctx context.Context, id string) (*models.Tournament, error)
	GetTeams(ctx context.Context, tournamentID string) ([]*models.TournamentTeam, error)
	GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error)
	UpdateTournamentTeam(ctx context.Context, team *models.TournamentTeam) error
}

// BracketService manages tournament bracket operations
type BracketService struct {
	tournamentRepo TournamentTeamRepository
	builder        *BracketBuilder
	validator      *models.BracketValidator
}

// NewBracketService creates a new bracket service
func NewBracketService(tournamentRepo TournamentTeamRepository) *BracketService {
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

// SelectWinner selects a winner for a game by incrementing their wins and rebuilding the bracket
func (s *BracketService) SelectWinner(ctx context.Context, tournamentID, gameID, winnerTeamID string) (*models.BracketStructure, error) {
	// Get current bracket to validate the selection
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

	// Get the winning team from database
	winningTeam, err := s.tournamentRepo.GetTournamentTeam(ctx, winnerTeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	if winningTeam == nil {
		return nil, errors.New("team not found")
	}

	// Increment wins for winner
	winningTeam.Wins++
	if err := s.tournamentRepo.UpdateTournamentTeam(ctx, winningTeam); err != nil {
		return nil, fmt.Errorf("failed to update team wins: %w", err)
	}

	// Mark losing team as eliminated
	var losingTeamID string
	if game.Team1 != nil && game.Team1.TeamID != winnerTeamID {
		losingTeamID = game.Team1.TeamID
	} else if game.Team2 != nil && game.Team2.TeamID != winnerTeamID {
		losingTeamID = game.Team2.TeamID
	}

	if losingTeamID != "" {
		losingTeam, err := s.tournamentRepo.GetTournamentTeam(ctx, losingTeamID)
		if err != nil {
			return nil, fmt.Errorf("failed to get losing team: %w", err)
		}
		if losingTeam != nil {
			losingTeam.Eliminated = true
			if err := s.tournamentRepo.UpdateTournamentTeam(ctx, losingTeam); err != nil {
				return nil, fmt.Errorf("failed to mark losing team as eliminated: %w", err)
			}
		}
	}

	// Rebuild bracket from updated state
	bracket, err = s.GetBracket(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to rebuild bracket: %w", err)
	}

	return bracket, nil
}

// UnselectWinner removes a winner selection by decrementing their wins and rebuilding the bracket
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

	// Get the winning team from database
	winningTeam, err := s.tournamentRepo.GetTournamentTeam(ctx, game.Winner.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	if winningTeam == nil {
		return nil, errors.New("team not found")
	}

	// Decrement wins (but not below 0)
	if winningTeam.Wins > 0 {
		winningTeam.Wins--
	}
	if err := s.tournamentRepo.UpdateTournamentTeam(ctx, winningTeam); err != nil {
		return nil, fmt.Errorf("failed to update team wins: %w", err)
	}

	// Reactivate the losing team (they're no longer eliminated)
	var losingTeamID string
	if game.Team1 != nil && game.Team1.TeamID != game.Winner.TeamID {
		losingTeamID = game.Team1.TeamID
	} else if game.Team2 != nil && game.Team2.TeamID != game.Winner.TeamID {
		losingTeamID = game.Team2.TeamID
	}

	if losingTeamID != "" {
		losingTeam, err := s.tournamentRepo.GetTournamentTeam(ctx, losingTeamID)
		if err != nil {
			return nil, fmt.Errorf("failed to get losing team: %w", err)
		}
		if losingTeam != nil {
			losingTeam.Eliminated = false
			if err := s.tournamentRepo.UpdateTournamentTeam(ctx, losingTeam); err != nil {
				return nil, fmt.Errorf("failed to reactivate losing team: %w", err)
			}
		}
	}

	// Rebuild bracket from updated state
	bracket, err = s.GetBracket(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to rebuild bracket: %w", err)
	}

	return bracket, nil
}

// applyCurrentResults reconstructs the bracket state from stored team wins
func (s *BracketService) applyCurrentResults(ctx context.Context, bracket *models.BracketStructure, teams []*models.TournamentTeam) error {
	// Create a map of team ID to team for quick lookup
	teamMap := make(map[string]*models.TournamentTeam)
	for _, team := range teams {
		teamMap[team.ID] = team
	}

	// Process games in order: First Four -> Round of 64 -> Round of 32 -> Sweet 16 -> Elite 8 -> Final Four -> Championship
	rounds := []models.BracketRound{
		models.RoundFirstFour,
		models.RoundOf64,
		models.RoundOf32,
		models.RoundSweet16,
		models.RoundElite8,
		models.RoundFinalFour,
		models.RoundChampionship,
	}

	for _, round := range rounds {
		// Get all games for this round
		for _, game := range bracket.Games {
			if game.Round != round {
				continue
			}

			// Skip if game doesn't have both teams yet
			if game.Team1 == nil || game.Team2 == nil {
				continue
			}

			// Get team data from database
			team1 := teamMap[game.Team1.TeamID]
			team2 := teamMap[game.Team2.TeamID]

			if team1 == nil || team2 == nil {
				continue
			}

			// Determine winner based on minimum required progress for this round
			// Each round requires a minimum (wins + byes) to participate
			// To WIN this game, a team must have MORE than the minimum (they advanced past it)
			minRequired := s.getMinProgressForRound(round)
			team1Progress := team1.Wins + team1.Byes
			team2Progress := team2.Wins + team2.Byes

			var winner *models.BracketTeam
			// Check if one team has progressed beyond this round (won the game)
			if team1Progress > minRequired && team2Progress >= minRequired {
				// team1 won this game and advanced
				winner = game.Team1
			} else if team2Progress > minRequired && team1Progress >= minRequired {
				// team2 won this game and advanced
				winner = game.Team2
			}
			// If both are exactly at minRequired, they just arrived - game not played yet
			// If either hasn't reached minRequired, they're not in this game yet

			// Handle next game progression
			if game.NextGameID != "" {
				nextGame := bracket.Games[game.NextGameID]
				if nextGame != nil {
					if winner != nil {
						// Set winner in this game
						game.Winner = winner

						// Progress winner to next game
						winnerCopy := *winner
						lowestInGame := game.Team1.LowestSeedSeen
						if game.Team2.LowestSeedSeen < lowestInGame {
							lowestInGame = game.Team2.LowestSeedSeen
						}
						winnerCopy.LowestSeedSeen = lowestInGame

						if game.NextGameSlot == 1 {
							nextGame.Team1 = &winnerCopy
						} else if game.NextGameSlot == 2 {
							nextGame.Team2 = &winnerCopy
						}
					} else {
						// No winner yet - clear the slot in next game
						if game.NextGameSlot == 1 {
							nextGame.Team1 = nil
						} else if game.NextGameSlot == 2 {
							nextGame.Team2 = nil
						}
					}
				}
			} else if winner != nil {
				// No next game but there is a winner (e.g., championship)
				game.Winner = winner
			}
		}
	}

	return nil
}

// getMinProgressForRound returns the minimum (wins + byes) required to participate in a round
func (s *BracketService) getMinProgressForRound(round models.BracketRound) int {
	switch round {
	case models.RoundFirstFour:
		return 0 // First Four is the starting point for play-in teams
	case models.RoundOf64:
		return 1 // Need 1 (either 1 bye or 1 win from First Four)
	case models.RoundOf32:
		return 2 // Need 2 (1 bye + 1 win from Round of 64, or 2 wins from First Four + Round of 64)
	case models.RoundSweet16:
		return 3 // Need 3 total progress
	case models.RoundElite8:
		return 4 // Need 4 total progress
	case models.RoundFinalFour:
		return 5 // Need 5 total progress
	case models.RoundChampionship:
		return 6 // Need 6 total progress
	default:
		return 0
	}
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
