package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestSyntheticEntryRoutes_CandidateAliasesAreRegistered(t *testing.T) {
	s := &Server{}

	r := mux.NewRouter()
	s.registerSyntheticEntryRoutes(r)

	id := "00000000-0000-0000-0000-000000000000"
	syntheticCalcuttaID := "11111111-1111-1111-1111-111111111111"

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "legacy_patch_flat",
			method: http.MethodPatch,
			path:   "/api/synthetic-entries/" + id,
		},
		{
			name:   "legacy_delete_flat",
			method: http.MethodDelete,
			path:   "/api/synthetic-entries/" + id,
		},
		{
			name:   "attachment_patch_flat",
			method: http.MethodPatch,
			path:   "/api/synthetic-calcutta-candidates/" + id,
		},
		{
			name:   "attachment_delete_flat",
			method: http.MethodDelete,
			path:   "/api/synthetic-calcutta-candidates/" + id,
		},
		{
			name:   "candidate_patch_nested",
			method: http.MethodPatch,
			path:   "/api/synthetic-calcuttas/" + syntheticCalcuttaID + "/candidates/" + id,
		},
		{
			name:   "candidate_delete_nested",
			method: http.MethodDelete,
			path:   "/api/synthetic-calcuttas/" + syntheticCalcuttaID + "/candidates/" + id,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			r.ServeHTTP(w, req)

			if got, want := w.Result().StatusCode, http.StatusUnauthorized; got != want {
				t.Fatalf("expected status %d for %s %s, got %d", want, tc.method, tc.path, got)
			}
		})
	}
}
