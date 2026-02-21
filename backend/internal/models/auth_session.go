package models

import "time"

// AuthSession represents a user's authentication session.
type AuthSession struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
}
