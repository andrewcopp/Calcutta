package prediction

import (
	"math"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type teamRoundKey struct {
	teamID     string
	roundOrder int
}

// buildIncByRound computes the incremental points for reaching each progress level.
func buildIncByRound(rules []scoring.Rule) map[int]float64 {
	incByRound := make(map[int]float64)
	for r := 1; r <= models.MaxRounds; r++ {
		ptsR := float64(scoring.PointsForProgress(rules, r, 0))
		ptsRMinus1 := float64(scoring.PointsForProgress(rules, r-1, 0))
		incByRound[r] = ptsR - ptsRMinus1
	}
	return incByRound
}

// aggregatePWinByRound sums pMatchup*pWin by (teamID, roundOrder) from matchups.
func aggregatePWinByRound(matchups []PredictedMatchup) map[teamRoundKey]float64 {
	result := make(map[teamRoundKey]float64)
	for _, m := range matchups {
		result[teamRoundKey{m.Team1ID, m.RoundOrder}] += m.PMatchup * m.PTeam1WinsGivenMatchup
		result[teamRoundKey{m.Team2ID, m.RoundOrder}] += m.PMatchup * m.PTeam2WinsGivenMatchup
	}
	return result
}

// enforceMonotonicity ensures each probability is <= the previous one.
func enforceMonotonicity(probs []float64) {
	for i := 1; i < len(probs); i++ {
		if probs[i] > probs[i-1] {
			probs[i] = probs[i-1]
		}
	}
}

// GenerateTournamentValues computes expected points and round-by-round advancement
// probabilities for all teams. This handles both pre-tournament (throughRound=0)
// and mid-tournament checkpoint scenarios:
//   - Eliminated teams (progress < throughRound): pRound = 1.0 up to progress, 0 after
//   - Alive teams (progress >= throughRound): pRound = 1.0 up to throughRound, matchup probs after
//   - ExpectedPoints = actualPoints + sum(pRound[r] * incrementalPoints[r]) for future rounds
//
// PRound semantics (shifted): PRound[r] = P(reached progress level r).
//   - PRound1 = P(survived play-in): 1.0 for bye teams, <1.0 for FF teams
//   - PRound2..PRound7 map to matchup rounds 1..6 (R64 through Championship)
//
// pPlayinSurvival maps teamID -> P(survived play-in). If nil, all teams get 1.0.
func GenerateTournamentValues(
	allTeams []TeamInput,
	matchups []PredictedMatchup,
	throughRound int,
	rules []scoring.Rule,
	pPlayinSurvival map[string]float64,
) []PredictedTeamValue {
	pWinByRound := aggregatePWinByRound(matchups)
	incByRound := buildIncByRound(rules)

	favMap := ComputeFavoritesBracket(allTeams, matchups, throughRound, rules)

	var results []PredictedTeamValue
	for _, team := range allTeams {
		progress := team.Wins + team.Byes
		actualPoints := float64(scoring.PointsForProgress(rules, team.Wins, team.Byes))

		var probs [models.MaxRounds]float64
		var expectedPoints float64
		var variancePoints float64

		if throughRound > 0 && progress < throughRound {
			// Eliminated: pRound = 1.0 for rounds survived, 0.0 for rest.
			for r := 1; r <= models.MaxRounds; r++ {
				if r <= progress {
					probs[r-1] = 1.0
				}
			}
			expectedPoints = actualPoints
		} else {
			// Alive (or pre-tournament where all teams start at progress 0).
			if pPlayinSurvival != nil {
				// Shifted mapping (NCAA): PRound1 = play-in survival, PRound r maps to matchup round r-1.
				probs[0] = pPlayinSurvival[team.ID]
				for r := 2; r <= models.MaxRounds; r++ {
					matchupRound := r - 1
					if throughRound > 0 && r <= throughRound {
						probs[r-1] = 1.0
					} else {
						probs[r-1] = pWinByRound[teamRoundKey{team.ID, matchupRound}]
					}
				}
			} else {
				// Direct mapping: PRound r maps to matchup round r (no play-in shift).
				for r := 1; r <= models.MaxRounds; r++ {
					if throughRound > 0 && r < throughRound {
						probs[r-1] = 1.0
					} else {
						probs[r-1] = pWinByRound[teamRoundKey{team.ID, r}]
					}
				}
			}

			enforceMonotonicity(probs[:])

			// Expected points = actual + sum of future conditional points.
			// Start from max(throughRound+1, progress+1) so that rounds already
			// reflected in actualPoints are not counted again in the probability sum.
			expectedPoints = actualPoints
			startRound := throughRound + 1
			if progress >= startRound {
				startRound = progress + 1
			}
			for r := startRound; r <= models.MaxRounds; r++ {
				expectedPoints += probs[r-1] * incByRound[r]
			}

			// Variance (pre-tournament only).
			if throughRound == 0 {
				for r := 1; r <= models.MaxRounds; r++ {
					p := probs[r-1]
					inc := incByRound[r]
					variancePoints += p * (1 - p) * inc * inc
				}
			}
		}

		results = append(results, PredictedTeamValue{
			TeamID:               team.ID,
			ExpectedPoints:       expectedPoints,
			VariancePoints:       variancePoints,
			StdPoints:            math.Sqrt(variancePoints),
			PRound1:              probs[0],
			PRound2:              probs[1],
			PRound3:              probs[2],
			PRound4:              probs[3],
			PRound5:              probs[4],
			PRound6:              probs[5],
			PRound7:              probs[6],
			FavoritesTotalPoints: favMap[team.ID],
		})
	}

	return results
}

// DefaultScoringRules returns the standard NCAA tournament scoring rules.
// WinIndex 1 = play-in survival (bye or First Four win) which awards 0 points.
// WinIndex 2-7 map to R64 through Championship wins.
func DefaultScoringRules() []scoring.Rule {
	return []scoring.Rule{
		{WinIndex: 1, PointsAwarded: 0},   // Play-in survival (bye or FF win)
		{WinIndex: 2, PointsAwarded: 10},  // Round of 64 win
		{WinIndex: 3, PointsAwarded: 20},  // Round of 32 win
		{WinIndex: 4, PointsAwarded: 40},  // Sweet 16 win
		{WinIndex: 5, PointsAwarded: 80},  // Elite 8 win
		{WinIndex: 6, PointsAwarded: 160}, // Final Four win
		{WinIndex: 7, PointsAwarded: 320}, // Championship win
	}
}
