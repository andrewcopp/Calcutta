package models

import "time"

// School represents a college or university in the system
type School struct {
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	Created time.Time  `json:"created"`
	Updated time.Time  `json:"updated"`
	Deleted *time.Time `json:"deleted,omitempty"`
}
