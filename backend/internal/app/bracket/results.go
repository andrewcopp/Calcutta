package bracket

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

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

			minRequired := round.MinProgressRequired()
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
