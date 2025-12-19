package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type tournamentResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Rounds  int    `json:"rounds"`
	Winner  string `json:"winner,omitempty"`
	Created string `json:"created"`
}

func (s *Server) tournamentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tournaments, err := s.tournamentService.GetAllTournaments(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]tournamentResponse, 0)
	for _, tournament := range tournaments {
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

		response = append(response, tournamentResponse{
			ID:      tournament.ID,
			Name:    tournament.Name,
			Rounds:  tournament.Rounds,
			Winner:  winnerName,
			Created: tournament.Created.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	json.NewEncoder(w).Encode(response)
}
