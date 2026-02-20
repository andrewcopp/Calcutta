package models

import "time"

// Calcutta represents an auction system for a tournament
type Calcutta struct {
	ID              string     `json:"id"`
	TournamentID    string     `json:"tournamentId"`
	OwnerID         string     `json:"ownerId"`
	Name            string     `json:"name"`
	MinTeams        int        `json:"minTeams"`
	MaxTeams        int        `json:"maxTeams"`
	MaxBid          int        `json:"maxBid"`
	BiddingOpen     bool       `json:"biddingOpen"`
	BiddingLockedAt *time.Time `json:"biddingLockedAt,omitempty"`
	Visibility      string     `json:"visibility"`
	Created         time.Time  `json:"created"`
	Updated         time.Time  `json:"updated"`
	Deleted         *time.Time `json:"deleted,omitempty"`
}
