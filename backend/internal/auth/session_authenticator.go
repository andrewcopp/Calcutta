package auth

import (
	"context"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

var _ ports.Authenticator = (*SessionAuthenticator)(nil)

// SessionAuthenticator validates HS256 session access tokens.
type SessionAuthenticator struct {
	tm       *TokenManager
	sessions ports.AuthSessionRepository
	users    ports.UserRepository
	now      func() time.Time
}

func NewSessionAuthenticator(tm *TokenManager, sessions ports.AuthSessionRepository, users ports.UserRepository) *SessionAuthenticator {
	return &SessionAuthenticator{tm: tm, sessions: sessions, users: users, now: time.Now}
}

func (a *SessionAuthenticator) Authenticate(ctx context.Context, token string) (*ports.AuthIdentity, error) {
	now := a.now()

	claims, err := a.tm.VerifyAccessToken(token, now)
	if err != nil {
		return nil, nil
	}

	sess, err := a.sessions.GetSessionByID(ctx, claims.Sid)
	if err != nil {
		return nil, err
	}
	if sess == nil || sess.RevokedAt != nil || now.After(sess.ExpiresAt) || sess.UserID != claims.Sub {
		return nil, nil
	}

	user, err := a.users.GetByID(ctx, claims.Sub)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Status != "active" {
		return nil, nil
	}

	return &ports.AuthIdentity{UserID: claims.Sub, SessionID: claims.Sid}, nil
}
