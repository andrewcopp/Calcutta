package models

import "time"

// School represents a college or university in the system
type School struct {
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}
