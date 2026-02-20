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

	invitation, err := h.app.Calcutta.GetPendingInvitationByCalcuttaAndUser(r.Context(), calcuttaID, userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if invitation.ID != invitationID {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "Invitation not found", "")
		return
	}

	if err := h.app.Calcutta.AcceptInvitation(r.Context(), invitationID); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	invitation.Status = "accepted"
	response.WriteJSON(w, http.StatusOK, dtos.NewInvitationResponse(invitation))
}

func (h *Handler) HandleRevokeInvitation(w http.ResponseWriter, r *http.Request) {
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

	if err := h.app.Calcutta.RevokeInvitation(r.Context(), invitationID); err != nil {
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

	invitations, err := h.app.Calcutta.ListPendingInvitationsByUserID(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewInvitationListResponse(invitations))
}
