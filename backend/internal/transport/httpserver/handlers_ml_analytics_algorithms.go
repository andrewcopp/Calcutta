package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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
