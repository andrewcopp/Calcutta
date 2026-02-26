package pools

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleCreateInvitation(w http.ResponseWriter, r *http.Request) {
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

	var req dtos.CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	pool, err := h.app.Pool.GetPoolByID(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanInviteToPool(r.Context(), h.authz, userID, pool)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	invitation := &models.PoolInvitation{
		PoolID:    poolID,
		UserID:    strings.TrimSpace(req.UserID),
		InvitedBy: userID,
	}

	if err := h.app.Pool.InviteUser(r.Context(), invitation); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusCreated, dtos.NewInvitationResponse(invitation))
}

func (h *Handler) HandleListInvitations(w http.ResponseWriter, r *http.Request) {
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

	decision, err := policy.CanInviteToPool(r.Context(), h.authz, userID, pool)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	invitations, err := h.app.Pool.ListInvitations(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{"items": dtos.NewInvitationListResponse(invitations)})
}

func (h *Handler) HandleAcceptInvitation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	poolID := vars["id"]
	invitationID := vars["invitationId"]
	if poolID == "" || invitationID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Pool ID and Invitation ID are required", "")
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

	invitation, err := h.app.Pool.GetPendingInvitationByPoolAndUser(r.Context(), poolID, userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if invitation.ID != invitationID {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "Invitation not found", "")
		return
	}

	pool, err := h.app.Pool.GetPoolByID(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	tournament, err := h.app.Tournament.GetByID(r.Context(), pool.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	decision, err := policy.CanAcceptInvitation(r.Context(), h.authz, userID, pool, tournament, time.Now())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	if err := h.app.Pool.AcceptInvitation(r.Context(), invitationID); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if h.granter != nil {
		_ = h.granter.GrantRole(r.Context(), userID, "player", "pool", poolID)
	}

	invitation.Status = "accepted"
	response.WriteJSON(w, http.StatusOK, dtos.NewInvitationResponse(invitation))
}

func (h *Handler) HandleRevokeInvitation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	poolID := vars["id"]
	invitationID := vars["invitationId"]
	if poolID == "" || invitationID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Pool ID and Invitation ID are required", "")
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

	decision, err := policy.CanInviteToPool(r.Context(), h.authz, userID, pool)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	if err := h.app.Pool.RevokeInvitation(r.Context(), invitationID); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleListMyInvitations(w http.ResponseWriter, r *http.Request) {
	userID := ""
	if h.authUserID != nil {
		userID = h.authUserID(r.Context())
	}
	if userID == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	invitations, err := h.app.Pool.ListPendingInvitationsByUserID(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{"items": dtos.NewInvitationListResponse(invitations)})
}
