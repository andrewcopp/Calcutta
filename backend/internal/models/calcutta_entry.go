package models

import "time"

// CalcuttaEntry represents a user's entry in a Calcutta tournament
type CalcuttaEntry struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	CalcuttaID string     `json:"calcuttaId"`
	Created    time.Time  `json:"created"`
	Updated    time.Time  `json:"updated"`
	Deleted    *time.Time `json:"deleted,omitempty"`
}
