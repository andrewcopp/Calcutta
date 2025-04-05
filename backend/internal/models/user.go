package models

import "time"

// User represents a participant in the Calcutta system
type User struct {
	ID      string     `json:"id"`
	Email   string     `json:"email"`
	Created time.Time  `json:"created"`
	Updated time.Time  `json:"updated"`
	Deleted *time.Time `json:"deleted,omitempty"`
}
