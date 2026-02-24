package prediction

import (
	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// TournamentData holds raw database results loaded once and reused across checkpoints.
type TournamentData struct {
	Teams    []TeamInput
	Rules    []scoring.Rule
	FFConfig *models.FinalFourConfig
}

// TournamentState is a checkpoint-specific snapshot with survivors partitioned from eliminated teams.
type TournamentState struct {
	ThroughRound int
	AllTeams     []TeamInput
	Survivors    []TeamInput
	Rules        []scoring.Rule
	FFConfig     *models.FinalFourConfig
}

// NewTournamentState creates a checkpoint snapshot by partitioning teams into survivors
// (wins + byes >= throughRound) and the full roster.
func NewTournamentState(data *TournamentData, throughRound int) *TournamentState {
	var survivors []TeamInput
	for _, t := range data.Teams {
		if t.Wins+t.Byes >= throughRound {
			survivors = append(survivors, t)
		}
	}

	return &TournamentState{
		ThroughRound: throughRound,
		AllTeams:     data.Teams,
		Survivors:    survivors,
		Rules:        data.Rules,
		FFConfig:     data.FFConfig,
	}
}
