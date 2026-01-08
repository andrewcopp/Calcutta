package mlanalytics

import (
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

// HandleGetTournamentSimStatsByID handles GET /analytics/tournaments/{id}/simulations
// This is a convenience endpoint that accepts tournament ID instead of year
func (h *Handler) HandleGetTournamentSimStatsByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
		return
	}

	stats, err := h.app.MLAnalytics.GetTournamentSimStatsByCoreTournamentID(ctx, tournamentID)
	if err != nil {
		log.Printf("Error getting tournament sim stats by ID: %v", err)
		httperr.Write(w, r, http.StatusInternalServerError, "database_error", "Failed to get tournament sim stats", "")
		return
	}
	if stats == nil {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "Tournament not found or no simulation data available", "")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
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
