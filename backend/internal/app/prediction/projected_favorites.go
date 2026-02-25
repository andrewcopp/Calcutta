package prediction

import (
	"strings"
)

// computeFavoritesFutureWins resolves every remaining game in favor of the
// higher-probability team and returns the number of additional wins each team
// earns beyond their current progress. This is purely about future matchup
// resolution — scoring is the caller's responsibility.
//
// Parameters:
//   - allTeams: all teams in the tournament (including eliminated)
//   - matchups: predicted matchups from GenerateMatchups (only future rounds)
//   - throughRound: the checkpoint round (rounds <= throughRound are resolved)
//
// Returns a map of teamID -> future wins from the favorites bracket.
func computeFavoritesFutureWins(
	allTeams []TeamInput,
	matchups []PredictedMatchup,
	throughRound int,
) map[string]int {
	// Track how many additional wins each team gets in the favorites bracket.
	favoritesWins := make(map[string]int)

	// alive tracks which teams survive each round in the favorites bracket.
	alive := make(map[string]bool, len(allTeams))
	for _, t := range allTeams {
		if t.Wins+t.Byes >= throughRound {
			alive[t.ID] = true
		}
	}
	// Add BYE sentinels from matchups so R128 games can be resolved.
	for _, m := range matchups {
		if strings.HasPrefix(m.Team1ID, byePrefix) {
			alive[m.Team1ID] = true
		}
		if strings.HasPrefix(m.Team2ID, byePrefix) {
			alive[m.Team2ID] = true
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

	return favoritesWins
}
