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

	// Query to get predicted market investment from ridge regression model
	// Naive = ML model prediction of market investment
	// Edge = Same as naive for now (will incorporate market inefficiencies later)
	// Delta = Edge - Naive (currently 0)
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
		latest_calcutta AS (
			SELECT id as calcutta_id
			FROM bronze_calcuttas
			WHERE tournament_id = (SELECT id FROM bronze_tournament)
			ORDER BY created_at DESC
			LIMIT 1
		),
		entry_count AS (
			SELECT COUNT(*) as num_entries
			FROM bronze_entries
			WHERE calcutta_id = (SELECT calcutta_id FROM latest_calcutta)
		),
		total_pool AS (
			SELECT (SELECT num_entries FROM entry_count) * 100.0 as pool_size
		)
		SELECT 
			t.id as team_id,
			t.school_name,
			t.seed,
			t.region,
			-- Naive: ML model prediction of market share × total pool
			COALESCE(spms.predicted_share_of_pool, 0.0) * (SELECT pool_size FROM total_pool) as naive,
			0.0 as delta,  -- For now, delta is always 0
			-- Edge: Same as naive for now (will incorporate market inefficiencies later)
			COALESCE(spms.predicted_share_of_pool, 0.0) * (SELECT pool_size FROM total_pool) as edge
		FROM bronze_teams t
		LEFT JOIN latest_calcutta lc ON true
		LEFT JOIN silver_predicted_market_share spms 
			ON spms.calcutta_id = lc.calcutta_id AND spms.team_id = t.id
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
