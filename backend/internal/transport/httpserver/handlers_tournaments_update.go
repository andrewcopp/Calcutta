package httpserver

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/gorilla/mux"
)

func (s *Server) updateTournamentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	var req dtos.UpdateTournamentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if err := s.app.Tournament.UpdateStartingAt(r.Context(), tournamentID, req.StartingAt.Value); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	tournament, err := s.app.Tournament.GetByID(r.Context(), tournamentID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if tournament == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Tournament not found", "id")
		return
	}

	team, err := s.app.Tournament.GetWinningTeam(r.Context(), tournament.ID)
	if err != nil {
		log.Printf("Error getting winning team for tournament %s: %v", tournament.ID, err)
	}

	winnerName := ""
	if team != nil {
		school, err := s.app.School.GetByID(r.Context(), team.SchoolID)
		if err != nil {
			log.Printf("Error getting school for team %s: %v", team.ID, err)
		} else if school != nil {
			winnerName = school.Name
		}
	}

	writeJSON(w, http.StatusOK, dtos.NewTournamentResponse(tournament, winnerName))
}
