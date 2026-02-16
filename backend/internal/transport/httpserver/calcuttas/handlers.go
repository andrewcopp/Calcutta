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
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/gorilla/mux"
)

type Handler struct {
	app        *app.App
	authz      policy.AuthorizationChecker
	authUserID func(context.Context) string
}

func NewHandlerWithAuthUserID(a *app.App, authz policy.AuthorizationChecker, authUserID func(context.Context) string) *Handler {
	return &Handler{app: a, authz: authz, authUserID: authUserID}
}

func (h *Handler) HandleListCalcuttas(w http.ResponseWriter, r *http.Request) {
	userID := ""
	if h.authUserID != nil {
		userID = h.authUserID(r.Context())
	}
	if userID == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	// Check if user is admin
	isAdmin, err := h.authz.HasPermission(r.Context(), userID, "global", "", "calcutta.config.write")
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	var result []*models.Calcutta
	if isAdmin {
		result, err = h.app.Calcutta.GetAllCalcuttas(r.Context())
	} else {
		result, err = h.app.Calcutta.GetCalcuttasByUser(r.Context(), userID)
	}
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	resp := dtos.NewCalcuttaListResponse(result)
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

	// Validate tournament has a start time (required for bidding-lock logic)
	tournament, err := h.app.Tournament.GetByID(r.Context(), calcutta.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if tournament.StartingAt == nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament must have a start time before creating a pool", "tournamentId")
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
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "id")
		return
	}

	calcutta, err := h.app.Calcutta.GetCalcuttaByID(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

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
	if req.BiddingOpen != nil {
		calcutta.BiddingOpen = *req.BiddingOpen
		if !*req.BiddingOpen && calcutta.BiddingLockedAt == nil {
			now := time.Now()
			calcutta.BiddingLockedAt = &now
		}
		if *req.BiddingOpen {
			calcutta.BiddingLockedAt = nil
		}
	}

	if err := h.app.Calcutta.UpdateCalcutta(r.Context(), calcutta); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewCalcuttaResponse(calcutta))
}

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

func (h *Handler) HandleCreateInvitation(w http.ResponseWriter, r *http.Request) {
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

	var req dtos.CreateInvitationRequest
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

	decision, err := policy.CanInviteToCalcutta(r.Context(), h.authz, userID, calcutta)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	invitation := &models.CalcuttaInvitation{
		CalcuttaID: calcuttaID,
		UserID:     strings.TrimSpace(req.UserID),
		InvitedBy:  userID,
	}

	if err := h.app.Calcutta.InviteUser(r.Context(), invitation); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusCreated, dtos.NewInvitationResponse(invitation))
}

func (h *Handler) HandleListInvitations(w http.ResponseWriter, r *http.Request) {
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

	decision, err := policy.CanInviteToCalcutta(r.Context(), h.authz, userID, calcutta)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	invitations, err := h.app.Calcutta.ListInvitations(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewInvitationListResponse(invitations))
}

func (h *Handler) HandleAcceptInvitation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	invitationID := vars["invitationId"]
	if calcuttaID == "" || invitationID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID and Invitation ID are required", "")
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

	invitation, err := h.app.Calcutta.GetInvitationByCalcuttaAndUser(r.Context(), calcuttaID, userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if invitation.ID != invitationID {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "Invitation not found", "")
		return
	}
	if invitation.Status != "pending" {
		httperr.Write(w, r, http.StatusBadRequest, "already_processed", "Invitation has already been processed", "")
		return
	}

	if err := h.app.Calcutta.AcceptInvitation(r.Context(), invitationID); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	invitation.Status = "accepted"
	response.WriteJSON(w, http.StatusOK, dtos.NewInvitationResponse(invitation))
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

type payoutItem struct {
	Position    int `json:"position"`
	AmountCents int `json:"amountCents"`
}

type replacePayoutsRequest struct {
	Payouts []payoutItem `json:"payouts"`
}

func (h *Handler) HandleListPayouts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "id")
		return
	}

	payouts, err := h.app.Calcutta.GetPayouts(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	items := make([]payoutItem, 0, len(payouts))
	for _, p := range payouts {
		items = append(items, payoutItem{Position: p.Position, AmountCents: p.AmountCents})
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"payouts": items})
}

func (h *Handler) HandleReplacePayouts(w http.ResponseWriter, r *http.Request) {
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

	// Only owner or admin can modify payouts
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

	var req replacePayoutsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	payouts := make([]*models.CalcuttaPayout, 0, len(req.Payouts))
	for _, p := range req.Payouts {
		payouts = append(payouts, &models.CalcuttaPayout{
			CalcuttaID:  calcuttaID,
			Position:    p.Position,
			AmountCents: p.AmountCents,
		})
	}

	if err := h.app.Calcutta.ReplacePayouts(r.Context(), calcuttaID, payouts); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	updated, err := h.app.Calcutta.GetPayouts(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	items := make([]payoutItem, 0, len(updated))
	for _, p := range updated {
		items = append(items, payoutItem{Position: p.Position, AmountCents: p.AmountCents})
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"payouts": items})
}

func (h *Handler) HandleReinvite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceCalcuttaID := vars["id"]
	if sourceCalcuttaID == "" {
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

	// Check ownership or admin
	source, err := h.app.Calcutta.GetCalcuttaByID(r.Context(), sourceCalcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if source.OwnerID != userID {
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

	var req dtos.ReinviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	newCalcutta := &models.Calcutta{
		Name:         strings.TrimSpace(req.Name),
		TournamentID: strings.TrimSpace(req.TournamentID),
		OwnerID:      userID,
	}

	created, invitations, err := h.app.Calcutta.ReinviteFromCalcutta(r.Context(), sourceCalcuttaID, newCalcutta, userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	invResp := make([]*dtos.InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		invResp = append(invResp, dtos.NewInvitationResponse(inv))
	}

	response.WriteJSON(w, http.StatusCreated, dtos.ReinviteResponse{
		Calcutta:    dtos.NewCalcuttaResponse(created),
		Invitations: invResp,
	})
}
