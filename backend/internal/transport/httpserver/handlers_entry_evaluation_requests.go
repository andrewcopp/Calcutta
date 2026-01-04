package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/gorilla/mux"
)

type createEntryEvaluationRequestResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (s *Server) registerEntryEvaluationRequestRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/entry-evaluation-requests",
		s.requirePermission("analytics.entry_evaluation_requests.write", s.createEntryEvaluationRequestHandler),
	).Methods("POST", "OPTIONS")
}

func (s *Server) createEntryEvaluationRequestHandler(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateEntryEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	q := `
		INSERT INTO derived.entry_evaluation_requests (
			calcutta_id,
			entry_candidate_id,
			excluded_entry_name,
			starting_state_key,
			n_sims,
			seed,
			experiment_key,
			request_source
		)
		VALUES ($1::uuid, $2::uuid, $3::text, $4::text, $5::int, $6::int, $7::text, $8::text)
		RETURNING id, status
	`

	var id string
	var status string
	if err := s.pool.QueryRow(
		r.Context(),
		q,
		req.CalcuttaID,
		req.EntryCandidateID,
		req.ExcludedEntryName,
		req.StartingStateKey,
		req.NSims,
		req.Seed,
		req.ExperimentKey,
		req.RequestSource,
	).Scan(&id, &status); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, createEntryEvaluationRequestResponse{ID: id, Status: status})
}
