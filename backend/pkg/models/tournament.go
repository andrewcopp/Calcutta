package models

import "time"

// Tournament represents a basketball tournament in the real world
type Tournament struct {
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	Rounds  int        `json:"rounds"` // Total number of rounds in the tournament
	Created time.Time  `json:"created"`
	Updated time.Time  `json:"updated"`
	Deleted *time.Time `json:"deleted,omitempty"`
}
