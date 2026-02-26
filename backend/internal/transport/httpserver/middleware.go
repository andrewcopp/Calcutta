package httpserver

import (
	"net"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/middleware"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/requestctx"
)

func requestIDMiddleware(next http.Handler) http.Handler {
	return requestctx.Middleware(next)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return middleware.LoggingMiddleware(
		httpMetrics.Observe,
		requestctx.Logger,
		clientIP(s.cfg.TrustProxyHeaders),
		authUserID,
	)(next)
}

func (s *Server) rateLimitMiddleware(rpm int) func(http.Handler) http.Handler {
	return middleware.RateLimitMiddleware(rpm, clientIP(s.cfg.TrustProxyHeaders), httperr.Write)
}

func clientIP(trustProxy bool) func(r *http.Request) string {
	return func(r *http.Request) string {
		if trustProxy {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				ip := strings.TrimSpace(strings.Split(xff, ",")[0])
				if net.ParseIP(ip) != nil {
					return ip
				}
			}
			if xri := r.Header.Get("X-Real-IP"); xri != "" {
				if net.ParseIP(xri) != nil {
					return xri
				}
			}
		}
		h, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil && h != "" {
			return h
		}
		return r.RemoteAddr
	}
}
