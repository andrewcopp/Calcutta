package ports

import "github.com/andrewcopp/Calcutta/backend/internal/models"

// BracketBuilder builds bracket structures from tournament teams.
type BracketBuilder interface {
	BuildBracket(tournamentID string, teams []*models.TournamentTeam, finalFour *models.FinalFourConfig) (*models.BracketStructure, error)
}
