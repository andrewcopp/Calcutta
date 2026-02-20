package calcuttas

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

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
