package models

// EntryBid represents a bid placed by a user on a team
type EntryBid struct {
	ID      string `json:"id"`
	EntryID string `json:"entryId"`
	TeamID  string `json:"teamId"`
	Amount  int    `json:"amount"`
}
