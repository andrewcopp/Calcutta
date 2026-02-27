package pools

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	poolapp "github.com/andrewcopp/Calcutta/backend/internal/app/pool"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleCreatePortfolio(w http.ResponseWriter, r *http.Request) {
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

	var req dtos.CreatePortfolioRequest
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

	tournament, err := h.app.Tournament.GetByID(r.Context(), pool.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanCreatePortfolio(r.Context(), h.authz, userID, pool, tournament, req.UserID, time.Now())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	portfolioUserID := &userID
	if req.UserID != nil {
		portfolioUserID = req.UserID
	}

	portfolio := &models.Portfolio{
		Name:   strings.TrimSpace(req.Name),
		UserID: portfolioUserID,
		PoolID: poolID,
		Status: "draft",
	}

	if err := h.app.Pool.CreatePortfolio(r.Context(), portfolio); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusCreated, dtos.NewPortfolioResponse(portfolio, nil))
}

func (h *Handler) HandleListPortfolios(w http.ResponseWriter, r *http.Request) {
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

	portfolios, standings, err := h.app.Pool.GetPortfolios(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	standingsByID := make(map[string]*models.PortfolioStanding, len(standings))
	for _, s := range standings {
		standingsByID[s.PortfolioID] = s
	}

	tournament, err := h.app.Tournament.GetByID(r.Context(), pool.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !tournament.HasStarted(time.Now()) {
		manageDecision, err := policy.CanManagePool(r.Context(), h.authz, userID, pool)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
		if !manageDecision.Allowed {
			filtered := make([]*models.Portfolio, 0)
			for _, p := range portfolios {
				if p.UserID != nil && *p.UserID == userID {
					filtered = append(filtered, p)
				}
			}
			portfolios = filtered
		}
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{"items": dtos.NewPortfolioListResponse(portfolios, standingsByID)})
}

func (h *Handler) HandleListInvestments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	poolID := vars["poolId"]
	portfolioID := vars["portfolioId"]

	if poolID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Pool ID is required", "poolId")
		return
	}
	if portfolioID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Portfolio ID is required", "portfolioId")
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

	portfolio, err := h.app.Pool.GetPortfolio(r.Context(), portfolioID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if portfolio == nil {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "portfolio not found", "")
		return
	}
	if portfolio.PoolID != poolID {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "portfolio not found", "")
		return
	}

	pool, err := h.app.Pool.GetPoolByID(r.Context(), portfolio.PoolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanViewPortfolioData(r.Context(), h.authz, userID, portfolio, pool)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	tournament, err := h.app.Tournament.GetByID(r.Context(), pool.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !policy.IsBiddingPhaseViewAllowed(userID, portfolio, tournament, time.Now(), decision.IsAdmin) {
		httperr.Write(w, r, http.StatusForbidden, "investing_active", "Portfolio data is sealed until tip-off", "")
		return
	}

	investments, err := h.app.Pool.GetInvestments(r.Context(), portfolioID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": dtos.NewInvestmentListResponse(investments)})
}

func (h *Handler) HandleUpdatePortfolio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["portfolioId"]
	if portfolioID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Portfolio ID is required", "portfolioId")
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

	portfolio, err := h.app.Pool.GetPortfolio(r.Context(), portfolioID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	pool, err := h.app.Pool.GetPoolByID(r.Context(), portfolio.PoolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	tournament, err := h.app.Tournament.GetByID(r.Context(), pool.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanEditPortfolioInvestments(r.Context(), h.authz, userID, portfolio, pool, tournament, time.Now())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	var req dtos.UpdatePortfolioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	investments := make([]*models.Investment, 0, len(req.Teams))
	for _, t := range req.Teams {
		investments = append(investments, &models.Investment{PortfolioID: portfolioID, TeamID: t.TeamID, Credits: t.Credits})
	}

	if err := poolapp.ValidatePortfolio(pool, portfolio, investments); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", err.Error(), "teams")
		return
	}

	if err := h.app.Pool.ReplaceInvestments(r.Context(), portfolioID, investments); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	// Best-effort audit trail
	snapshotEntries := make([]models.InvestmentSnapshotEntry, len(investments))
	for i, inv := range investments {
		snapshotEntries[i] = models.InvestmentSnapshotEntry{TeamID: inv.TeamID, Credits: inv.Credits}
	}
	reason := ""
	if decision.IsAdmin {
		reason = "admin_override"
	}
	snapshot := &models.InvestmentSnapshot{
		PortfolioID: portfolioID,
		ChangedBy:   userID,
		Reason:      reason,
		Investments: snapshotEntries,
	}
	if err := h.app.Pool.CreateInvestmentSnapshot(r.Context(), snapshot); err != nil {
		slog.Error("failed to create investment snapshot", "portfolio_id", portfolioID, "error", err)
	}

	if err := h.app.Pool.UpdatePortfolioStatus(r.Context(), portfolioID, "submitted"); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	updatedPortfolio, err := h.app.Pool.GetPortfolio(r.Context(), portfolioID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	_, standings, err := h.app.Pool.GetPortfolios(r.Context(), updatedPortfolio.PoolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	var standing *models.PortfolioStanding
	for _, s := range standings {
		if s.PortfolioID == portfolioID {
			standing = s
			break
		}
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewPortfolioResponse(updatedPortfolio, standing))
}
