package httpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

func (s *Server) handleListAlgorithms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var kind *string
	if v := r.URL.Query().Get("kind"); v != "" {
		kind = &v
	}

	items, err := s.app.Analytics.ListAlgorithms(ctx, kind)
	if err != nil {
		log.Printf("Error listing algorithms: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to list algorithms", "")
		return
	}

	data := make([]map[string]interface{}, 0, len(items))
	for _, a := range items {
		var params interface{} = nil
		if len(a.ParamsJSON) > 0 {
			params = json.RawMessage(a.ParamsJSON)
		}
		data = append(data, map[string]interface{}{
			"id":          a.ID,
			"kind":        a.Kind,
			"name":        a.Name,
			"description": a.Description,
			"params_json": params,
			"created_at":  a.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"kind":  kind,
		"items": data,
		"count": len(data),
	})
}

func (s *Server) handleListGameOutcomeRunsForTournament(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
		return
	}

	runs, err := s.app.Analytics.ListGameOutcomeRunsByTournamentID(ctx, tournamentID)
	if err != nil {
		log.Printf("Error listing game outcome runs: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to list game outcome runs", "")
		return
	}

	data := make([]map[string]interface{}, 0, len(runs))
	for _, run := range runs {
		var params interface{} = nil
		if len(run.ParamsJSON) > 0 {
			params = json.RawMessage(run.ParamsJSON)
		}
		data = append(data, map[string]interface{}{
			"id":            run.ID,
			"algorithm_id":  run.AlgorithmID,
			"tournament_id": run.TournamentID,
			"params_json":   params,
			"git_sha":       run.GitSHA,
			"created_at":    run.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id": tournamentID,
		"runs":          data,
		"count":         len(data),
	})
}

func (s *Server) handleListMarketShareRunsForCalcutta(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	runs, err := s.app.Analytics.ListMarketShareRunsByCalcuttaID(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error listing market share runs: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to list market share runs", "")
		return
	}

	data := make([]map[string]interface{}, 0, len(runs))
	for _, run := range runs {
		var params interface{} = nil
		if len(run.ParamsJSON) > 0 {
			params = json.RawMessage(run.ParamsJSON)
		}
		data = append(data, map[string]interface{}{
			"id":           run.ID,
			"algorithm_id": run.AlgorithmID,
			"calcutta_id":  run.CalcuttaID,
			"params_json":  params,
			"git_sha":      run.GitSHA,
			"created_at":   run.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id": calcuttaID,
		"runs":        data,
		"count":       len(data),
	})
}

func (s *Server) handleGetLatestPredictionRunsForCalcutta(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	latest, err := s.app.Analytics.GetLatestPredictionRunsForCalcutta(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error getting latest prediction runs: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to get latest prediction runs", "")
		return
	}
	if latest == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "No prediction run metadata found for calcutta", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":         calcuttaID,
		"tournament_id":       latest.TournamentID,
		"game_outcome_run_id": latest.GameOutcomeRunID,
		"market_share_run_id": latest.MarketShareRunID,
	})
}

func (s *Server) handleBulkCreateGameOutcomeRunsForAlgorithm(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		writeError(w, r, http.StatusServiceUnavailable, "unavailable", "Database not available", "")
		return
	}

	vars := mux.Vars(r)
	algorithmID := strings.TrimSpace(vars["id"])
	if algorithmID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing algorithm ID", "id")
		return
	}
	if _, err := uuid.Parse(algorithmID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Algorithm ID must be a valid UUID", "id")
		return
	}

	ctx := r.Context()

	var exists bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM derived.algorithms a
			WHERE a.id = $1::uuid
				AND a.deleted_at IS NULL
			LIMIT 1
		)
	`, algorithmID).Scan(&exists); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if !exists {
		writeError(w, r, http.StatusNotFound, "not_found", "Algorithm not found", "id")
		return
	}

	gitSHA := strings.TrimSpace(os.Getenv("GIT_SHA"))
	var gitSHAParam any
	if gitSHA != "" {
		gitSHAParam = gitSHA
	} else {
		gitSHAParam = nil
	}

	paramsJSON, _ := json.Marshal(map[string]any{"source": "api_bulk"})

	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM core.tournaments
		WHERE deleted_at IS NULL
	`).Scan(&total); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	// Bulk insert one run per tournament (skip if a run already exists for algorithm+tournament).
	cmdTag, err := s.pool.Exec(ctx, `
		INSERT INTO derived.game_outcome_runs (
			algorithm_id,
			tournament_id,
			params_json,
			git_sha
		)
		SELECT
			$1::uuid,
			t.id,
			$2::jsonb,
			$3
		FROM core.tournaments t
		WHERE t.deleted_at IS NULL
			AND NOT EXISTS (
				SELECT 1
				FROM derived.game_outcome_runs r
				WHERE r.algorithm_id = $1::uuid
					AND r.tournament_id = t.id
					AND r.deleted_at IS NULL
			)
	`, algorithmID, string(paramsJSON), gitSHAParam)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	created := int(cmdTag.RowsAffected())
	skipped := total - created
	if skipped < 0 {
		skipped = 0
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"algorithm_id":      algorithmID,
		"total_tournaments": total,
		"created":           created,
		"skipped":           skipped,
	})
}

func (s *Server) handleBulkCreateMarketShareRunsForAlgorithm(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		writeError(w, r, http.StatusServiceUnavailable, "unavailable", "Database not available", "")
		return
	}

	vars := mux.Vars(r)
	algorithmID := strings.TrimSpace(vars["id"])
	if algorithmID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing algorithm ID", "id")
		return
	}
	if _, err := uuid.Parse(algorithmID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Algorithm ID must be a valid UUID", "id")
		return
	}

	ctx := r.Context()

	var kind, name string
	if err := s.pool.QueryRow(ctx, `
		SELECT kind, name
		FROM derived.algorithms
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, algorithmID).Scan(&kind, &name); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Algorithm not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	if kind != "market_share" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Algorithm kind must be market_share", "id")
		return
	}

	// For now, only the ridge Python runner is supported for bulk execution.
	if name != "ridge" {
		writeError(w, r, http.StatusBadRequest, "unsupported_algorithm", "Bulk execution is only supported for market_share algorithm 'ridge'", "id")
		return
	}

	// Merge any JSON body params into the market_share_runs.params_json.
	params := map[string]any{"source": "api_bulk"}
	if r.Body != nil {
		defer r.Body.Close()
		var payload map[string]any
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&payload); err != nil && err != io.EOF {
			writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid JSON body", "")
			return
		}
		for k, v := range payload {
			params[k] = v
		}
	}

	excludedEntryName := ""
	if v, ok := params["excluded_entry_name"]; ok && v != nil {
		s, ok := v.(string)
		if ok {
			excludedEntryName = strings.TrimSpace(s)
		} else {
			excludedEntryName = strings.TrimSpace(fmt.Sprintf("%v", v))
		}
	}
	if excludedEntryName == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "excluded_entry_name is required", "excluded_entry_name")
		return
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	gitSHA := strings.TrimSpace(os.Getenv("GIT_SHA"))
	var gitSHAParam any
	if gitSHA != "" {
		gitSHAParam = gitSHA
	} else {
		gitSHAParam = nil
	}

	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM core.calcuttas
		WHERE deleted_at IS NULL
	`).Scan(&total); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	// Requeue failed jobs for existing runs so users can retry after fixing worker/runtime issues.
	// This avoids needing to delete runs or create duplicates when the first attempt failed due to infra (e.g. missing python deps).
	requeueTag, err := s.pool.Exec(ctx, `
		UPDATE derived.run_jobs j
		SET status = 'queued',
			claimed_at = NULL,
			claimed_by = NULL,
			finished_at = NULL,
			error_message = NULL,
			updated_at = NOW()
		WHERE j.run_kind = 'market_share'
			AND j.status = 'failed'
			AND j.run_id IN (
				SELECT r.id
				FROM derived.market_share_runs r
				WHERE r.algorithm_id = $1::uuid
					AND r.deleted_at IS NULL
					AND COALESCE(r.params_json->>'excluded_entry_name', '') = $2::text
			)
	`, algorithmID, excludedEntryName)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	requeued := int(requeueTag.RowsAffected())

	// Bulk insert one run per calcutta (skip if a matching run already exists for algorithm+calcutta+excluded_entry_name).
	cmdTag, err := s.pool.Exec(ctx, `
		INSERT INTO derived.market_share_runs (
			algorithm_id,
			calcutta_id,
			params_json,
			git_sha
		)
		SELECT
			$1::uuid,
			c.id,
			$2::jsonb,
			$3
		FROM core.calcuttas c
		WHERE c.deleted_at IS NULL
			AND NOT EXISTS (
				SELECT 1
				FROM derived.market_share_runs r
				WHERE r.algorithm_id = $1::uuid
					AND r.calcutta_id = c.id
					AND r.deleted_at IS NULL
					AND COALESCE(r.params_json->>'excluded_entry_name', '') = $4::text
			)
	`, algorithmID, string(paramsJSON), gitSHAParam, excludedEntryName)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	created := int(cmdTag.RowsAffected())
	skipped := total - created
	if skipped < 0 {
		skipped = 0
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"algorithm_id":        algorithmID,
		"total_calcuttas":     total,
		"created":             created,
		"skipped":             skipped,
		"requeued":            requeued,
		"excluded_entry_name": excludedEntryName,
	})
}
