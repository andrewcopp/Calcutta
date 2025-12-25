package httpserver

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/gorilla/mux"
)

func (s *Server) portfoliosHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entryID := vars["id"]
	if entryID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Entry ID is required", "id")
		return
	}

	portfolios, err := s.calcuttaService.GetPortfoliosByEntry(r.Context(), entryID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewPortfolioListResponse(portfolios))
}

func (s *Server) portfolioTeamsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["id"]
	if portfolioID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Portfolio ID is required", "id")
		return
	}

	teams, err := s.calcuttaService.GetPortfolioTeams(r.Context(), portfolioID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewPortfolioTeamListResponse(teams))
}
