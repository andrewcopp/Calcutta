package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type runProgressEventListItem struct {
	ID        string          `json:"id"`
	EventKind string          `json:"event_kind"`
	Status    *string         `json:"status,omitempty"`
	Percent   *float64        `json:"percent,omitempty"`
	Phase     *string         `json:"phase,omitempty"`
	Message   *string         `json:"message,omitempty"`
	Source    string          `json:"source"`
	Payload   json.RawMessage `json:"payload_json"`
	CreatedAt string          `json:"created_at"`
}

type getRunProgressResponse struct {
	RunKind           string                     `json:"run_kind"`
	RunID             string                     `json:"run_id"`
	RunKey            *string                    `json:"run_key,omitempty"`
	Status            string                     `json:"status"`
	Attempt           int                        `json:"attempt"`
	ParamsJSON        json.RawMessage            `json:"params_json"`
	ProgressJSON      json.RawMessage            `json:"progress_json"`
	ProgressUpdatedAt *string                    `json:"progress_updated_at,omitempty"`
	ClaimedAt         *string                    `json:"claimed_at,omitempty"`
	ClaimedBy         *string                    `json:"claimed_by,omitempty"`
	StartedAt         *string                    `json:"started_at,omitempty"`
	FinishedAt        *string                    `json:"finished_at,omitempty"`
	ErrorMessage      *string                    `json:"error_message,omitempty"`
	CreatedAt         string                     `json:"created_at"`
	UpdatedAt         string                     `json:"updated_at"`
	Events            []runProgressEventListItem `json:"events"`
}

func (s *Server) registerRunProgressRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/runs/{runKind}/{runId}/progress",
		s.requirePermission("analytics.run_jobs.read", s.handleGetRunProgress),
	).Methods("GET", "OPTIONS")
}

func (s *Server) handleGetRunProgress(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		writeError(w, r, http.StatusServiceUnavailable, "unavailable", "Database not available", "")
		return
	}

	vars := mux.Vars(r)
	runKind := strings.TrimSpace(vars["runKind"])
	runID := strings.TrimSpace(vars["runId"])
	if runKind == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "runKind is required", "runKind")
		return
	}
	if runID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "runId is required", "runId")
		return
	}

	ctx := r.Context()

	var resp getRunProgressResponse
	resp.RunKind = runKind
	resp.RunID = runID

	var runKey *string
	var paramsText string
	var progressText string
	var progressUpdatedAt *string
	var claimedAt *string
	var claimedBy *string
	var startedAt *string
	var finishedAt *string
	var errorMessage *string

	var attempt int
	var createdAt string
	var updatedAt string

	q := `
		SELECT
			run_key::text,
			status,
			attempt::int,
			params_json::text,
			progress_json::text,
			progress_updated_at::text,
			claimed_at::text,
			claimed_by,
			started_at::text,
			finished_at::text,
			error_message,
			created_at::text,
			updated_at::text
		FROM derived.run_jobs
		WHERE run_kind = $1
			AND run_id = $2::uuid
		LIMIT 1
	`

	if err := s.pool.QueryRow(ctx, q, runKind, runID).Scan(
		&runKey,
		&resp.Status,
		&attempt,
		&paramsText,
		&progressText,
		&progressUpdatedAt,
		&claimedAt,
		&claimedBy,
		&startedAt,
		&finishedAt,
		&errorMessage,
		&createdAt,
		&updatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Run job not found", "")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	if runKey != nil && strings.TrimSpace(*runKey) != "" {
		v := strings.TrimSpace(*runKey)
		resp.RunKey = &v
	}
	resp.Attempt = attempt
	resp.ParamsJSON = json.RawMessage([]byte(paramsText))
	resp.ProgressJSON = json.RawMessage([]byte(progressText))
	resp.ProgressUpdatedAt = progressUpdatedAt
	resp.ClaimedAt = claimedAt
	if claimedBy != nil && strings.TrimSpace(*claimedBy) != "" {
		v := strings.TrimSpace(*claimedBy)
		resp.ClaimedBy = &v
	}
	resp.StartedAt = startedAt
	resp.FinishedAt = finishedAt
	if errorMessage != nil && strings.TrimSpace(*errorMessage) != "" {
		v := strings.TrimSpace(*errorMessage)
		resp.ErrorMessage = &v
	}
	resp.CreatedAt = createdAt
	resp.UpdatedAt = updatedAt

	rows, err := s.pool.Query(ctx, `
		SELECT
			id::text,
			event_kind,
			status,
			percent,
			phase,
			message,
			source,
			payload_json::text,
			created_at::text
		FROM derived.run_progress_events
		WHERE run_kind = $1
			AND run_id = $2::uuid
		ORDER BY created_at ASC
	`, runKind, runID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	events := make([]runProgressEventListItem, 0)
	for rows.Next() {
		var it runProgressEventListItem
		var status *string
		var percent *float64
		var phase *string
		var message *string
		var payloadText string
		if err := rows.Scan(
			&it.ID,
			&it.EventKind,
			&status,
			&percent,
			&phase,
			&message,
			&it.Source,
			&payloadText,
			&it.CreatedAt,
		); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if status != nil && strings.TrimSpace(*status) != "" {
			v := strings.TrimSpace(*status)
			it.Status = &v
		}
		if percent != nil {
			it.Percent = percent
		}
		if phase != nil && strings.TrimSpace(*phase) != "" {
			v := strings.TrimSpace(*phase)
			it.Phase = &v
		}
		if message != nil && strings.TrimSpace(*message) != "" {
			v := strings.TrimSpace(*message)
			it.Message = &v
		}
		it.Payload = json.RawMessage([]byte(payloadText))
		events = append(events, it)
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	resp.Events = events
	writeJSON(w, http.StatusOK, resp)
}
