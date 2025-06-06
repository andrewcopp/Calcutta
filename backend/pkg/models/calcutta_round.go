package models

import "time"

// CalcuttaRound represents the points awarded for a specific round in a Calcutta
// Note: In the NCAA tournament, the First Four is round 1, and the First Round is round 2
type CalcuttaRound struct {
	ID         string     `json:"id"`
	CalcuttaID string     `json:"calcuttaId"`
	Round      int        `json:"round"` // 1 = First Four, 2 = First Round, 3 = Sweet 16, etc.
	Points     int        `json:"points"`
	Created    time.Time  `json:"created"`
	Updated    time.Time  `json:"updated"`
	Deleted    *time.Time `json:"deleted,omitempty"`
}
