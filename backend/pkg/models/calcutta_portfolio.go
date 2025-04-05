package models

import "time"

// CalcuttaPortfolio represents a user's portfolio of team investments in a Calcutta
type CalcuttaPortfolio struct {
	ID      string     `json:"id"`
	EntryID string     `json:"entryId"` // References CalcuttaEntry
	Created time.Time  `json:"created"`
	Updated time.Time  `json:"updated"`
	Deleted *time.Time `json:"deleted,omitempty"`
}
