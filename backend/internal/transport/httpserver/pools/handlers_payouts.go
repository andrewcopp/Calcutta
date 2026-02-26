package pools

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
	poolID := vars["id"]
	if poolID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Pool ID is required", "id")
		return
	}

	pool, err := h.app.Pool.GetPoolByID(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	userID := ""
	if h.authUserID != nil {
		userID = h.authUserID(r.Context())
	}

	participantIDs, err := h.app.Pool.GetDistinctUserIDsByPool(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanViewPool(r.Context(), h.authz, userID, pool, participantIDs)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	payouts, err := h.app.Pool.GetPayouts(r.Context(), poolID)
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
	poolID := vars["id"]
	if poolID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Pool ID is required", "id")
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

	pool, err := h.app.Pool.GetPoolByID(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanManagePool(r.Context(), h.authz, userID, pool)
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

	payouts := make([]*models.PoolPayout, 0, len(req.Payouts))
	for _, p := range req.Payouts {
		payouts = append(payouts, &models.PoolPayout{
			PoolID:      poolID,
			Position:    p.Position,
			AmountCents: p.AmountCents,
		})
	}

	if err := h.app.Pool.ReplacePayouts(r.Context(), poolID, payouts); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	updated, err := h.app.Pool.GetPayouts(r.Context(), poolID)
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
	sourcePoolID := vars["id"]
	if sourcePoolID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Pool ID is required", "id")
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

	source, err := h.app.Pool.GetPoolByID(r.Context(), sourcePoolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanManagePool(r.Context(), h.authz, userID, source)
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

	newPool := &models.Pool{
		Name:         strings.TrimSpace(req.Name),
		TournamentID: strings.TrimSpace(req.TournamentID),
		OwnerID:      userID,
		CreatedBy:    userID,
	}

	created, invitations, err := h.app.Pool.ReinviteFromPool(r.Context(), sourcePoolID, newPool, userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	invResp := make([]*dtos.InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		invResp = append(invResp, dtos.NewInvitationResponse(inv))
	}

	response.WriteJSON(w, http.StatusCreated, dtos.ReinviteResponse{
		Pool:        dtos.NewPoolResponse(created),
		Invitations: invResp,
	})
}
