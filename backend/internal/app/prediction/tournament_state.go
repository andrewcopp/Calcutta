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

// capTeamProgress caps a team's progress (Wins + Byes) to throughRound.
// Byes are preserved first (up to the cap), then Wins fill the remainder.
func capTeamProgress(team TeamInput, throughRound int) TeamInput {
	progress := team.Wins + team.Byes
	if progress <= throughRound {
		return team
	}
	capped := team
	capped.Byes = team.Byes
	if capped.Byes > throughRound {
		capped.Byes = throughRound
	}
	capped.Wins = throughRound - capped.Byes
	return capped
}

// NewTournamentState creates a checkpoint snapshot by capping team progress to throughRound
// and partitioning teams into survivors (wins + byes >= throughRound) and the full roster.
func NewTournamentState(data *TournamentData, throughRound int) *TournamentState {
	cappedTeams := make([]TeamInput, len(data.Teams))
	for i, t := range data.Teams {
		cappedTeams[i] = capTeamProgress(t, throughRound)
	}

	var survivors []TeamInput
	for _, t := range cappedTeams {
		if t.Wins+t.Byes >= throughRound {
			survivors = append(survivors, t)
		}
	}

	return &TournamentState{
		ThroughRound: throughRound,
		AllTeams:     cappedTeams,
		Survivors:    survivors,
		Rules:        data.Rules,
		FFConfig:     data.FFConfig,
	}
}
