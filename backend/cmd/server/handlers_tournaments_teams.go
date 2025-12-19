package main

import (
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
	"github.com/gorilla/mux"
)

func (s *Server) tournamentTeamsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	teams, err := s.tournamentService.GetTeams(r.Context(), tournamentID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	response := make([]*dtos.TournamentTeamResponse, 0, len(teams))
	for _, team := range teams {
		school, err := s.schoolService.GetSchoolByID(r.Context(), team.SchoolID)
		if err != nil {
			log.Printf("Error getting school for team %s: %v", team.ID, err)
			response = append(response, dtos.NewTournamentTeamResponse(team, nil))
			continue
		}
		s := school
		response = append(response, dtos.NewTournamentTeamResponse(team, &s))
	}
	writeJSON(w, http.StatusOK, response)
}
