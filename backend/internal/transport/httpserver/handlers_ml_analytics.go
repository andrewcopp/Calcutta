package httpserver

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// handleGetTournamentSimStats handles GET /tournaments/{year}/simulations
func (s *Server) handleGetTournamentSimStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}

	stats, err := s.app.MLAnalytics.GetTournamentSimStats(ctx, year)
	if err != nil {
		log.Printf("Error getting tournament sim stats: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	if stats == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Tournament simulations not found", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id": stats.TournamentID,
		"season":        stats.Season,
		"n_sims":        stats.NSims,
		"n_teams":       stats.NTeams,
		"avg_progress":  stats.AvgProgress,
		"max_progress":  stats.MaxProgress,
	})
}

// handleGetTeamPerformance handles GET /tournaments/{year}/teams/{team_id}/performance
func (s *Server) handleGetTeamPerformance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}
	teamID := vars["team_id"]
	if teamID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing team_id parameter", "team_id")
		return
	}

	perf, err := s.app.MLAnalytics.GetTeamPerformance(ctx, year, teamID)
	if err != nil {
		log.Printf("Error getting team performance: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	if perf == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Team performance not found", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"team_id":            perf.TeamID,
		"school_name":        perf.SchoolName,
		"seed":               perf.Seed,
		"region":             perf.Region,
		"kenpom_net":         perf.KenpomNet,
		"total_sims":         perf.TotalSims,
		"avg_wins":           perf.AvgWins,
		"round_distribution": perf.RoundDistribution,
	})
}

// handleGetTeamPerformanceByCalcutta handles GET /api/v1/analytics/calcuttas/{calcutta_id}/teams/{team_id}/performance
func (s *Server) handleGetTeamPerformanceByCalcutta(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["calcutta_id"]
	teamID := vars["team_id"]
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta_id parameter", "calcutta_id")
		return
	}
	if teamID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing team_id parameter", "team_id")
		return
	}

	perf, err := s.app.MLAnalytics.GetTeamPerformanceByCalcutta(ctx, calcuttaID, teamID)
	if err != nil {
		log.Printf("Error getting team performance by calcutta: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to get team performance", "")
		return
	}
	if perf == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Team performance not found", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":        calcuttaID,
		"team_id":            perf.TeamID,
		"school_name":        perf.SchoolName,
		"seed":               perf.Seed,
		"region":             perf.Region,
		"kenpom_net":         perf.KenpomNet,
		"total_sims":         perf.TotalSims,
		"avg_wins":           perf.AvgWins,
		"avg_points":         perf.AvgPoints,
		"round_distribution": perf.RoundDistribution,
	})
}

// handleGetTeamPredictions handles GET /tournaments/{year}/teams/predictions
func (s *Server) handleGetTeamPredictions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}

	// Optional run_id query parameter
	var runID *string
	if rid := r.URL.Query().Get("run_id"); rid != "" {
		runID = &rid
	}

	predictions, err := s.app.MLAnalytics.GetTeamPredictions(ctx, year, runID)
	if err != nil {
		log.Printf("Error getting team predictions: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	teams := make([]map[string]interface{}, len(predictions))
	for i, pred := range predictions {
		teams[i] = map[string]interface{}{
			"team_id":     pred.TeamID,
			"school_name": pred.SchoolName,
			"seed":        pred.Seed,
			"region":      pred.Region,
			"kenpom_net":  pred.KenpomNet,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"year":  year,
		"teams": teams,
	})
}
