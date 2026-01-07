package httpserver

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

func (s *Server) registerSyntheticCalcuttaCohortRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/cohorts",
		s.requirePermission("analytics.suites.read", s.listSyntheticCalcuttaCohortsHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/cohorts/{id}",
		s.requirePermission("analytics.suites.read", s.getSyntheticCalcuttaCohortHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/cohorts/{id}",
		s.requirePermission("analytics.suites.write", s.patchSyntheticCalcuttaCohortHandler),
	).Methods("PATCH", "OPTIONS")
}

func (s *Server) listSyntheticCalcuttaCohortsHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) getSyntheticCalcuttaCohortHandler(w http.ResponseWriter, r *http.Request) {
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
			writeError(w, r, http.StatusNotFound, "not_found", "Cohort not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, it)
}

func (s *Server) patchSyntheticCalcuttaCohortHandler(w http.ResponseWriter, r *http.Request) {
	// Keep compatibility: re-use suite patch handler for cohort patch.
	s.updateSuiteHandler(w, r)
}
