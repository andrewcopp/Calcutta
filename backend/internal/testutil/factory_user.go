package testutil

import "github.com/andrewcopp/Calcutta/backend/internal/models"

// NewUser returns a fully-populated User with sensible defaults.
func NewUser() *models.User {
	return &models.User{
		ID:        "user-1",
		Email:     StringPtr("test@example.com"),
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
		CreatedAt: DefaultTime,
		UpdatedAt: DefaultTime,
	}
}

// NewAuthSession returns a fully-populated AuthSession with sensible defaults.
func NewAuthSession() *models.AuthSession {
	return &models.AuthSession{
		ID:               "session-1",
		UserID:           "user-1",
		RefreshTokenHash: "hash",
		ExpiresAt:        DefaultTime.AddDate(0, 0, 7),
	}
}
