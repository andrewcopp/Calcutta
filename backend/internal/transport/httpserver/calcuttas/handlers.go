package calcuttas

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
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
	if calcutta.MaxBid < 1 {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "MaxBid must be at least 1", "maxBid")
		return
	}

	rounds := req.ToScoringRules()
	if err := h.app.Calcutta.CreateCalcuttaWithRounds(r.Context(), calcutta, rounds); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
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
