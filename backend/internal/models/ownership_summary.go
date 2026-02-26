package models

import "time"

// OwnershipSummary represents the ownership summary for a portfolio
type OwnershipSummary struct {
	ID             string     `json:"id"`
	PortfolioID    string     `json:"portfolioId"`
	MaximumReturns float64    `json:"maximumReturns"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"deletedAt,omitempty"`
}
