package ports

import (
	"context"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// AuthSessionRepository manages authentication sessions.
type AuthSessionRepository interface {
	CreateSession(ctx context.Context, userID, refreshTokenHash, userAgent, ipAddress string, expiresAt time.Time) (string, error)
	GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*models.AuthSession, error)
	RotateRefreshToken(ctx context.Context, sessionID, newRefreshTokenHash string, newExpiresAt time.Time) error
	RevokeSession(ctx context.Context, sessionID string) error
	IsUserActive(ctx context.Context, userID string) (bool, error)
}

// AuthorizationChecker verifies user permissions.
type AuthorizationChecker interface {
	HasPermission(ctx context.Context, userID, scopeType, scopeID, permissionKey string) (bool, error)
	GrantGlobalAdmin(ctx context.Context, userID string) error
}
