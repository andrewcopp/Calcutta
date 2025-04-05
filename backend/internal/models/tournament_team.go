package models

import (
	"errors"
	"time"
)

// TournamentTeam represents a team participating in a tournament
type TournamentTeam struct {
	ID           string     `json:"id"`
	SchoolID     string     `json:"schoolId"`
	TournamentID string     `json:"tournamentId"`
	Seed         int        `json:"seed"` // The team's seed in the tournament (1-16)
	Byes         int        `json:"byes"` // Number of byes the team received (0 = no byes, 1 = first round bye, etc.)
	Wins         int        `json:"wins"` // Number of wins in the tournament
	Created      time.Time  `json:"created"`
	Updated      time.Time  `json:"updated"`
	Deleted      *time.Time `json:"deleted,omitempty"`
}

// Validate checks if the TournamentTeam is valid
func (t *TournamentTeam) Validate() error {
	if t.Seed < 1 || t.Seed > 16 {
		return errors.New("seed must be between 1 and 16")
	}
	return nil
}
