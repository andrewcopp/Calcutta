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

type createSuiteExecutionRequest struct {
	SuiteID           string   `json:"suiteId"`
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
	SuiteID          string    `json:"suite_id"`
	SuiteName        string    `json:"suite_name"`
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

func (s *Server) registerSuiteExecutionRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/suite-executions",
		s.requirePermission("analytics.suite_executions.write", s.createSuiteExecutionHandler),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/suite-executions",
		s.requirePermission("analytics.suite_executions.read", s.listSuiteExecutionsHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/suite-executions/{id}",
		s.requirePermission("analytics.suite_executions.read", s.getSuiteExecutionHandler),
	).Methods("GET", "OPTIONS")
}

func (s *Server) createSuiteExecutionHandler(w http.ResponseWriter, r *http.Request) {
	var req createSuiteExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	req.SuiteID = strings.TrimSpace(req.SuiteID)
	if req.SuiteID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "suiteId is required", "suiteId")
		return
	}
	if _, err := uuid.Parse(req.SuiteID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "suiteId must be a valid UUID", "suiteId")
		return
	}
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

	startingStateKey := "post_first_four"
	if req.StartingStateKey != nil && *req.StartingStateKey != "" {
		startingStateKey = *req.StartingStateKey
	}
	if startingStateKey != "post_first_four" && startingStateKey != "current" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "startingStateKey must be 'current' or 'post_first_four'", "startingStateKey")
		return
	}

	ctx := r.Context()

	var goAlgID string
	var msAlgID string
	var suiteOptimizerKey string
	var suiteNSims int
	var suiteSeed int
	if err := s.pool.QueryRow(ctx, `
		SELECT
			game_outcomes_algorithm_id::text,
			market_share_algorithm_id::text,
			COALESCE(optimizer_key, ''::text) AS optimizer_key,
			COALESCE(n_sims, 0)::int AS n_sims,
			COALESCE(seed, 0)::int AS seed
		FROM derived.suites
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, req.SuiteID).Scan(&goAlgID, &msAlgID, &suiteOptimizerKey, &suiteNSims, &suiteSeed); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Suite not found", "suiteId")
			return
		}
		writeErrorFromErr(w, r, err)
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
		excluded = *req.ExcludedEntryName
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
		INSERT INTO derived.suite_executions (
			suite_id,
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
	`, req.SuiteID, name, effOptimizerKey, effNSims, effSeed, startingStateKey, excluded).Scan(&executionID, &status); err != nil {
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
			INSERT INTO derived.suite_calcutta_evaluations (
				suite_execution_id,
				suite_id,
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
		`, executionID, req.SuiteID, calcuttaID, goRunID, msRunID, effOptimizerKey, effNSims, effSeed, startingStateKey, excluded)
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
	suiteID := r.URL.Query().Get("suite_id")
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
			e.suite_id::text,
			COALESCE(s.name, ''::text) AS suite_name,
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
		FROM derived.suite_executions e
		LEFT JOIN derived.suites s
			ON s.id = e.suite_id
			AND s.deleted_at IS NULL
		WHERE e.deleted_at IS NULL
			AND ($1::uuid IS NULL OR e.suite_id = $1::uuid)
		ORDER BY e.created_at DESC
		LIMIT $2::int
		OFFSET $3::int
	`, nullUUIDParam(suiteID), limit, offset)
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
			&it.SuiteID,
			&it.SuiteName,
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
			e.suite_id::text,
			COALESCE(s.name, ''::text) AS suite_name,
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
		FROM derived.suite_executions e
		LEFT JOIN derived.suites s
			ON s.id = e.suite_id
			AND s.deleted_at IS NULL
		WHERE e.id = $1::uuid
			AND e.deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.SuiteID,
		&it.SuiteName,
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
