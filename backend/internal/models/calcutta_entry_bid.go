package models

// CalcuttaEntryBid represents a bid placed by a user on a team
type CalcuttaEntryBid struct {
	ID      string `json:"id"`
	EntryID string `json:"entryId"` // References CalcuttaEntry
	TeamID  string `json:"teamId"`
	Amount  int    `json:"amount"`
}
