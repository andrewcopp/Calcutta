package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type cohortListItem struct {
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

type listCohortsResponse struct {
	Items []cohortListItem `json:"items"`
}

func (s *Server) registerSyntheticCalcuttaCohortRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/cohorts",
		s.requirePermission("analytics.suites.read", s.listSyntheticCalcuttaCohortsHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/cohorts",
		s.requirePermission("analytics.suites.write", s.createSyntheticCalcuttaCohortHandler),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/cohorts/{id}",
		s.requirePermission("analytics.suites.read", s.getSyntheticCalcuttaCohortHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/cohorts/{id}",
		s.requirePermission("analytics.suites.write", s.patchSyntheticCalcuttaCohortHandler),
	).Methods("PATCH", "OPTIONS")
}

func (s *Server) createSyntheticCalcuttaCohortHandler(w http.ResponseWriter, r *http.Request) {
	type createCohortRequest struct {
		Name                  string  `json:"name"`
		Description           *string `json:"description"`
		GameOutcomesAlgorithm *string `json:"gameOutcomesAlgorithmId"`
		MarketShareAlgorithm  *string `json:"marketShareAlgorithmId"`
		OptimizerKey          *string `json:"optimizerKey"`
		NSims                 *int    `json:"nSims"`
		Seed                  *int    `json:"seed"`
		StartingStateKey      *string `json:"startingStateKey"`
		ExcludedEntryName     *string `json:"excludedEntryName"`
	}

	var req createCohortRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "name is required", "name")
		return
	}

	var desc any
	if req.Description != nil {
		v := strings.TrimSpace(*req.Description)
		if v == "" {
			desc = nil
		} else {
			desc = v
		}
	} else {
		desc = nil
	}

	ctx := r.Context()

	goAlgID := ""
	if req.GameOutcomesAlgorithm != nil {
		v := strings.TrimSpace(*req.GameOutcomesAlgorithm)
		if v != "" {
			if _, err := uuid.Parse(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "gameOutcomesAlgorithmId must be a valid UUID", "gameOutcomesAlgorithmId")
				return
			}
			goAlgID = v
		}
	}
	if goAlgID == "" {
		id, err := s.getDefaultAlgorithmID(ctx, "game_outcomes", "kenpom-v1-go")
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		goAlgID = id
	}

	msAlgID := ""
	if req.MarketShareAlgorithm != nil {
		v := strings.TrimSpace(*req.MarketShareAlgorithm)
		if v != "" {
			if _, err := uuid.Parse(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "marketShareAlgorithmId must be a valid UUID", "marketShareAlgorithmId")
				return
			}
			msAlgID = v
		}
	}
	if msAlgID == "" {
		id, err := s.getDefaultAlgorithmID(ctx, "market_share", "naive-ev-baseline")
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		msAlgID = id
	}

	// Validate algorithm IDs exist and match expected kind.
	{
		var ok bool
		if err := s.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1
				FROM derived.algorithms
				WHERE id = $1::uuid
					AND kind = 'game_outcomes'
					AND deleted_at IS NULL
			)
		`, goAlgID).Scan(&ok); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if !ok {
			writeError(w, r, http.StatusBadRequest, "validation_error", "gameOutcomesAlgorithmId not found", "gameOutcomesAlgorithmId")
			return
		}
	}
	{
		var ok bool
		if err := s.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1
				FROM derived.algorithms
				WHERE id = $1::uuid
					AND kind = 'market_share'
					AND deleted_at IS NULL
			)
		`, msAlgID).Scan(&ok); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if !ok {
			writeError(w, r, http.StatusBadRequest, "validation_error", "marketShareAlgorithmId not found", "marketShareAlgorithmId")
			return
		}
	}

	// Default algorithm IDs already resolved above.

	optimizerKey := "minlp_v1"
	if req.OptimizerKey != nil {
		v := strings.TrimSpace(*req.OptimizerKey)
		if v != "" {
			optimizerKey = v
		}
	}
	// If optimizer key wasn't explicitly provided, default to first non-deprecated optimizer.
	if req.OptimizerKey == nil || strings.TrimSpace(*req.OptimizerKey) == "" {
		if s.app != nil && s.app.ModelCatalogs != nil {
			descs := s.app.ModelCatalogs.ListEntryOptimizers()
			for _, d := range descs {
				if d.Deprecated {
					continue
				}
				if strings.TrimSpace(d.ID) == "" {
					continue
				}
				optimizerKey = strings.TrimSpace(d.ID)
				break
			}
		}
	}

	nSims := 5000
	if req.NSims != nil {
		if *req.NSims <= 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "nSims must be positive", "nSims")
			return
		}
		nSims = *req.NSims
	}

	seed := 42
	if req.Seed != nil {
		if *req.Seed == 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "seed cannot be 0", "seed")
			return
		}
		seed = *req.Seed
	}

	startingStateKey := "post_first_four"
	if req.StartingStateKey != nil {
		v := strings.TrimSpace(*req.StartingStateKey)
		if v != "" {
			if v != "post_first_four" && v != "current" {
				writeError(w, r, http.StatusBadRequest, "validation_error", "startingStateKey must be 'current' or 'post_first_four'", "startingStateKey")
				return
			}
			startingStateKey = v
		}
	}

	excludedEntryName := (*string)(nil)
	if req.ExcludedEntryName != nil {
		v := strings.TrimSpace(*req.ExcludedEntryName)
		if v != "" {
			excludedEntryName = &v
		}
	}

	var createdID string
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO derived.simulation_cohorts (
			name,
			description,
			game_outcomes_algorithm_id,
			market_share_algorithm_id,
			optimizer_key,
			n_sims,
			seed,
			starting_state_key,
			excluded_entry_name,
			params_json
		)
		VALUES (
			$1,
			$2,
			$3::uuid,
			$4::uuid,
			$5,
			$6::int,
			$7::int,
			$8,
			$9,
			'{"auto": false}'::jsonb
		)
		RETURNING id::text
	`, req.Name, desc, goAlgID, msAlgID, optimizerKey, nSims, seed, startingStateKey, excludedEntryName).Scan(&createdID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// unique violation
			if pgErr.Code == "23505" {
				writeError(w, r, http.StatusConflict, "conflict", "A cohort with that name already exists", "name")
				return
			}
			// foreign key violation
			if pgErr.Code == "23503" {
				writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid algorithm ID(s)", "")
				return
			}
		}
		writeErrorFromErr(w, r, err)
		return
	}

	var it cohortListItem
	if err := s.pool.QueryRow(ctx, `
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
		FROM derived.simulation_cohorts c
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
	`, createdID).Scan(
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

	writeJSON(w, http.StatusCreated, it)
}

func (s *Server) getDefaultAlgorithmID(ctx context.Context, kind string, preferredName string) (string, error) {
	// Preferred algorithm by name.
	var id string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text
		FROM derived.algorithms
		WHERE kind = $1
			AND name = $2
			AND deleted_at IS NULL
		LIMIT 1
	`, kind, preferredName).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != pgx.ErrNoRows {
		return "", err
	}

	// Fallback: first algorithm for the kind.
	if err := s.pool.QueryRow(ctx, `
		SELECT id::text
		FROM derived.algorithms
		WHERE kind = $1
			AND deleted_at IS NULL
		ORDER BY name ASC
		LIMIT 1
	`, kind).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
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
		FROM derived.simulation_cohorts c
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

	items := make([]cohortListItem, 0)
	for rows.Next() {
		var it cohortListItem
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

	writeJSON(w, http.StatusOK, listCohortsResponse{Items: items})
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

	var it cohortListItem
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
		FROM derived.simulation_cohorts c
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

	type patchCohortRequest struct {
		OptimizerKey      *string `json:"optimizerKey"`
		NSims             *int    `json:"nSims"`
		Seed              *int    `json:"seed"`
		StartingStateKey  *string `json:"startingStateKey"`
		ExcludedEntryName *string `json:"excludedEntryName"`
	}

	var req patchCohortRequest
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
		UPDATE derived.simulation_cohorts
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
		writeError(w, r, http.StatusNotFound, "not_found", "Cohort not found", "id")
		return
	}

	s.getSyntheticCalcuttaCohortHandler(w, r)
}
