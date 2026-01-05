package httpserver

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type TeamPredictedMarketShare struct {
	TeamID         string  `json:"team_id"`
	SchoolName     string  `json:"school_name"`
	Seed           int     `json:"seed"`
	Region         string  `json:"region"`
	RationalShare  float64 `json:"rational_share"`
	PredictedShare float64 `json:"predicted_share"`
	DeltaPercent   float64 `json:"delta_percent"`
}

// handleGetCalcuttaPredictedMarketShare handles GET /analytics/calcuttas/{id}/predicted-market-share
func (s *Server) handleGetCalcuttaPredictedMarketShare(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	var marketShareRunID *string
	if v := r.URL.Query().Get("market_share_run_id"); v != "" {
		marketShareRunID = &v
	}

	var gameOutcomeRunID *string
	if v := r.URL.Query().Get("game_outcome_run_id"); v != "" {
		gameOutcomeRunID = &v
	}

	marketShareSelectedID, gameOutcomeSelectedID, data, err := s.app.Analytics.GetCalcuttaPredictedMarketShare(ctx, calcuttaID, marketShareRunID, gameOutcomeRunID)
	if err != nil {
		log.Printf("Error querying predicted market share: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to query predicted market share", "")
		return
	}
	if len(data) == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "No predicted market share found for calcutta", "")
		return
	}

	results := make([]TeamPredictedMarketShare, 0, len(data))
	for _, d := range data {
		results = append(results, TeamPredictedMarketShare{
			TeamID:         d.TeamID,
			SchoolName:     d.SchoolName,
			Seed:           d.Seed,
			Region:         d.Region,
			RationalShare:  d.RationalShare,
			PredictedShare: d.PredictedShare,
			DeltaPercent:   d.DeltaPercent,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":         calcuttaID,
		"market_share_run_id": marketShareSelectedID,
		"game_outcome_run_id": gameOutcomeSelectedID,
		"teams":               results,
		"count":               len(results),
	})
}
