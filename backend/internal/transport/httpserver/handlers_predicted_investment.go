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
	Rational   float64 `json:"rational"`  // Rational market investment (equal ROI baseline)
	Predicted  float64 `json:"predicted"` // ML model prediction (ridge regression)
	Delta      float64 `json:"delta"`     // Percentage difference (market inefficiency)
}

// handleGetCalcuttaPredictedInvestment handles GET /analytics/calcuttas/{id}/predicted-investment
func (s *Server) handleGetCalcuttaPredictedInvestment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]

	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	// Query to get predicted market investment from ridge regression model
	// Predicted market share is currently tournament-scoped in the analytics schema (silver_predicted_market_share.tournament_id)
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
				bcc.tournament_id AS bronze_tournament_id
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
		total_expected_points AS (
			SELECT SUM(expected_points) as total_ev
			FROM team_expected_points
		)
		SELECT
			t.id as team_id,
			t.school_name,
			t.seed,
			t.region,
			-- Rational: Proportional investment based on expected points (equal ROI scenario)
			(tep.expected_points / NULLIF((SELECT total_ev FROM total_expected_points), 0)) * (SELECT pool_size FROM total_pool) as rational,
			-- Predicted: ML model prediction of market share × total pool
			COALESCE(spms_c.predicted_share, spms_t.predicted_share, 0.0) * (SELECT pool_size FROM total_pool) as predicted,
			-- Delta: Percentage difference (Predicted - Rational) / Rational * 100
			CASE
				WHEN (tep.expected_points / NULLIF((SELECT total_ev FROM total_expected_points), 0)) * (SELECT pool_size FROM total_pool) > 0
				THEN ((COALESCE(spms_c.predicted_share, spms_t.predicted_share, 0.0) * (SELECT pool_size FROM total_pool)) -
				      ((tep.expected_points / NULLIF((SELECT total_ev FROM total_expected_points), 0)) * (SELECT pool_size FROM total_pool))) /
				     ((tep.expected_points / NULLIF((SELECT total_ev FROM total_expected_points), 0)) * (SELECT pool_size FROM total_pool)) * 100
				ELSE 0
			END as delta
		FROM bronze.teams t
		LEFT JOIN team_expected_points tep ON t.id = tep.team_id
		LEFT JOIN silver.predicted_market_share spms_c
			ON spms_c.calcutta_id = (SELECT bronze_calcutta_id FROM bronze_calcutta) AND spms_c.team_id = t.id
		LEFT JOIN silver.predicted_market_share spms_t
			ON spms_t.tournament_id = (SELECT id FROM bronze_tournament)
			AND spms_t.calcutta_id IS NULL
			AND spms_t.team_id = t.id
		WHERE t.tournament_id = (SELECT id FROM bronze_tournament)
		ORDER BY predicted DESC, t.seed ASC
	`

	rows, err := s.pool.Query(ctx, query, calcuttaID)
	if err != nil {
		log.Printf("Error querying predicted investment: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to query predicted investment", "")
		return
	}
	defer rows.Close()

	var results []TeamPredictedInvestment
	for rows.Next() {
		var team TeamPredictedInvestment
		if err := rows.Scan(
			&team.TeamID,
			&team.SchoolName,
			&team.Seed,
			&team.Region,
			&team.Rational,
			&team.Predicted,
			&team.Delta,
		); err != nil {
			log.Printf("Error scanning predicted investment row: %v", err)
			continue
		}
		results = append(results, team)
	}

	if len(results) == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "No predicted investment found for calcutta", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id": calcuttaID,
		"teams":       results,
		"count":       len(results),
	})
}
