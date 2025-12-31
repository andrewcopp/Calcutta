package httpserver

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// EntryRanking represents an entry's performance in simulated calcuttas
type EntryRanking struct {
	Rank             int     `json:"rank"`
	EntryName        string  `json:"entry_name"`
	IsOurStrategy    bool    `json:"is_our_strategy"`
	MeanPayout       float64 `json:"mean_payout"`
	MedianPayout     float64 `json:"median_payout"`
	PTop1            float64 `json:"p_top1"`
	PInMoney         float64 `json:"p_in_money"`
	TotalSimulations int     `json:"total_simulations"`
}

// handleGetTournamentSimulatedCalcuttas handles GET /analytics/tournaments/{id}/simulated-calcuttas
func (s *Server) handleGetTournamentSimulatedCalcuttas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
		return
	}

	// Get the latest optimization run for this tournament
	runID, err := s.getLatestOptimizationRun(ctx, tournamentID)
	if err != nil {
		log.Printf("Error getting latest optimization run: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to get optimization run", "")
		return
	}

	if runID == "" {
		writeError(w, r, http.StatusNotFound, "not_found", "No optimization run found for tournament", "")
		return
	}

	// Query entry performance rankings
	query := `
		SELECT 
			ROW_NUMBER() OVER (ORDER BY mean_payout DESC) as rank,
			entry_name,
			mean_payout,
			median_payout,
			p_top1,
			p_in_money,
			(SELECT COUNT(*) FROM gold_entry_simulation_outcomes WHERE run_id = $1 AND entry_name = gep.entry_name) as total_sims
		FROM gold_entry_performance gep
		WHERE run_id = $1
		ORDER BY mean_payout DESC
	`

	rows, err := s.pool.Query(ctx, query, runID)
	if err != nil {
		log.Printf("Error querying simulated calcuttas: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to query simulated calcuttas", "")
		return
	}
	defer rows.Close()

	var results []EntryRanking
	for rows.Next() {
		var entry EntryRanking
		if err := rows.Scan(
			&entry.Rank,
			&entry.EntryName,
			&entry.MeanPayout,
			&entry.MedianPayout,
			&entry.PTop1,
			&entry.PInMoney,
			&entry.TotalSimulations,
		); err != nil {
			log.Printf("Error scanning simulated calcutta row: %v", err)
			continue
		}

		// Mark our strategy
		entry.IsOurStrategy = entry.EntryName == "Our Strategy"

		results = append(results, entry)
	}

	if len(results) == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "No simulated calcutta data found for tournament", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id": tournamentID,
		"run_id":        runID,
		"entries":       results,
	})
}

func (s *Server) getLatestOptimizationRun(ctx context.Context, tournamentID string) (string, error) {
	// First try to find runs via gold_entry_performance (simulated calcuttas)
	// Need to map tournaments.id -> bronze_tournaments.id via year/name
	query := `
		SELECT DISTINCT gep.run_id
		FROM gold_entry_performance gep
		WHERE EXISTS (
			SELECT 1 FROM gold_recommended_entry_bids greb
			JOIN bronze_teams bt ON greb.team_id = bt.id
			JOIN bronze_tournaments btr ON bt.tournament_id = btr.id
			JOIN tournaments t ON t.name LIKE '%' || btr.season || '%'
			WHERE greb.run_id = gep.run_id
			AND t.id = $1
		)
		ORDER BY gep.run_id DESC
		LIMIT 1
	`

	var runID string
	err := s.pool.QueryRow(ctx, query, tournamentID).Scan(&runID)
	if err == nil {
		return runID, nil
	}

	// Fallback: try via bronze_calcuttas (for older data)
	fallbackQuery := `
		SELECT gor.run_id
		FROM gold_optimization_runs gor
		JOIN bronze_calcuttas bc ON gor.calcutta_id = bc.id
		WHERE bc.tournament_id = $1
		ORDER BY gor.created_at DESC
		LIMIT 1
	`

	err = s.pool.QueryRow(ctx, fallbackQuery, tournamentID).Scan(&runID)
	if err != nil {
		return "", err
	}

	return runID, nil
}
