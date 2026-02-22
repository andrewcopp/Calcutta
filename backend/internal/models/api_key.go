package models

import "time"

// APIKey represents an API key for programmatic access.
type APIKey struct {
	ID         string
	UserID     string
	Label      *string
	CreatedAt  time.Time
	RevokedAt  *time.Time
	LastUsedAt *time.Time
}
