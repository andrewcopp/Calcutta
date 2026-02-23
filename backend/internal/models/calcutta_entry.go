package models

import "time"

// CalcuttaEntry represents a user's entry in a Calcutta tournament
type CalcuttaEntry struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`             // Name of the participant (required for historical data)
	UserID     *string    `json:"userId,omitempty"` // Optional reference to a User (null for historical data)
	CalcuttaID string     `json:"calcuttaId"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt,omitempty"`
}
