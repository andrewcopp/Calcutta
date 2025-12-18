package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func tournamentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tournaments, err := tournamentService.GetAllTournaments(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a response that includes tournament winners
	type TournamentResponse struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Rounds  int    `json:"rounds"`
		Winner  string `json:"winner,omitempty"`
		Created string `json:"created"`
	}

	response := make([]TournamentResponse, 0)
	for _, tournament := range tournaments {
		// Get the winning team for this tournament
		team, err := tournamentService.GetWinningTeam(r.Context(), tournament.ID)
		if err != nil {
			log.Printf("Error getting winning team for tournament %s: %v", tournament.ID, err)
			continue
		}

		winnerName := ""
		if team != nil {
			// Get the school name
			school, err := schoolService.GetSchoolByID(r.Context(), team.SchoolID)
			if err != nil {
				log.Printf("Error getting school for team %s: %v", team.ID, err)
				continue
			}
			winnerName = school.Name
		}

		// Log tournament data
		log.Printf("Processing tournament: ID=%s, Name=%s", tournament.ID, tournament.Name)

		response = append(response, TournamentResponse{
			ID:      tournament.ID,
			Name:    tournament.Name,
			Rounds:  tournament.Rounds,
			Winner:  winnerName,
			Created: tournament.Created.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	json.NewEncoder(w).Encode(response)
}

func tournamentTeamsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract tournament ID from URL path
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		http.Error(w, "Tournament ID is required", http.StatusBadRequest)
		return
	}

	// Get teams for the tournament
	teams, err := tournamentRepo.GetTeams(r.Context(), tournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get school information for each team
	response := make([]map[string]interface{}, 0)
	for _, team := range teams {
		// Get the school
		school, err := schoolService.GetSchoolByID(r.Context(), team.SchoolID)
		if err != nil {
			log.Printf("Error getting school for team %s: %v", team.ID, err)
			continue
		}

		// Create a response object with team and school information
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

func updateTeamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract team ID from URL path
	vars := mux.Vars(r)
	teamID := vars["id"]
	if teamID == "" {
		http.Error(w, "Team ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var request struct {
		Wins       *int  `json:"wins,omitempty"`
		Byes       *int  `json:"byes,omitempty"`
		Eliminated *bool `json:"eliminated,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the team
	team, err := calcuttaRepo.GetTournamentTeam(r.Context(), teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the team fields
	if request.Wins != nil {
		team.Wins = *request.Wins
	}
	if request.Byes != nil {
		team.Byes = *request.Byes
	}
	if request.Eliminated != nil {
		team.Eliminated = *request.Eliminated
	}

	// Validate the team
	if err := team.ValidateDefault(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the team in the database
	err = tournamentRepo.UpdateTournamentTeam(r.Context(), team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the updated team with school information
	updatedTeam, err := tournamentRepo.GetTeams(r.Context(), team.TournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Find the updated team in the list
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

	// Get the school information
	school, err := schoolService.GetSchoolByID(r.Context(), responseTeam.SchoolID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create response with team and school information
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

func recalculatePortfoliosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract tournament ID from URL path
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		http.Error(w, "Tournament ID is required", http.StatusBadRequest)
		return
	}

	// Get all calcuttas for this tournament
	calcuttas, err := calcuttaRepo.GetCalcuttasByTournament(r.Context(), tournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Recalculate portfolios for each calcutta
	for _, calcutta := range calcuttas {
		if err := calcuttaService.RecalculatePortfolio(r.Context(), calcutta.ID); err != nil {
			log.Printf("Error recalculating portfolio for calcutta %s: %v", calcutta.ID, err)
			continue
		}
	}

	w.WriteHeader(http.StatusOK)
}

func createTournamentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var request struct {
		Name   string `json:"name"`
		Rounds int    `json:"rounds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if request.Rounds <= 0 {
		http.Error(w, "Rounds must be greater than 0", http.StatusBadRequest)
		return
	}

	// Create tournament
	tournament, err := tournamentService.CreateTournament(r.Context(), request.Name, request.Rounds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created tournament
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tournament)
}

func createTournamentTeamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract tournament ID from URL path
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		http.Error(w, "Tournament ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var request struct {
		SchoolID string `json:"schoolId"`
		Seed     int    `json:"seed"`
		Region   string `json:"region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
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

	// Create the team
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

	if err := tournamentRepo.CreateTeam(r.Context(), team); err != nil {
		log.Printf("Error creating team: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created team
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

func tournamentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract tournament ID from URL path
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		http.Error(w, "Tournament ID is required", http.StatusBadRequest)
		return
	}

	// Get tournament by ID
	tournament, err := tournamentService.GetTournamentByID(r.Context(), tournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tournament == nil {
		http.Error(w, "Tournament not found", http.StatusNotFound)
		return
	}

	// Get the winning team for this tournament
	team, err := tournamentService.GetWinningTeam(r.Context(), tournament.ID)
	if err != nil {
		log.Printf("Error getting winning team for tournament %s: %v", tournament.ID, err)
	}

	// Create response with tournament and winning team info
	type TournamentResponse struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Rounds  int    `json:"rounds"`
		Winner  string `json:"winner,omitempty"`
		Created string `json:"created"`
	}

	winnerName := ""
	if team != nil {
		// Get the school name
		school, err := schoolService.GetSchoolByID(r.Context(), team.SchoolID)
		if err != nil {
			log.Printf("Error getting school for team %s: %v", team.ID, err)
		} else {
			winnerName = school.Name
		}
	}

	response := TournamentResponse{
		ID:      tournament.ID,
		Name:    tournament.Name,
		Rounds:  tournament.Rounds,
		Winner:  winnerName,
		Created: tournament.Created.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Return the tournament response
	json.NewEncoder(w).Encode(response)
}
