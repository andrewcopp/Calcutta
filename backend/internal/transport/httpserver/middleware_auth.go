package httpserver

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type authContextKey string

const (
	authUserIDKey    authContextKey = "authUserID"
	authSessionIDKey authContextKey = "authSessionID"
)

func (s *Server) authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.tokenManager == nil {
			next.ServeHTTP(w, r)
			return
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

		claims, err := s.tokenManager.VerifyAccessToken(tok, time.Now())
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		sess, err := s.authRepo.GetSessionByID(r.Context(), claims.Sid)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		if sess == nil || sess.RevokedAt != nil || time.Now().After(sess.ExpiresAt) {
			next.ServeHTTP(w, r)
			return
		}
		if sess.UserID != claims.Sub {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), authUserIDKey, claims.Sub)
		ctx = context.WithValue(ctx, authSessionIDKey, claims.Sid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func authUserID(ctx context.Context) string {
	if v, ok := ctx.Value(authUserIDKey).(string); ok {
		return v
	}
	return ""
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
