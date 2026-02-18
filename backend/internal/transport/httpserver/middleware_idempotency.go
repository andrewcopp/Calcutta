package httpserver

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
)

// idempotencyWriter captures the response for later caching.
type idempotencyWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *idempotencyWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *idempotencyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// idempotencyMiddleware wraps a handler to support Idempotency-Key header.
// If the header is absent, the request passes through unchanged.
func idempotencyMiddleware(repo *dbadapters.IdempotencyRepository, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if key == "" {
			next(w, r)
			return
		}

		userID := authUserID(r.Context())
		if userID == "" {
			next(w, r)
			return
		}

		// Check for cached response
		rec, err := repo.Get(r.Context(), key, userID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if rec != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Idempotent-Replayed", "true")
			w.WriteHeader(rec.ResponseStatus)
			w.Write(rec.ResponseBody)
			return
		}

		// Reserve the key
		reserved, err := repo.Reserve(r.Context(), key, userID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if !reserved {
			// Key exists but no response yet – concurrent request in progress
			writeError(w, r, http.StatusConflict, "idempotency_conflict", "A request with this idempotency key is already being processed", "")
			return
		}

		// Capture the response
		iw := &idempotencyWriter{ResponseWriter: w, status: http.StatusOK}
		next(iw, r)

		// Store the response (best-effort; don't fail the request if caching fails)
		if err := repo.Complete(r.Context(), key, userID, iw.status, json.RawMessage(iw.body.Bytes())); err != nil {
			slog.Error("idempotency cache store failed", "key", key, "user_id", userID, "error", err)
		}
	}
}
