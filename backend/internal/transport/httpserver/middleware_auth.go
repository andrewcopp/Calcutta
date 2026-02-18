package httpserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
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
				slog.Warn("dev_auth_bypass", "user_id", userID, "header", "X-Dev-User")
				if s.pool == nil {
					ctx := context.WithValue(r.Context(), authUserIDKey, userID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				active, err := s.isUserActive(r.Context(), userID)
				if err != nil {
					slog.Error("auth_user_active_check_failed", "user_id", userID, "error", err)
					http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
					return
				}
				if active {
					ctx := context.WithValue(r.Context(), authUserIDKey, userID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
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

						created := &models.User{ID: id, Email: &email, FirstName: first, LastName: last, Status: "active"}
						if err := s.userRepo.Create(r.Context(), created); err != nil {
							slog.Error("cognito_auto_provision_failed", "email", email, "error", err)
							http.Error(w, "Internal Server Error", http.StatusInternalServerError)
							return
						}
						user = created
					}
					if err != nil {
						slog.Error("auth_cognito_user_lookup_failed", "email", email, "error", err)
						http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
						return
					}
					if user != nil {
						active, err := s.isUserActive(r.Context(), user.ID)
						if err != nil {
							slog.Error("auth_user_active_check_failed", "user_id", user.ID, "error", err)
							http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
							return
						}
						if active {
							ctx := context.WithValue(r.Context(), authUserIDKey, user.ID)
							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}
					}
				}

				if s.cfg.CognitoAllowUnprovisioned {
					active, err := s.isUserActive(r.Context(), claims.Sub)
					if err != nil {
						slog.Error("auth_user_active_check_failed", "user_id", claims.Sub, "error", err)
						http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
						return
					}
					if active {
						ctx := context.WithValue(r.Context(), authUserIDKey, claims.Sub)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}
		}

		if s.tokenManager != nil {
			claims, err := s.tokenManager.VerifyAccessToken(tok, time.Now())
			if err == nil {
				sess, err := s.authRepo.GetSessionByID(r.Context(), claims.Sid)
				if err != nil {
					slog.Error("auth_session_lookup_failed", "session_id", claims.Sid, "error", err)
					http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
					return
				}
				if sess != nil && sess.RevokedAt == nil && !time.Now().After(sess.ExpiresAt) && sess.UserID == claims.Sub {
					active, err := s.isUserActive(r.Context(), claims.Sub)
					if err != nil {
						slog.Error("auth_user_active_check_failed", "user_id", claims.Sub, "error", err)
						http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
						return
					}
					if active {
						ctx := context.WithValue(r.Context(), authUserIDKey, claims.Sub)
						ctx = context.WithValue(ctx, authSessionIDKey, claims.Sid)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}
		}

		if s.apiKeysRepo != nil {
			sum := sha256.Sum256([]byte(tok))
			h := hex.EncodeToString(sum[:])
			k, err := s.apiKeysRepo.GetActiveByHash(r.Context(), h, time.Now().UTC())
			if err != nil {
				slog.Error("auth_api_key_lookup_failed", "error", err)
				http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
				return
			}
			if k != nil {
				active, err := s.isUserActive(r.Context(), k.UserID)
				if err != nil {
					slog.Error("auth_user_active_check_failed", "user_id", k.UserID, "error", err)
					http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
					return
				}
				if active {
					ctx := context.WithValue(r.Context(), authUserIDKey, k.UserID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
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

func (s *Server) isUserActive(ctx context.Context, userID string) (bool, error) {
	if s.pool == nil || strings.TrimSpace(userID) == "" {
		return false, nil
	}
	var status string
	err := s.pool.QueryRow(ctx, `
		SELECT status
		FROM core.users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return status == "active", nil
}

func (s *Server) requireAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authUserID(r.Context()) == "" {
			httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requirePermission(permissionKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := authUserID(r.Context())
		if userID == "" {
			httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
			return
		}

		ok, err := s.authzRepo.HasPermission(r.Context(), userID, "global", "", permissionKey)
		if err != nil {
			httperr.WriteFromErr(w, r, err, authUserID)
			return
		}
		if !ok {
			httperr.Write(w, r, http.StatusForbidden, "forbidden", "Insufficient permissions", "")
			return
		}

		next(w, r)
	}
}

// requirePermissionWithScope checks global grants first, then falls back to
// a tournament-scoped grant extracted from the URL path variable.
func (s *Server) requirePermissionWithScope(permissionKey, pathVar string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := authUserID(r.Context())
		if userID == "" {
			httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
			return
		}

		// Check global permission first.
		ok, err := s.authzRepo.HasPermission(r.Context(), userID, "global", "", permissionKey)
		if err != nil {
			httperr.WriteFromErr(w, r, err, authUserID)
			return
		}
		if ok {
			next(w, r)
			return
		}

		// Fall back to tournament-scoped permission.
		scopeID := mux.Vars(r)[pathVar]
		if scopeID != "" {
			ok, err = s.authzRepo.HasPermission(r.Context(), userID, "tournament", scopeID, permissionKey)
			if err != nil {
				httperr.WriteFromErr(w, r, err, authUserID)
				return
			}
			if ok {
				next(w, r)
				return
			}
		}

		httperr.Write(w, r, http.StatusForbidden, "forbidden", "Insufficient permissions", "")
	}
}
