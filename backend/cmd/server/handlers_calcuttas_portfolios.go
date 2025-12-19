package main

import (
	"encoding/json"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
	"github.com/gorilla/mux"
)

func (s *Server) portfoliosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	entryID := vars["id"]
	if entryID == "" {
		http.Error(w, "Entry ID is required", http.StatusBadRequest)
		return
	}

	portfolios, err := s.calcuttaService.GetPortfoliosByEntry(r.Context(), entryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(dtos.NewPortfolioListResponse(portfolios))
}

func (s *Server) portfolioTeamsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	portfolioID := vars["id"]
	if portfolioID == "" {
		http.Error(w, "Portfolio ID is required", http.StatusBadRequest)
		return
	}

	teams, err := s.calcuttaService.GetPortfolioTeams(r.Context(), portfolioID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(dtos.NewPortfolioTeamListResponse(teams))
}
