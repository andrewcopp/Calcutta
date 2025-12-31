package httpserver

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// TeamPredictedInvestment represents predicted investment for a single team
type TeamPredictedInvestment struct {
	TeamID     string  `json:"team_id"`
	SchoolName string  `json:"school_name"`
	Seed       int     `json:"seed"`
	Region     string  `json:"region"`
	Naive      float64 `json:"naive"` // Naive expected value (what you might eyeball)
	Delta      float64 `json:"delta"` // Difference between edge and naive (positive = overinvested, negative = underinvested)
	Edge       float64 `json:"edge"`  // Our edge calculation (opportunities for under/over investment)
}

// handleGetTournamentPredictedInvestment handles GET /analytics/tournaments/{id}/predicted-investment
func (s *Server) handleGetTournamentPredictedInvestment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
		return
	}

	// Query to calculate investment metrics for each team
	// For now, Naive and Edge are both set to EV, Delta is 0
	// In the future, Edge will incorporate market inefficiencies
	query := `
		WITH main_tournament AS (
			SELECT 
				id,
				CAST(SUBSTRING(name FROM '[0-9]{4}') AS INTEGER) as season
			FROM tournaments
			WHERE id = $1
		),
		bronze_tournament AS (
			SELECT id
			FROM bronze_tournaments
			WHERE season = (SELECT season FROM main_tournament)
		),
		team_win_counts AS (
			SELECT 
				st.team_id,
				st.wins,
				COUNT(*) as sim_count
			FROM silver_simulated_tournaments st
			WHERE st.tournament_id = (SELECT id FROM bronze_tournament)
			GROUP BY st.team_id, st.wins
		),
		team_probabilities AS (
			SELECT 
				team_id,
				SUM(sim_count)::float as total_sims,
				-- Probability of winning each specific round (not cumulative)
				SUM(CASE WHEN wins = 0 THEN sim_count ELSE 0 END)::float as win_pi,
				SUM(CASE WHEN wins = 1 THEN sim_count ELSE 0 END)::float as win_r64,
				SUM(CASE WHEN wins = 2 THEN sim_count ELSE 0 END)::float as win_r32,
				SUM(CASE WHEN wins = 3 THEN sim_count ELSE 0 END)::float as win_s16,
				SUM(CASE WHEN wins = 4 THEN sim_count ELSE 0 END)::float as win_e8,
				SUM(CASE WHEN wins = 5 THEN sim_count ELSE 0 END)::float as win_ff,
				SUM(CASE WHEN wins = 6 THEN sim_count ELSE 0 END)::float as win_champ
			FROM team_win_counts
			GROUP BY team_id
		)
		SELECT 
			t.id as team_id,
			t.school_name,
			t.seed,
			t.region,
			-- Naive: Predicted market investment (proportional share of 6800 point pool)
			(COALESCE(tp.win_r64 / NULLIF(tp.total_sims, 0), 0) * 50 + 
			 COALESCE(tp.win_r32 / NULLIF(tp.total_sims, 0), 0) * 150 + 
			 COALESCE(tp.win_s16 / NULLIF(tp.total_sims, 0), 0) * 300 + 
			 COALESCE(tp.win_e8 / NULLIF(tp.total_sims, 0), 0) * 500 + 
			 COALESCE(tp.win_ff / NULLIF(tp.total_sims, 0), 0) * 750 + 
			 COALESCE(tp.win_champ / NULLIF(tp.total_sims, 0), 0) * 1050) / 6000.0 * 6800.0 as naive,
			0.0 as delta,  -- For now, delta is always 0
			-- Edge is same as naive for now (will incorporate market inefficiencies later)
			(COALESCE(tp.win_r64 / NULLIF(tp.total_sims, 0), 0) * 50 + 
			 COALESCE(tp.win_r32 / NULLIF(tp.total_sims, 0), 0) * 150 + 
			 COALESCE(tp.win_s16 / NULLIF(tp.total_sims, 0), 0) * 300 + 
			 COALESCE(tp.win_e8 / NULLIF(tp.total_sims, 0), 0) * 500 + 
			 COALESCE(tp.win_ff / NULLIF(tp.total_sims, 0), 0) * 750 + 
			 COALESCE(tp.win_champ / NULLIF(tp.total_sims, 0), 0) * 1050) / 6000.0 * 6800.0 as edge
		FROM bronze_teams t
		LEFT JOIN team_probabilities tp ON t.id = tp.team_id
		WHERE t.tournament_id = (SELECT id FROM bronze_tournament)
		ORDER BY naive DESC, t.seed ASC
	`

	rows, err := s.pool.Query(ctx, query, tournamentID)
	if err != nil {
		log.Printf("Error querying predicted investment: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to query predicted investment", "")
		return
	}
	defer rows.Close()

	var results []TeamPredictedInvestment
	for rows.Next() {
		var ti TeamPredictedInvestment
		err := rows.Scan(
			&ti.TeamID,
			&ti.SchoolName,
			&ti.Seed,
			&ti.Region,
			&ti.Naive,
			&ti.Delta,
			&ti.Edge,
		)
		if err != nil {
			log.Printf("Error scanning predicted investment row: %v", err)
			continue
		}
		results = append(results, ti)
	}

	if len(results) == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "No predicted investment found for tournament", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id": tournamentID,
		"teams":         results,
		"count":         len(results),
	})
}
