package models

import "time"

// Portfolio represents a user's portfolio in a pool
type Portfolio struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	UserID    *string    `json:"userId,omitempty"`
	PoolID    string     `json:"poolId"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}
