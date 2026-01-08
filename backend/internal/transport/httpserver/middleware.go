package httpserver

import (
	"net"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/middleware"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/requestctx"
)

func requestIDMiddleware(next http.Handler) http.Handler {
	return requestctx.Middleware(next)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return middleware.LoggingMiddleware(
		httpMetrics.Observe,
		requestctx.Logger,
		clientIP,
		authUserID,
	)(next)
}

func rateLimitMiddleware(rpm int) func(http.Handler) http.Handler {
	return middleware.RateLimitMiddleware(rpm, clientIP, writeError)
}

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	host := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if host != "" {
		return host
	}

	h, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && h != "" {
		return h
	}
	return r.RemoteAddr
}
