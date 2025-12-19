package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) tournamentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract tournament ID from URL path
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		http.Error(w, "Tournament ID is required", http.StatusBadRequest)
		return
	}

	// Get tournament by ID
	tournament, err := s.tournamentService.GetTournamentByID(r.Context(), tournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tournament == nil {
		http.Error(w, "Tournament not found", http.StatusNotFound)
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

	response := tournamentResponse{
		ID:      tournament.ID,
		Name:    tournament.Name,
		Rounds:  tournament.Rounds,
		Winner:  winnerName,
		Created: tournament.Created.Format("2006-01-02T15:04:05Z07:00"),
	}

	json.NewEncoder(w).Encode(response)
}
