package models

import "time"

// CalcuttaPortfolioTeam represents a team investment within a user's Calcutta portfolio
type CalcuttaPortfolioTeam struct {
	ID                  string          `json:"id"`
	PortfolioID         string          `json:"portfolioId"`
	TeamID              string          `json:"teamId"` // References TournamentTeam
	OwnershipPercentage float64         `json:"ownershipPercentage"`
	ActualPoints        float64         `json:"actualPoints"`
	ExpectedPoints      float64         `json:"expectedPoints"`
	Created             time.Time       `json:"created"`
	Updated             time.Time       `json:"updated"`
	Deleted             *time.Time      `json:"deleted,omitempty"`
	Team                *TournamentTeam `json:"team,omitempty"`
}
