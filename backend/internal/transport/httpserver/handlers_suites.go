package httpserver

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type suiteListItem struct {
	ID                       string     `json:"id"`
	Name                     string     `json:"name"`
	Description              *string    `json:"description,omitempty"`
	GameOutcomesAlgID        string     `json:"game_outcomes_algorithm_id"`
	MarketShareAlgID         string     `json:"market_share_algorithm_id"`
	OptimizerKey             string     `json:"optimizer_key"`
	NSims                    int        `json:"n_sims"`
	Seed                     int        `json:"seed"`
	LatestExecutionID        *string    `json:"latest_execution_id,omitempty"`
	LatestExecutionName      *string    `json:"latest_execution_name,omitempty"`
	LatestExecutionStatus    *string    `json:"latest_execution_status,omitempty"`
	LatestExecutionCreatedAt *time.Time `json:"latest_execution_created_at,omitempty"`
	LatestExecutionUpdatedAt *time.Time `json:"latest_execution_updated_at,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type listSuitesResponse struct {
	Items []suiteListItem `json:"items"`
}

func (s *Server) registerSuiteRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/suites",
		s.requirePermission("analytics.suite_executions.read", s.listSuitesHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/suites/{id}",
		s.requirePermission("analytics.suite_executions.read", s.getSuiteHandler),
	).Methods("GET", "OPTIONS")
}

func (s *Server) listSuitesHandler(w http.ResponseWriter, r *http.Request) {
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
			s.id::text,
			COALESCE(s.name, ''::text) AS name,
			s.description,
			s.game_outcomes_algorithm_id::text,
			s.market_share_algorithm_id::text,
			s.optimizer_key,
			s.n_sims,
			s.seed,
			le.id,
			le.name,
			le.status,
			le.created_at,
			le.updated_at,
			s.created_at,
			s.updated_at
		FROM derived.suites s
		LEFT JOIN LATERAL (
			SELECT
				e.id::text,
				e.name,
				e.status,
				e.created_at,
				e.updated_at
			FROM derived.suite_executions e
			WHERE e.suite_id = s.id
				AND e.deleted_at IS NULL
			ORDER BY e.created_at DESC
			LIMIT 1
		) le ON TRUE
		WHERE s.deleted_at IS NULL
		ORDER BY s.created_at DESC
		LIMIT $1::int
		OFFSET $2::int
	`, limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]suiteListItem, 0)
	for rows.Next() {
		var it suiteListItem
		if err := rows.Scan(
			&it.ID,
			&it.Name,
			&it.Description,
			&it.GameOutcomesAlgID,
			&it.MarketShareAlgID,
			&it.OptimizerKey,
			&it.NSims,
			&it.Seed,
			&it.LatestExecutionID,
			&it.LatestExecutionName,
			&it.LatestExecutionStatus,
			&it.LatestExecutionCreatedAt,
			&it.LatestExecutionUpdatedAt,
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

	writeJSON(w, http.StatusOK, listSuitesResponse{Items: items})
}

func (s *Server) getSuiteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := strings.TrimSpace(vars["id"])
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	var it suiteListItem
	if err := s.pool.QueryRow(r.Context(), `
		SELECT
			s.id::text,
			COALESCE(s.name, ''::text) AS name,
			s.description,
			s.game_outcomes_algorithm_id::text,
			s.market_share_algorithm_id::text,
			s.optimizer_key,
			s.n_sims,
			s.seed,
			le.id,
			le.name,
			le.status,
			le.created_at,
			le.updated_at,
			s.created_at,
			s.updated_at
		FROM derived.suites s
		LEFT JOIN LATERAL (
			SELECT
				e.id::text,
				e.name,
				e.status,
				e.created_at,
				e.updated_at
			FROM derived.suite_executions e
			WHERE e.suite_id = s.id
				AND e.deleted_at IS NULL
			ORDER BY e.created_at DESC
			LIMIT 1
		) le ON TRUE
		WHERE s.id = $1::uuid
			AND s.deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.Name,
		&it.Description,
		&it.GameOutcomesAlgID,
		&it.MarketShareAlgID,
		&it.OptimizerKey,
		&it.NSims,
		&it.Seed,
		&it.LatestExecutionID,
		&it.LatestExecutionName,
		&it.LatestExecutionStatus,
		&it.LatestExecutionCreatedAt,
		&it.LatestExecutionUpdatedAt,
		&it.CreatedAt,
		&it.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Suite not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, it)
}
