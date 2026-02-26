package models

import "time"

// OwnershipDetail represents a team investment within a portfolio's ownership
type OwnershipDetail struct {
	ID                  string          `json:"id"`
	PortfolioID         string          `json:"portfolioId"`
	TeamID              string          `json:"teamId"`
	OwnershipPercentage float64         `json:"ownershipPercentage"`
	ActualReturns       float64         `json:"actualReturns"`
	ExpectedReturns     float64         `json:"expectedReturns"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
	DeletedAt           *time.Time      `json:"deletedAt,omitempty"`
	Team                *TournamentTeam `json:"team,omitempty"`
}
