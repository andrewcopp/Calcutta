package prediction

import (
	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
)

// TeamProgress represents the current tournament progress of a team.
type TeamProgress struct {
	Wins         int
	Byes         int
	IsEliminated bool
}

// ProjectedTeamEV computes the projected expected value for a team given its
// predicted round-by-round probabilities, the scoring rules, and current progress.
//
// - Eliminated team: returns actual points earned so far
// - Pre-tournament (0 wins, 0 byes): returns the full predicted expected points
// - Alive mid-tournament: returns actual points + conditional expected remaining points
//
// When throughRound > 0, the batch was generated from a checkpoint and pRound values
// are already conditional on survival. The /pAlive division is skipped.
func ProjectedTeamEV(ptv PredictedTeamValue, rules []scoring.Rule, tp TeamProgress, throughRound int) float64 {
	actualPoints := float64(scoring.PointsForProgress(rules, tp.Wins, tp.Byes))

	if tp.IsEliminated {
		return actualPoints
	}

	progress := tp.Wins + tp.Byes
	if progress == 0 {
		return ptv.ExpectedPoints
	}

	pAlive := pRoundByIndex(ptv, progress)
	if pAlive <= 0 {
		return actualPoints
	}

	maxRound := maxRoundFromRules(rules)

	var conditionalRemaining float64
	for r := progress + 1; r <= maxRound; r++ {
		pReachRound := pRoundByIndex(ptv, r)
		incPoints := float64(scoring.PointsForProgress(rules, r, 0) - scoring.PointsForProgress(rules, r-1, 0))
		if throughRound > 0 {
			conditionalRemaining += pReachRound * incPoints
		} else {
			conditionalRemaining += (pReachRound / pAlive) * incPoints
		}
	}

	return actualPoints + conditionalRemaining
}

// pRoundByIndex returns the advancement probability for a given progress index.
// Progress 1 = survived round 1 (PRound1), progress 2 = survived round 2 (PRound2), etc.
func pRoundByIndex(ptv PredictedTeamValue, round int) float64 {
	switch round {
	case 1:
		return ptv.PRound1
	case 2:
		return ptv.PRound2
	case 3:
		return ptv.PRound3
	case 4:
		return ptv.PRound4
	case 5:
		return ptv.PRound5
	case 6:
		return ptv.PRound6
	case 7:
		return ptv.PRound7
	default:
		return 0
	}
}

// maxRoundFromRules returns the highest WinIndex present in the scoring rules.
func maxRoundFromRules(rules []scoring.Rule) int {
	max := 0
	for _, r := range rules {
		if r.WinIndex > max {
			max = r.WinIndex
		}
	}
	return max
}
