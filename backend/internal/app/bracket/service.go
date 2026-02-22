package bracket

import (
	"context"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type TournamentRepo interface {
	GetByID(ctx context.Context, id string) (*models.Tournament, error)
	GetTeams(ctx context.Context, tournamentID string) ([]*models.TournamentTeam, error)
	GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error)
	UpdateTournamentTeam(ctx context.Context, team *models.TournamentTeam) error
}

type Service struct {
	tournamentRepo TournamentRepo
}

func New(tournamentRepo TournamentRepo) *Service {
	return &Service{
		tournamentRepo: tournamentRepo,
	}
}

func (s *Service) GetBracket(ctx context.Context, tournamentID string) (*models.BracketStructure, error) {
	tournament, err := s.tournamentRepo.GetByID(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}
	if tournament == nil {
		return nil, &apperrors.NotFoundError{Resource: "tournament", ID: tournamentID}
	}

	teams, err := s.tournamentRepo.GetTeams(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	if len(teams) == 0 {
		return nil, fmt.Errorf("tournament has no teams")
	}

	finalFour := &models.FinalFourConfig{
		TopLeftRegion:     tournament.FinalFourTopLeft,
		BottomLeftRegion:  tournament.FinalFourBottomLeft,
		TopRightRegion:    tournament.FinalFourTopRight,
		BottomRightRegion: tournament.FinalFourBottomRight,
	}
	if err := finalFour.ApplyDefaults(); err != nil {
		return nil, fmt.Errorf("failed to apply final four defaults: %w", err)
	}

	bracket, err := BuildBracketStructure(tournamentID, teams, finalFour)
	if err != nil {
		return nil, fmt.Errorf("failed to build bracket: %w", err)
	}

	if err := s.applyCurrentResults(ctx, bracket, teams); err != nil {
		return nil, fmt.Errorf("failed to apply results: %w", err)
	}

	return bracket, nil
}

func (s *Service) SelectWinner(ctx context.Context, tournamentID, gameID, winnerTeamID string) (*models.BracketStructure, error) {
	bracket, err := s.GetBracket(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	game, exists := bracket.Games[gameID]
	if !exists {
		return nil, &apperrors.NotFoundError{Resource: "game", ID: gameID}
	}

	if err := models.ValidateWinnerSelection(game, winnerTeamID); err != nil {
		return nil, err
	}

	winningTeam, err := s.tournamentRepo.GetTournamentTeam(ctx, winnerTeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	if winningTeam == nil {
		return nil, &apperrors.NotFoundError{Resource: "team", ID: winnerTeamID}
	}

	winningTeam.Wins++
	if err := s.tournamentRepo.UpdateTournamentTeam(ctx, winningTeam); err != nil {
		return nil, fmt.Errorf("failed to update team wins: %w", err)
	}

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
			losingTeam.IsEliminated = true
			if err := s.tournamentRepo.UpdateTournamentTeam(ctx, losingTeam); err != nil {
				return nil, fmt.Errorf("failed to mark losing team as eliminated: %w", err)
			}
		}
	}

	bracket, err = s.GetBracket(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to rebuild bracket: %w", err)
	}

	return bracket, nil
}

func (s *Service) UnselectWinner(ctx context.Context, tournamentID, gameID string) (*models.BracketStructure, error) {
	bracket, err := s.GetBracket(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	game, exists := bracket.Games[gameID]
	if !exists {
		return nil, &apperrors.NotFoundError{Resource: "game", ID: gameID}
	}

	if game.Winner == nil {
		return bracket, nil
	}

	winningTeam, err := s.tournamentRepo.GetTournamentTeam(ctx, game.Winner.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	if winningTeam == nil {
		return nil, &apperrors.NotFoundError{Resource: "team", ID: game.Winner.TeamID}
	}

	if winningTeam.Wins > 0 {
		winningTeam.Wins--
	}
	if err := s.tournamentRepo.UpdateTournamentTeam(ctx, winningTeam); err != nil {
		return nil, fmt.Errorf("failed to update team wins: %w", err)
	}

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
			losingTeam.IsEliminated = false
			if err := s.tournamentRepo.UpdateTournamentTeam(ctx, losingTeam); err != nil {
				return nil, fmt.Errorf("failed to reactivate losing team: %w", err)
			}
		}
	}

	bracket, err = s.GetBracket(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to rebuild bracket: %w", err)
	}

	return bracket, nil
}

func (s *Service) ValidateBracketSetup(ctx context.Context, tournamentID string) error {
	teams, err := s.tournamentRepo.GetTeams(ctx, tournamentID)
	if err != nil {
		return err
	}

	return ValidateBracketSetupTeams(teams)
}

// ValidateBracketSetupTeams validates that the teams slice represents a valid 68-team tournament bracket.
// This is a pure function that can be tested without mocking repositories.
func ValidateBracketSetupTeams(teams []*models.TournamentTeam) error {
	if len(teams) != 68 {
		return fmt.Errorf("tournament must have exactly 68 teams, has %d", len(teams))
	}

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

	type regionSeed struct {
		region string
		seed   int
	}
	regionSeedCounts := make(map[regionSeed]int)

	for _, team := range teams {
		key := regionSeed{region: team.Region, seed: team.Seed}
		regionSeedCounts[key]++
	}

	for _, region := range expectedRegions {
		for seed := 1; seed <= 16; seed++ {
			key := regionSeed{region: region, seed: seed}
			count := regionSeedCounts[key]

			if count == 0 {
				return fmt.Errorf("region %s is missing seed %d", region, seed)
			}
			if count > 2 {
				return fmt.Errorf("region %s has %d teams with seed %d (max 2 for play-in game)", region, count, seed)
			}
		}
	}

	for _, team := range teams {
		key := regionSeed{region: team.Region, seed: team.Seed}
		count := regionSeedCounts[key]
		if count == 2 {
			if team.Byes != 0 {
				return fmt.Errorf("play-in team must have byes=0 (region=%s seed=%d team_id=%s byes=%d)", team.Region, team.Seed, team.ID, team.Byes)
			}
		} else if count == 1 {
			if team.Byes != 1 {
				return fmt.Errorf("non-play-in team must have byes=1 (region=%s seed=%d team_id=%s byes=%d)", team.Region, team.Seed, team.ID, team.Byes)
			}
		}
	}

	return nil
}
