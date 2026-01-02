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

// handleGetCalcuttaSimulatedEntry handles GET /analytics/calcuttas/{id}/simulated-entry
func (s *Server) handleGetCalcuttaSimulatedEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]

	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	// Query to calculate simulated entry metrics for each team
	// Uses ridge regression predictions from silver_predicted_market_share for expected_market
	// Reads Our Bid from gold_recommended_entry_bids if available (from MINLP optimizer)
	query := `
		WITH calcutta AS (
			SELECT c.id AS calcutta_id
			FROM core.calcuttas c
			WHERE c.id = $1
			  AND c.deleted_at IS NULL
			LIMIT 1
		),
		bronze_calcutta AS (
			SELECT
				bcc.id AS bronze_calcutta_id,
				bcc.tournament_id AS bronze_tournament_id,
				bcc.season AS season
			FROM bronze_calcuttas_core_ctx bcc
			JOIN calcutta c ON c.calcutta_id = bcc.core_calcutta_id
			LIMIT 1
		),
		bronze_tournament AS (
			SELECT bronze_tournament_id AS id
			FROM bronze_calcutta
		),
		entry_count AS (
			SELECT COUNT(DISTINCT entry_name) as num_entries
			FROM bronze.entry_bids beb
			WHERE beb.calcutta_id = (SELECT bronze_calcutta_id FROM bronze_calcutta)
		),
		total_pool AS (
			SELECT COALESCE(NULLIF((SELECT num_entries FROM entry_count), 0), 47) * 100.0 as pool_size
		),
		team_expected_points AS (
			SELECT
				st.team_id,
				AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta), st.wins, st.byes))::float AS expected_points
			FROM silver.simulated_tournaments st
			WHERE st.tournament_id = (SELECT id FROM bronze_tournament)
			GROUP BY st.team_id
		),
		latest_optimization AS (
			SELECT gor.run_id
			FROM gold.optimization_runs gor
			WHERE gor.calcutta_id = (SELECT bronze_calcutta_id FROM bronze_calcutta)
			ORDER BY gor.created_at DESC
			LIMIT 1
		)
		SELECT
			t.id as team_id,
			t.school_name,
			t.seed,
			t.region,
			COALESCE(tep.expected_points, 0.0) as expected_points,
			-- Expected market: ML model prediction × total pool (based on actual entry count)
			COALESCE(spms_c.predicted_share, spms_t.predicted_share, 0.0) * (SELECT pool_size FROM total_pool) as expected_market,
			-- Our bid from MINLP optimizer (0 if not available)
			COALESCE(reb.recommended_bid_points, 0.0) as our_bid
		FROM bronze.teams t
		LEFT JOIN team_expected_points tep ON t.id = tep.team_id
		LEFT JOIN silver.predicted_market_share spms_c
			ON spms_c.calcutta_id = (SELECT bronze_calcutta_id FROM bronze_calcutta) AND spms_c.team_id = t.id
		LEFT JOIN silver.predicted_market_share spms_t
			ON spms_t.tournament_id = (SELECT id FROM bronze_tournament)
			AND spms_t.calcutta_id IS NULL
			AND spms_t.team_id = t.id
		LEFT JOIN latest_optimization lo ON true
		LEFT JOIN gold.recommended_entry_bids reb ON reb.run_id = lo.run_id AND reb.team_id = t.id
		WHERE t.tournament_id = (SELECT id FROM bronze_tournament)
		ORDER BY t.seed ASC, t.school_name ASC
	`

	rows, err := s.pool.Query(ctx, query, calcuttaID)
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
		writeError(w, r, http.StatusNotFound, "not_found", "No simulated entry found for calcutta", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id": calcuttaID,
		"teams":       results,
		"count":       len(results),
	})
}
