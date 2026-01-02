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

// handleGetCalcuttaPredictedReturns handles GET /analytics/calcuttas/{id}/predicted-returns
func (s *Server) handleGetCalcuttaPredictedReturns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]

	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	// Query to calculate probabilities for each team
	// Wins mapping: 0=eliminated before R64, 1=won R64, 2=won R32, 3=won S16, 4=won E8, 5=won FF, 6=won Championship
	query := `
		WITH calcutta AS (
			SELECT c.id AS calcutta_id
			FROM core.calcuttas c
			WHERE c.id = $1
			  AND c.deleted_at IS NULL
			LIMIT 1
		),
		bronze_calcutta AS (
			SELECT bcc.tournament_id AS bronze_tournament_id
			FROM bronze_calcuttas_core_ctx bcc
			JOIN calcutta c ON c.calcutta_id = bcc.core_calcutta_id
			LIMIT 1
		),
		bronze_tournament AS (
			SELECT bronze_tournament_id AS id
			FROM bronze_calcutta
		),
		team_win_counts AS (
			SELECT
				st.team_id,
				st.wins,
				COUNT(*) as sim_count
			FROM silver.simulated_tournaments st
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
		),
		team_expected_value AS (
			SELECT
				st.team_id,
				AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta), st.wins, st.byes))::float AS expected_value
			FROM silver.simulated_tournaments st
			WHERE st.tournament_id = (SELECT id FROM bronze_tournament)
			GROUP BY st.team_id
		)
		SELECT
			t.id as team_id,
			t.school_name,
			t.seed,
			t.region,
			COALESCE(tp.win_pi / NULLIF(tp.total_sims, 0), 0) as prob_pi,
			COALESCE(tp.win_r64 / NULLIF(tp.total_sims, 0), 0) as prob_r64,
			COALESCE(tp.win_r32 / NULLIF(tp.total_sims, 0), 0) as prob_r32,
			COALESCE(tp.win_s16 / NULLIF(tp.total_sims, 0), 0) as prob_s16,
			COALESCE(tp.win_e8 / NULLIF(tp.total_sims, 0), 0) as prob_e8,
			COALESCE(tp.win_ff / NULLIF(tp.total_sims, 0), 0) as prob_ff,
			COALESCE(tp.win_champ / NULLIF(tp.total_sims, 0), 0) as prob_champ,
			COALESCE(tev.expected_value, 0.0) as expected_value
		FROM bronze.teams t
		LEFT JOIN team_probabilities tp ON t.id = tp.team_id
		LEFT JOIN team_expected_value tev ON t.id = tev.team_id
		WHERE t.tournament_id = (SELECT id FROM bronze_tournament)
		ORDER BY expected_value DESC, t.seed ASC
	`

	rows, err := s.pool.Query(ctx, query, calcuttaID)
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
		writeError(w, r, http.StatusNotFound, "not_found", "No predicted returns found for calcutta", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id": calcuttaID,
		"teams":       results,
		"count":       len(results),
	})
}
