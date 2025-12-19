package main

import (
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
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
	tournament, err := s.tournamentService.GetTournamentByID(r.Context(), tournamentID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if tournament == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Tournament not found", "id")
		return
	}

	// Get the winning team for this tournament
	team, err := s.tournamentService.GetWinningTeam(r.Context(), tournament.ID)
	if err != nil {
		log.Printf("Error getting winning team for tournament %s: %v", tournament.ID, err)
	}

	winnerName := ""
	if team != nil {
		// Get the school name
		school, err := s.schoolService.GetSchoolByID(r.Context(), team.SchoolID)
		if err != nil {
			log.Printf("Error getting school for team %s: %v", team.ID, err)
		} else {
			winnerName = school.Name
		}
	}

	writeJSON(w, http.StatusOK, dtos.NewTournamentResponse(tournament, winnerName))
}
