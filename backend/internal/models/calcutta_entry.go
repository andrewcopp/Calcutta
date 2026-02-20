package models

import "time"

// CalcuttaEntry represents a user's entry in a Calcutta tournament
type CalcuttaEntry struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`             // Name of the participant (required for historical data)
	UserID         *string    `json:"userId,omitempty"` // Optional reference to a User (null for historical data)
	CalcuttaID     string     `json:"calcuttaId"`
	Status         string     `json:"status"`
	TotalPoints    float64    `json:"totalPoints"`
	FinishPosition int        `json:"finishPosition"`
	InTheMoney     bool       `json:"inTheMoney"`
	PayoutCents    int        `json:"payoutCents"`
	IsTied         bool       `json:"isTied"`
	Created        time.Time  `json:"created"`
	Updated        time.Time  `json:"updated"`
	Deleted        *time.Time `json:"deleted,omitempty"`
}
