package httpserver

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type createStrategyGenerationRunRequest struct {
	CalcuttaID       string  `json:"calcuttaId"`
	SuiteScenarioID  *string `json:"suiteScenarioId"`
	Name             *string `json:"name"`
	OptimizerKey     *string `json:"optimizerKey"`
	MarketShareRunID *string `json:"marketShareRunId"`
}

type createStrategyGenerationRunResponse struct {
	StrategyGenerationRunID string  `json:"strategy_generation_run_id"`
	RunKey                  string  `json:"run_key"`
	SuiteScenarioID         *string `json:"suite_scenario_id,omitempty"`
	CalcuttaSnapshotID      *string `json:"calcutta_snapshot_id,omitempty"`
}

type strategyGenerationRunListItem struct {
	ID                    string  `json:"id"`
	RunKey                *string `json:"run_key,omitempty"`
	Name                  *string `json:"name,omitempty"`
	CalcuttaID            *string `json:"calcutta_id,omitempty"`
	SimulatedTournamentID *string `json:"simulated_tournament_id,omitempty"`
	Purpose               string  `json:"purpose"`
	ReturnsModelKey       string  `json:"returns_model_key"`
	InvestmentModelKey    string  `json:"investment_model_key"`
	OptimizerKey          string  `json:"optimizer_key"`
	CreatedAt             string  `json:"created_at"`
}

type listStrategyGenerationRunsResponse struct {
	Items []strategyGenerationRunListItem `json:"items"`
}

func (s *Server) registerStrategyGenerationRunRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/strategy-generation-runs",
		s.requirePermission("analytics.strategy_generation_runs.write", s.createStrategyGenerationRunHandler),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/strategy-generation-runs",
		s.requirePermission("analytics.strategy_generation_runs.read", s.listStrategyGenerationRunsHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/strategy-generation-runs/{id}",
		s.requirePermission("analytics.strategy_generation_runs.read", s.getStrategyGenerationRunHandler),
	).Methods("GET", "OPTIONS")
}

func (s *Server) createStrategyGenerationRunHandler(w http.ResponseWriter, r *http.Request) {
	var req createStrategyGenerationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	req.CalcuttaID = strings.TrimSpace(req.CalcuttaID)
	if req.CalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaId is required", "calcuttaId")
		return
	}
	if _, err := uuid.Parse(req.CalcuttaID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaId must be a valid UUID", "calcuttaId")
		return
	}

	suiteScenarioID := ""
	if req.SuiteScenarioID != nil {
		v := strings.TrimSpace(*req.SuiteScenarioID)
		if v != "" {
			if _, err := uuid.Parse(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "suiteScenarioId must be a valid UUID", "suiteScenarioId")
				return
			}
			suiteScenarioID = v
		}
	}

	name := "api_strategy_generation"
	if req.Name != nil {
		v := strings.TrimSpace(*req.Name)
		if v != "" {
			name = v
		}
	}

	optimizerKey := ""
	if req.OptimizerKey != nil {
		optimizerKey = strings.TrimSpace(*req.OptimizerKey)
	}

	marketShareRunID := (*string)(nil)
	if req.MarketShareRunID != nil {
		v := strings.TrimSpace(*req.MarketShareRunID)
		if v != "" {
			if _, err := uuid.Parse(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "marketShareRunId must be a valid UUID", "marketShareRunId")
				return
			}
			marketShareRunID = &v
		}
	}

	ctx := r.Context()
	runKeyUUID := uuid.NewString()
	runKeyText := runKeyUUID

	if optimizerKey == "" {
		optimizerKey = "minlp_v1"
	}

	params := map[string]any{}
	if marketShareRunID != nil {
		params["market_share_run_id"] = *marketShareRunID
	}
	params["source"] = "api_strategy_generation"
	paramsJSON, _ := json.Marshal(params)

	gitSHA := strings.TrimSpace(os.Getenv("GIT_SHA"))
	var gitSHAParam any
	if gitSHA != "" {
		gitSHAParam = gitSHA
	} else {
		gitSHAParam = nil
	}

	var runID string
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO derived.strategy_generation_runs (
			run_key,
			run_key_uuid,
			name,
			simulated_tournament_id,
			calcutta_id,
			purpose,
			returns_model_key,
			investment_model_key,
			optimizer_key,
			params_json,
			git_sha
		)
		VALUES ($1, $2::uuid, $3, NULL, $4::uuid, 'go_recommended_entry_bids', 'legacy', 'predicted_market_share', $5, $6::jsonb, $7)
		RETURNING id::text
	`, runKeyText, runKeyUUID, name, req.CalcuttaID, optimizerKey, string(paramsJSON), gitSHAParam).Scan(&runID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	resp := createStrategyGenerationRunResponse{
		StrategyGenerationRunID: runID,
		RunKey:                  runKeyText,
	}

	if suiteScenarioID != "" {
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

		var scenarioCalcuttaID string
		var excludedEntryName *string
		var startingStateKey *string
		if err := tx.QueryRow(ctx, `
			SELECT calcutta_id::text, excluded_entry_name, starting_state_key
			FROM derived.suite_scenarios
			WHERE id = $1::uuid
				AND deleted_at IS NULL
			LIMIT 1
		`, suiteScenarioID).Scan(&scenarioCalcuttaID, &excludedEntryName, &startingStateKey); err != nil {
			if err == pgx.ErrNoRows {
				writeError(w, r, http.StatusNotFound, "not_found", "Suite scenario not found", "suiteScenarioId")
				return
			}
			writeErrorFromErr(w, r, err)
			return
		}
		if scenarioCalcuttaID != req.CalcuttaID {
			writeError(w, r, http.StatusBadRequest, "validation_error", "suiteScenarioId does not match calcuttaId", "suiteScenarioId")
			return
		}

		snapshotID, err := createSuiteScenarioSnapshot(ctx, tx, scenarioCalcuttaID, excludedEntryName, runID, nil)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		_, err = tx.Exec(ctx, `
			UPDATE derived.suite_scenarios
			SET focus_strategy_generation_run_id = $2::uuid,
				calcutta_snapshot_id = $3::uuid,
				focus_entry_name = COALESCE(focus_entry_name, 'Our Strategy'),
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, suiteScenarioID, runID, snapshotID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		if err := tx.Commit(ctx); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		committed = true

		resp.SuiteScenarioID = &suiteScenarioID
		resp.CalcuttaSnapshotID = &snapshotID
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) listStrategyGenerationRunsHandler(w http.ResponseWriter, r *http.Request) {
	calcuttaID := strings.TrimSpace(r.URL.Query().Get("calcutta_id"))
	if calcuttaID != "" {
		if _, err := uuid.Parse(calcuttaID); err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "calcutta_id must be a valid UUID", "calcutta_id")
			return
		}
	}

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
			id::text,
			run_key,
			name,
			calcutta_id::text,
			simulated_tournament_id::text,
			purpose,
			returns_model_key,
			investment_model_key,
			optimizer_key,
			created_at::text
		FROM derived.strategy_generation_runs
		WHERE deleted_at IS NULL
			AND ($1::uuid IS NULL OR calcutta_id = $1::uuid)
		ORDER BY created_at DESC
		LIMIT $2::int
		OFFSET $3::int
	`, nullUUIDParam(calcuttaID), limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]strategyGenerationRunListItem, 0)
	for rows.Next() {
		var it strategyGenerationRunListItem
		if err := rows.Scan(
			&it.ID,
			&it.RunKey,
			&it.Name,
			&it.CalcuttaID,
			&it.SimulatedTournamentID,
			&it.Purpose,
			&it.ReturnsModelKey,
			&it.InvestmentModelKey,
			&it.OptimizerKey,
			&it.CreatedAt,
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

	writeJSON(w, http.StatusOK, listStrategyGenerationRunsResponse{Items: items})
}

func (s *Server) getStrategyGenerationRunHandler(w http.ResponseWriter, r *http.Request) {
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

	var it strategyGenerationRunListItem
	if err := s.pool.QueryRow(r.Context(), `
		SELECT
			id::text,
			run_key,
			name,
			calcutta_id::text,
			simulated_tournament_id::text,
			purpose,
			returns_model_key,
			investment_model_key,
			optimizer_key,
			created_at::text
		FROM derived.strategy_generation_runs
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.RunKey,
		&it.Name,
		&it.CalcuttaID,
		&it.SimulatedTournamentID,
		&it.Purpose,
		&it.ReturnsModelKey,
		&it.InvestmentModelKey,
		&it.OptimizerKey,
		&it.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Strategy generation run not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, it)
}
