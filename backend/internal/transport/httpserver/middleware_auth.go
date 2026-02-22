package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/gorilla/mux"
)

type authContextKey string

const (
	authUserIDKey    authContextKey = "authUserID"
	authSessionIDKey authContextKey = "authSessionID"
)

func (s *Server) authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Dev auth bypass (header-based, not token-based).
		if s.devMode {
			if identity, err := auth.DevAuthenticate(r, s.userRepo); err != nil {
				slog.Error("auth_dev_failed", "error", err)
				httperr.Write(w, r, http.StatusServiceUnavailable, "service_unavailable", "Service Unavailable", "")
				return
			} else if identity != nil {
				ctx := context.WithValue(r.Context(), authUserIDKey, identity.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		tok := extractBearerToken(r)
		if tok == "" {
			next.ServeHTTP(w, r)
			return
		}

		identity, err := s.authenticator.Authenticate(r.Context(), tok)
		if err != nil {
			slog.Error("auth_failed", "error", err)
			httperr.Write(w, r, http.StatusServiceUnavailable, "service_unavailable", "Service Unavailable", "")
			return
		}
		if identity == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), authUserIDKey, identity.UserID)
		if identity.SessionID != "" {
			ctx = context.WithValue(ctx, authSessionIDKey, identity.SessionID)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
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

func authUserID(ctx context.Context) string {
	if v, ok := ctx.Value(authUserIDKey).(string); ok {
		return v
	}
	return ""
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
