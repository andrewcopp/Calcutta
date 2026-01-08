package httpserver

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type TeamPredictedAdvancement struct {
	TeamID     string  `json:"team_id"`
	SchoolName string  `json:"school_name"`
	Seed       int     `json:"seed"`
	Region     string  `json:"region"`
	ProbPI     float64 `json:"prob_pi"`
	ReachR64   float64 `json:"reach_r64"`
	ReachR32   float64 `json:"reach_r32"`
	ReachS16   float64 `json:"reach_s16"`
	ReachE8    float64 `json:"reach_e8"`
	ReachFF    float64 `json:"reach_ff"`
	ReachChamp float64 `json:"reach_champ"`
	WinChamp   float64 `json:"win_champ"`
}

// handleGetTournamentPredictedAdvancement handles GET /analytics/tournaments/{id}/predicted-advancement
func (s *Server) handleGetTournamentPredictedAdvancement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
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

	selectedRunID, data, err := s.app.Analytics.GetTournamentPredictedAdvancement(ctx, tournamentID, gameOutcomeRunID)
	if err != nil {
		log.Printf("Error querying tournament predicted advancement: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to query predicted advancement", "")
		return
	}
	if len(data) == 0 {
		writeError(w, r, http.StatusNotFound, "not_found", "No predicted advancement found for tournament", "")
		return
	}

	results := make([]TeamPredictedAdvancement, 0, len(data))
	for _, d := range data {
		results = append(results, TeamPredictedAdvancement{
			TeamID:     d.TeamID,
			SchoolName: d.SchoolName,
			Seed:       d.Seed,
			Region:     d.Region,
			ProbPI:     d.ProbPI,
			ReachR64:   d.ReachR64,
			ReachR32:   d.ReachR32,
			ReachS16:   d.ReachS16,
			ReachE8:    d.ReachE8,
			ReachFF:    d.ReachFF,
			ReachChamp: d.ReachChamp,
			WinChamp:   d.WinChamp,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id":       tournamentID,
		"game_outcome_run_id": selectedRunID,
		"teams":               results,
		"count":               len(results),
	})
}
