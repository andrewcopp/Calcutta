package httpserver

import (
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
)

func (s *Server) tournamentsHandler(w http.ResponseWriter, r *http.Request) {
	tournaments, err := s.app.Tournament.List(r.Context())
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	response := make([]*dtos.TournamentResponse, 0, len(tournaments))
	for _, tournament := range tournaments {
		tournament := tournament
		// Get the winning team for this tournament
		team, err := s.app.Tournament.GetWinningTeam(r.Context(), tournament.ID)
		if err != nil {
			log.Printf("Error getting winning team for tournament %s: %v", tournament.ID, err)
			continue
		}

		winnerName := ""
		if team != nil {
			// Get the school name
			school, err := s.app.School.GetByID(r.Context(), team.SchoolID)
			if err != nil {
				log.Printf("Error getting school for team %s: %v", team.ID, err)
				continue
			}
			if school != nil {
				winnerName = school.Name
			}
		}

		// Log tournament data
		log.Printf("Processing tournament: ID=%s, Name=%s", tournament.ID, tournament.Name)

		response = append(response, dtos.NewTournamentResponse(&tournament, winnerName))
	}
	writeJSON(w, http.StatusOK, response)
}
