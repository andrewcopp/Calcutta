package models

import "time"

// PredictionBatchSummary is a lightweight view of a prediction batch for listing.
type PredictionBatchSummary struct {
	ID                   string
	ProbabilitySourceKey string
	ThroughRound         int
	CreatedAt            time.Time
}

// PredictedTeamValue holds tournament outcome predictions for a single team,
// derived from KenPom-based simulations. These are NOT market predictions —
// they represent how a team is expected to perform in the tournament
// (advancement probabilities and expected scoring points).
type PredictedTeamValue struct {
	TeamID         string
	ExpectedPoints float64
	VariancePoints float64
	StdPoints      float64
	PRound1        float64
	PRound2        float64
	PRound3        float64
	PRound4        float64
	PRound5        float64
	PRound6        float64
	PRound7        float64
}

// MaxRounds is the total number of rounds in the NCAA tournament (including First Four).
const MaxRounds = 7

// PRoundByIndex returns the advancement probability for a given round index (1-7).
func (v PredictedTeamValue) PRoundByIndex(round int) float64 {
	switch round {
	case 1:
		return v.PRound1
	case 2:
		return v.PRound2
	case 3:
		return v.PRound3
	case 4:
		return v.PRound4
	case 5:
		return v.PRound5
	case 6:
		return v.PRound6
	case 7:
		return v.PRound7
	default:
		return 0
	}
}

// PredictionTeamInput represents the input data needed for a team to generate predictions.
type PredictionTeamInput struct {
	ID        string
	Seed      int
	Region    string
	KenPomNet float64
	Wins      int
	Byes      int
}
