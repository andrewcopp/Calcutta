package httpserver

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/gorilla/mux"
)

func (s *Server) portfolioTeamsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["id"]
	if portfolioID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Portfolio ID is required", "id")
		return
	}

	userID := authUserID(r.Context())
	if userID == "" {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	portfolio, err := s.app.Calcutta.GetPortfolio(r.Context(), portfolioID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if portfolio == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "portfolio not found", "")
		return
	}

	entry, err := s.app.Calcutta.GetEntry(r.Context(), portfolio.EntryID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if entry == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "entry not found", "")
		return
	}

	calcutta, err := s.app.Calcutta.GetCalcuttaByID(r.Context(), entry.CalcuttaID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	decision, err := policy.CanViewEntryData(r.Context(), s.authzRepo, userID, entry, calcutta)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if !decision.Allowed {
		writeError(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	teams, err := s.app.Calcutta.GetPortfolioTeams(r.Context(), portfolioID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewPortfolioTeamListResponse(teams))
}
