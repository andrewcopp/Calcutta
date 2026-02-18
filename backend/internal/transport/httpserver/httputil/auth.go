package httputil

import (
	"context"
	"errors"
)

// ErrNoAuthUser is returned when authentication is required but no user ID is present.
var ErrNoAuthUser = errors.New("authentication required")

// RequireAuthUserID extracts the authenticated user ID from the context using the
// provided extractor function. Returns ErrNoAuthUser if no user is authenticated.
func RequireAuthUserID(ctx context.Context, extractor func(context.Context) string) (string, error) {
	if extractor == nil {
		return "", ErrNoAuthUser
	}
	userID := extractor(ctx)
	if userID == "" {
		return "", ErrNoAuthUser
	}
	return userID, nil
}
