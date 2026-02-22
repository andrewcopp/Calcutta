package calcuttas

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

// RoleGranter assigns roles to users with scope.
type RoleGranter interface {
	GrantRole(ctx context.Context, userID, roleKey, scopeType, scopeID string) error
}

type Handler struct {
	app        *app.App
	authz      policy.AuthorizationChecker
	granter    RoleGranter
	authUserID func(context.Context) string
}

func NewHandlerWithAuthUserID(a *app.App, authz policy.AuthorizationChecker, granter RoleGranter, authUserID func(context.Context) string) *Handler {
	return &Handler{app: a, authz: authz, granter: granter, authUserID: authUserID}
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
	var req dtos.CreateCalcuttaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Debug("create_calcutta_decode_failed", "error", err)
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if err := req.Validate(); err != nil {
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
	calcutta.CreatedBy = calcutta.OwnerID

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

	calcutta.ApplyDefaults()

	// Validate constraints
	if calcutta.MinTeams < 1 || calcutta.MinTeams > 68 {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "MinTeams must be between 1 and 68", "minTeams")
		return
	}
	if calcutta.MaxTeams < 1 || calcutta.MaxTeams > 68 {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "MaxTeams must be between 1 and 68", "maxTeams")
		return
	}
	if calcutta.MinTeams > calcutta.MaxTeams {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "MinTeams cannot exceed MaxTeams", "minTeams")
		return
	}
	if calcutta.MaxBidPoints < 1 {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "MaxBid must be at least 1", "maxBid")
		return
	}

	rounds := req.ToScoringRules()
	if err := h.app.Calcutta.CreateCalcuttaWithRounds(r.Context(), calcutta, rounds); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if h.granter != nil {
		_ = h.granter.GrantRole(r.Context(), calcutta.OwnerID, "calcutta_admin", "calcutta", calcutta.ID)
	}
	slog.Info("calcutta_created", "calcutta_id", calcutta.ID)
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

	resp := dtos.NewCalcuttaResponse(calcutta)
	resp.Abilities = computeAbilities(r.Context(), h.authz, userID, calcutta)
	response.WriteJSON(w, http.StatusOK, resp)
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

	decision, err := policy.CanManageCalcutta(r.Context(), h.authz, userID, calcutta)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
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
	if req.MaxBidPoints != nil {
		calcutta.MaxBidPoints = *req.MaxBidPoints
	}
	if err := h.app.Calcutta.UpdateCalcutta(r.Context(), calcutta); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewCalcuttaResponse(calcutta))
}
