package httpserver

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// TeamSimulatedEntry represents simulated entry data for a single team
type TeamSimulatedEntry struct {
	TeamID         string  `json:"team_id"`
	SchoolName     string  `json:"school_name"`
	Seed           int     `json:"seed"`
	Region         string  `json:"region"`
	ExpectedPoints float64 `json:"expected_points"` // Expected points from simulations
	ExpectedMarket float64 `json:"expected_market"` // Predicted market investment
	ExpectedROI    float64 `json:"expected_roi"`    // Expected ROI (points / market)
	OurBid         float64 `json:"our_bid"`         // Our recommended bid (0 for now)
	OurROI         float64 `json:"our_roi"`         // Our ROI accounting for our bid
}

// handleGetCalcuttaSimulatedEntry handles GET /analytics/calcuttas/{id}/simulated-entry
func (s *Server) handleGetCalcuttaSimulatedEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]

	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	data, err := s.app.Analytics.GetCalcuttaSimulatedEntry(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error querying simulated entry: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to query simulated entry", "")
		return
	}

	if len(data) == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "No simulated entry found for calcutta", "")
		return
	}

	results := make([]TeamSimulatedEntry, 0, len(data))
	for _, d := range data {
		se := TeamSimulatedEntry{
			TeamID:         d.TeamID,
			SchoolName:     d.SchoolName,
			Seed:           d.Seed,
			Region:         d.Region,
			ExpectedPoints: d.ExpectedPoints,
			ExpectedMarket: d.ExpectedMarket,
			OurBid:         d.OurBid,
		}

		if se.ExpectedMarket > 0 {
			se.ExpectedROI = se.ExpectedPoints / se.ExpectedMarket
		} else {
			se.ExpectedROI = 0.0
		}

		totalMarket := se.ExpectedMarket + se.OurBid
		if totalMarket > 0 {
			se.OurROI = se.ExpectedPoints / totalMarket
		} else {
			se.OurROI = 0.0
		}

		results = append(results, se)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id": calcuttaID,
		"teams":       results,
		"count":       len(results),
	})
}
