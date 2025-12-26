package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/gorilla/mux"
)

func (s *Server) calculatePortfolioScoresHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["id"]

	if err := s.app.Calcutta.CalculatePortfolioScores(r.Context(), portfolioID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) updatePortfolioTeamScoresHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["id"]
	teamID := vars["teamId"]

	var req dtos.UpdatePortfolioTeamScoresRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	teams, err := s.app.Calcutta.GetPortfolioTeams(r.Context(), portfolioID)
	if err != nil {
		writeErrorFromErr(w, r, err)
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
		writeError(w, r, http.StatusNotFound, "not_found", "Portfolio team not found", "")
		return
	}

	portfolioTeam.ExpectedPoints = req.ExpectedPoints
	portfolioTeam.PredictedPoints = req.PredictedPoints
	portfolioTeam.Updated = time.Now()

	if err := s.app.Calcutta.UpdatePortfolioTeam(r.Context(), portfolioTeam); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) updatePortfolioMaximumScoreHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["id"]

	var req dtos.UpdatePortfolioMaximumScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if err := s.app.Calcutta.UpdatePortfolioScores(r.Context(), portfolioID, req.MaximumPoints); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
