package mlanalytics

import (
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleGetCalcuttaSimulatedCalcuttas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]

	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	var calcuttaEvaluationRunID *string
	if v := r.URL.Query().Get("calcutta_evaluation_run_id"); v != "" {
		calcuttaEvaluationRunID = &v
	}

	runID, evalRunID, data, err := h.app.MLAnalytics.GetSimulatedCalcuttaEntryRankings(ctx, calcuttaID, calcuttaEvaluationRunID)
	if err != nil {
		log.Printf("Error querying simulated calcuttas: %v", err)
		httperr.Write(w, r, http.StatusInternalServerError, "database_error", "Failed to query simulated calcuttas", "")
		return
	}

	if (runID == "" && evalRunID == nil) || len(data) == 0 {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "No simulated calcutta data found for calcutta", "")
		return
	}

	var focusEntryName *string
	if evalRunID != nil && *evalRunID != "" {
		var name string
		if err := h.pool.QueryRow(ctx, `
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
		if err := h.pool.QueryRow(ctx, `
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

	results := make([]map[string]interface{}, 0, len(data))
	for _, d := range data {
		isOurStrategy := focusEntryName != nil && d.EntryName == *focusEntryName
		results = append(results, map[string]interface{}{
			"rank":                     d.Rank,
			"entry_name":               d.EntryName,
			"is_our_strategy":          isOurStrategy,
			"mean_normalized_payout":   d.MeanNormalizedPayout,
			"median_normalized_payout": d.MedianNormalizedPayout,
			"p_top1":                   d.PTop1,
			"p_in_money":               d.PInMoney,
			"total_simulations":        d.TotalSimulations,
		})
	}

	var runIDOut interface{} = runID
	if runID == "" {
		runIDOut = nil
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":                calcuttaID,
		"run_id":                     runIDOut,
		"calcutta_evaluation_run_id": evalRunID,
		"entries":                    results,
	})
}
