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

	results := make([]EntryRanking, 0, len(data))
	for _, d := range data {
		isOurStrategy := d.EntryName == "Out Strategy" || d.EntryName == "our_strategy" || d.EntryName == "Our Strategy"
		results = append(results, EntryRanking{
			Rank:             d.Rank,
			EntryName:        d.EntryName,
			IsOurStrategy:    isOurStrategy,
			MeanPayout:       d.MeanPayout,
			MedianPayout:     d.MedianPayout,
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
