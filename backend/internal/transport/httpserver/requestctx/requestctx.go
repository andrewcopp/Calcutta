package requestctx

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "requestID"

const requestLoggerKey contextKey = "requestLogger"

// Middleware attaches a request id (from X-Request-ID or generated) and a request-scoped logger to the request context.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		ctx = context.WithValue(ctx, requestLoggerKey, slog.Default().With("request_id", requestID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return "unknown"
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func Logger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(requestLoggerKey).(*slog.Logger); ok && l != nil {
		return l
	}
	requestID := GetRequestID(ctx)
	if requestID != "" && requestID != "unknown" {
		return slog.Default().With("request_id", requestID)
	}
	return slog.Default()
}
