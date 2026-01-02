package httpserver

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// handleGetTournamentSimStatsByID handles GET /analytics/tournaments/{id}/simulations
// This is a convenience endpoint that accepts tournament ID instead of year
func (s *Server) handleGetTournamentSimStatsByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
		return
	}

	stats, err := s.app.MLAnalytics.GetTournamentSimStatsByCoreTournamentID(ctx, tournamentID)
	if err != nil {
		log.Printf("Error getting tournament sim stats by ID: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to get tournament sim stats", "")
		return
	}
	if stats == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Tournament not found or no simulation data available", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id":     stats.TournamentID,
		"season":            stats.Season,
		"total_simulations": stats.TotalSimulations,
		"total_predictions": stats.TotalPredictions,
		"mean_wins":         stats.MeanWins,
		"median_wins":       stats.MedianWins,
		"max_wins":          stats.MaxWins,
		"last_updated":      stats.LastUpdated,
	})
}
