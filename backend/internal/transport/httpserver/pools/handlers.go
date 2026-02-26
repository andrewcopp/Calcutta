package pools

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

func (h *Handler) HandleListPools(w http.ResponseWriter, r *http.Request) {
	userID := ""
	if h.authUserID != nil {
		userID = h.authUserID(r.Context())
	}
	if userID == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	// Check if user is admin
	isAdmin, err := h.authz.HasPermission(r.Context(), userID, "global", "", "pool.config.write")
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	var pools []*models.Pool
	if isAdmin {
		pools, err = h.app.Pool.GetAllPools(r.Context())
	} else {
		pools, err = h.app.Pool.GetPoolsByUser(r.Context(), userID)
	}
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	if r.URL.Query().Get("include") == "rankings" {
		h.listPoolsWithRankings(w, r, userID, pools)
		return
	}

	resp := dtos.NewPoolListResponse(pools)
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func (h *Handler) HandleCreatePool(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreatePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Debug("create_pool_decode_failed", "error", err)
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	pool := req.ToModel()
	pool.OwnerID = ""
	if h.authUserID != nil {
		pool.OwnerID = h.authUserID(r.Context())
	}
	if pool.OwnerID == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}
	pool.CreatedBy = pool.OwnerID

	// Validate tournament has a start time (required for investing-lock logic)
	tournament, err := h.app.Tournament.GetByID(r.Context(), pool.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if tournament.StartingAt == nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament must have a start time before creating a pool", "tournamentId")
		return
	}

	pool.ApplyDefaults()

	// Validate constraints
	if pool.MinTeams < 1 || pool.MinTeams > 68 {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "MinTeams must be between 1 and 68", "minTeams")
		return
	}
	if pool.MaxTeams < 1 || pool.MaxTeams > 68 {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "MaxTeams must be between 1 and 68", "maxTeams")
		return
	}
	if pool.MinTeams > pool.MaxTeams {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "MinTeams cannot exceed MaxTeams", "minTeams")
		return
	}
	if pool.MaxInvestmentCredits < 1 {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "MaxInvestmentCredits must be at least 1", "maxInvestmentCredits")
		return
	}

	scoringRules := req.ToScoringRules()
	if err := h.app.Pool.CreatePoolWithScoringRules(r.Context(), pool, scoringRules); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if h.granter != nil {
		_ = h.granter.GrantRole(r.Context(), pool.OwnerID, "pool_admin", "pool", pool.ID)
	}
	slog.Info("pool_created", "pool_id", pool.ID)
	response.WriteJSON(w, http.StatusCreated, dtos.NewPoolResponse(pool))
}

func (h *Handler) HandleGetPool(w http.ResponseWriter, r *http.Request) {
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

	resp := dtos.NewPoolResponse(pool)
	resp.Abilities = computeAbilities(r.Context(), h.authz, userID, pool)
	response.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleUpdatePool(w http.ResponseWriter, r *http.Request) {
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

	var req dtos.UpdatePoolRequest
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
		pool.Name = name
	}
	if req.MinTeams != nil {
		pool.MinTeams = *req.MinTeams
	}
	if req.MaxTeams != nil {
		pool.MaxTeams = *req.MaxTeams
	}
	if req.MaxInvestmentCredits != nil {
		pool.MaxInvestmentCredits = *req.MaxInvestmentCredits
	}
	if err := h.app.Pool.UpdatePool(r.Context(), pool); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewPoolResponse(pool))
}
