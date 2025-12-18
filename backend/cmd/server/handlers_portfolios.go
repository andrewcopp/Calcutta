package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/gorilla/mux"
)

func calculatePortfolioScoresHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["id"]

	if err := calcuttaService.CalculatePortfolioScores(r.Context(), portfolioID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func updatePortfolioTeamScoresHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["id"]
	teamID := vars["teamId"]

	var request struct {
		ExpectedPoints  float64 `json:"expectedPoints"`
		PredictedPoints float64 `json:"predictedPoints"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	teams, err := calcuttaService.GetPortfolioTeams(r.Context(), portfolioID)
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

	portfolioTeam.ExpectedPoints = request.ExpectedPoints
	portfolioTeam.PredictedPoints = request.PredictedPoints
	portfolioTeam.Updated = time.Now()

	if err := calcuttaService.UpdatePortfolioTeam(r.Context(), portfolioTeam); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func updatePortfolioMaximumScoreHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["id"]

	var request struct {
		MaximumPoints float64 `json:"maximumPoints"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := calcuttaService.UpdatePortfolioScores(r.Context(), portfolioID, request.MaximumPoints); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
