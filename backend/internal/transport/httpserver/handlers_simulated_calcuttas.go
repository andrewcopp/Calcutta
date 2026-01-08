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
	MeanPayout       float64 `json:"mean_normalized_payout"`
	MedianPayout     float64 `json:"median_normalized_payout"`
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

	var calcuttaEvaluationRunID *string
	if v := r.URL.Query().Get("calcutta_evaluation_run_id"); v != "" {
		calcuttaEvaluationRunID = &v
	}

	runID, evalRunID, data, err := s.app.MLAnalytics.GetSimulatedCalcuttaEntryRankings(ctx, calcuttaID, calcuttaEvaluationRunID)
	if err != nil {
		log.Printf("Error querying simulated calcuttas: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to query simulated calcuttas", "")
		return
	}

	if (runID == "" && evalRunID == nil) || len(data) == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "No simulated calcutta data found for calcutta", "")
		return
	}

	var focusEntryName *string
	if evalRunID != nil && *evalRunID != "" {
		var name string
		if err := s.pool.QueryRow(ctx, `
			WITH focus AS (
				SELECT sr.focus_snapshot_entry_id
				FROM derived.simulation_runs sr
				WHERE sr.calcutta_evaluation_run_id = $1::uuid
					AND sr.deleted_at IS NULL
				LIMIT 1
			)
			SELECT se.display_name
			FROM core.calcutta_snapshot_entries se
			WHERE se.id = (SELECT focus_snapshot_entry_id FROM focus)
				AND se.deleted_at IS NULL
			LIMIT 1
		`, *evalRunID).Scan(&name); err == nil {
			focusEntryName = &name
		}
	} else {
		var name string
		if err := s.pool.QueryRow(ctx, `
			WITH latest AS (
				SELECT sr.focus_snapshot_entry_id
				FROM derived.simulation_runs sr
				WHERE sr.calcutta_id = $1::uuid
					AND sr.deleted_at IS NULL
					AND sr.focus_snapshot_entry_id IS NOT NULL
				ORDER BY sr.created_at DESC
				LIMIT 1
			)
			SELECT se.display_name
			FROM core.calcutta_snapshot_entries se
			WHERE se.id = (SELECT focus_snapshot_entry_id FROM latest)
				AND se.deleted_at IS NULL
			LIMIT 1
		`, calcuttaID).Scan(&name); err == nil {
			focusEntryName = &name
		}
	}

	results := make([]EntryRanking, 0, len(data))
	for _, d := range data {
		isOurStrategy := focusEntryName != nil && d.EntryName == *focusEntryName
		results = append(results, EntryRanking{
			Rank:             d.Rank,
			EntryName:        d.EntryName,
			IsOurStrategy:    isOurStrategy,
			MeanPayout:       d.MeanNormalizedPayout,
			MedianPayout:     d.MedianNormalizedPayout,
			PTop1:            d.PTop1,
			PInMoney:         d.PInMoney,
			TotalSimulations: d.TotalSimulations,
		})
	}

	var runIDOut interface{} = runID
	if runID == "" {
		runIDOut = nil
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":                calcuttaID,
		"run_id":                     runIDOut,
		"calcutta_evaluation_run_id": evalRunID,
		"entries":                    results,
	})
}
