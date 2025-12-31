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

	// Query the database to get simulation statistics for this tournament
	query := `
		WITH tournament_info AS (
			SELECT id, season
			FROM bronze_tournaments
			WHERE id = $1
		),
		sim_stats AS (
			SELECT 
				COUNT(DISTINCT sim_id) as total_simulations,
				COUNT(DISTINCT team_id) as total_teams
			FROM silver_simulated_tournaments st
			JOIN tournament_info ti ON st.tournament_id = ti.id
		),
		prediction_stats AS (
			SELECT COUNT(*) as total_predictions
			FROM silver_predicted_game_outcomes pgo
			JOIN tournament_info ti ON pgo.tournament_id = ti.id
		),
		win_stats AS (
			SELECT 
				AVG(wins)::numeric as mean_wins,
				PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY wins) as median_wins,
				MAX(wins) as max_wins
			FROM silver_simulated_tournaments st
			JOIN tournament_info ti ON st.tournament_id = ti.id
		)
		SELECT 
			ti.id as tournament_id,
			ti.season,
			COALESCE(ss.total_simulations, 0) as total_simulations,
			COALESCE(ps.total_predictions, 0) as total_predictions,
			COALESCE(ws.mean_wins, 0) as mean_wins,
			COALESCE(ws.median_wins, 0) as median_wins,
			COALESCE(ws.max_wins, 0) as max_wins,
			NOW() as last_updated
		FROM tournament_info ti
		LEFT JOIN sim_stats ss ON true
		LEFT JOIN prediction_stats ps ON true
		LEFT JOIN win_stats ws ON true
	`

	var stats struct {
		TournamentID     string  `db:"tournament_id"`
		Season           int     `db:"season"`
		TotalSimulations int     `db:"total_simulations"`
		TotalPredictions int     `db:"total_predictions"`
		MeanWins         float64 `db:"mean_wins"`
		MedianWins       float64 `db:"median_wins"`
		MaxWins          int     `db:"max_wins"`
		LastUpdated      string  `db:"last_updated"`
	}

	err := s.pool.QueryRow(ctx, query, tournamentID).Scan(
		&stats.TournamentID,
		&stats.Season,
		&stats.TotalSimulations,
		&stats.TotalPredictions,
		&stats.MeanWins,
		&stats.MedianWins,
		&stats.MaxWins,
		&stats.LastUpdated,
	)

	if err != nil {
		log.Printf("Error getting tournament sim stats by ID: %v", err)
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
