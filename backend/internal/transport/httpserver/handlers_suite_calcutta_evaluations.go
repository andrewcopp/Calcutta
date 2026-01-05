package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/gorilla/mux"
)

type createSuiteCalcuttaEvaluationResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type suiteCalcuttaEvaluationListItem struct {
	ID                      string     `json:"id"`
	SuiteID                 string     `json:"suite_id"`
	SuiteName               string     `json:"suite_name"`
	OptimizerKey            string     `json:"optimizer_key"`
	NSims                   int        `json:"n_sims"`
	Seed                    int        `json:"seed"`
	CalcuttaID              string     `json:"calcutta_id"`
	GameOutcomeRunID        *string    `json:"game_outcome_run_id,omitempty"`
	MarketShareRunID        *string    `json:"market_share_run_id,omitempty"`
	StrategyGenerationRunID *string    `json:"strategy_generation_run_id,omitempty"`
	CalcuttaEvaluationRunID *string    `json:"calcutta_evaluation_run_id,omitempty"`
	StartingStateKey        string     `json:"starting_state_key"`
	ExcludedEntryName       *string    `json:"excluded_entry_name,omitempty"`
	Status                  string     `json:"status"`
	ClaimedAt               *time.Time `json:"claimed_at,omitempty"`
	ClaimedBy               *string    `json:"claimed_by,omitempty"`
	ErrorMessage            *string    `json:"error_message,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

type suiteCalcuttaEvaluationListResponse struct {
	Items []suiteCalcuttaEvaluationListItem `json:"items"`
}

func (s *Server) registerSuiteCalcuttaEvaluationRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/suite-calcutta-evaluations",
		s.requirePermission("analytics.suite_calcutta_evaluations.write", s.createSuiteCalcuttaEvaluationHandler),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/suite-calcutta-evaluations",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.listSuiteCalcuttaEvaluationsHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/suite-calcutta-evaluations/{id}",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.getSuiteCalcuttaEvaluationHandler),
	).Methods("GET", "OPTIONS")
}

func (s *Server) createSuiteCalcuttaEvaluationHandler(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateSuiteCalcuttaEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	ctx := r.Context()

	// Resolve (or create) suite_id.
	suiteID := ""
	if req.SuiteID != nil {
		suiteID = *req.SuiteID
	} else {
		// Resolve algorithm ids from runs.
		var goAlgID string
		if err := s.pool.QueryRow(ctx, `
			SELECT algorithm_id::text
			FROM derived.game_outcome_runs
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, *req.GameOutcomeRunID).Scan(&goAlgID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		var msAlgID string
		if err := s.pool.QueryRow(ctx, `
			SELECT algorithm_id::text
			FROM derived.market_share_runs
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, *req.MarketShareRunID).Scan(&msAlgID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		var insertedID string
		if err := s.pool.QueryRow(ctx, `
			INSERT INTO derived.suites (
				name,
				description,
				game_outcomes_algorithm_id,
				market_share_algorithm_id,
				optimizer_key,
				n_sims,
				seed,
				params_json
			)
			VALUES ($1, NULL, $2::uuid, $3::uuid, $4, $5, $6, '{}'::jsonb)
			ON CONFLICT (name) WHERE deleted_at IS NULL
			DO UPDATE SET
				game_outcomes_algorithm_id = EXCLUDED.game_outcomes_algorithm_id,
				market_share_algorithm_id = EXCLUDED.market_share_algorithm_id,
				optimizer_key = EXCLUDED.optimizer_key,
				n_sims = EXCLUDED.n_sims,
				seed = EXCLUDED.seed,
				updated_at = NOW(),
				deleted_at = NULL
			RETURNING id
		`, *req.SuiteName, goAlgID, msAlgID, *req.OptimizerKey, req.NSims, req.Seed).Scan(&insertedID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		suiteID = insertedID
	}

	var evalID string
	var status string
	var goRun any
	if req.GameOutcomeRunID != nil {
		goRun = *req.GameOutcomeRunID
	} else {
		goRun = nil
	}
	var msRun any
	if req.MarketShareRunID != nil {
		msRun = *req.MarketShareRunID
	} else {
		msRun = nil
	}
	var excluded any
	if req.ExcludedEntryName != nil {
		excluded = *req.ExcludedEntryName
	} else {
		excluded = nil
	}

	q := `
		INSERT INTO derived.suite_calcutta_evaluations (
			suite_id,
			calcutta_id,
			game_outcome_run_id,
			market_share_run_id,
			starting_state_key,
			excluded_entry_name
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6::text)
		RETURNING id, status
	`
	if err := s.pool.QueryRow(ctx, q, suiteID, req.CalcuttaID, goRun, msRun, req.StartingStateKey, excluded).Scan(&evalID, &status); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, createSuiteCalcuttaEvaluationResponse{ID: evalID, Status: status})
}

func (s *Server) listSuiteCalcuttaEvaluationsHandler(w http.ResponseWriter, r *http.Request) {
	calcuttaID := r.URL.Query().Get("calcutta_id")
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
			r.id,
			r.suite_id,
			COALESCE(s.name, '') AS suite_name,
			COALESCE(s.optimizer_key, '') AS optimizer_key,
			COALESCE(s.n_sims, 0) AS n_sims,
			COALESCE(s.seed, 0) AS seed,
			r.calcutta_id,
			r.game_outcome_run_id,
			r.market_share_run_id,
			r.strategy_generation_run_id,
			r.calcutta_evaluation_run_id,
			r.starting_state_key,
			r.excluded_entry_name,
			r.status,
			r.claimed_at,
			r.claimed_by,
			r.error_message,
			r.created_at,
			r.updated_at
		FROM derived.suite_calcutta_evaluations r
		LEFT JOIN derived.suites s
			ON s.id = r.suite_id
			AND s.deleted_at IS NULL
		WHERE r.deleted_at IS NULL
			AND ($1::uuid IS NULL OR r.calcutta_id = $1::uuid)
			AND ($2::uuid IS NULL OR r.suite_id = $2::uuid)
		ORDER BY r.created_at DESC
		LIMIT $3::int
		OFFSET $4::int
	`, nullUUIDParam(calcuttaID), nullUUIDParam(suiteID), limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]suiteCalcuttaEvaluationListItem, 0)
	for rows.Next() {
		var it suiteCalcuttaEvaluationListItem
		if err := rows.Scan(
			&it.ID,
			&it.SuiteID,
			&it.SuiteName,
			&it.OptimizerKey,
			&it.NSims,
			&it.Seed,
			&it.CalcuttaID,
			&it.GameOutcomeRunID,
			&it.MarketShareRunID,
			&it.StrategyGenerationRunID,
			&it.CalcuttaEvaluationRunID,
			&it.StartingStateKey,
			&it.ExcludedEntryName,
			&it.Status,
			&it.ClaimedAt,
			&it.ClaimedBy,
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

	writeJSON(w, http.StatusOK, suiteCalcuttaEvaluationListResponse{Items: items})
}

func (s *Server) getSuiteCalcuttaEvaluationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	var it suiteCalcuttaEvaluationListItem
	err := s.pool.QueryRow(r.Context(), `
		SELECT
			r.id,
			r.suite_id,
			COALESCE(s.name, '') AS suite_name,
			COALESCE(s.optimizer_key, '') AS optimizer_key,
			COALESCE(s.n_sims, 0) AS n_sims,
			COALESCE(s.seed, 0) AS seed,
			r.calcutta_id,
			r.game_outcome_run_id,
			r.market_share_run_id,
			r.strategy_generation_run_id,
			r.calcutta_evaluation_run_id,
			r.starting_state_key,
			r.excluded_entry_name,
			r.status,
			r.claimed_at,
			r.claimed_by,
			r.error_message,
			r.created_at,
			r.updated_at
		FROM derived.suite_calcutta_evaluations r
		LEFT JOIN derived.suites s
			ON s.id = r.suite_id
			AND s.deleted_at IS NULL
		WHERE r.id = $1::uuid
			AND r.deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.SuiteID,
		&it.SuiteName,
		&it.OptimizerKey,
		&it.NSims,
		&it.Seed,
		&it.CalcuttaID,
		&it.GameOutcomeRunID,
		&it.MarketShareRunID,
		&it.StrategyGenerationRunID,
		&it.CalcuttaEvaluationRunID,
		&it.StartingStateKey,
		&it.ExcludedEntryName,
		&it.Status,
		&it.ClaimedAt,
		&it.ClaimedBy,
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

func nullUUIDParam(v string) any {
	if v == "" {
		return nil
	}
	return v
}
