package models

import (
	"time"
)

// TournamentGame represents a game in a tournament
type TournamentGame struct {
	ID           string     `json:"id"`
	TournamentID string     `json:"tournamentId"`
	Team1ID      *string    `json:"team1Id,omitempty"` // Nullable reference to TournamentTeam
	Team2ID      *string    `json:"team2Id,omitempty"` // Nullable reference to TournamentTeam
	TipoffTime   time.Time  `json:"tipoffTime"`
	SortOrder    int        `json:"sortOrder"`              // Order in the bracket
	Team1Score   int        `json:"team1Score"`             // Score for team 1 (default 0)
	Team2Score   int        `json:"team2Score"`             // Score for team 2 (default 0)
	NextGameID   *string    `json:"nextGameId,omitempty"`   // Reference to the next game in the bracket
	NextGameSlot *int       `json:"nextGameSlot,omitempty"` // Slot in the next game (1 or 2), null for championship game
	IsFinal      bool       `json:"isFinal"`                // Whether the game is final
	Created      time.Time  `json:"created"`
	Updated      time.Time  `json:"updated"`
	Deleted      *time.Time `json:"deleted,omitempty"`
}
