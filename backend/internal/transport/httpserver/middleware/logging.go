package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type statusCapturingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *statusCapturingResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(
	observe func(method string, status int, duration time.Duration),
	logger func(ctx context.Context) *slog.Logger,
	clientIP func(r *http.Request) string,
	authUserID func(ctx context.Context) string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &statusCapturingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			if observe != nil {
				observe(r.Method, wrapped.statusCode, duration)
			}

			l := slog.Default()
			if logger != nil {
				l = logger(r.Context())
			}

			var userID string
			if authUserID != nil {
				userID = authUserID(r.Context())
			}

			var ip string
			if clientIP != nil {
				ip = clientIP(r)
			}

			l.InfoContext(
				r.Context(),
				"http_request",
				"event", "http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"client_ip", ip,
				"user_id", userID,
			)
		})
	}
}
