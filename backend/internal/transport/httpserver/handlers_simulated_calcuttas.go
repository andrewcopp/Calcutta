package httpserver

import (
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

// handleGetCalcuttaSimulatedCalcuttas handles GET /analytics/calcuttas/{id}/simulated-calcuttas
func (s *Server) handleGetCalcuttaSimulatedCalcuttas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]

	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	// Get the latest optimization run for this calcutta.
	// This requires a mapping from core.calcuttas.id -> bronze_calcuttas.id via bronze_calcuttas.core_calcutta_id.
	queryRun := `
		SELECT gor.run_id
		FROM gold.optimization_runs gor
		JOIN bronze_calcuttas_core_ctx bcc ON bcc.id = gor.calcutta_id
		WHERE bcc.core_calcutta_id = $1
		ORDER BY gor.created_at DESC
		LIMIT 1
	`

	var runID string
	err := s.pool.QueryRow(ctx, queryRun, calcuttaID).Scan(&runID)
	if err != nil {
		log.Printf("Error getting latest optimization run: %v", err)
		writeError(w, r, http.StatusNotFound, "not_found", "No optimization run found for calcutta", "")
		return
	}

	if runID == "" {
		writeError(w, r, http.StatusNotFound, "not_found", "No optimization run found for calcutta", "")
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
			(SELECT COUNT(*) FROM gold.entry_simulation_outcomes WHERE run_id = $1 AND entry_name = gep.entry_name) as total_sims
		FROM gold.entry_performance gep
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
		writeError(w, r, http.StatusNotFound, "not_found", "No simulated calcutta data found for calcutta", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id": calcuttaID,
		"run_id":      runID,
		"entries":     results,
	})
}
