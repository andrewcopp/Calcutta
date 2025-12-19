package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (s *Server) createTournamentTeamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}
	var req dtos.CreateTournamentTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if req.Region == "" {
		req.Region = "Unknown"
	}

	team := &models.TournamentTeam{
		ID:           uuid.New().String(),
		TournamentID: tournamentID,
		SchoolID:     req.SchoolID,
		Seed:         req.Seed,
		Region:       req.Region,
		Byes:         0,
		Wins:         0,
		Eliminated:   false,
	}

	if err := s.tournamentService.CreateTeam(r.Context(), team); err != nil {
		log.Printf("Error creating team: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, dtos.NewTournamentTeamResponse(team, nil))
}

func (s *Server) updateTeamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["teamId"]
	if teamID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Team ID is required", "teamId")
		return
	}
	var req dtos.UpdateTournamentTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	team, err := s.calcuttaService.GetTournamentTeam(r.Context(), teamID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if req.Wins != nil {
		team.Wins = *req.Wins
	}
	if req.Byes != nil {
		team.Byes = *req.Byes
	}
	if req.Eliminated != nil {
		team.Eliminated = *req.Eliminated
	}

	if err := team.ValidateDefault(); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", err.Error(), "")
		return
	}

	err = s.tournamentService.UpdateTournamentTeam(r.Context(), team)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	updatedTeam, err := s.tournamentService.GetTeams(r.Context(), team.TournamentID)
	if err != nil {
		writeErrorFromErr(w, r, err)
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
		writeError(w, r, http.StatusInternalServerError, "internal_error", "Failed to retrieve updated team", "")
		return
	}

	school, err := s.schoolService.GetSchoolByID(r.Context(), responseTeam.SchoolID)
	if err != nil {
		log.Printf("Error getting school for team %s: %v", responseTeam.ID, err)
		writeJSON(w, http.StatusOK, dtos.NewTournamentTeamResponse(responseTeam, nil))
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewTournamentTeamResponse(responseTeam, &school))
}
