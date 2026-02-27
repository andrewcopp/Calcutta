package models

import "time"

// InvestmentSnapshot records the state of investments at the time of a change.
type InvestmentSnapshot struct {
	ID          string
	PortfolioID string
	ChangedBy   string
	Reason      string
	Investments []InvestmentSnapshotEntry
	CreatedAt   time.Time
}

// InvestmentSnapshotEntry is a single team+credits pair within a snapshot.
type InvestmentSnapshotEntry struct {
	TeamID  string `json:"teamId"`
	Credits int    `json:"credits"`
}
