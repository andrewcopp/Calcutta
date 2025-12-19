package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/gorilla/mux"
)

func (s *Server) calculatePortfolioScoresHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["id"]

	if err := s.calcuttaService.CalculatePortfolioScores(r.Context(), portfolioID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) updatePortfolioTeamScoresHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["id"]
	teamID := vars["teamId"]

	var req dtos.UpdatePortfolioTeamScoresRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	teams, err := s.calcuttaService.GetPortfolioTeams(r.Context(), portfolioID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var portfolioTeam *models.CalcuttaPortfolioTeam
	for _, team := range teams {
		if team.TeamID == teamID {
			portfolioTeam = team
			break
		}
	}

	if portfolioTeam == nil {
		http.Error(w, "Portfolio team not found", http.StatusNotFound)
		return
	}

	portfolioTeam.ExpectedPoints = req.ExpectedPoints
	portfolioTeam.PredictedPoints = req.PredictedPoints
	portfolioTeam.Updated = time.Now()

	if err := s.calcuttaService.UpdatePortfolioTeam(r.Context(), portfolioTeam); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) updatePortfolioMaximumScoreHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["id"]

	var req dtos.UpdatePortfolioMaximumScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.calcuttaService.UpdatePortfolioScores(r.Context(), portfolioID, req.MaximumPoints); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
