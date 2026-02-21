package httpserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type authContextKey string

const (
	authUserIDKey    authContextKey = "authUserID"
	authSessionIDKey authContextKey = "authSessionID"
)

func (s *Server) authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if userID, handled := s.tryDevAuth(w, r); handled {
			if userID != "" {
				ctx := context.WithValue(r.Context(), authUserIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
			return
		}

		tok := extractBearerToken(r)
		if tok == "" {
			next.ServeHTTP(w, r)
			return
		}

		if userID, handled := s.tryCognitoAuth(w, r, tok); handled {
			if userID != "" {
				ctx := context.WithValue(r.Context(), authUserIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
			return
		}

		if userID, sessionID, handled := s.trySessionAuth(w, r, tok); handled {
			if userID != "" {
				ctx := context.WithValue(r.Context(), authUserIDKey, userID)
				ctx = context.WithValue(ctx, authSessionIDKey, sessionID)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
			return
		}

		if userID, handled := s.tryAPIKeyAuth(w, r, tok); handled {
			if userID != "" {
				ctx := context.WithValue(r.Context(), authUserIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// extractBearerToken parses the Authorization header and returns the bearer
// token, or an empty string when the header is missing or malformed.
func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	tok := strings.TrimSpace(parts[1])
	return tok
}

// tryDevAuth handles the dev/local auth bypass via the X-Dev-User header.
// Returns (userID, true) on success, ("", true) if an error was written, or
// ("", false) when this auth method does not apply.
func (s *Server) tryDevAuth(w http.ResponseWriter, r *http.Request) (string, bool) {
	if s.cfg.AuthMode != "dev" {
		return "", false
	}
	userID := strings.TrimSpace(r.Header.Get("X-Dev-User"))
	if userID == "" {
		return "", false
	}
	slog.Warn("dev_auth_bypass", "user_id", userID, "header", "X-Dev-User")
	if s.pool == nil {
		return userID, true
	}
	active, err := s.isUserActive(r.Context(), userID)
	if err != nil {
		slog.Error("auth_user_active_check_failed", "user_id", userID, "error", err)
		httperr.Write(w, r, http.StatusServiceUnavailable, "service_unavailable", "Service Unavailable", "")
		return "", true
	}
	if active {
		return userID, true
	}
	return "", false
}

// tryCognitoAuth handles AWS Cognito JWT token authentication.
// Returns (userID, true) on success, ("", true) if an error was written, or
// ("", false) when this auth method does not apply.
func (s *Server) tryCognitoAuth(w http.ResponseWriter, r *http.Request, tok string) (string, bool) {
	if s.cfg.AuthMode != "cognito" || s.cognitoJWT == nil {
		return "", false
	}
	claims, err := s.cognitoJWT.Verify(tok, time.Now())
	if err != nil {
		return "", false
	}

	cognitoSub := strings.TrimSpace(claims.Sub)
	if cognitoSub != "" && s.userRepo != nil {
		user, err := s.userRepo.GetByExternalProvider(r.Context(), "cognito", cognitoSub)
		if err != nil {
			slog.Error("auth_cognito_user_lookup_failed", "cognito_sub", cognitoSub, "error", err)
			httperr.Write(w, r, http.StatusServiceUnavailable, "service_unavailable", "Service Unavailable", "")
			return "", true
		}

		// Auto-provision if enabled and user not found
		if user == nil && s.cfg.CognitoAutoProvision {
			id := cognitoSub
			if _, err := uuid.Parse(id); err != nil {
				id = uuid.NewString()
			}

			email := strings.TrimSpace(claims.Email)
			var emailPtr *string
			if email != "" {
				emailPtr = &email
			}

			first := strings.TrimSpace(claims.GivenName)
			last := strings.TrimSpace(claims.FamilyName)
			if first == "" {
				first = "User"
			}
			if last == "" {
				last = "User"
			}

			provider := "cognito"
			created := &models.User{
				ID:                 id,
				Email:              emailPtr,
				FirstName:          first,
				LastName:           last,
				Status:             "active",
				ExternalProvider:   &provider,
				ExternalProviderID: &cognitoSub,
			}
			if err := s.userRepo.Create(r.Context(), created); err != nil {
				slog.Error("cognito_auto_provision_failed", "cognito_sub", cognitoSub, "error", err)
				httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "Internal Server Error", "")
				return "", true
			}
			user = created
		}

		if user != nil {
			active, err := s.isUserActive(r.Context(), user.ID)
			if err != nil {
				slog.Error("auth_user_active_check_failed", "user_id", user.ID, "error", err)
				httperr.Write(w, r, http.StatusServiceUnavailable, "service_unavailable", "Service Unavailable", "")
				return "", true
			}
			if active {
				return user.ID, true
			}
		}
	}

	if s.cfg.CognitoAllowUnprovisioned {
		active, err := s.isUserActive(r.Context(), claims.Sub)
		if err != nil {
			slog.Error("auth_user_active_check_failed", "user_id", claims.Sub, "error", err)
			httperr.Write(w, r, http.StatusServiceUnavailable, "service_unavailable", "Service Unavailable", "")
			return "", true
		}
		if active {
			return claims.Sub, true
		}
		httperr.Write(w, r, http.StatusForbidden, "forbidden", "Forbidden", "")
		return "", true
	}

	return "", false
}

// trySessionAuth handles session-based access token authentication.
// Returns (userID, sessionID, true) on success, ("", "", true) if an error
// was written, or ("", "", false) when this auth method does not apply.
func (s *Server) trySessionAuth(w http.ResponseWriter, r *http.Request, tok string) (string, string, bool) {
	if s.tokenManager == nil {
		return "", "", false
	}
	claims, err := s.tokenManager.VerifyAccessToken(tok, time.Now())
	if err != nil {
		return "", "", false
	}
	sess, err := s.authRepo.GetSessionByID(r.Context(), claims.Sid)
	if err != nil {
		slog.Error("auth_session_lookup_failed", "session_id", claims.Sid, "error", err)
		httperr.Write(w, r, http.StatusServiceUnavailable, "service_unavailable", "Service Unavailable", "")
		return "", "", true
	}
	if sess == nil || sess.RevokedAt != nil || time.Now().After(sess.ExpiresAt) || sess.UserID != claims.Sub {
		return "", "", false
	}
	active, err := s.isUserActive(r.Context(), claims.Sub)
	if err != nil {
		slog.Error("auth_user_active_check_failed", "user_id", claims.Sub, "error", err)
		httperr.Write(w, r, http.StatusServiceUnavailable, "service_unavailable", "Service Unavailable", "")
		return "", "", true
	}
	if active {
		return claims.Sub, claims.Sid, true
	}
	return "", "", false
}

// tryAPIKeyAuth handles API key authentication via SHA-256 hash lookup.
// Returns (userID, true) on success, ("", true) if an error was written, or
// ("", false) when this auth method does not apply.
func (s *Server) tryAPIKeyAuth(w http.ResponseWriter, r *http.Request, tok string) (string, bool) {
	if s.apiKeysRepo == nil {
		return "", false
	}
	sum := sha256.Sum256([]byte(tok))
	h := hex.EncodeToString(sum[:])
	k, err := s.apiKeysRepo.GetActiveByHash(r.Context(), h, time.Now().UTC())
	if err != nil {
		slog.Error("auth_api_key_lookup_failed", "error", err)
		httperr.Write(w, r, http.StatusServiceUnavailable, "service_unavailable", "Service Unavailable", "")
		return "", true
	}
	if k == nil {
		return "", false
	}
	active, err := s.isUserActive(r.Context(), k.UserID)
	if err != nil {
		slog.Error("auth_user_active_check_failed", "user_id", k.UserID, "error", err)
		httperr.Write(w, r, http.StatusServiceUnavailable, "service_unavailable", "Service Unavailable", "")
		return "", true
	}
	if active {
		return k.UserID, true
	}
	return "", false
}

func authUserID(ctx context.Context) string {
	if v, ok := ctx.Value(authUserIDKey).(string); ok {
		return v
	}
	return ""
}

func (s *Server) isUserActive(ctx context.Context, userID string) (bool, error) {
	if s.userRepo == nil || strings.TrimSpace(userID) == "" {
		return false, nil
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return user.Status == "active", nil
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
// a scoped grant extracted from the URL path variable.
func (s *Server) requirePermissionWithScope(permissionKey, scopeType, pathVar string, next http.HandlerFunc) http.HandlerFunc {
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

		// Fall back to scoped permission.
		scopeID := mux.Vars(r)[pathVar]
		if scopeID != "" {
			ok, err = s.authzRepo.HasPermission(r.Context(), userID, scopeType, scopeID, permissionKey)
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
