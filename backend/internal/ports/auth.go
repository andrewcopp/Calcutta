package ports

import (
	"context"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// AuthIdentity represents a successfully authenticated user.
type AuthIdentity struct {
	UserID    string
	SessionID string // empty for non-session auth (Cognito, API key, dev)
}

// Authenticator validates a bearer token and returns an identity.
// Convention: (nil, nil) = "not my token, try next"; (nil, err) = "recognized but failed, stop chain".
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*AuthIdentity, error)
}

// APIKeyReader looks up active API keys by hash.
type APIKeyReader interface {
	GetActiveByHash(ctx context.Context, keyHash string, now time.Time) (*models.APIKey, error)
}

// AuthSessionRepository manages authentication sessions.
type AuthSessionRepository interface {
	CreateSession(ctx context.Context, userID, refreshTokenHash, userAgent, ipAddress string, expiresAt time.Time) (string, error)
	GetSessionByID(ctx context.Context, id string) (*models.AuthSession, error)
	GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*models.AuthSession, error)
	RotateRefreshToken(ctx context.Context, sessionID, newRefreshTokenHash string, newExpiresAt time.Time) error
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllSessionsForUser(ctx context.Context, userID string) error
	IsUserActive(ctx context.Context, userID string) (bool, error)
}

// AuthorizationChecker verifies user permissions.
type AuthorizationChecker interface {
	HasPermission(ctx context.Context, userID, scopeType, scopeID, permissionKey string) (bool, error)
	GrantGlobalAdmin(ctx context.Context, userID string) error
}
