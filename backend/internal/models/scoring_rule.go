package models

import "time"

// ScoringRule represents the points awarded for a specific win index in a Calcutta.
// Note: In the NCAA tournament, win_index 0 is a bye, 1 is the First Four, 2 is the First Round, etc.
type ScoringRule struct {
	ID            string     `json:"id"`
	CalcuttaID    string     `json:"calcuttaId"`
	WinIndex      int        `json:"winIndex"`
	PointsAwarded int        `json:"pointsAwarded"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	DeletedAt     *time.Time `json:"deletedAt,omitempty"`
}
