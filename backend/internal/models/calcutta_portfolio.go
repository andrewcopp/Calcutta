package models

import "time"

// CalcuttaPortfolio represents a user's portfolio of team investments in a Calcutta
type CalcuttaPortfolio struct {
	ID            string     `json:"id"`
	EntryID       string     `json:"entryId"` // References CalcuttaEntry
	MaximumPoints float64    `json:"maximumPoints"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	DeletedAt     *time.Time `json:"deletedAt,omitempty"`
}
