package prediction

// PredictedMatchup represents a potential game between two teams with probabilities.
type PredictedMatchup struct {
	GameID                 string
	RoundOrder             int
	Team1ID                string
	Team2ID                string
	PMatchup               float64 // Probability this matchup occurs
	PTeam1WinsGivenMatchup float64 // P(team1 wins | matchup occurs)
	PTeam2WinsGivenMatchup float64 // P(team2 wins | matchup occurs) = 1 - PTeam1WinsGivenMatchup
}

// PredictedTeamValue is the output of prediction calculations.
// Contains expected points and round-by-round advancement probabilities.
type PredictedTeamValue struct {
	TeamID         string
	ExpectedPoints float64
	VariancePoints float64
	StdPoints      float64
	PRound1        float64 // P(win first four or bye) - always 1.0 for non-first-four teams
	PRound2        float64 // P(reach round of 32)
	PRound3        float64 // P(reach sweet 16)
	PRound4        float64 // P(reach elite 8)
	PRound5        float64 // P(reach final four)
	PRound6        float64 // P(reach championship game)
	PRound7        float64 // P(win championship)
}

// TeamInput represents the input data needed for a team to generate predictions.
type TeamInput struct {
	ID        string
	Seed      int
	Region    string
	KenPomNet float64
}
