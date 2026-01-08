package mlanalytics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) HandleBulkCreateGameOutcomeRunsForAlgorithm(w http.ResponseWriter, r *http.Request) {
	if h.pool == nil {
		httperr.Write(w, r, http.StatusServiceUnavailable, "unavailable", "Database not available", "")
		return
	}

	vars := mux.Vars(r)
	algorithmID := strings.TrimSpace(vars["id"])
	if algorithmID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing algorithm ID", "id")
		return
	}
	if _, err := uuid.Parse(algorithmID); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Algorithm ID must be a valid UUID", "id")
		return
	}

	ctx := r.Context()

	var exists bool
	if err := h.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM derived.algorithms a
			WHERE a.id = $1::uuid
				AND a.deleted_at IS NULL
			LIMIT 1
		)
	`, algorithmID).Scan(&exists); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !exists {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "Algorithm not found", "id")
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
	if err := h.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM core.tournaments
		WHERE deleted_at IS NULL
	`).Scan(&total); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	// Bulk insert one run per tournament (skip if a run already exists for algorithm+tournament).
	cmdTag, err := h.pool.Exec(ctx, `
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
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	created := int(cmdTag.RowsAffected())
	skipped := total - created
	if skipped < 0 {
		skipped = 0
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"algorithm_id":      algorithmID,
		"total_tournaments": total,
		"created":           created,
		"skipped":           skipped,
	})
}

func (h *Handler) HandleBulkCreateMarketShareRunsForAlgorithm(w http.ResponseWriter, r *http.Request) {
	if h.pool == nil {
		httperr.Write(w, r, http.StatusServiceUnavailable, "unavailable", "Database not available", "")
		return
	}

	vars := mux.Vars(r)
	algorithmID := strings.TrimSpace(vars["id"])
	if algorithmID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing algorithm ID", "id")
		return
	}
	if _, err := uuid.Parse(algorithmID); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Algorithm ID must be a valid UUID", "id")
		return
	}

	ctx := r.Context()

	var kind, name string
	if err := h.pool.QueryRow(ctx, `
		SELECT kind, name
		FROM derived.algorithms
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, algorithmID).Scan(&kind, &name); err != nil {
		if err == pgx.ErrNoRows {
			httperr.Write(w, r, http.StatusNotFound, "not_found", "Algorithm not found", "id")
			return
		}
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if kind != "market_share" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Algorithm kind must be market_share", "id")
		return
	}

	// For now, only the ridge Python runner is supported for bulk execution.
	if name != "ridge" {
		httperr.Write(w, r, http.StatusBadRequest, "unsupported_algorithm", "Bulk execution is only supported for market_share algorithm 'ridge'", "id")
		return
	}

	// Merge any JSON body params into the market_share_runs.params_json.
	params := map[string]any{"source": "api_bulk"}
	if r.Body != nil {
		defer r.Body.Close()
		var payload map[string]any
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&payload); err != nil && err != io.EOF {
			httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Invalid JSON body", "")
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
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "excluded_entry_name is required", "excluded_entry_name")
		return
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
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
	if err := h.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM core.calcuttas
		WHERE deleted_at IS NULL
	`).Scan(&total); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	// Requeue failed jobs for existing runs so users can retry after fixing worker/runtime issues.
	// This avoids needing to delete runs or create duplicates when the first attempt failed due to infra (e.g. missing python deps).
	requeueTag, err := h.pool.Exec(ctx, `
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
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	requeued := int(requeueTag.RowsAffected())

	// Bulk insert one run per calcutta (skip if a matching run already exists for algorithm+calcutta+excluded_entry_name).
	cmdTag, err := h.pool.Exec(ctx, `
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
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	created := int(cmdTag.RowsAffected())
	skipped := total - created
	if skipped < 0 {
		skipped = 0
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"algorithm_id":        algorithmID,
		"total_calcuttas":     total,
		"created":             created,
		"skipped":             skipped,
		"requeued":            requeued,
		"excluded_entry_name": excludedEntryName,
	})
}
