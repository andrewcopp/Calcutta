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
	Seed         int        `json:"seed"`       // The team's seed in the tournament (1-16)
	Region       string     `json:"region"`     // The team's region in the tournament
	Byes         int        `json:"byes"`       // Number of byes the team received (0 = no byes, 1 = first round bye, etc.)
	Wins         int        `json:"wins"`       // Number of wins in the tournament
	Eliminated   bool       `json:"eliminated"` // Whether the team has been eliminated from the tournament
	Created      time.Time  `json:"created"`
	Updated      time.Time  `json:"updated"`
	Deleted      *time.Time `json:"deleted,omitempty"`
}

// TournamentConfig holds configuration for tournament rules and validation
type TournamentConfig struct {
	MinSeed int
	MaxSeed int
	MinByes int
	MaxByes int
	MinWins int
	MaxWins int
}

// DefaultTournamentConfig returns the default configuration for tournament rules
func DefaultTournamentConfig() *TournamentConfig {
	return &TournamentConfig{
		MinSeed: 1,
		MaxSeed: 16,
		MinByes: 0,
		MaxByes: 1,
		MinWins: 0,
		MaxWins: 7,
	}
}

// Validate checks if the TournamentTeam is valid using the provided tournament configuration
func (t *TournamentTeam) Validate(config *TournamentConfig) error {
	if config == nil {
		config = DefaultTournamentConfig()
	}

	if t.Seed < config.MinSeed || t.Seed > config.MaxSeed {
		return errors.New("seed must be between 1 and 16")
	}

	if t.Byes < config.MinByes || t.Byes > config.MaxByes {
		return errors.New("byes must be between 0 and 1")
	}

	if t.Wins < config.MinWins || t.Wins > config.MaxWins {
		return errors.New("wins must be between 0 and 7")
	}

	return nil
}

// ValidateDefault checks if the TournamentTeam is valid using default tournament configuration
func (t *TournamentTeam) ValidateDefault() error {
	return t.Validate(DefaultTournamentConfig())
}
