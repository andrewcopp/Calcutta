package httpserver

import (
	"encoding/json"
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
	StartingStateKey         string     `json:"starting_state_key"`
	ExcludedEntryName        *string    `json:"excluded_entry_name,omitempty"`
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
		s.requirePermission("analytics.suites.read", s.listSuitesHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/suites/{id}",
		s.requirePermission("analytics.suites.read", s.getSuiteHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/suites/{id}",
		s.requirePermission("analytics.suites.write", s.updateSuiteHandler),
	).Methods("PATCH", "OPTIONS")
}

type updateSuiteRequest struct {
	OptimizerKey      *string `json:"optimizerKey"`
	NSims             *int    `json:"nSims"`
	Seed              *int    `json:"seed"`
	StartingStateKey  *string `json:"startingStateKey"`
	ExcludedEntryName *string `json:"excludedEntryName"`
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
			c.id::text,
			COALESCE(c.name, ''::text) AS name,
			c.description,
			c.game_outcomes_algorithm_id::text,
			c.market_share_algorithm_id::text,
			c.optimizer_key,
			c.n_sims,
			c.seed,
			COALESCE(NULLIF(c.starting_state_key, ''), 'post_first_four') AS starting_state_key,
			c.excluded_entry_name,
			le.id,
			le.name,
			le.status,
			le.created_at,
			le.updated_at,
			c.created_at,
			c.updated_at
		FROM derived.synthetic_calcutta_cohorts c
		LEFT JOIN LATERAL (
			SELECT
				e.id::text,
				e.name,
				e.status,
				e.created_at,
				e.updated_at
			FROM derived.simulation_run_batches e
			WHERE e.cohort_id = c.id
				AND e.deleted_at IS NULL
			ORDER BY e.created_at DESC
			LIMIT 1
		) le ON TRUE
		WHERE c.deleted_at IS NULL
		ORDER BY c.created_at DESC
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
			&it.StartingStateKey,
			&it.ExcludedEntryName,
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
			c.id::text,
			COALESCE(c.name, ''::text) AS name,
			c.description,
			c.game_outcomes_algorithm_id::text,
			c.market_share_algorithm_id::text,
			c.optimizer_key,
			c.n_sims,
			c.seed,
			COALESCE(NULLIF(c.starting_state_key, ''), 'post_first_four') AS starting_state_key,
			c.excluded_entry_name,
			le.id,
			le.name,
			le.status,
			le.created_at,
			le.updated_at,
			c.created_at,
			c.updated_at
		FROM derived.synthetic_calcutta_cohorts c
		LEFT JOIN LATERAL (
			SELECT
				e.id::text,
				e.name,
				e.status,
				e.created_at,
				e.updated_at
			FROM derived.simulation_run_batches e
			WHERE e.cohort_id = c.id
				AND e.deleted_at IS NULL
			ORDER BY e.created_at DESC
			LIMIT 1
		) le ON TRUE
		WHERE c.id = $1::uuid
			AND c.deleted_at IS NULL
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
		&it.StartingStateKey,
		&it.ExcludedEntryName,
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

func (s *Server) updateSuiteHandler(w http.ResponseWriter, r *http.Request) {
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

	var req updateSuiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	var optimizer any
	if req.OptimizerKey != nil {
		v := strings.TrimSpace(*req.OptimizerKey)
		if v == "" {
			optimizer = nil
		} else {
			optimizer = v
		}
	} else {
		optimizer = nil
	}

	var nSims any
	if req.NSims != nil {
		if *req.NSims <= 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "nSims must be positive", "nSims")
			return
		}
		nSims = *req.NSims
	} else {
		nSims = nil
	}

	var seed any
	if req.Seed != nil {
		seed = *req.Seed
	} else {
		seed = nil
	}

	var starting any
	if req.StartingStateKey != nil {
		v := strings.TrimSpace(*req.StartingStateKey)
		if v != "" && v != "post_first_four" && v != "current" {
			writeError(w, r, http.StatusBadRequest, "validation_error", "startingStateKey must be 'current' or 'post_first_four'", "startingStateKey")
			return
		}
		if v == "" {
			starting = nil
		} else {
			starting = v
		}
	} else {
		starting = nil
	}

	var excluded any
	if req.ExcludedEntryName != nil {
		v := strings.TrimSpace(*req.ExcludedEntryName)
		if v == "" {
			excluded = nil
		} else {
			excluded = v
		}
	} else {
		excluded = nil
	}

	ct, err := s.pool.Exec(r.Context(), `
		UPDATE derived.synthetic_calcutta_cohorts
		SET optimizer_key = COALESCE($2::text, optimizer_key),
			n_sims = COALESCE($3::int, n_sims),
			seed = COALESCE($4::int, seed),
			starting_state_key = COALESCE($5::text, starting_state_key),
			excluded_entry_name = CASE
				WHEN $6::text IS NULL THEN excluded_entry_name
				ELSE $6::text
			END,
			updated_at = NOW(),
			deleted_at = NULL
		WHERE id = $1::uuid
			AND deleted_at IS NULL
	`, id, optimizer, nSims, seed, starting, excluded)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if ct.RowsAffected() == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "Suite not found", "id")
		return
	}

	s.getSuiteHandler(w, r)
}
