package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (s *Server) createTournamentTeamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		http.Error(w, "Tournament ID is required", http.StatusBadRequest)
		return
	}

	var request struct {
		SchoolID string `json:"schoolId"`
		Seed     int    `json:"seed"`
		Region   string `json:"region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.SchoolID == "" {
		http.Error(w, "School ID is required", http.StatusBadRequest)
		return
	}
	if request.Seed < 1 || request.Seed > 16 {
		http.Error(w, "Seed must be between 1 and 16", http.StatusBadRequest)
		return
	}
	if request.Region == "" {
		request.Region = "Unknown" // Default to "Unknown" if not provided
	}

	team := &models.TournamentTeam{
		ID:           uuid.New().String(),
		TournamentID: tournamentID,
		SchoolID:     request.SchoolID,
		Seed:         request.Seed,
		Region:       request.Region,
		Byes:         0,
		Wins:         0,
		Eliminated:   false,
	}

	if err := s.tournamentService.CreateTeam(r.Context(), team); err != nil {
		log.Printf("Error creating team: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

func (s *Server) updateTeamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	teamID := vars["id"]
	if teamID == "" {
		http.Error(w, "Team ID is required", http.StatusBadRequest)
		return
	}

	var request struct {
		Wins       *int  `json:"wins,omitempty"`
		Byes       *int  `json:"byes,omitempty"`
		Eliminated *bool `json:"eliminated,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	team, err := s.calcuttaService.GetTournamentTeam(r.Context(), teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if request.Wins != nil {
		team.Wins = *request.Wins
	}
	if request.Byes != nil {
		team.Byes = *request.Byes
	}
	if request.Eliminated != nil {
		team.Eliminated = *request.Eliminated
	}

	if err := team.ValidateDefault(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.tournamentService.UpdateTournamentTeam(r.Context(), team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedTeam, err := s.tournamentService.GetTeams(r.Context(), team.TournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var responseTeam *models.TournamentTeam
	for _, t := range updatedTeam {
		if t.ID == team.ID {
			responseTeam = t
			break
		}
	}

	if responseTeam == nil {
		http.Error(w, "Failed to retrieve updated team", http.StatusInternalServerError)
		return
	}

	school, err := s.schoolService.GetSchoolByID(r.Context(), responseTeam.SchoolID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":         responseTeam.ID,
		"schoolId":   responseTeam.SchoolID,
		"seed":       responseTeam.Seed,
		"byes":       responseTeam.Byes,
		"wins":       responseTeam.Wins,
		"eliminated": responseTeam.Eliminated,
		"created":    responseTeam.Created.Format("2006-01-02T15:04:05Z07:00"),
		"updated":    responseTeam.Updated.Format("2006-01-02T15:04:05Z07:00"),
		"school": map[string]interface{}{
			"id":   school.ID,
			"name": school.Name,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
