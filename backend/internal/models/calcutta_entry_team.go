package models

// CalcuttaEntryTeam represents a team selected by a user in their Calcutta entry
type CalcuttaEntryTeam struct {
	ID      string `json:"id"`
	EntryID string `json:"entryId"` // References CalcuttaEntry
	TeamID  string `json:"teamId"`  // References TournamentTeam
	Amount  int    `json:"amount"`
}
