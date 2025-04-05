package models

// Calcutta represents an auction system for a tournament
type Calcutta struct {
	ID           string `json:"id"`
	TournamentID string `json:"tournamentId"`
	OwnerID      string `json:"ownerId"`
	Name         string `json:"name"`
}
