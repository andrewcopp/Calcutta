package pools

import (
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListOwnership(w http.ResponseWriter, r *http.Request) {
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
	if portfolio == nil {
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

	ownershipSummaries, err := h.app.Pool.GetOwnershipSummariesByPortfolio(r.Context(), portfolioID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": dtos.NewOwnershipSummaryListResponse(ownershipSummaries)})
}
