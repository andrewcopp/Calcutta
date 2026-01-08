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
	ID            string               `json:"id"`
	CandidateID   string               `json:"candidate_id"`
	SnapshotEntry string               `json:"snapshot_entry_id"`
	EntryID       *string              `json:"entry_id,omitempty"`
	DisplayName   string               `json:"display_name"`
	IsSynthetic   bool                 `json:"is_synthetic"`
	Rank          *int                 `json:"rank"`
	Mean          *float64             `json:"mean_normalized_payout"`
	PTop1         *float64             `json:"p_top1"`
	PInMoney      *float64             `json:"p_in_money"`
	Teams         []syntheticEntryTeam `json:"teams"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
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
	EntryArtifactID string  `json:"entryArtifactId"`
	DisplayName     *string `json:"displayName"`
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

	// Candidate alias routes (preferred naming).
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}/candidates",
		s.requirePermission("analytics.suite_scenarios.read", s.handleListSyntheticEntries),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}/candidates",
		s.requirePermission("analytics.suite_scenarios.write", s.handleCreateSyntheticEntry),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}/candidates/import",
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

	// Candidate attachment alias routes.
	r.HandleFunc(
		"/api/synthetic-calcutta-candidates/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handlePatchSyntheticEntry),
	).Methods("PATCH", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcutta-candidates/{id}",
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

	ctx := r.Context()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var snapshotID *string
	if err := tx.QueryRow(ctx, `
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

	rows, err := tx.Query(ctx, `
		WITH latest_eval AS (
			SELECT sr.calcutta_evaluation_run_id
			FROM derived.simulation_runs sr
			WHERE sr.synthetic_calcutta_id = $2::uuid
				AND sr.deleted_at IS NULL
				AND sr.calcutta_evaluation_run_id IS NOT NULL
			ORDER BY sr.created_at DESC
			LIMIT 1
		),
		perf AS (
			SELECT
				ROW_NUMBER() OVER (ORDER BY COALESCE(ep.mean_normalized_payout, 0.0) DESC)::int AS rank,
				ep.entry_name,
				COALESCE(ep.mean_normalized_payout, 0.0)::double precision AS mean_normalized_payout,
				COALESCE(ep.p_top1, 0.0)::double precision AS p_top1,
				COALESCE(ep.p_in_money, 0.0)::double precision AS p_in_money
			FROM derived.entry_performance ep
			JOIN latest_eval le
				ON le.calcutta_evaluation_run_id = ep.calcutta_evaluation_run_id
			WHERE ep.deleted_at IS NULL
		)
		SELECT
			scc.id::text,
			scc.candidate_id::text,
			scc.snapshot_entry_id::text,
			e.entry_id::text,
			e.display_name,
			e.is_synthetic,
			e.created_at,
			e.updated_at,
			et.team_id::text,
			et.bid_points,
			p.rank,
			p.mean_normalized_payout,
			p.p_top1,
			p.p_in_money
		FROM derived.synthetic_calcutta_candidates scc
		JOIN derived.candidates c
			ON c.id = scc.candidate_id
			AND c.deleted_at IS NULL
		JOIN core.calcutta_snapshot_entries e
			ON e.id = scc.snapshot_entry_id
			AND e.calcutta_snapshot_id = $1::uuid
			AND e.deleted_at IS NULL
		LEFT JOIN core.calcutta_snapshot_entry_teams et
			ON et.calcutta_snapshot_entry_id = e.id
			AND et.deleted_at IS NULL
		LEFT JOIN perf p
			ON p.entry_name = e.display_name
		WHERE scc.synthetic_calcutta_id = $2::uuid
			AND scc.deleted_at IS NULL
		ORDER BY scc.created_at ASC, et.bid_points DESC
	`, *snapshotID, syntheticCalcuttaID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	byID := make(map[string]*syntheticEntryListItem)
	order := make([]string, 0)
	for rows.Next() {
		var attachmentID string
		var candidateID string
		var snapshotEntryID string
		var sourceEntryID *string
		var displayName string
		var isSynthetic bool
		var createdAt time.Time
		var updatedAt time.Time
		var teamID *string
		var bidPoints *int
		var rank *int
		var mean *float64
		var pTop1 *float64
		var pInMoney *float64

		if err := rows.Scan(
			&attachmentID,
			&candidateID,
			&snapshotEntryID,
			&sourceEntryID,
			&displayName,
			&isSynthetic,
			&createdAt,
			&updatedAt,
			&teamID,
			&bidPoints,
			&rank,
			&mean,
			&pTop1,
			&pInMoney,
		); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		it, ok := byID[attachmentID]
		if !ok {
			it = &syntheticEntryListItem{
				ID:            attachmentID,
				CandidateID:   candidateID,
				SnapshotEntry: snapshotEntryID,
				EntryID:       sourceEntryID,
				DisplayName:   displayName,
				IsSynthetic:   isSynthetic,
				Rank:          rank,
				Mean:          mean,
				PTop1:         pTop1,
				PInMoney:      pInMoney,
				Teams:         make([]syntheticEntryTeam, 0),
				CreatedAt:     createdAt,
				UpdatedAt:     updatedAt,
			}
			byID[attachmentID] = it
			order = append(order, attachmentID)
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

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
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

	metadata := map[string]any{"teams": req.Teams}
	metadataJSON, _ := json.Marshal(metadata)

	var candidateID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.candidates (source_kind, source_entry_artifact_id, display_name, metadata_json)
		VALUES ('manual', NULL, $1, $2::jsonb)
		RETURNING id::text
	`, req.DisplayName, string(metadataJSON)).Scan(&candidateID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	var snapshotEntryID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
		VALUES ($1::uuid, NULL, $2, true)
		RETURNING id::text
	`, *snapshotID, req.DisplayName).Scan(&snapshotEntryID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	for _, t := range req.Teams {
		if _, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			VALUES ($1::uuid, $2::uuid, $3::int)
		`, snapshotEntryID, t.TeamID, t.BidPoints); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	var attachmentID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.synthetic_calcutta_candidates (synthetic_calcutta_id, candidate_id, snapshot_entry_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		RETURNING id::text
	`, syntheticCalcuttaID, candidateID, snapshotEntryID).Scan(&attachmentID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

	writeJSON(w, http.StatusCreated, createSyntheticEntryResponse{ID: attachmentID})
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

	req.EntryArtifactID = strings.TrimSpace(req.EntryArtifactID)
	if req.EntryArtifactID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "entryArtifactId is required", "entryArtifactId")
		return
	}
	if _, err := uuid.Parse(req.EntryArtifactID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "entryArtifactId must be a valid UUID", "entryArtifactId")
		return
	}

	ctx := r.Context()

	strategyGenerationRunID := ""
	artifactKind := ""
	if err := s.pool.QueryRow(ctx, `
		SELECT run_id::text, artifact_kind
		FROM derived.run_artifacts
		WHERE id = $1::uuid
			AND run_kind = 'strategy_generation'
			AND deleted_at IS NULL
		LIMIT 1
	`, req.EntryArtifactID).Scan(&strategyGenerationRunID, &artifactKind); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Entry artifact not found", "entryArtifactId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	strategyGenerationRunID = strings.TrimSpace(strategyGenerationRunID)
	if strategyGenerationRunID == "" {
		writeError(w, r, http.StatusConflict, "invalid_state", "Entry artifact has no run_id", "entryArtifactId")
		return
	}
	if strings.TrimSpace(artifactKind) != "metrics" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "entryArtifactId must reference a metrics artifact", "entryArtifactId")
		return
	}

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
		`, strategyGenerationRunID).Scan(&resolved)
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
	`, strategyGenerationRunID)
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
		writeError(w, r, http.StatusConflict, "invalid_state", "No recommended_entry_bids found for that entryArtifactId", "entryArtifactId")
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

	var candidateID string
	if err := tx.QueryRow(ctx, `
		SELECT id::text
		FROM derived.candidates
		WHERE source_kind = 'entry_artifact'
			AND source_entry_artifact_id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, req.EntryArtifactID).Scan(&candidateID); err != nil {
		if err == pgx.ErrNoRows {
			if err := tx.QueryRow(ctx, `
				INSERT INTO derived.candidates (source_kind, source_entry_artifact_id, display_name, metadata_json)
				VALUES ('entry_artifact', $1::uuid, $2, '{}'::jsonb)
				RETURNING id::text
			`, req.EntryArtifactID, displayName).Scan(&candidateID); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
		} else {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	// If already attached, return existing attachment.
	var existingAttachmentID string
	if err := tx.QueryRow(ctx, `
		SELECT id::text
		FROM derived.synthetic_calcutta_candidates
		WHERE synthetic_calcutta_id = $1::uuid
			AND candidate_id = $2::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, syntheticCalcuttaID, candidateID).Scan(&existingAttachmentID); err == nil {
		if err := tx.Commit(ctx); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		committed = true
		writeJSON(w, http.StatusCreated, importSyntheticEntryResponse{ID: existingAttachmentID, NTeams: len(teams)})
		return
	} else if err != pgx.ErrNoRows {
		writeErrorFromErr(w, r, err)
		return
	}

	var snapshotEntryID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
		VALUES ($1::uuid, NULL, $2, true)
		RETURNING id::text
	`, *snapshotID, displayName).Scan(&snapshotEntryID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	for _, t := range teams {
		if _, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			VALUES ($1::uuid, $2::uuid, $3::int)
		`, snapshotEntryID, t.TeamID, t.BidPoints); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	var attachmentID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.synthetic_calcutta_candidates (synthetic_calcutta_id, candidate_id, snapshot_entry_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		RETURNING id::text
	`, syntheticCalcuttaID, candidateID, snapshotEntryID).Scan(&attachmentID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

	writeJSON(w, http.StatusCreated, importSyntheticEntryResponse{ID: attachmentID, NTeams: len(teams)})
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

	var candidateID string
	var snapshotEntryID string
	var sourceKind string
	if err := tx.QueryRow(ctx, `
		SELECT
			scc.candidate_id::text,
			scc.snapshot_entry_id::text,
			c.source_kind
		FROM derived.synthetic_calcutta_candidates scc
		JOIN derived.candidates c
			ON c.id = scc.candidate_id
			AND c.deleted_at IS NULL
		WHERE scc.id = $1::uuid
			AND scc.deleted_at IS NULL
		LIMIT 1
	`, id).Scan(&candidateID, &snapshotEntryID, &sourceKind); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic entry not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	if displayName != nil {
		if _, err := tx.Exec(ctx, `
			UPDATE derived.candidates
			SET display_name = $2,
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, candidateID, *displayName); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if _, err := tx.Exec(ctx, `
			UPDATE core.calcutta_snapshot_entries
			SET display_name = $2,
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, snapshotEntryID, *displayName); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	if req.Teams != nil {
		if sourceKind != "manual" {
			writeError(w, r, http.StatusBadRequest, "validation_error", "Only manual candidates can be edited", "teams")
			return
		}

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

		metadata := map[string]any{"teams": teams}
		metadataJSON, _ := json.Marshal(metadata)
		if _, err := tx.Exec(ctx, `
			UPDATE derived.candidates
			SET metadata_json = $2::jsonb,
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, candidateID, string(metadataJSON)); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM core.calcutta_snapshot_entry_teams
			WHERE calcutta_snapshot_entry_id = $1::uuid
		`, snapshotEntryID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		for _, t := range teams {
			if _, err := tx.Exec(ctx, `
				INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
				VALUES ($1::uuid, $2::uuid, $3::int)
			`, snapshotEntryID, t.TeamID, t.BidPoints); err != nil {
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

	ctx := r.Context()

	var snapshotEntryID string
	if err := s.pool.QueryRow(ctx, `
		SELECT snapshot_entry_id::text
		FROM derived.synthetic_calcutta_candidates
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, id).Scan(&snapshotEntryID); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic entry not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
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

	if _, err := tx.Exec(ctx, `
		UPDATE derived.synthetic_calcutta_candidates
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1::uuid
			AND deleted_at IS NULL
	`, id); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if _, err := tx.Exec(ctx, `
		UPDATE core.calcutta_snapshot_entries
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1::uuid
			AND deleted_at IS NULL
	`, snapshotEntryID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	// If this entry is highlighted anywhere, clear the highlight.
	if _, err := tx.Exec(ctx, `
		UPDATE derived.synthetic_calcuttas
		SET highlighted_snapshot_entry_id = NULL,
			updated_at = NOW()
		WHERE highlighted_snapshot_entry_id = $1::uuid
			AND deleted_at IS NULL
	`, snapshotEntryID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
