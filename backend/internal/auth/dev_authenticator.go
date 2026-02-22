package auth

import (
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

// DevAuthenticate checks the X-Dev-User header for dev-mode auth bypass.
// Returns (nil, nil) when the header is absent or the user is not active.
func DevAuthenticate(r *http.Request, users ports.UserRepository) (*ports.AuthIdentity, error) {
	userID := strings.TrimSpace(r.Header.Get("X-Dev-User"))
	if userID == "" {
		return nil, nil
	}

	user, err := users.GetByID(r.Context(), userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Status != "active" {
		return nil, nil
	}

	return &ports.AuthIdentity{UserID: userID}, nil
}
