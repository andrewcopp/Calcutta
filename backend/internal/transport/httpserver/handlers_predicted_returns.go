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

	var entryRunID *string
	if v := r.URL.Query().Get("entry_run_id"); v != "" {
		entryRunID = &v
	} else if v := r.URL.Query().Get("strategy_generation_run_id"); v != "" {
		// Backward compat.
		entryRunID = &v
	}

	var gameOutcomeRunID *string
	if v := r.URL.Query().Get("game_outcome_run_id"); v != "" {
		gameOutcomeRunID = &v
	}

	selectedID, gameOutcomeSelectedID, data, err := s.app.Analytics.GetCalcuttaPredictedReturns(ctx, calcuttaID, entryRunID, gameOutcomeRunID)
	if err != nil {
		log.Printf("Error querying predicted returns: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to query predicted returns", "")
		return
	}

	if len(data) == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "No predicted returns found for calcutta", "")
		return
	}

	results := make([]TeamPredictedReturns, 0, len(data))
	for _, d := range data {
		results = append(results, TeamPredictedReturns{
			TeamID:        d.TeamID,
			SchoolName:    d.SchoolName,
			Seed:          d.Seed,
			Region:        d.Region,
			ProbPI:        d.ProbPI,
			ProbR64:       d.ProbR64,
			ProbR32:       d.ProbR32,
			ProbS16:       d.ProbS16,
			ProbE8:        d.ProbE8,
			ProbFF:        d.ProbFF,
			ProbChamp:     d.ProbChamp,
			ExpectedValue: d.ExpectedValue,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":                calcuttaID,
		"entry_run_id":               selectedID,
		"strategy_generation_run_id": selectedID,
		"game_outcome_run_id":        gameOutcomeSelectedID,
		"teams":                      results,
		"count":                      len(results),
	})
}
