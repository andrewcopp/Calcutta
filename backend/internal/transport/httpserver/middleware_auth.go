package httpserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/google/uuid"
)

type authContextKey string

const (
	authUserIDKey    authContextKey = "authUserID"
	authSessionIDKey authContextKey = "authSessionID"
)

func (s *Server) authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.AuthMode == "dev" {
			if userID := strings.TrimSpace(r.Header.Get("X-Dev-User")); userID != "" {
				ctx := context.WithValue(r.Context(), authUserIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		h := r.Header.Get("Authorization")
		if h == "" {
			next.ServeHTTP(w, r)
			return
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			next.ServeHTTP(w, r)
			return
		}
		tok := strings.TrimSpace(parts[1])
		if tok == "" {
			next.ServeHTTP(w, r)
			return
		}

		if s.cfg.AuthMode == "cognito" && s.cognitoJWT != nil {
			claims, err := s.cognitoJWT.Verify(tok, time.Now())
			if err == nil {
				email := strings.TrimSpace(claims.Email)
				if email != "" && s.userRepo != nil {
					user, err := s.userRepo.GetByEmail(r.Context(), email)
					if err == nil && user == nil && s.cfg.CognitoAutoProvision {
						id := strings.TrimSpace(claims.Sub)
						if _, err := uuid.Parse(id); err != nil {
							id = uuid.NewString()
						}

						first := strings.TrimSpace(claims.GivenName)
						last := strings.TrimSpace(claims.FamilyName)
						if first == "" {
							first = "User"
						}
						if last == "" {
							last = "User"
						}

						created := &models.User{ID: id, Email: email, FirstName: first, LastName: last}
						_ = s.userRepo.Create(r.Context(), created)
						user, _ = s.userRepo.GetByEmail(r.Context(), email)
					}
					if err == nil && user != nil {
						ctx := context.WithValue(r.Context(), authUserIDKey, user.ID)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}

				if s.cfg.CognitoAllowUnprovisioned {
					ctx := context.WithValue(r.Context(), authUserIDKey, claims.Sub)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		if s.tokenManager != nil {
			claims, err := s.tokenManager.VerifyAccessToken(tok, time.Now())
			if err == nil {
				sess, err := s.authRepo.GetSessionByID(r.Context(), claims.Sid)
				if err == nil && sess != nil && sess.RevokedAt == nil && !time.Now().After(sess.ExpiresAt) && sess.UserID == claims.Sub {
					ctx := context.WithValue(r.Context(), authUserIDKey, claims.Sub)
					ctx = context.WithValue(ctx, authSessionIDKey, claims.Sid)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		if s.apiKeysRepo != nil {
			sum := sha256.Sum256([]byte(tok))
			h := hex.EncodeToString(sum[:])
			k, err := s.apiKeysRepo.GetActiveByHash(r.Context(), h, time.Now().UTC())
			if err == nil && k != nil {
				ctx := context.WithValue(r.Context(), authUserIDKey, k.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func authUserID(ctx context.Context) string {
	if v, ok := ctx.Value(authUserIDKey).(string); ok {
		return v
	}
	return ""
}

func (s *Server) requireAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authUserID(r.Context()) == "" {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requirePermission(permissionKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := authUserID(r.Context())
		if userID == "" {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
			return
		}

		ok, err := s.authzRepo.HasPermission(r.Context(), userID, "global", "", permissionKey)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if !ok {
			writeError(w, r, http.StatusForbidden, "forbidden", "Insufficient permissions", "")
			return
		}

		next(w, r)
	}
}
