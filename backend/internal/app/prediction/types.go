package prediction

import "github.com/andrewcopp/Calcutta/backend/internal/models"

// Type aliases preserve existing import paths (prediction.PredictedTeamValue, etc.)
// while the canonical definitions live in models.
type PredictionBatch = models.PredictionBatch
type PredictedTeamValue = models.PredictedTeamValue
type TeamInput = models.PredictionTeamInput

// PredictedMatchup represents a potential game between two teams with probabilities.
// This is internal to the prediction computation and not persisted.
type PredictedMatchup struct {
	GameID                 string
	RoundOrder             int
	Team1ID                string
	Team2ID                string
	PMatchup               float64 // Probability this matchup occurs
	PTeam1WinsGivenMatchup float64 // P(team1 wins | matchup occurs)
	PTeam2WinsGivenMatchup float64 // P(team2 wins | matchup occurs) = 1 - PTeam1WinsGivenMatchup
}
