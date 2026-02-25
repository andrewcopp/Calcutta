package prediction

import (
	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
)

// ComputeFavoritesBracket computes projected total points for each team by
// resolving every remaining game in favor of the higher-probability team.
// This produces a single deterministic bracket outcome.
//
// Parameters:
//   - allTeams: all teams in the tournament (including eliminated)
//   - matchups: predicted matchups from GenerateMatchups (only future rounds)
//   - throughRound: the checkpoint round (rounds <= throughRound are resolved)
//   - rules: scoring rules for point calculation
//
// Returns a map of teamID -> total points under the favorites bracket.
func ComputeFavoritesBracket(
	allTeams []TeamInput,
	matchups []PredictedMatchup,
	throughRound int,
	rules []scoring.Rule,
) map[string]float64 {
	// Track how many additional wins each team gets in the favorites bracket.
	favoritesWins := make(map[string]int)

	// alive tracks which teams survive each round in the favorites bracket.
	alive := make(map[string]bool, len(allTeams))
	for _, t := range allTeams {
		if t.Wins+t.Byes >= throughRound {
			alive[t.ID] = true
		}
	}

	// Group matchups by round order.
	matchupsByRound := make(map[int][]PredictedMatchup)
	maxRound := 0
	for _, m := range matchups {
		matchupsByRound[m.RoundOrder] = append(matchupsByRound[m.RoundOrder], m)
		if m.RoundOrder > maxRound {
			maxRound = m.RoundOrder
		}
	}

	// Process each round in order. Start from round 1 so that all generated
	// matchups are processed; rounds before the checkpoint have no matchups
	// (the matchup generator skips them), so the loop body is a no-op.
	for round := 1; round <= maxRound; round++ {
		roundMatchups := matchupsByRound[round]

		// Group matchups by game ID.
		matchupsByGame := make(map[string][]PredictedMatchup)
		for _, m := range roundMatchups {
			matchupsByGame[m.GameID] = append(matchupsByGame[m.GameID], m)
		}

		// For each game, find the matchup where both teams are alive.
		for _, gameMatchups := range matchupsByGame {
			var bestMatchup *PredictedMatchup
			for i := range gameMatchups {
				m := &gameMatchups[i]
				if alive[m.Team1ID] && alive[m.Team2ID] {
					if bestMatchup == nil || m.PMatchup > bestMatchup.PMatchup {
						bestMatchup = m
					}
				}
			}
			if bestMatchup == nil {
				continue
			}

			// Deterministic tiebreak: Team1 wins if P >= 0.5.
			var winnerID, loserID string
			if bestMatchup.PTeam1WinsGivenMatchup >= 0.5 {
				winnerID = bestMatchup.Team1ID
				loserID = bestMatchup.Team2ID
			} else {
				winnerID = bestMatchup.Team2ID
				loserID = bestMatchup.Team1ID
			}

			favoritesWins[winnerID]++
			delete(alive, loserID)
		}
	}

	// Compute total points for each team.
	result := make(map[string]float64, len(allTeams))
	for _, t := range allTeams {
		totalWins := t.Wins + favoritesWins[t.ID]
		totalProgress := totalWins + t.Byes
		result[t.ID] = float64(scoring.PointsForProgress(rules, totalProgress, 0))
	}

	return result
}
