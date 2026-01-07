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

type syntheticEntryTeam struct {
	TeamID    string `json:"team_id"`
	BidPoints int    `json:"bid_points"`
}

type syntheticEntryListItem struct {
	ID          string               `json:"id"`
	EntryID     *string              `json:"entry_id,omitempty"`
	DisplayName string               `json:"display_name"`
	IsSynthetic bool                 `json:"is_synthetic"`
	Teams       []syntheticEntryTeam `json:"teams"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type listSyntheticEntriesResponse struct {
	Items []syntheticEntryListItem `json:"items"`
}

type createSyntheticEntryRequest struct {
	DisplayName string               `json:"displayName"`
	Teams       []syntheticEntryTeam `json:"teams"`
}

type createSyntheticEntryResponse struct {
	ID string `json:"id"`
}

type importSyntheticEntryRequest struct {
	StrategyGenerationRunID string  `json:"strategyGenerationRunId"`
	DisplayName             *string `json:"displayName"`
}

type importSyntheticEntryResponse struct {
	ID     string `json:"id"`
	NTeams int    `json:"nTeams"`
}

type patchSyntheticEntryRequest struct {
	DisplayName *string               `json:"displayName"`
	Teams       *[]syntheticEntryTeam `json:"teams"`
}

func (s *Server) registerSyntheticEntryRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}/synthetic-entries",
		s.requirePermission("analytics.suite_scenarios.read", s.handleListSyntheticEntries),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}/synthetic-entries",
		s.requirePermission("analytics.suite_scenarios.write", s.handleCreateSyntheticEntry),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}/synthetic-entries/import",
		s.requirePermission("analytics.suite_scenarios.write", s.handleImportSyntheticEntry),
	).Methods("POST", "OPTIONS")
	// TODO: prefer nested routes long-term; keep flat resource routes for now.
	r.HandleFunc(
		"/api/synthetic-entries/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handlePatchSyntheticEntry),
	).Methods("PATCH", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-entries/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handleDeleteSyntheticEntry),
	).Methods("DELETE", "OPTIONS")
}

func (s *Server) handleListSyntheticEntries(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	syntheticCalcuttaID := strings.TrimSpace(vars["id"])
	if syntheticCalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(syntheticCalcuttaID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	var snapshotID *string
	if err := s.pool.QueryRow(r.Context(), `
		SELECT calcutta_snapshot_id::text
		FROM derived.synthetic_calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, syntheticCalcuttaID).Scan(&snapshotID); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic calcutta not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	if snapshotID == nil || strings.TrimSpace(*snapshotID) == "" {
		writeError(w, r, http.StatusConflict, "invalid_state", "Synthetic calcutta has no snapshot", "id")
		return
	}

	rows, err := s.pool.Query(r.Context(), `
		SELECT
			e.id::text,
			e.entry_id::text,
			e.display_name,
			e.is_synthetic,
			e.created_at,
			e.updated_at,
			et.team_id::text,
			et.bid_points
		FROM core.calcutta_snapshot_entries e
		LEFT JOIN core.calcutta_snapshot_entry_teams et
			ON et.calcutta_snapshot_entry_id = e.id
		WHERE e.calcutta_snapshot_id = $1::uuid
			AND e.deleted_at IS NULL
		ORDER BY e.created_at ASC, et.bid_points DESC
	`, *snapshotID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	byID := make(map[string]*syntheticEntryListItem)
	order := make([]string, 0)
	for rows.Next() {
		var entryID string
		var sourceEntryID *string
		var displayName string
		var isSynthetic bool
		var createdAt time.Time
		var updatedAt time.Time
		var teamID *string
		var bidPoints *int

		if err := rows.Scan(
			&entryID,
			&sourceEntryID,
			&displayName,
			&isSynthetic,
			&createdAt,
			&updatedAt,
			&teamID,
			&bidPoints,
		); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		it, ok := byID[entryID]
		if !ok {
			it = &syntheticEntryListItem{
				ID:          entryID,
				EntryID:     sourceEntryID,
				DisplayName: displayName,
				IsSynthetic: isSynthetic,
				Teams:       make([]syntheticEntryTeam, 0),
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			}
			byID[entryID] = it
			order = append(order, entryID)
		}

		if teamID != nil && bidPoints != nil {
			it.Teams = append(it.Teams, syntheticEntryTeam{TeamID: *teamID, BidPoints: *bidPoints})
		}
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	items := make([]syntheticEntryListItem, 0, len(order))
	for _, id := range order {
		items = append(items, *byID[id])
	}

	writeJSON(w, http.StatusOK, listSyntheticEntriesResponse{Items: items})
}

func (s *Server) handleCreateSyntheticEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	syntheticCalcuttaID := strings.TrimSpace(vars["id"])
	if syntheticCalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(syntheticCalcuttaID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	var req createSyntheticEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.DisplayName == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "displayName is required", "displayName")
		return
	}
	for i := range req.Teams {
		req.Teams[i].TeamID = strings.TrimSpace(req.Teams[i].TeamID)
		if req.Teams[i].TeamID == "" {
			writeError(w, r, http.StatusBadRequest, "validation_error", "teams.team_id is required", "teams")
			return
		}
		if _, err := uuid.Parse(req.Teams[i].TeamID); err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "teams.team_id must be a valid UUID", "teams")
			return
		}
		if req.Teams[i].BidPoints <= 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "teams.bid_points must be positive", "teams")
			return
		}
	}

	ctx := r.Context()
	var snapshotID *string
	if err := s.pool.QueryRow(ctx, `
		SELECT calcutta_snapshot_id::text
		FROM derived.synthetic_calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, syntheticCalcuttaID).Scan(&snapshotID); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic calcutta not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	if snapshotID == nil || strings.TrimSpace(*snapshotID) == "" {
		writeError(w, r, http.StatusConflict, "invalid_state", "Synthetic calcutta has no snapshot", "id")
		return
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

	var entryID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
		VALUES ($1::uuid, NULL, $2, true)
		RETURNING id::text
	`, *snapshotID, req.DisplayName).Scan(&entryID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	for _, t := range req.Teams {
		if _, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			VALUES ($1::uuid, $2::uuid, $3::int)
		`, entryID, t.TeamID, t.BidPoints); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

	writeJSON(w, http.StatusCreated, createSyntheticEntryResponse{ID: entryID})
}

func (s *Server) handleImportSyntheticEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	syntheticCalcuttaID := strings.TrimSpace(vars["id"])
	if syntheticCalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(syntheticCalcuttaID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	var req importSyntheticEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	req.StrategyGenerationRunID = strings.TrimSpace(req.StrategyGenerationRunID)
	if req.StrategyGenerationRunID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "strategyGenerationRunId is required", "strategyGenerationRunId")
		return
	}
	if _, err := uuid.Parse(req.StrategyGenerationRunID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "strategyGenerationRunId must be a valid UUID", "strategyGenerationRunId")
		return
	}

	ctx := r.Context()

	var snapshotID *string
	if err := s.pool.QueryRow(ctx, `
		SELECT calcutta_snapshot_id::text
		FROM derived.synthetic_calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, syntheticCalcuttaID).Scan(&snapshotID); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic calcutta not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	if snapshotID == nil || strings.TrimSpace(*snapshotID) == "" {
		writeError(w, r, http.StatusConflict, "invalid_state", "Synthetic calcutta has no snapshot", "id")
		return
	}

	displayName := ""
	if req.DisplayName != nil {
		displayName = strings.TrimSpace(*req.DisplayName)
	}
	if displayName == "" {
		var resolved string
		_ = s.pool.QueryRow(ctx, `
			SELECT COALESCE(name, ''::text)
			FROM derived.strategy_generation_runs
			WHERE id = $1::uuid
				AND deleted_at IS NULL
			LIMIT 1
		`, req.StrategyGenerationRunID).Scan(&resolved)
		resolved = strings.TrimSpace(resolved)
		if resolved == "" {
			displayName = "Imported Strategy"
		} else {
			displayName = resolved
		}
	}

	rows, err := s.pool.Query(ctx, `
		SELECT team_id::text, bid_points::int
		FROM derived.recommended_entry_bids
		WHERE strategy_generation_run_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY bid_points DESC
	`, req.StrategyGenerationRunID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	teams := make([]syntheticEntryTeam, 0)
	for rows.Next() {
		var t syntheticEntryTeam
		if err := rows.Scan(&t.TeamID, &t.BidPoints); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		teams = append(teams, t)
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if len(teams) == 0 {
		writeError(w, r, http.StatusConflict, "invalid_state", "No recommended_entry_bids found for that strategyGenerationRunId", "strategyGenerationRunId")
		return
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

	var entryID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
		VALUES ($1::uuid, NULL, $2, true)
		RETURNING id::text
	`, *snapshotID, displayName).Scan(&entryID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	for _, t := range teams {
		if _, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			VALUES ($1::uuid, $2::uuid, $3::int)
		`, entryID, t.TeamID, t.BidPoints); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

	writeJSON(w, http.StatusCreated, importSyntheticEntryResponse{ID: entryID, NTeams: len(teams)})
}

func (s *Server) handlePatchSyntheticEntry(w http.ResponseWriter, r *http.Request) {
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

	var req patchSyntheticEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	displayName := (*string)(nil)
	if req.DisplayName != nil {
		v := strings.TrimSpace(*req.DisplayName)
		if v == "" {
			writeError(w, r, http.StatusBadRequest, "validation_error", "displayName must be non-empty", "displayName")
			return
		}
		displayName = &v
	}

	ctx := r.Context()
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

	if displayName != nil {
		ct, err := tx.Exec(ctx, `
			UPDATE core.calcutta_snapshot_entries
			SET display_name = $2,
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, id, *displayName)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if ct.RowsAffected() == 0 {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic entry not found", "id")
			return
		}
	}

	if req.Teams != nil {
		teams := *req.Teams
		for i := range teams {
			teams[i].TeamID = strings.TrimSpace(teams[i].TeamID)
			if teams[i].TeamID == "" {
				writeError(w, r, http.StatusBadRequest, "validation_error", "teams.team_id is required", "teams")
				return
			}
			if _, err := uuid.Parse(teams[i].TeamID); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "teams.team_id must be a valid UUID", "teams")
				return
			}
			if teams[i].BidPoints <= 0 {
				writeError(w, r, http.StatusBadRequest, "validation_error", "teams.bid_points must be positive", "teams")
				return
			}
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM core.calcutta_snapshot_entry_teams
			WHERE calcutta_snapshot_entry_id = $1::uuid
		`, id); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		for _, t := range teams {
			if _, err := tx.Exec(ctx, `
				INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
				VALUES ($1::uuid, $2::uuid, $3::int)
			`, id, t.TeamID, t.BidPoints); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleDeleteSyntheticEntry(w http.ResponseWriter, r *http.Request) {
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

	ct, err := s.pool.Exec(r.Context(), `
		UPDATE core.calcutta_snapshot_entries
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1::uuid
			AND deleted_at IS NULL
	`, id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if ct.RowsAffected() == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "Synthetic entry not found", "id")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
