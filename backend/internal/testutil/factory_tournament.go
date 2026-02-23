package testutil

import "github.com/andrewcopp/Calcutta/backend/internal/models"

// NewTournament returns a fully-populated Tournament with sensible defaults.
func NewTournament() *models.Tournament {
	startingAt := DefaultTime.AddDate(0, 3, 0)
	return &models.Tournament{
		ID:                   "tournament-1",
		Name:                 "Test Tournament",
		Rounds:               7,
		FinalFourTopLeft:     "East",
		FinalFourBottomLeft:  "West",
		FinalFourTopRight:    "South",
		FinalFourBottomRight: "Midwest",
		StartingAt:           &startingAt,
		CreatedAt:            DefaultTime,
		UpdatedAt:            DefaultTime,
	}
}

// NewTournamentTeam returns a fully-populated TournamentTeam with sensible defaults.
func NewTournamentTeam() *models.TournamentTeam {
	return &models.TournamentTeam{
		ID:           "team-1",
		SchoolID:     "school-1",
		TournamentID: "tournament-1",
		Seed:         1,
		Region:       "East",
		CreatedAt:    DefaultTime,
		UpdatedAt:    DefaultTime,
		School:       NewSchool(),
	}
}

// NewSchool returns a fully-populated School with sensible defaults.
func NewSchool() *models.School {
	return &models.School{
		ID:        "school-1",
		Name:      "Test School",
		CreatedAt: DefaultTime,
		UpdatedAt: DefaultTime,
	}
}
