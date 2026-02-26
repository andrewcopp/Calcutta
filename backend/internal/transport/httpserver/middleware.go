package httpserver

import (
	"net"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
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
	return middleware.RateLimitMiddleware(rpm, clientIP, httperr.Write)
}

func clientIP(r *http.Request) string {
	h, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && h != "" {
		return h
	}
	return r.RemoteAddr
}
