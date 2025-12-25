package httpserver

import (
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/gorilla/mux"
)

func (s *Server) tournamentHandler(w http.ResponseWriter, r *http.Request) {
	// Extract tournament ID from URL path
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	// Get tournament by ID
	tournament, err := s.app.Tournament.GetByID(r.Context(), tournamentID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if tournament == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Tournament not found", "id")
		return
	}

	// Get the winning team for this tournament
	team, err := s.app.Tournament.GetWinningTeam(r.Context(), tournament.ID)
	if err != nil {
		log.Printf("Error getting winning team for tournament %s: %v", tournament.ID, err)
	}

	winnerName := ""
	if team != nil {
		// Get the school name
		school, err := s.app.School.GetByID(r.Context(), team.SchoolID)
		if err != nil {
			log.Printf("Error getting school for team %s: %v", team.ID, err)
		} else {
			if school != nil {
				winnerName = school.Name
			}
		}
	}

	writeJSON(w, http.StatusOK, dtos.NewTournamentResponse(tournament, winnerName))
}
