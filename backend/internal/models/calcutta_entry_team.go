package models

import "time"

// CalcuttaEntryTeam represents a team selected by a user in their Calcutta entry
type CalcuttaEntryTeam struct {
	ID      string          `json:"id"`
	EntryID string          `json:"entryId"` // References CalcuttaEntry
	TeamID  string          `json:"teamId"`  // References TournamentTeam
	BidPoints int             `json:"bidPoints"`
	Created time.Time       `json:"created"`
	Updated time.Time       `json:"updated"`
	Deleted *time.Time      `json:"deleted,omitempty"`
	Team    *TournamentTeam `json:"team,omitempty"`
}
