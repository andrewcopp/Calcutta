package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/gorilla/mux"
)

func (s *Server) calcuttaEntriesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "id")
		return
	}

	entries, err := s.calcuttaService.GetEntries(r.Context(), calcuttaID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewEntryListResponse(entries))
}

func (s *Server) calcuttaEntryTeamsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	calcuttaID := vars["calcuttaId"]
	entryID := vars["entryId"]

	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "calcuttaId")
		return
	}
	if entryID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Entry ID is required", "entryId")
		return
	}

	teams, err := s.calcuttaService.GetEntryTeams(r.Context(), entryID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewEntryTeamListResponse(teams))
}

func (s *Server) updateEntryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entryID := vars["id"]
	if entryID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Entry ID is required", "id")
		return
	}

	userID := authUserID(r.Context())
	if userID == "" {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	entry, err := s.calcuttaService.GetEntry(r.Context(), entryID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	calcutta, err := s.calcuttaService.GetCalcuttaByID(r.Context(), entry.CalcuttaID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	tournament, err := s.app.Tournament.GetByID(r.Context(), calcutta.TournamentID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	decision, err := policy.CanEditEntryBids(r.Context(), s.authzRepo, userID, entry, calcutta, tournament, time.Now())
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if !decision.Allowed {
		writeError(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	var req dtos.UpdateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	teams := make([]*models.CalcuttaEntryTeam, 0, len(req.Teams))
	for _, t := range req.Teams {
		teams = append(teams, &models.CalcuttaEntryTeam{
			EntryID: entryID,
			TeamID:  t.TeamID,
			Bid:     t.Bid,
		})
	}

	if err := s.calcuttaService.ValidateEntry(entry, teams); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", err.Error(), "teams")
		return
	}

	if err := s.calcuttaService.ReplaceEntryTeams(r.Context(), entryID, teams); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if err := s.calcuttaService.EnsurePortfoliosAndRecalculate(r.Context(), calcutta.ID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	updatedTeams, err := s.calcuttaService.GetEntryTeams(r.Context(), entryID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewEntryTeamListResponse(updatedTeams))
}
