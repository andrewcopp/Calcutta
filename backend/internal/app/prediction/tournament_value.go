package prediction

import (
	"math"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
)

// GenerateTournamentValues computes expected points and round probabilities from matchups.
// This is a port of the Python tournament_value.py module.
func GenerateTournamentValues(matchups []PredictedMatchup, rules []scoring.Rule) []PredictedTeamValue {
	// Calculate win probabilities by round for each team
	// p_win = p_matchup * p_team_wins_given_matchup
	type winRecord struct {
		teamID     string
		roundOrder int
		pWin       float64
	}

	var wins []winRecord
	for _, m := range matchups {
		if m.RoundOrder < 1 || m.RoundOrder > 7 {
			continue
		}
		// Team 1 win probability
		wins = append(wins, winRecord{
			teamID:     m.Team1ID,
			roundOrder: m.RoundOrder,
			pWin:       m.PMatchup * m.PTeam1WinsGivenMatchup,
		})
		// Team 2 win probability
		wins = append(wins, winRecord{
			teamID:     m.Team2ID,
			roundOrder: m.RoundOrder,
			pWin:       m.PMatchup * m.PTeam2WinsGivenMatchup,
		})
	}

	// Group by (team_id, round_order) and sum p_win
	type teamRoundKey struct {
		teamID     string
		roundOrder int
	}
	pByRound := make(map[teamRoundKey]float64)
	teamIDs := make(map[string]bool)

	for _, w := range wins {
		key := teamRoundKey{teamID: w.teamID, roundOrder: w.roundOrder}
		pByRound[key] += w.pWin
		teamIDs[w.teamID] = true
	}

	// Calculate incremental points for each round
	// inc_by_round[r] = points_for_progress(r) - points_for_progress(r-1)
	incByRound := make(map[int]float64)
	for r := 1; r <= 7; r++ {
		// For predictions, we use wins=r, byes=0 to get cumulative points
		// Then subtract previous round's points
		ptsR := float64(scoring.PointsForProgress(rules, r, 0))
		ptsRMinus1 := float64(scoring.PointsForProgress(rules, r-1, 0))
		incByRound[r] = ptsR - ptsRMinus1
	}

	// Calculate expected points and variance per team
	type teamStats struct {
		expectedPoints float64
		variancePoints float64
		pByRound       map[int]float64
	}
	statsByTeam := make(map[string]*teamStats)

	for teamID := range teamIDs {
		statsByTeam[teamID] = &teamStats{
			pByRound: make(map[int]float64),
		}
	}

	for key, pWin := range pByRound {
		stats := statsByTeam[key.teamID]
		if stats == nil {
			continue
		}

		pointsInc := incByRound[key.roundOrder]
		stats.expectedPoints += pWin * pointsInc
		stats.variancePoints += pWin * (1 - pWin) * (pointsInc * pointsInc)
		stats.pByRound[key.roundOrder] = pWin
	}

	// Build output
	var results []PredictedTeamValue
	for teamID, stats := range statsByTeam {
		result := PredictedTeamValue{
			TeamID:         teamID,
			ExpectedPoints: stats.expectedPoints,
			VariancePoints: stats.variancePoints,
			StdPoints:      math.Sqrt(stats.variancePoints),
			PRound1:        stats.pByRound[1],
			PRound2:        stats.pByRound[2],
			PRound3:        stats.pByRound[3],
			PRound4:        stats.pByRound[4],
			PRound5:        stats.pByRound[5],
			PRound6:        stats.pByRound[6],
			PRound7:        stats.pByRound[7],
		}
		results = append(results, result)
	}

	return results
}

// DefaultScoringRules returns the standard NCAA tournament scoring rules.
// This matches the default used in the Python implementation.
func DefaultScoringRules() []scoring.Rule {
	return []scoring.Rule{
		{WinIndex: 1, PointsAwarded: 10},  // Round of 64 win
		{WinIndex: 2, PointsAwarded: 20},  // Round of 32 win
		{WinIndex: 3, PointsAwarded: 40},  // Sweet 16 win
		{WinIndex: 4, PointsAwarded: 80},  // Elite 8 win
		{WinIndex: 5, PointsAwarded: 160}, // Final Four win
		{WinIndex: 6, PointsAwarded: 320}, // Championship win
	}
}
