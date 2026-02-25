package prediction

import (
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/app/winprob"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// generatePredictions is a pure computation pipeline: matchups -> tournament values.
// No context, no repo, no side effects.
func generatePredictions(state *TournamentState, spec *winprob.Model) ([]PredictedTeamValue, error) {
	var matchups []PredictedMatchup
	if state.ThroughRound < models.MaxRounds {
		var err error
		matchups, err = GenerateMatchups(state.Survivors, state.ThroughRound, spec, state.FFConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to generate matchups: %w", err)
		}
	}

	teamValues := GenerateTournamentValues(state.AllTeams, matchups, state.ThroughRound, state.Rules)
	return teamValues, nil
}
