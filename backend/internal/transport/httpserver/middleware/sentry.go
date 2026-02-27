package middleware

import (
	"net/http"

	"github.com/getsentry/sentry-go"
)

// SentryMiddleware clones the Sentry hub per request to prevent concurrent
// scope pollution, attaches the HTTP request for context, and recovers panics.
func SentryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub := sentry.GetHubFromContext(r.Context())
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
		}
		hub.Scope().SetRequest(r)
		ctx := sentry.SetHubOnContext(r.Context(), hub)

		defer func() {
			if err := recover(); err != nil {
				hub.RecoverWithContext(ctx, err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
