package httpserver

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// TeamPredictedReturns represents predicted returns for a single team
type TeamPredictedReturns struct {
	TeamID        string  `json:"team_id"`
	SchoolName    string  `json:"school_name"`
	Seed          int     `json:"seed"`
	Region        string  `json:"region"`
	ProbPI        float64 `json:"prob_pi"`        // Probability of winning Play-In
	ProbR64       float64 `json:"prob_r64"`       // Probability of winning R64
	ProbR32       float64 `json:"prob_r32"`       // Probability of winning R32
	ProbS16       float64 `json:"prob_s16"`       // Probability of winning Sweet 16
	ProbE8        float64 `json:"prob_e8"`        // Probability of winning Elite 8
	ProbFF        float64 `json:"prob_ff"`        // Probability of winning Final Four
	ProbChamp     float64 `json:"prob_champ"`     // Probability of winning Championship
	ExpectedValue float64 `json:"expected_value"` // Expected value in points
}

// handleGetTournamentPredictedReturns handles GET /analytics/tournaments/{id}/predicted-returns
func (s *Server) handleGetTournamentPredictedReturns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
		return
	}

	// Query to calculate probabilities for each team
	// Wins mapping: 0=eliminated before R64, 1=won R64, 2=won R32, 3=won S16, 4=won E8, 5=won FF, 6=won Championship
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
				-- Probability of reaching each round (cumulative)
				SUM(CASE WHEN wins >= 0 THEN sim_count ELSE 0 END)::float as reach_pi,
				SUM(CASE WHEN wins >= 1 THEN sim_count ELSE 0 END)::float as reach_r64,
				SUM(CASE WHEN wins >= 2 THEN sim_count ELSE 0 END)::float as reach_r32,
				SUM(CASE WHEN wins >= 3 THEN sim_count ELSE 0 END)::float as reach_s16,
				SUM(CASE WHEN wins >= 4 THEN sim_count ELSE 0 END)::float as reach_e8,
				SUM(CASE WHEN wins >= 5 THEN sim_count ELSE 0 END)::float as reach_ff,
				SUM(CASE WHEN wins >= 6 THEN sim_count ELSE 0 END)::float as reach_champ
			FROM team_win_counts
			GROUP BY team_id
		)
		SELECT 
			t.id as team_id,
			t.school_name,
			t.seed,
			t.region,
			COALESCE(tp.reach_pi / NULLIF(tp.total_sims, 0), 0) as prob_pi,
			COALESCE(tp.reach_r64 / NULLIF(tp.total_sims, 0), 0) as prob_r64,
			COALESCE(tp.reach_r32 / NULLIF(tp.total_sims, 0), 0) as prob_r32,
			COALESCE(tp.reach_s16 / NULLIF(tp.total_sims, 0), 0) as prob_s16,
			COALESCE(tp.reach_e8 / NULLIF(tp.total_sims, 0), 0) as prob_e8,
			COALESCE(tp.reach_ff / NULLIF(tp.total_sims, 0), 0) as prob_ff,
			COALESCE(tp.reach_champ / NULLIF(tp.total_sims, 0), 0) as prob_champ,
			-- Expected value calculation (cumulative points)
			-- 0 for PI, 50 for R32, 150 for S16, 300 for E8, 500 for FF, 750 for Champ, 1050 for Winner
			(COALESCE(tp.reach_r32 / NULLIF(tp.total_sims, 0), 0) * 50 + 
			 COALESCE(tp.reach_s16 / NULLIF(tp.total_sims, 0), 0) * 100 + 
			 COALESCE(tp.reach_e8 / NULLIF(tp.total_sims, 0), 0) * 150 + 
			 COALESCE(tp.reach_ff / NULLIF(tp.total_sims, 0), 0) * 200 + 
			 COALESCE(tp.reach_champ / NULLIF(tp.total_sims, 0), 0) * 250 + 
			 COALESCE(tp.reach_champ / NULLIF(tp.total_sims, 0), 0) * 300) as expected_value
		FROM bronze_teams t
		LEFT JOIN team_probabilities tp ON t.id = tp.team_id
		WHERE t.tournament_id = (SELECT id FROM bronze_tournament)
		ORDER BY expected_value DESC, t.seed ASC
	`

	rows, err := s.pool.Query(ctx, query, tournamentID)
	if err != nil {
		log.Printf("Error querying predicted returns: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to query predicted returns", "")
		return
	}
	defer rows.Close()

	var results []TeamPredictedReturns
	for rows.Next() {
		var tr TeamPredictedReturns
		err := rows.Scan(
			&tr.TeamID,
			&tr.SchoolName,
			&tr.Seed,
			&tr.Region,
			&tr.ProbPI,
			&tr.ProbR64,
			&tr.ProbR32,
			&tr.ProbS16,
			&tr.ProbE8,
			&tr.ProbFF,
			&tr.ProbChamp,
			&tr.ExpectedValue,
		)
		if err != nil {
			log.Printf("Error scanning predicted returns row: %v", err)
			continue
		}
		results = append(results, tr)
	}

	if len(results) == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "No predicted returns found for tournament", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id": tournamentID,
		"teams":         results,
		"count":         len(results),
	})
}
