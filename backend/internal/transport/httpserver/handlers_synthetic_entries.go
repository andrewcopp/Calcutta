package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/synthetic_scenarios"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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
	r.HandleFunc(
		"/api/synthetic-calcuttas/{syntheticCalcuttaId}/candidates/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handlePatchSyntheticEntry),
	).Methods("PATCH", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas/{syntheticCalcuttaId}/candidates/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handleDeleteSyntheticEntry),
	).Methods("DELETE", "OPTIONS")

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
	items, err := s.app.SyntheticScenarios.ListSyntheticEntries(ctx, syntheticCalcuttaID)
	if err != nil {
		if errors.Is(err, synthetic_scenarios.ErrSyntheticCalcuttaNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic calcutta not found", "id")
			return
		}
		if errors.Is(err, synthetic_scenarios.ErrSyntheticCalcuttaHasNoSnapshot) {
			writeError(w, r, http.StatusConflict, "invalid_state", "Synthetic calcutta has no snapshot", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	resp := make([]syntheticEntryListItem, 0, len(items))
	for _, it := range items {
		teams := make([]syntheticEntryTeam, 0, len(it.Teams))
		for _, t := range it.Teams {
			teams = append(teams, syntheticEntryTeam{TeamID: t.TeamID, BidPoints: t.BidPoints})
		}
		resp = append(resp, syntheticEntryListItem{
			ID:            it.ID,
			CandidateID:   it.CandidateID,
			SnapshotEntry: it.SnapshotEntry,
			EntryID:       it.EntryID,
			DisplayName:   it.DisplayName,
			IsSynthetic:   it.IsSynthetic,
			Rank:          it.Rank,
			Mean:          it.Mean,
			PTop1:         it.PTop1,
			PInMoney:      it.PInMoney,
			Teams:         teams,
			CreatedAt:     it.CreatedAt,
			UpdatedAt:     it.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, listSyntheticEntriesResponse{Items: resp})
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
	teams := make([]synthetic_scenarios.SyntheticEntryTeam, 0, len(req.Teams))
	for _, t := range req.Teams {
		teams = append(teams, synthetic_scenarios.SyntheticEntryTeam{TeamID: t.TeamID, BidPoints: t.BidPoints})
	}

	createdID, err := s.app.SyntheticScenarios.CreateSyntheticEntry(ctx, synthetic_scenarios.CreateSyntheticEntryParams{
		SyntheticCalcuttaID: syntheticCalcuttaID,
		DisplayName:         req.DisplayName,
		Teams:               teams,
	})
	if err != nil {
		if errors.Is(err, synthetic_scenarios.ErrSyntheticCalcuttaNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic calcutta not found", "id")
			return
		}
		if errors.Is(err, synthetic_scenarios.ErrSyntheticCalcuttaHasNoSnapshot) {
			writeError(w, r, http.StatusConflict, "invalid_state", "Synthetic calcutta has no snapshot", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, createSyntheticEntryResponse{ID: createdID})
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

	attachmentID, nTeams, err := s.app.SyntheticScenarios.ImportSyntheticEntry(ctx, synthetic_scenarios.ImportSyntheticEntryParams{
		SyntheticCalcuttaID: syntheticCalcuttaID,
		EntryArtifactID:     req.EntryArtifactID,
		DisplayName:         req.DisplayName,
	})
	if err != nil {
		if errors.Is(err, synthetic_scenarios.ErrEntryArtifactNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Entry artifact not found", "entryArtifactId")
			return
		}
		if errors.Is(err, synthetic_scenarios.ErrEntryArtifactHasNoRunID) {
			writeError(w, r, http.StatusConflict, "invalid_state", "Entry artifact has no run_id", "entryArtifactId")
			return
		}
		if errors.Is(err, synthetic_scenarios.ErrEntryArtifactNotMetrics) {
			writeError(w, r, http.StatusBadRequest, "validation_error", "entryArtifactId must reference a metrics artifact", "entryArtifactId")
			return
		}
		if errors.Is(err, synthetic_scenarios.ErrNoStrategyGenerationRunBids) {
			writeError(w, r, http.StatusConflict, "invalid_state", "No strategy_generation_run_bids found for that entryArtifactId", "entryArtifactId")
			return
		}
		if errors.Is(err, synthetic_scenarios.ErrSyntheticCalcuttaNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic calcutta not found", "id")
			return
		}
		if errors.Is(err, synthetic_scenarios.ErrSyntheticCalcuttaHasNoSnapshot) {
			writeError(w, r, http.StatusConflict, "invalid_state", "Synthetic calcutta has no snapshot", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, importSyntheticEntryResponse{ID: attachmentID, NTeams: nTeams})
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

	teams := (*[]synthetic_scenarios.SyntheticEntryTeam)(nil)
	if req.Teams != nil {
		parsed := make([]synthetic_scenarios.SyntheticEntryTeam, 0, len(*req.Teams))
		for _, t := range *req.Teams {
			parsed = append(parsed, synthetic_scenarios.SyntheticEntryTeam{TeamID: t.TeamID, BidPoints: t.BidPoints})
		}
		teams = &parsed
	}

	if err := s.app.SyntheticScenarios.PatchSyntheticEntry(r.Context(), synthetic_scenarios.PatchSyntheticEntryParams{AttachmentID: id, DisplayName: displayName, Teams: teams}); err != nil {
		if errors.Is(err, synthetic_scenarios.ErrSyntheticEntryNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic entry not found", "id")
			return
		}
		if errors.Is(err, synthetic_scenarios.ErrOnlyManualCandidatesCanBeEdited) {
			writeError(w, r, http.StatusBadRequest, "validation_error", "Only manual candidates can be edited", "teams")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
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

	if err := s.app.SyntheticScenarios.DeleteSyntheticEntry(r.Context(), synthetic_scenarios.DeleteSyntheticEntryParams{AttachmentID: id}); err != nil {
		if errors.Is(err, synthetic_scenarios.ErrSyntheticEntryNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic entry not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
