package httpserver

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type simulatedEntryTeam struct {
	TeamID    string `json:"team_id"`
	BidPoints int    `json:"bid_points"`
}

type simulatedEntryListItem struct {
	ID                  string               `json:"id"`
	SimulatedCalcuttaID string               `json:"simulated_calcutta_id"`
	DisplayName         string               `json:"display_name"`
	SourceKind          string               `json:"source_kind"`
	SourceEntryID       *string              `json:"source_entry_id,omitempty"`
	SourceCandidateID   *string              `json:"source_candidate_id,omitempty"`
	Teams               []simulatedEntryTeam `json:"teams"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

type listSimulatedEntriesResponse struct {
	Items []simulatedEntryListItem `json:"items"`
}

type createSimulatedEntryRequest struct {
	DisplayName string               `json:"displayName"`
	Teams       []simulatedEntryTeam `json:"teams"`
}

type createSimulatedEntryResponse struct {
	ID string `json:"simulatedEntryId"`
}

type patchSimulatedEntryRequest struct {
	DisplayName *string               `json:"displayName"`
	Teams       *[]simulatedEntryTeam `json:"teams"`
}

type importCandidateAsSimulatedEntryRequest struct {
	CandidateID string  `json:"candidateId"`
	DisplayName *string `json:"displayName"`
}

type importCandidateAsSimulatedEntryResponse struct {
	SimulatedEntryID string `json:"simulatedEntryId"`
	NTeams           int    `json:"nTeams"`
}

func (s *Server) registerSimulatedEntryRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/simulated-calcuttas/{id}/entries",
		s.requirePermission("analytics.suite_scenarios.read", s.handleListSimulatedEntries),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/simulated-calcuttas/{id}/entries",
		s.requirePermission("analytics.suite_scenarios.write", s.handleCreateSimulatedEntry),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/simulated-calcuttas/{simulatedCalcuttaId}/entries/{entryId}",
		s.requirePermission("analytics.suite_scenarios.write", s.handlePatchSimulatedEntry),
	).Methods("PATCH", "OPTIONS")
	r.HandleFunc(
		"/api/simulated-calcuttas/{simulatedCalcuttaId}/entries/{entryId}",
		s.requirePermission("analytics.suite_scenarios.write", s.handleDeleteSimulatedEntry),
	).Methods("DELETE", "OPTIONS")
	r.HandleFunc(
		"/api/simulated-calcuttas/{id}/entries/import-candidate",
		s.requirePermission("analytics.suite_scenarios.write", s.handleImportCandidateAsSimulatedEntry),
	).Methods("POST", "OPTIONS")
}

func (s *Server) handleListSimulatedEntries(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	simulatedCalcuttaID := strings.TrimSpace(vars["id"])
	if simulatedCalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(simulatedCalcuttaID); err != nil {
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

	var exists bool
	if err := tx.QueryRow(ctx, `
 		SELECT EXISTS(
 			SELECT 1
 			FROM derived.simulated_calcuttas
 			WHERE id = $1::uuid
 				AND deleted_at IS NULL
 		)
 	`, simulatedCalcuttaID).Scan(&exists); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if !exists {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulated calcutta not found", "id")
		return
	}

	rows, err := tx.Query(ctx, `
 		SELECT
 			e.id::text,
 			e.simulated_calcutta_id::text,
 			e.display_name,
 			e.source_kind,
 			e.source_entry_id::text,
 			e.source_candidate_id::text,
 			e.created_at,
 			e.updated_at,
 			et.team_id::text,
 			et.bid_points::int
 		FROM derived.simulated_entries e
 		LEFT JOIN derived.simulated_entry_teams et
 			ON et.simulated_entry_id = e.id
 			AND et.deleted_at IS NULL
 		WHERE e.simulated_calcutta_id = $1::uuid
 			AND e.deleted_at IS NULL
 		ORDER BY e.created_at ASC, et.bid_points DESC
 	`, simulatedCalcuttaID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	byID := make(map[string]*simulatedEntryListItem)
	order := make([]string, 0)
	for rows.Next() {
		var entryID string
		var scID string
		var displayName string
		var sourceKind string
		var sourceEntryID *string
		var sourceCandidateID *string
		var createdAt time.Time
		var updatedAt time.Time
		var teamID *string
		var bidPoints *int

		if err := rows.Scan(
			&entryID,
			&scID,
			&displayName,
			&sourceKind,
			&sourceEntryID,
			&sourceCandidateID,
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
			it = &simulatedEntryListItem{
				ID:                  entryID,
				SimulatedCalcuttaID: scID,
				DisplayName:         displayName,
				SourceKind:          sourceKind,
				SourceEntryID:       sourceEntryID,
				SourceCandidateID:   sourceCandidateID,
				Teams:               make([]simulatedEntryTeam, 0),
				CreatedAt:           createdAt,
				UpdatedAt:           updatedAt,
			}
			byID[entryID] = it
			order = append(order, entryID)
		}
		if teamID != nil && bidPoints != nil {
			it.Teams = append(it.Teams, simulatedEntryTeam{TeamID: *teamID, BidPoints: *bidPoints})
		}
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	items := make([]simulatedEntryListItem, 0, len(order))
	for _, id := range order {
		items = append(items, *byID[id])
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, listSimulatedEntriesResponse{Items: items})
}

func (s *Server) handleCreateSimulatedEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	simulatedCalcuttaID := strings.TrimSpace(vars["id"])
	if simulatedCalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(simulatedCalcuttaID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	var req createSimulatedEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.DisplayName == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "displayName is required", "displayName")
		return
	}
	if len(req.Teams) == 0 {
		writeError(w, r, http.StatusBadRequest, "validation_error", "teams is required", "teams")
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

	var exists bool
	if err := tx.QueryRow(ctx, `
 		SELECT EXISTS(
 			SELECT 1
 			FROM derived.simulated_calcuttas
 			WHERE id = $1::uuid
 				AND deleted_at IS NULL
 		)
 	`, simulatedCalcuttaID).Scan(&exists); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if !exists {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulated calcutta not found", "id")
		return
	}

	var entryID string
	if err := tx.QueryRow(ctx, `
 		INSERT INTO derived.simulated_entries (
 			simulated_calcutta_id,
 			display_name,
 			source_kind,
 			source_entry_id,
 			source_candidate_id
 		)
 		VALUES ($1::uuid, $2, 'manual', NULL, NULL)
 		RETURNING id::text
 	`, simulatedCalcuttaID, req.DisplayName).Scan(&entryID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	for _, t := range req.Teams {
		if _, err := tx.Exec(ctx, `
 			INSERT INTO derived.simulated_entry_teams (
 				simulated_entry_id,
 				team_id,
 				bid_points
 			)
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

	writeJSON(w, http.StatusCreated, createSimulatedEntryResponse{ID: entryID})
}

func (s *Server) handlePatchSimulatedEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	simulatedCalcuttaID := strings.TrimSpace(vars["simulatedCalcuttaId"])
	entryID := strings.TrimSpace(vars["entryId"])
	if simulatedCalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "simulatedCalcuttaId is required", "simulatedCalcuttaId")
		return
	}
	if _, err := uuid.Parse(simulatedCalcuttaID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "simulatedCalcuttaId must be a valid UUID", "simulatedCalcuttaId")
		return
	}
	if entryID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "entryId is required", "entryId")
		return
	}
	if _, err := uuid.Parse(entryID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "entryId must be a valid UUID", "entryId")
		return
	}

	var req patchSimulatedEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if req.DisplayName != nil {
		v := strings.TrimSpace(*req.DisplayName)
		req.DisplayName = &v
		if v == "" {
			req.DisplayName = nil
		}
	}
	if req.Teams != nil {
		teams := *req.Teams
		if len(teams) == 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "teams cannot be empty", "teams")
			return
		}
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
		*req.Teams = teams
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

	var existingSimulatedCalcuttaID string
	if err := tx.QueryRow(ctx, `
 		SELECT simulated_calcutta_id::text
 		FROM derived.simulated_entries
 		WHERE id = $1::uuid
 			AND deleted_at IS NULL
 		LIMIT 1
 	`, entryID).Scan(&existingSimulatedCalcuttaID); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Simulated entry not found", "entryId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	if existingSimulatedCalcuttaID != simulatedCalcuttaID {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulated entry not found", "entryId")
		return
	}

	if req.DisplayName != nil {
		if _, err := tx.Exec(ctx, `
 			UPDATE derived.simulated_entries
 			SET display_name = $2,
 				updated_at = NOW()
 			WHERE id = $1::uuid
 				AND deleted_at IS NULL
 		`, entryID, *req.DisplayName); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	if req.Teams != nil {
		teams := *req.Teams
		sort.Slice(teams, func(i, j int) bool { return teams[i].BidPoints > teams[j].BidPoints })
		if _, err := tx.Exec(ctx, `
 			UPDATE derived.simulated_entry_teams
 			SET deleted_at = NOW(),
 				updated_at = NOW()
 			WHERE simulated_entry_id = $1::uuid
 				AND deleted_at IS NULL
 		`, entryID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		for _, t := range teams {
			if _, err := tx.Exec(ctx, `
 				INSERT INTO derived.simulated_entry_teams (
 					simulated_entry_id,
 					team_id,
 					bid_points
 				)
 				VALUES ($1::uuid, $2::uuid, $3::int)
 			`, entryID, t.TeamID, t.BidPoints); err != nil {
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

func (s *Server) handleDeleteSimulatedEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	simulatedCalcuttaID := strings.TrimSpace(vars["simulatedCalcuttaId"])
	entryID := strings.TrimSpace(vars["entryId"])
	if simulatedCalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "simulatedCalcuttaId is required", "simulatedCalcuttaId")
		return
	}
	if _, err := uuid.Parse(simulatedCalcuttaID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "simulatedCalcuttaId must be a valid UUID", "simulatedCalcuttaId")
		return
	}
	if entryID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "entryId is required", "entryId")
		return
	}
	if _, err := uuid.Parse(entryID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "entryId must be a valid UUID", "entryId")
		return
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

	var existingSimulatedCalcuttaID string
	if err := tx.QueryRow(ctx, `
 		SELECT simulated_calcutta_id::text
 		FROM derived.simulated_entries
 		WHERE id = $1::uuid
 			AND deleted_at IS NULL
 		LIMIT 1
 	`, entryID).Scan(&existingSimulatedCalcuttaID); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Simulated entry not found", "entryId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	if existingSimulatedCalcuttaID != simulatedCalcuttaID {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulated entry not found", "entryId")
		return
	}

	if _, err := tx.Exec(ctx, `
 		UPDATE derived.simulated_entry_teams
 		SET deleted_at = NOW(),
 			updated_at = NOW()
 		WHERE simulated_entry_id = $1::uuid
 			AND deleted_at IS NULL
 	`, entryID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if _, err := tx.Exec(ctx, `
 		UPDATE derived.simulated_entries
 		SET deleted_at = NOW(),
 			updated_at = NOW()
 		WHERE id = $1::uuid
 			AND deleted_at IS NULL
 	`, entryID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if _, err := tx.Exec(ctx, `
 		UPDATE derived.simulated_calcuttas
 		SET highlighted_simulated_entry_id = NULL,
 			updated_at = NOW()
 		WHERE id = $1::uuid
 			AND highlighted_simulated_entry_id = $2::uuid
 			AND deleted_at IS NULL
 	`, simulatedCalcuttaID, entryID); err != nil {
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

func (s *Server) handleImportCandidateAsSimulatedEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	simulatedCalcuttaID := strings.TrimSpace(vars["id"])
	if simulatedCalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(simulatedCalcuttaID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	var req importCandidateAsSimulatedEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	req.CandidateID = strings.TrimSpace(req.CandidateID)
	if req.CandidateID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "candidateId is required", "candidateId")
		return
	}
	if _, err := uuid.Parse(req.CandidateID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "candidateId must be a valid UUID", "candidateId")
		return
	}
	if req.DisplayName != nil {
		v := strings.TrimSpace(*req.DisplayName)
		req.DisplayName = &v
		if v == "" {
			req.DisplayName = nil
		}
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

	var exists bool
	if err := tx.QueryRow(ctx, `
 		SELECT EXISTS(
 			SELECT 1
 			FROM derived.simulated_calcuttas
 			WHERE id = $1::uuid
 				AND deleted_at IS NULL
 		)
 	`, simulatedCalcuttaID).Scan(&exists); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if !exists {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulated calcutta not found", "id")
		return
	}

	var candidateDisplayName string
	var strategyGenerationRunID *string
	var metadataJSON []byte
	if err := tx.QueryRow(ctx, `
 		SELECT display_name, strategy_generation_run_id::text, metadata_json
 		FROM derived.candidates
 		WHERE id = $1::uuid
 			AND deleted_at IS NULL
 		LIMIT 1
 	`, req.CandidateID).Scan(&candidateDisplayName, &strategyGenerationRunID, &metadataJSON); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Candidate not found", "candidateId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	name := strings.TrimSpace(candidateDisplayName)
	if req.DisplayName != nil {
		name = *req.DisplayName
	}
	if name == "" {
		name = "Candidate"
	}

	teams := make([]simulatedEntryTeam, 0)
	if strategyGenerationRunID != nil && strings.TrimSpace(*strategyGenerationRunID) != "" {
		rows, err := tx.Query(ctx, `
 			SELECT team_id::text, bid_points::int
			FROM derived.strategy_generation_run_bids
			WHERE strategy_generation_run_id = $1::uuid
				AND deleted_at IS NULL
			ORDER BY bid_points DESC
		`, *strategyGenerationRunID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		for rows.Next() {
			var teamID string
			var bidPoints int
			if err := rows.Scan(&teamID, &bidPoints); err != nil {
				rows.Close()
				writeErrorFromErr(w, r, err)
				return
			}
			teams = append(teams, simulatedEntryTeam{TeamID: teamID, BidPoints: bidPoints})
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			writeErrorFromErr(w, r, err)
			return
		}
		rows.Close()
	}

	if len(teams) == 0 {
		var parsed struct {
			Teams []simulatedEntryTeam `json:"teams"`
		}
		_ = json.Unmarshal(metadataJSON, &parsed)
		teams = parsed.Teams
	}

	if len(teams) == 0 {
		writeError(w, r, http.StatusConflict, "invalid_state", "Candidate has no bids to import", "candidateId")
		return
	}
	for i := range teams {
		teams[i].TeamID = strings.TrimSpace(teams[i].TeamID)
		if teams[i].TeamID == "" {
			writeError(w, r, http.StatusConflict, "invalid_state", "Candidate has invalid team_id", "candidateId")
			return
		}
		if _, err := uuid.Parse(teams[i].TeamID); err != nil {
			writeError(w, r, http.StatusConflict, "invalid_state", "Candidate has invalid team_id", "candidateId")
			return
		}
		if teams[i].BidPoints <= 0 {
			writeError(w, r, http.StatusConflict, "invalid_state", "Candidate has invalid bid_points", "candidateId")
			return
		}
	}

	var entryID string
	if err := tx.QueryRow(ctx, `
 		INSERT INTO derived.simulated_entries (
 			simulated_calcutta_id,
 			display_name,
 			source_kind,
 			source_entry_id,
 			source_candidate_id
 		)
 		VALUES ($1::uuid, $2, 'from_candidate', NULL, $3::uuid)
 		RETURNING id::text
 	`, simulatedCalcuttaID, name, req.CandidateID).Scan(&entryID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	sort.Slice(teams, func(i, j int) bool { return teams[i].BidPoints > teams[j].BidPoints })
	for _, t := range teams {
		if _, err := tx.Exec(ctx, `
 			INSERT INTO derived.simulated_entry_teams (
 				simulated_entry_id,
 				team_id,
 				bid_points
 			)
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

	writeJSON(w, http.StatusCreated, importCandidateAsSimulatedEntryResponse{SimulatedEntryID: entryID, NTeams: len(teams)})
}
