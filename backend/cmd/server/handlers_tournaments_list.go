package main

import (
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
)

func (s *Server) tournamentsHandler(w http.ResponseWriter, r *http.Request) {
	tournaments, err := s.tournamentService.GetAllTournaments(r.Context())
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	response := make([]*dtos.TournamentResponse, 0, len(tournaments))
	for _, tournament := range tournaments {
		tournament := tournament
		// Get the winning team for this tournament
		team, err := s.tournamentService.GetWinningTeam(r.Context(), tournament.ID)
		if err != nil {
			log.Printf("Error getting winning team for tournament %s: %v", tournament.ID, err)
			continue
		}

		winnerName := ""
		if team != nil {
			// Get the school name
			school, err := s.schoolService.GetSchoolByID(r.Context(), team.SchoolID)
			if err != nil {
				log.Printf("Error getting school for team %s: %v", team.ID, err)
				continue
			}
			winnerName = school.Name
		}

		// Log tournament data
		log.Printf("Processing tournament: ID=%s, Name=%s", tournament.ID, tournament.Name)

		response = append(response, dtos.NewTournamentResponse(&tournament, winnerName))
	}
	writeJSON(w, http.StatusOK, response)
}
