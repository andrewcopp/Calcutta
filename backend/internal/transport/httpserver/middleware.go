package httpserver

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "requestID"

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return "unknown"
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := getRequestID(r.Context())

		log.Printf("[%s] Started %s %s", requestID, r.Method, r.URL.Path)

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("[%s] Completed %s %s - %d in %v",
			requestID, r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
		if allowedOriginsEnv == "" {
			allowedOriginsEnv = os.Getenv("ALLOWED_ORIGIN")
		}
		if allowedOriginsEnv == "" && os.Getenv("NODE_ENV") == "development" {
			allowedOriginsEnv = "http://localhost:3000"
		}

		allowedOrigins := make([]string, 0)
		for _, o := range strings.Split(allowedOriginsEnv, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
		if len(allowedOrigins) == 0 && os.Getenv("NODE_ENV") == "development" {
			allowedOrigins = []string{"http://localhost:3000"}
		}

		origin := r.Header.Get("Origin")
		if origin != "" {
			for _, ao := range allowedOrigins {
				if ao == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Add("Vary", "Origin")
					break
				}
			}
		} else if len(allowedOrigins) == 1 {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0])
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization, X-Requested-With, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func maxBodyBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes > 0 {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

type rateLimitEntry struct {
	windowStart time.Time
	count       int
	lastSeen    time.Time
}

type rateLimiter struct {
	rpm         int
	mu          sync.Mutex
	byKey       map[string]*rateLimitEntry
	lastCleanup time.Time
}

func rateLimitMiddleware(rpm int) func(http.Handler) http.Handler {
	lim := &rateLimiter{rpm: rpm, byKey: map[string]*rateLimitEntry{}}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if lim.rpm <= 0 || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			key := clientIP(r)
			now := time.Now()

			allowed := lim.allow(key, now)
			if !allowed {
				w.Header().Set("Retry-After", "60")
				writeError(w, r, http.StatusTooManyRequests, "rate_limited", "Too Many Requests", "")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (l *rateLimiter) allow(key string, now time.Time) bool {
	windowStart := now.Truncate(time.Minute)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.lastCleanup.IsZero() {
		l.lastCleanup = now
	}
	if now.Sub(l.lastCleanup) > 10*time.Minute {
		for k, v := range l.byKey {
			if now.Sub(v.lastSeen) > 15*time.Minute {
				delete(l.byKey, k)
			}
		}
		l.lastCleanup = now
	}

	e := l.byKey[key]
	if e == nil {
		e = &rateLimitEntry{windowStart: windowStart, count: 0, lastSeen: now}
		l.byKey[key] = e
	}

	if !e.windowStart.Equal(windowStart) {
		e.windowStart = windowStart
		e.count = 0
	}

	e.lastSeen = now
	e.count++
	return e.count <= l.rpm
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
