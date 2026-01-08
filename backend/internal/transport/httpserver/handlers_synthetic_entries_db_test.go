package httpserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func TestHandleListSyntheticEntries_IsReadOnlyTransaction(t *testing.T) {
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

	syntheticCalcuttaID := uuid.NewString()
	r := httptest.NewRequest(http.MethodGet, "/api/synthetic-calcuttas/"+syntheticCalcuttaID+"/synthetic-entries", nil)
	r = mux.SetURLVars(r, map[string]string{"id": syntheticCalcuttaID})
	w := httptest.NewRecorder()

	s.handleListSyntheticEntries(w, r)

	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		body := w.Body.String()
		t.Fatalf("expected status %d, got %d: %s", want, got, body)
	}
}
