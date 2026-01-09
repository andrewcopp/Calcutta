package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/suite_scenarios"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

	ok, entries, err := s.app.SuiteScenarios.ListSimulatedEntries(r.Context(), simulatedCalcuttaID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulated calcutta not found", "id")
		return
	}

	items := make([]simulatedEntryListItem, 0, len(entries))
	for _, e := range entries {
		teams := make([]simulatedEntryTeam, 0, len(e.Teams))
		for _, t := range e.Teams {
			teams = append(teams, simulatedEntryTeam{TeamID: t.TeamID, BidPoints: t.BidPoints})
		}
		items = append(items, simulatedEntryListItem{
			ID:                  e.ID,
			SimulatedCalcuttaID: e.SimulatedCalcuttaID,
			DisplayName:         e.DisplayName,
			SourceKind:          e.SourceKind,
			SourceEntryID:       e.SourceEntryID,
			SourceCandidateID:   e.SourceCandidateID,
			Teams:               teams,
			CreatedAt:           e.CreatedAt,
			UpdatedAt:           e.UpdatedAt,
		})
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

	params := suite_scenarios.CreateSimulatedEntryParams{
		SimulatedCalcuttaID: simulatedCalcuttaID,
		DisplayName:         req.DisplayName,
		Teams:               make([]suite_scenarios.SimulatedEntryTeam, 0, len(req.Teams)),
	}
	for _, t := range req.Teams {
		params.Teams = append(params.Teams, suite_scenarios.SimulatedEntryTeam{TeamID: t.TeamID, BidPoints: t.BidPoints})
	}

	entryID, err := s.app.SuiteScenarios.CreateSimulatedEntry(r.Context(), params)
	if err != nil {
		if errors.Is(err, suite_scenarios.ErrSimulatedCalcuttaNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Simulated calcutta not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

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

	params := suite_scenarios.PatchSimulatedEntryParams{
		SimulatedCalcuttaID: simulatedCalcuttaID,
		EntryID:             entryID,
		DisplayName:         req.DisplayName,
		Teams: func() *[]suite_scenarios.SimulatedEntryTeam {
			if req.Teams == nil {
				return nil
			}
			teams := make([]suite_scenarios.SimulatedEntryTeam, 0, len(*req.Teams))
			for _, t := range *req.Teams {
				teams = append(teams, suite_scenarios.SimulatedEntryTeam{TeamID: t.TeamID, BidPoints: t.BidPoints})
			}
			return &teams
		}(),
	}

	if err := s.app.SuiteScenarios.PatchSimulatedEntry(r.Context(), params); err != nil {
		if errors.Is(err, suite_scenarios.ErrSimulatedEntryNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Simulated entry not found", "entryId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

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

	params := suite_scenarios.DeleteSimulatedEntryParams{SimulatedCalcuttaID: simulatedCalcuttaID, EntryID: entryID}
	if err := s.app.SuiteScenarios.DeleteSimulatedEntry(r.Context(), params); err != nil {
		if errors.Is(err, suite_scenarios.ErrSimulatedEntryNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Simulated entry not found", "entryId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

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

	params := suite_scenarios.ImportCandidateAsSimulatedEntryParams{
		SimulatedCalcuttaID: simulatedCalcuttaID,
		CandidateID:         req.CandidateID,
		DisplayName:         req.DisplayName,
	}

	entryID, nTeams, err := s.app.SuiteScenarios.ImportCandidateAsSimulatedEntry(r.Context(), params)
	if err != nil {
		if errors.Is(err, suite_scenarios.ErrSimulatedCalcuttaNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Simulated calcutta not found", "id")
			return
		}
		if errors.Is(err, suite_scenarios.ErrCandidateNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Candidate not found", "candidateId")
			return
		}
		if errors.Is(err, suite_scenarios.ErrCandidateHasNoBids) {
			writeError(w, r, http.StatusConflict, "invalid_state", "Candidate has no bids to import", "candidateId")
			return
		}
		if errors.Is(err, suite_scenarios.ErrCandidateInvalidTeamID) {
			writeError(w, r, http.StatusConflict, "invalid_state", "Candidate has invalid team_id", "candidateId")
			return
		}
		if errors.Is(err, suite_scenarios.ErrCandidateInvalidBidPoints) {
			writeError(w, r, http.StatusConflict, "invalid_state", "Candidate has invalid bid_points", "candidateId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, importCandidateAsSimulatedEntryResponse{SimulatedEntryID: entryID, NTeams: nTeams})
}
