package middleware

import (
	"net/http"
	"sync"
	"time"
)

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

func RateLimitMiddleware(
	rpm int,
	clientIP func(r *http.Request) string,
	writeError func(w http.ResponseWriter, r *http.Request, status int, code string, message string, field string),
) func(http.Handler) http.Handler {
	lim := &rateLimiter{rpm: rpm, byKey: map[string]*rateLimitEntry{}}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if lim.rpm <= 0 || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			if clientIP == nil {
				next.ServeHTTP(w, r)
				return
			}

			key := clientIP(r)
			now := time.Now()

			allowed := lim.allow(key, now)
			if !allowed {
				w.Header().Set("Retry-After", "60")
				if writeError != nil {
					writeError(w, r, http.StatusTooManyRequests, "rate_limited", "Too Many Requests", "")
					return
				}
				w.WriteHeader(http.StatusTooManyRequests)
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
