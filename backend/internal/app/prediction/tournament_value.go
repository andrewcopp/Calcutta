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
// probabilities for all teams. The computation is split into three phases:
//
//   - ACTUAL: deterministic points from team progress at this checkpoint. No model needed.
//   - PROJECTED: forward-looking probabilities and expected points for remaining rounds.
//   - COMBINE: ExpectedPoints = actualPoints + projectedEV,
//     FavoritesTotalPoints = PointsForProgress(progress + futureWins).
//
// PRound semantics (128-team symmetric model): PRound[r] maps directly to matchup round r.
//   - PRound1 = P(wins R128): 1.0 for bye teams, <1.0 for FF teams
//   - PRound2..PRound7 map to R64 through Championship
func GenerateTournamentValues(
	allTeams []TeamInput,
	matchups []PredictedMatchup,
	throughRound int,
	rules []scoring.Rule,
) []PredictedTeamValue {
	pWinByRound := aggregatePWinByRound(matchups)
	incByRound := buildIncByRound(rules)

	futureWins := computeFavoritesFutureWins(allTeams, matchups, throughRound)

	var results []PredictedTeamValue
	for _, team := range allTeams {
		progress := team.Wins + team.Byes
		isEliminated := throughRound > 0 && progress < throughRound

		// ── ACTUAL (what has happened) ──
		actualPoints := float64(scoring.PointsForProgress(rules, team.Wins, team.Byes))

		// ── PROJECTED (what we think will happen) ──
		var probs [models.MaxRounds]float64
		var projectedEV float64
		var variancePoints float64

		// Fill actual rounds with 1.0.
		for r := 1; r <= progress; r++ {
			probs[r-1] = 1.0
		}

		if !isEliminated {
			// Fill projected rounds from matchup model.
			for r := throughRound + 1; r <= models.MaxRounds; r++ {
				probs[r-1] = pWinByRound[teamRoundKey{team.ID, r}]
			}
			enforceMonotonicity(probs[:])

			// Projected EV = sum of future probability * incremental points.
			for r := progress + 1; r <= models.MaxRounds; r++ {
				projectedEV += probs[r-1] * incByRound[r]
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

		// ── COMBINE ──
		expectedPoints := actualPoints + projectedEV
		favoritesTotalPoints := float64(scoring.PointsForProgress(
			rules, progress+futureWins[team.ID], 0))

		results = append(results, PredictedTeamValue{
			TeamID:               team.ID,
			ActualPoints:         actualPoints,
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
			FavoritesTotalPoints: favoritesTotalPoints,
		})
	}

	return results
}

// DefaultScoringRules returns the standard NCAA tournament scoring rules.
// In the 128-team symmetric model:
// WinIndex 1 = R128 win (bye or First Four win) which awards 0 points.
// WinIndex 2-7 map to R64 through Championship wins.
func DefaultScoringRules() []scoring.Rule {
	return []scoring.Rule{
		{WinIndex: 1, PointsAwarded: 0},   // R128 win (bye or FF win)
		{WinIndex: 2, PointsAwarded: 10},  // Round of 64 win
		{WinIndex: 3, PointsAwarded: 20},  // Round of 32 win
		{WinIndex: 4, PointsAwarded: 40},  // Sweet 16 win
		{WinIndex: 5, PointsAwarded: 80},  // Elite 8 win
		{WinIndex: 6, PointsAwarded: 160}, // Final Four win
		{WinIndex: 7, PointsAwarded: 320}, // Championship win
	}
}
