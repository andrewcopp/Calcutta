package bracket

import (
	"context"
	"fmt"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type Service struct {
	tournamentRepo *dbadapters.TournamentRepository
	builder        *BracketBuilder
	validator      *models.BracketValidator
}

func New(tournamentRepo *dbadapters.TournamentRepository) *Service {
	return &Service{
		tournamentRepo: tournamentRepo,
		builder:        NewBracketBuilder(),
		validator:      models.NewBracketValidator(),
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

	if err := s.validator.ValidateWinnerSelection(game, winnerTeamID); err != nil {
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
			losingTeam.Eliminated = true
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
			losingTeam.Eliminated = false
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

	return nil
}

func (s *Service) applyCurrentResults(ctx context.Context, bracket *models.BracketStructure, teams []*models.TournamentTeam) error {
	teamMap := make(map[string]*models.TournamentTeam)
	for _, team := range teams {
		teamMap[team.ID] = team
	}

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
		for _, game := range bracket.Games {
			if game.Round != round {
				continue
			}
			if game.Team1 == nil || game.Team2 == nil {
				continue
			}

			team1 := teamMap[game.Team1.TeamID]
			team2 := teamMap[game.Team2.TeamID]
			if team1 == nil || team2 == nil {
				continue
			}

			minRequired := s.getMinProgressForRound(round)
			team1Progress := team1.Wins + team1.Byes
			team2Progress := team2.Wins + team2.Byes

			var winner *models.BracketTeam
			if team1Progress > minRequired && team2Progress >= minRequired {
				winner = game.Team1
			} else if team2Progress > minRequired && team1Progress >= minRequired {
				winner = game.Team2
			}

			if game.NextGameID != "" {
				nextGame := bracket.Games[game.NextGameID]
				if nextGame != nil {
					if winner != nil {
						game.Winner = winner

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
						if game.NextGameSlot == 1 {
							nextGame.Team1 = nil
						} else if game.NextGameSlot == 2 {
							nextGame.Team2 = nil
						}
					}
				}
			} else if winner != nil {
				game.Winner = winner
			}
		}
	}

	return nil
}

func (s *Service) getMinProgressForRound(round models.BracketRound) int {
	switch round {
	case models.RoundFirstFour:
		return 0
	case models.RoundOf64:
		return 1
	case models.RoundOf32:
		return 2
	case models.RoundSweet16:
		return 3
	case models.RoundElite8:
		return 4
	case models.RoundFinalFour:
		return 5
	case models.RoundChampionship:
		return 6
	default:
		return 0
	}
}
