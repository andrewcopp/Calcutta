package httpserver

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

func (s *Server) portfolioTeamsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["id"]
	if portfolioID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Portfolio ID is required", "id")
		return
	}

	userID := authUserID(r.Context())
	if userID == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	portfolio, err := s.app.Calcutta.GetPortfolio(r.Context(), portfolioID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if portfolio == nil {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "portfolio not found", "")
		return
	}

	entry, err := s.app.Calcutta.GetEntry(r.Context(), portfolio.EntryID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if entry == nil {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "entry not found", "")
		return
	}

	calcutta, err := s.app.Calcutta.GetCalcuttaByID(r.Context(), entry.CalcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	decision, err := policy.CanViewEntryData(r.Context(), s.authzRepo, userID, entry, calcutta)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	teams, err := s.app.Calcutta.GetPortfolioTeams(r.Context(), portfolioID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	response.WriteJSON(w, http.StatusOK, dtos.NewPortfolioTeamListResponse(teams))
}
