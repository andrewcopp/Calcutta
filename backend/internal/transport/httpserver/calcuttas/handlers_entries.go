package calcuttas

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	calcuttaapp "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleCreateEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "id")
		return
	}

	userID := ""
	if h.authUserID != nil {
		userID = h.authUserID(r.Context())
	}
	if userID == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	var req dtos.CreateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	calcutta, err := h.app.Calcutta.GetCalcuttaByID(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	tournament, err := h.app.Tournament.GetByID(r.Context(), calcutta.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanCreateEntry(r.Context(), h.authz, userID, calcutta, tournament, req.UserID, time.Now())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	entryUserID := &userID
	if req.UserID != nil {
		entryUserID = req.UserID
	}

	entry := &models.CalcuttaEntry{
		Name:       strings.TrimSpace(req.Name),
		UserID:     entryUserID,
		CalcuttaID: calcuttaID,
	}

	if err := h.app.Calcutta.CreateEntry(r.Context(), entry); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusCreated, dtos.NewEntryResponse(entry))
}

func (h *Handler) HandleListCalcuttaEntries(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "id")
		return
	}

	entries, err := h.app.Calcutta.GetEntries(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	response.WriteJSON(w, http.StatusOK, dtos.NewEntryListResponse(entries))
}

func (h *Handler) HandleListEntryTeams(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	calcuttaID := vars["calcuttaId"]
	entryID := vars["entryId"]

	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "calcuttaId")
		return
	}
	if entryID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Entry ID is required", "entryId")
		return
	}

	userID := ""
	if h.authUserID != nil {
		userID = h.authUserID(r.Context())
	}
	if userID == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	entry, err := h.app.Calcutta.GetEntry(r.Context(), entryID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if entry == nil {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "entry not found", "")
		return
	}
	if entry.CalcuttaID != calcuttaID {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "entry not found", "")
		return
	}

	calcutta, err := h.app.Calcutta.GetCalcuttaByID(r.Context(), entry.CalcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanViewEntryData(r.Context(), h.authz, userID, entry, calcutta)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	teams, err := h.app.Calcutta.GetEntryTeams(r.Context(), entryID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	response.WriteJSON(w, http.StatusOK, dtos.NewEntryTeamListResponse(teams))
}

func (h *Handler) HandleUpdateEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entryID := vars["id"]
	if entryID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Entry ID is required", "id")
		return
	}

	userID := ""
	if h.authUserID != nil {
		userID = h.authUserID(r.Context())
	}
	if userID == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	entry, err := h.app.Calcutta.GetEntry(r.Context(), entryID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	calcutta, err := h.app.Calcutta.GetCalcuttaByID(r.Context(), entry.CalcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	tournament, err := h.app.Tournament.GetByID(r.Context(), calcutta.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanEditEntryBids(r.Context(), h.authz, userID, entry, calcutta, tournament, time.Now())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	var req dtos.UpdateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	teams := make([]*models.CalcuttaEntryTeam, 0, len(req.Teams))
	for _, t := range req.Teams {
		teams = append(teams, &models.CalcuttaEntryTeam{EntryID: entryID, TeamID: t.TeamID, Bid: t.Bid})
	}

	if err := calcuttaapp.ValidateEntry(entry, teams); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", err.Error(), "teams")
		return
	}

	if err := h.app.Calcutta.ReplaceEntryTeams(r.Context(), entryID, teams); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	updatedTeams, err := h.app.Calcutta.GetEntryTeams(r.Context(), entryID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	response.WriteJSON(w, http.StatusOK, dtos.NewEntryTeamListResponse(updatedTeams))
}
