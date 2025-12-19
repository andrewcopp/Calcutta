package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) tournamentTeamsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		http.Error(w, "Tournament ID is required", http.StatusBadRequest)
		return
	}

	teams, err := s.tournamentService.GetTeams(r.Context(), tournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]map[string]interface{}, 0)
	for _, team := range teams {
		school, err := s.schoolService.GetSchoolByID(r.Context(), team.SchoolID)
		if err != nil {
			log.Printf("Error getting school for team %s: %v", team.ID, err)
			continue
		}

		teamResponse := map[string]interface{}{
			"id":         team.ID,
			"schoolId":   team.SchoolID,
			"seed":       team.Seed,
			"byes":       team.Byes,
			"wins":       team.Wins,
			"eliminated": team.Eliminated,
			"created":    team.Created.Format("2006-01-02T15:04:05Z07:00"),
			"updated":    team.Updated.Format("2006-01-02T15:04:05Z07:00"),
			"school": map[string]interface{}{
				"id":   school.ID,
				"name": school.Name,
			},
		}

		response = append(response, teamResponse)
	}

	json.NewEncoder(w).Encode(response)
}
