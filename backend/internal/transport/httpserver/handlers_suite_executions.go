package httpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type createSuiteExecutionRequest struct {
	CohortID          string   `json:"cohortId"`
	Name              *string  `json:"name"`
	CalcuttaIDs       []string `json:"calcuttaIds"`
	OptimizerKey      *string  `json:"optimizerKey"`
	NSims             *int     `json:"nSims"`
	Seed              *int     `json:"seed"`
	StartingStateKey  *string  `json:"startingStateKey"`
	ExcludedEntryName *string  `json:"excludedEntryName"`
}

type createSuiteExecutionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type suiteExecutionListItem struct {
	ID               string    `json:"id"`
	CohortID         string    `json:"cohort_id"`
	CohortName       string    `json:"cohort_name"`
	Name             *string   `json:"name,omitempty"`
	OptimizerKey     *string   `json:"optimizer_key,omitempty"`
	NSims            *int      `json:"n_sims,omitempty"`
	Seed             *int      `json:"seed,omitempty"`
	StartingStateKey string    `json:"starting_state_key"`
	ExcludedEntry    *string   `json:"excluded_entry_name,omitempty"`
	Status           string    `json:"status"`
	ErrorMessage     *string   `json:"error_message,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type suiteExecutionListResponse struct {
	Items []suiteExecutionListItem `json:"items"`
}

func (s *Server) registerCohortSimulationBatchRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/cohorts/{cohortId}/simulation-batches",
		s.requirePermission("analytics.suite_executions.write", s.createCohortSimulationBatchHandler),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/cohorts/{cohortId}/simulation-batches",
		s.requirePermission("analytics.suite_executions.read", s.listCohortSimulationBatchesHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/cohorts/{cohortId}/simulation-batches/{id}",
		s.requirePermission("analytics.suite_executions.read", s.getCohortSimulationBatchHandler),
	).Methods("GET", "OPTIONS")
}

func (s *Server) listCohortSimulationBatchesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	q := r.URL.Query()
	q.Set("cohort_id", cohortID)
	r.URL.RawQuery = q.Encode()
	s.listSuiteExecutionsHandler(w, r)
}

func (s *Server) getCohortSimulationBatchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}

	id := strings.TrimSpace(vars["id"])
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	var it suiteExecutionListItem
	if err := s.pool.QueryRow(r.Context(), `
		SELECT
			e.id::text,
			e.cohort_id::text,
			COALESCE(s.name, ''::text) AS cohort_name,
			e.name,
			e.optimizer_key,
			e.n_sims,
			e.seed,
			e.starting_state_key,
			e.excluded_entry_name,
			e.status,
			e.error_message,
			e.created_at,
			e.updated_at
		FROM derived.simulation_run_batches e
		LEFT JOIN derived.synthetic_calcutta_cohorts s
			ON s.id = e.cohort_id
			AND s.deleted_at IS NULL
		WHERE e.id = $1::uuid
			AND e.deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.CohortID,
		&it.CohortName,
		&it.Name,
		&it.OptimizerKey,
		&it.NSims,
		&it.Seed,
		&it.StartingStateKey,
		&it.ExcludedEntry,
		&it.Status,
		&it.ErrorMessage,
		&it.CreatedAt,
		&it.UpdatedAt,
	); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if strings.TrimSpace(it.CohortID) != cohortID {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulation batch not found", "id")
		return
	}

	writeJSON(w, http.StatusOK, it)
}

func (s *Server) createCohortSimulationBatchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}

	var req createSuiteExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if strings.TrimSpace(req.CohortID) == "" {
		req.CohortID = cohortID
	}

	b, _ := json.Marshal(req)
	cloned := r.Clone(r.Context())
	cloned.Body = io.NopCloser(bytes.NewReader(b))
	s.createSuiteExecutionHandler(w, cloned)
}

func (s *Server) createSuiteExecutionHandler(w http.ResponseWriter, r *http.Request) {
	var req createSuiteExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	req.CohortID = strings.TrimSpace(req.CohortID)
	if req.CohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(req.CohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}

	ctx := r.Context()

	var goAlgID string
	var msAlgID string
	var suiteOptimizerKey string
	var suiteNSims int
	var suiteSeed int
	var suiteStartingStateKey string
	var suiteExcludedEntryName *string
	if err := s.pool.QueryRow(ctx, `
		SELECT
			game_outcomes_algorithm_id::text,
			market_share_algorithm_id::text,
			COALESCE(optimizer_key, ''::text) AS optimizer_key,
			n_sims,
			seed,
			COALESCE(NULLIF(starting_state_key, ''), 'post_first_four') AS starting_state_key,
			excluded_entry_name
		FROM derived.synthetic_calcutta_cohorts
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, req.CohortID).Scan(&goAlgID, &msAlgID, &suiteOptimizerKey, &suiteNSims, &suiteSeed, &suiteStartingStateKey, &suiteExcludedEntryName); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Cohort not found", "cohortId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	// Resolve calcutta IDs if omitted.
	if len(req.CalcuttaIDs) == 0 {
		writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaIds is required", "calcuttaIds")
		return
	}
	for i := range req.CalcuttaIDs {
		req.CalcuttaIDs[i] = strings.TrimSpace(req.CalcuttaIDs[i])
		if req.CalcuttaIDs[i] == "" {
			writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaIds contains an empty id", "calcuttaIds")
			return
		}
		if _, err := uuid.Parse(req.CalcuttaIDs[i]); err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaIds must all be valid UUIDs", "calcuttaIds")
			return
		}
	}

	startingStateKey := suiteStartingStateKey
	if req.StartingStateKey != nil && strings.TrimSpace(*req.StartingStateKey) != "" {
		startingStateKey = strings.TrimSpace(*req.StartingStateKey)
	}
	if startingStateKey == "" {
		startingStateKey = "post_first_four"
	}
	if startingStateKey != "post_first_four" && startingStateKey != "current" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "startingStateKey must be 'current' or 'post_first_four'", "startingStateKey")
		return
	}

	effOptimizerKey := suiteOptimizerKey
	if req.OptimizerKey != nil && *req.OptimizerKey != "" {
		effOptimizerKey = *req.OptimizerKey
	}
	var effNSims any
	if req.NSims != nil {
		effNSims = *req.NSims
	} else if suiteNSims > 0 {
		effNSims = suiteNSims
	} else {
		effNSims = nil
	}
	var effSeed any
	if req.Seed != nil {
		effSeed = *req.Seed
	} else if suiteSeed != 0 {
		effSeed = suiteSeed
	} else {
		effSeed = nil
	}
	var excluded any
	if req.ExcludedEntryName != nil {
		v := strings.TrimSpace(*req.ExcludedEntryName)
		if v != "" {
			excluded = v
		} else {
			excluded = nil
		}
	} else if suiteExcludedEntryName != nil && strings.TrimSpace(*suiteExcludedEntryName) != "" {
		excluded = strings.TrimSpace(*suiteExcludedEntryName)
	} else {
		excluded = nil
	}
	var name any
	if req.Name != nil {
		name = *req.Name
	} else {
		name = nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	var executionID string
	var status string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.simulation_run_batches (
			cohort_id,
			name,
			optimizer_key,
			n_sims,
			seed,
			starting_state_key,
			excluded_entry_name,
			status
		)
		VALUES ($1::uuid, $2, $3, $4::int, $5::int, $6, $7::text, 'running')
		RETURNING id::text, status
	`, req.CohortID, name, effOptimizerKey, effNSims, effSeed, startingStateKey, excluded).Scan(&executionID, &status); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	for _, calcuttaID := range req.CalcuttaIDs {
		var tournamentID string
		if err := tx.QueryRow(ctx, `
			SELECT tournament_id::text
			FROM core.calcuttas
			WHERE id = $1::uuid
				AND deleted_at IS NULL
			LIMIT 1
		`, calcuttaID).Scan(&tournamentID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		var goRunID string
		if err := tx.QueryRow(ctx, `
			SELECT id::text
			FROM derived.game_outcome_runs
			WHERE tournament_id = $1::uuid
				AND algorithm_id = $2::uuid
				AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1
		`, tournamentID, goAlgID).Scan(&goRunID); err != nil {
			if err == pgx.ErrNoRows {
				writeError(w, r, http.StatusConflict, "missing_run", "Missing game-outcome run for suite execution", "gameOutcomeRunId")
				return
			}
			writeErrorFromErr(w, r, err)
			return
		}

		var msRunID string
		if err := tx.QueryRow(ctx, `
			SELECT id::text
			FROM derived.market_share_runs
			WHERE calcutta_id = $1::uuid
				AND algorithm_id = $2::uuid
				AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1
		`, calcuttaID, msAlgID).Scan(&msRunID); err != nil {
			if err == pgx.ErrNoRows {
				writeError(w, r, http.StatusConflict, "missing_run", "Missing market-share run for suite execution", "marketShareRunId")
				return
			}
			writeErrorFromErr(w, r, err)
			return
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO derived.simulation_runs (
				simulation_run_batch_id,
				cohort_id,
				calcutta_id,
				game_outcome_run_id,
				market_share_run_id,
				optimizer_key,
				n_sims,
				seed,
				starting_state_key,
				excluded_entry_name
			)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid, $6, $7::int, $8::int, $9, $10::text)
		`, executionID, req.CohortID, calcuttaID, goRunID, msRunID, effOptimizerKey, effNSims, effSeed, startingStateKey, excluded)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

	writeJSON(w, http.StatusCreated, createSuiteExecutionResponse{ID: executionID, Status: status})
}

func (s *Server) listSuiteExecutionsHandler(w http.ResponseWriter, r *http.Request) {
	cohortID := r.URL.Query().Get("cohort_id")
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
			e.id::text,
			e.cohort_id::text,
			COALESCE(s.name, ''::text) AS cohort_name,
			e.name,
			e.optimizer_key,
			e.n_sims,
			e.seed,
			e.starting_state_key,
			e.excluded_entry_name,
			e.status,
			e.error_message,
			e.created_at,
			e.updated_at
		FROM derived.simulation_run_batches e
		LEFT JOIN derived.synthetic_calcutta_cohorts s
			ON s.id = e.cohort_id
			AND s.deleted_at IS NULL
		WHERE e.deleted_at IS NULL
			AND ($1::uuid IS NULL OR e.cohort_id = $1::uuid)
		ORDER BY e.created_at DESC
		LIMIT $2::int
		OFFSET $3::int
	`, nullUUIDParam(cohortID), limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]suiteExecutionListItem, 0)
	for rows.Next() {
		var it suiteExecutionListItem
		if err := rows.Scan(
			&it.ID,
			&it.CohortID,
			&it.CohortName,
			&it.Name,
			&it.OptimizerKey,
			&it.NSims,
			&it.Seed,
			&it.StartingStateKey,
			&it.ExcludedEntry,
			&it.Status,
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

	writeJSON(w, http.StatusOK, suiteExecutionListResponse{Items: items})
}

func (s *Server) getSuiteExecutionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	var it suiteExecutionListItem
	if err := s.pool.QueryRow(r.Context(), `
		SELECT
			e.id::text,
			e.cohort_id::text,
			COALESCE(s.name, ''::text) AS cohort_name,
			e.name,
			e.optimizer_key,
			e.n_sims,
			e.seed,
			e.starting_state_key,
			e.excluded_entry_name,
			e.status,
			e.error_message,
			e.created_at,
			e.updated_at
		FROM derived.simulation_run_batches e
		LEFT JOIN derived.synthetic_calcutta_cohorts s
			ON s.id = e.cohort_id
			AND s.deleted_at IS NULL
		WHERE e.id = $1::uuid
			AND e.deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.CohortID,
		&it.CohortName,
		&it.Name,
		&it.OptimizerKey,
		&it.NSims,
		&it.Seed,
		&it.StartingStateKey,
		&it.ExcludedEntry,
		&it.Status,
		&it.ErrorMessage,
		&it.CreatedAt,
		&it.UpdatedAt,
	); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, it)
}
