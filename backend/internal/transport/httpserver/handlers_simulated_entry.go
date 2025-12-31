package httpserver

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// TeamSimulatedEntry represents simulated entry data for a single team
type TeamSimulatedEntry struct {
	TeamID         string  `json:"team_id"`
	SchoolName     string  `json:"school_name"`
	Seed           int     `json:"seed"`
	Region         string  `json:"region"`
	ExpectedPoints float64 `json:"expected_points"` // Expected points from simulations
	ExpectedMarket float64 `json:"expected_market"` // Predicted market investment
	ExpectedROI    float64 `json:"expected_roi"`    // Expected ROI (points / market)
	OurBid         float64 `json:"our_bid"`         // Our recommended bid (0 for now)
	OurROI         float64 `json:"our_roi"`         // Our ROI accounting for our bid
}

// handleGetTournamentSimulatedEntry handles GET /analytics/tournaments/{id}/simulated-entry
func (s *Server) handleGetTournamentSimulatedEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
		return
	}

	// Query to calculate simulated entry metrics for each team
	// Uses ridge regression predictions from silver_predicted_market_share for expected_market
	// Reads Our Bid from gold_recommended_entry_bids if available (from MINLP optimizer)
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
		entry_count AS (
			SELECT COUNT(DISTINCT entry_name) as num_entries
			FROM bronze_entry_bids beb
			JOIN bronze_calcuttas bc ON beb.calcutta_id = bc.id
			WHERE bc.tournament_id = (SELECT id FROM bronze_tournament)
		),
		total_pool AS (
			SELECT COALESCE(NULLIF((SELECT num_entries FROM entry_count), 0), 47) * 100.0 as pool_size
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
				SUM(CASE WHEN wins = 0 THEN sim_count ELSE 0 END)::float as win_pi,
				SUM(CASE WHEN wins = 1 THEN sim_count ELSE 0 END)::float as win_r64,
				SUM(CASE WHEN wins = 2 THEN sim_count ELSE 0 END)::float as win_r32,
				SUM(CASE WHEN wins = 3 THEN sim_count ELSE 0 END)::float as win_s16,
				SUM(CASE WHEN wins = 4 THEN sim_count ELSE 0 END)::float as win_e8,
				SUM(CASE WHEN wins = 5 THEN sim_count ELSE 0 END)::float as win_ff,
				SUM(CASE WHEN wins = 6 THEN sim_count ELSE 0 END)::float as win_champ
			FROM team_win_counts
			GROUP BY team_id
		),
		latest_optimization AS (
			SELECT run_id
			FROM gold_optimization_runs
			ORDER BY created_at DESC
			LIMIT 1
		)
		SELECT 
			t.id as team_id,
			t.school_name,
			t.seed,
			t.region,
			-- Expected points (EV calculation)
			(COALESCE(tp.win_r64 / NULLIF(tp.total_sims, 0), 0) * 50 + 
			 COALESCE(tp.win_r32 / NULLIF(tp.total_sims, 0), 0) * 150 + 
			 COALESCE(tp.win_s16 / NULLIF(tp.total_sims, 0), 0) * 300 + 
			 COALESCE(tp.win_e8 / NULLIF(tp.total_sims, 0), 0) * 500 + 
			 COALESCE(tp.win_ff / NULLIF(tp.total_sims, 0), 0) * 750 + 
			 COALESCE(tp.win_champ / NULLIF(tp.total_sims, 0), 0) * 1050) as expected_points,
			-- Expected market: ML model prediction × total pool (based on actual entry count)
			COALESCE(spms.predicted_share, 0.0) * (SELECT pool_size FROM total_pool) as expected_market,
			-- Our bid from MINLP optimizer (0 if not available)
			COALESCE(reb.recommended_bid_points, 0.0) as our_bid
		FROM bronze_teams t
		LEFT JOIN team_probabilities tp ON t.id = tp.team_id
		LEFT JOIN silver_predicted_market_share spms 
			ON spms.tournament_id = (SELECT id FROM bronze_tournament) AND spms.team_id = t.id
		LEFT JOIN latest_optimization lo ON true
		LEFT JOIN gold_recommended_entry_bids reb ON reb.run_id = lo.run_id AND reb.team_id = t.id
		WHERE t.tournament_id = (SELECT id FROM bronze_tournament)
		ORDER BY t.seed ASC, t.school_name ASC
	`

	rows, err := s.pool.Query(ctx, query, tournamentID)
	if err != nil {
		log.Printf("Error querying simulated entry: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to query simulated entry", "")
		return
	}
	defer rows.Close()

	var results []TeamSimulatedEntry
	for rows.Next() {
		var se TeamSimulatedEntry
		err := rows.Scan(
			&se.TeamID,
			&se.SchoolName,
			&se.Seed,
			&se.Region,
			&se.ExpectedPoints,
			&se.ExpectedMarket,
			&se.OurBid,
		)
		if err != nil {
			log.Printf("Error scanning simulated entry row: %v", err)
			continue
		}

		// Calculate expected ROI
		if se.ExpectedMarket > 0 {
			se.ExpectedROI = se.ExpectedPoints / se.ExpectedMarket
		} else {
			se.ExpectedROI = 0.0
		}

		// Calculate our ROI (accounting for our bid moving the market)
		totalMarket := se.ExpectedMarket + se.OurBid
		if totalMarket > 0 {
			se.OurROI = se.ExpectedPoints / totalMarket
		} else {
			se.OurROI = 0.0
		}

		results = append(results, se)
	}

	if len(results) == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "No simulated entry found for tournament", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id": tournamentID,
		"teams":         results,
		"count":         len(results),
	})
}
