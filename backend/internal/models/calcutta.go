package models

import "time"

// Default constraints for new calcuttas.
const (
	DefaultMinTeams     = 3
	DefaultMaxTeams     = 10
	DefaultMaxBidPoints = 50
)

// ApplyDefaults fills in zero-value constraint fields with sensible defaults.
func (c *Calcutta) ApplyDefaults() {
	if c.MinTeams == 0 {
		c.MinTeams = DefaultMinTeams
	}
	if c.MaxTeams == 0 {
		c.MaxTeams = DefaultMaxTeams
	}
	if c.MaxBidPoints == 0 {
		c.MaxBidPoints = DefaultMaxBidPoints
	}
}

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
