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

	var entryRunID *string
	if v := r.URL.Query().Get("entry_run_id"); v != "" {
		entryRunID = &v
	} else if v := r.URL.Query().Get("strategy_generation_run_id"); v != "" {
		// Backward compat.
		entryRunID = &v
	}

	var marketShareRunID *string
	if v := r.URL.Query().Get("market_share_run_id"); v != "" {
		marketShareRunID = &v
	}
	if marketShareRunID == nil || *marketShareRunID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "market_share_run_id is required", "market_share_run_id")
		return
	}

	var gameOutcomeRunID *string
	if v := r.URL.Query().Get("game_outcome_run_id"); v != "" {
		gameOutcomeRunID = &v
	}
	if gameOutcomeRunID == nil || *gameOutcomeRunID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "game_outcome_run_id is required", "game_outcome_run_id")
		return
	}

	selectedID, marketShareSelectedID, data, err := s.app.Analytics.GetCalcuttaPredictedInvestment(ctx, calcuttaID, entryRunID, marketShareRunID, gameOutcomeRunID)
	if err != nil {
		log.Printf("Error querying predicted investment: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to query predicted investment", "")
		return
	}

	if len(data) == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "No predicted investment found for calcutta", "")
		return
	}

	results := make([]TeamPredictedInvestment, 0, len(data))
	for _, d := range data {
		results = append(results, TeamPredictedInvestment{
			TeamID:     d.TeamID,
			SchoolName: d.SchoolName,
			Seed:       d.Seed,
			Region:     d.Region,
			Rational:   d.Rational,
			Predicted:  d.Predicted,
			Delta:      d.Delta,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":                calcuttaID,
		"entry_run_id":               selectedID,
		"strategy_generation_run_id": selectedID,
		"market_share_run_id":        marketShareSelectedID,
		"game_outcome_run_id":        *gameOutcomeRunID,
		"teams":                      results,
		"count":                      len(results),
	})
}
