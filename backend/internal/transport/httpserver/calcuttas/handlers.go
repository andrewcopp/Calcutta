package calcuttas

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/gorilla/mux"
)

type Handler struct {
	app        *app.App
	authz      policy.AuthorizationChecker
	authUserID func(context.Context) string
}

func NewHandler(a *app.App, authz policy.AuthorizationChecker) *Handler {
	return &Handler{app: a, authz: authz}
}

func NewHandlerWithAuthUserID(a *app.App, authz policy.AuthorizationChecker, authUserID func(context.Context) string) *Handler {
	return &Handler{app: a, authz: authz, authUserID: authUserID}
}

func (h *Handler) HandleListCalcuttas(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling GET request to /api/calcuttas")
	calcuttas, err := h.app.Calcutta.GetAllCalcuttas(r.Context())
	if err != nil {
		log.Printf("Error getting all calcuttas: %v", err)
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	log.Printf("Successfully retrieved %d calcuttas", len(calcuttas))

	resp := dtos.NewCalcuttaListResponse(calcuttas)
	response.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleCreateCalcutta(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling POST request to /api/calcuttas")

	var req dtos.CreateCalcuttaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if err := req.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	calcutta := req.ToModel()
	calcutta.OwnerID = ""
	if h.authUserID != nil {
		calcutta.OwnerID = h.authUserID(r.Context())
	}
	if calcutta.OwnerID == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}
	if calcutta.MinTeams == 0 {
		calcutta.MinTeams = 3
	}
	if calcutta.MaxTeams == 0 {
		calcutta.MaxTeams = 10
	}
	if calcutta.MaxBid == 0 {
		calcutta.MaxBid = 50
	}

	if err := h.app.Calcutta.CreateCalcuttaWithRounds(r.Context(), calcutta); err != nil {
		log.Printf("Error creating calcutta with rounds: %v", err)
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	log.Printf("Successfully created calcutta %s with rounds", calcutta.ID)
	response.WriteJSON(w, http.StatusCreated, dtos.NewCalcuttaResponse(calcutta))
}

func (h *Handler) HandleGetCalcutta(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling GET request to /api/calcuttas/{id}")

	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		log.Printf("Error: Calcutta ID is empty")
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "id")
		return
	}

	log.Printf("Fetching calcutta with ID: %s", calcuttaID)
	calcutta, err := h.app.Calcutta.GetCalcuttaByID(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	log.Printf("Successfully retrieved calcutta: %s", calcutta.ID)
	response.WriteJSON(w, http.StatusOK, dtos.NewCalcuttaResponse(calcutta))
}

func (h *Handler) HandleUpdateCalcutta(w http.ResponseWriter, r *http.Request) {
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

	calcutta, err := h.app.Calcutta.GetCalcuttaByID(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	if calcutta.OwnerID != userID {
		ok, err := h.authz.HasPermission(r.Context(), userID, "global", "", "calcutta.config.write")
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
		if !ok {
			httperr.Write(w, r, http.StatusForbidden, "forbidden", "Insufficient permissions", "")
			return
		}
	}

	var req dtos.UpdateCalcuttaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Name is required", "name")
			return
		}
		calcutta.Name = name
	}
	if req.MinTeams != nil {
		calcutta.MinTeams = *req.MinTeams
	}
	if req.MaxTeams != nil {
		calcutta.MaxTeams = *req.MaxTeams
	}
	if req.MaxBid != nil {
		calcutta.MaxBid = *req.MaxBid
	}

	if err := h.app.Calcutta.UpdateCalcutta(r.Context(), calcutta); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewCalcuttaResponse(calcutta))
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

	if err := h.app.Calcutta.ValidateEntry(entry, teams); err != nil {
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

func (h *Handler) HandleListEntryPortfolios(w http.ResponseWriter, r *http.Request) {
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
	if entry == nil {
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

	portfolios, err := h.app.Calcutta.GetPortfoliosByEntry(r.Context(), entryID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	response.WriteJSON(w, http.StatusOK, dtos.NewPortfolioListResponse(portfolios))
}
