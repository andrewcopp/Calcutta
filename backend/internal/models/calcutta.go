package models

import "time"

// Calcutta represents an auction system for a tournament
type Calcutta struct {
	ID           string     `json:"id"`
	TournamentID string     `json:"tournamentId"`
	OwnerID      string     `json:"ownerId"`
	CreatedBy    string     `json:"createdBy"`
	Name         string     `json:"name"`
	MinTeams     int        `json:"minTeams"`
	MaxTeams     int        `json:"maxTeams"`
	MaxBidPoints int        `json:"maxBidPoints"`
	BudgetPoints int        `json:"budgetPoints"`
	Visibility   string     `json:"visibility"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"deletedAt,omitempty"`
}
