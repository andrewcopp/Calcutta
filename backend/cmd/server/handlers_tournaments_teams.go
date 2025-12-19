package main

import (
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
		response = append(response, dtos.NewTournamentTeamResponse(team, team.School))
	}
	writeJSON(w, http.StatusOK, response)
}
