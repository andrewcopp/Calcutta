package calcuttas

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

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

	calcutta, err := h.app.Calcutta.GetCalcuttaByID(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	userID := ""
	if h.authUserID != nil {
		userID = h.authUserID(r.Context())
	}

	participantIDs, err := h.app.Calcutta.GetDistinctUserIDsByCalcutta(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanViewCalcutta(r.Context(), h.authz, userID, calcutta, participantIDs)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
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
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
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

	decision, err := policy.CanManageCalcutta(r.Context(), h.authz, userID, calcutta)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	var req replacePayoutsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if len(req.Payouts) > 100 {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "too many payout positions", "payouts")
		return
	}
	seen := make(map[int]bool, len(req.Payouts))
	for _, p := range req.Payouts {
		if p.Position < 1 {
			httperr.Write(w, r, http.StatusBadRequest, "validation_error", "position must be >= 1", "position")
			return
		}
		if p.AmountCents < 0 {
			httperr.Write(w, r, http.StatusBadRequest, "validation_error", "amountCents cannot be negative", "amountCents")
			return
		}
		if seen[p.Position] {
			httperr.Write(w, r, http.StatusBadRequest, "validation_error", "duplicate position", "position")
			return
		}
		seen[p.Position] = true
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
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
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

	source, err := h.app.Calcutta.GetCalcuttaByID(r.Context(), sourceCalcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanManageCalcutta(r.Context(), h.authz, userID, source)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
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
		CreatedBy:    userID,
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
