package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/gorilla/mux"
)

type createEntryEvaluationRequestResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type entryEvaluationRequestListItem struct {
	ID               string     `json:"id"`
	CalcuttaID       string     `json:"calcutta_id"`
	EntryCandidateID string     `json:"entry_candidate_id"`
	ExperimentKey    string     `json:"experiment_key"`
	RequestSource    string     `json:"request_source"`
	StartingStateKey string     `json:"starting_state_key"`
	ExcludedEntry    *string    `json:"excluded_entry_name,omitempty"`
	NSims            int        `json:"n_sims"`
	Seed             int        `json:"seed"`
	Status           string     `json:"status"`
	ClaimedAt        *time.Time `json:"claimed_at,omitempty"`
	ClaimedBy        *string    `json:"claimed_by,omitempty"`
	EvaluationRunID  *string    `json:"evaluation_run_id,omitempty"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type entryEvaluationRequestListResponse struct {
	Items []entryEvaluationRequestListItem `json:"items"`
}

func (s *Server) registerEntryEvaluationRequestRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/entry-evaluation-requests",
		s.requirePermission("analytics.entry_evaluation_requests.write", s.createEntryEvaluationRequestHandler),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/entry-evaluation-requests",
		s.requirePermission("analytics.entry_evaluation_requests.read", s.listEntryEvaluationRequestsHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/entry-evaluation-requests/{id}",
		s.requirePermission("analytics.entry_evaluation_requests.read", s.getEntryEvaluationRequestHandler),
	).Methods("GET", "OPTIONS")
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

func (s *Server) listEntryEvaluationRequestsHandler(w http.ResponseWriter, r *http.Request) {
	experimentKey := r.URL.Query().Get("experiment_key")
	limit := getLimit(r, 50)
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := getOffset(r, 0)
	if offset < 0 {
		offset = 0
	}

	rows, err := s.pool.Query(r.Context(), `
		SELECT
			id,
			calcutta_id,
			entry_candidate_id,
			experiment_key,
			request_source,
			starting_state_key,
			excluded_entry_name,
			n_sims,
			seed,
			status,
			claimed_at,
			claimed_by,
			evaluation_run_id,
			error_message,
			created_at,
			updated_at
		FROM derived.entry_evaluation_requests
		WHERE deleted_at IS NULL
			AND ($1::text = '' OR experiment_key = $1::text)
		ORDER BY created_at DESC
		LIMIT $2::int
		OFFSET $3::int
	`, experimentKey, limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]entryEvaluationRequestListItem, 0)
	for rows.Next() {
		var it entryEvaluationRequestListItem
		if err := rows.Scan(
			&it.ID,
			&it.CalcuttaID,
			&it.EntryCandidateID,
			&it.ExperimentKey,
			&it.RequestSource,
			&it.StartingStateKey,
			&it.ExcludedEntry,
			&it.NSims,
			&it.Seed,
			&it.Status,
			&it.ClaimedAt,
			&it.ClaimedBy,
			&it.EvaluationRunID,
			&it.ErrorMessage,
			&it.CreatedAt,
			&it.UpdatedAt,
		); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, entryEvaluationRequestListResponse{Items: items})
}

func (s *Server) getEntryEvaluationRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	var it entryEvaluationRequestListItem
	err := s.pool.QueryRow(r.Context(), `
		SELECT
			id,
			calcutta_id,
			entry_candidate_id,
			experiment_key,
			request_source,
			starting_state_key,
			excluded_entry_name,
			n_sims,
			seed,
			status,
			claimed_at,
			claimed_by,
			evaluation_run_id,
			error_message,
			created_at,
			updated_at
		FROM derived.entry_evaluation_requests
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.CalcuttaID,
		&it.EntryCandidateID,
		&it.ExperimentKey,
		&it.RequestSource,
		&it.StartingStateKey,
		&it.ExcludedEntry,
		&it.NSims,
		&it.Seed,
		&it.Status,
		&it.ClaimedAt,
		&it.ClaimedBy,
		&it.EvaluationRunID,
		&it.ErrorMessage,
		&it.CreatedAt,
		&it.UpdatedAt,
	)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, it)
}
