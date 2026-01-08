package httpserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
)

func TestListLabEntriesCoverageHandler_IsReadOnlyTransaction(t *testing.T) {
	if os.Getenv("CALCUTTA_RUN_DB_TESTS") != "1" {
		t.Skip("set CALCUTTA_RUN_DB_TESTS=1 to run DB-backed tests")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set DATABASE_URL to run DB-backed tests")
	}

	pool, err := platform.OpenPGXPool(t.Context(), platform.Config{DatabaseURL: databaseURL}, &platform.PGXPoolOptions{MaxConns: 2})
	if err != nil {
		t.Fatalf("failed to open db pool: %v", err)
	}
	defer pool.Close()

	s := &Server{pool: pool}

	r := httptest.NewRequest(http.MethodGet, "/api/lab/entries", nil)
	w := httptest.NewRecorder()

	s.listLabEntriesCoverageHandler(w, r)

	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		body := w.Body.String()
		t.Fatalf("expected status %d, got %d: %s", want, got, body)
	}
}
