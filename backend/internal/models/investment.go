package models

import "time"

// Investment represents a team investment within a portfolio
type Investment struct {
	ID          string          `json:"id"`
	PortfolioID string          `json:"portfolioId"`
	TeamID      string          `json:"teamId"`
	Credits     int             `json:"credits"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   *time.Time      `json:"deletedAt,omitempty"`
	Team        *TournamentTeam `json:"team,omitempty"`
}
