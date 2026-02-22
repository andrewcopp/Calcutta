package tournaments

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	apptournament "github.com/andrewcopp/Calcutta/backend/internal/app/tournament"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListTournamentTeams(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	teams, err := h.app.Tournament.GetTeams(r.Context(), tournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	resp := make([]*dtos.TournamentTeamResponse, 0, len(teams))
	for _, team := range teams {
		resp = append(resp, dtos.NewTournamentTeamResponse(team, team.School))
	}
	response.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleCreateTournamentTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}
	var req dtos.CreateTournamentTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
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
		IsEliminated: false,
	}

	if err := h.app.Tournament.CreateTeam(r.Context(), team); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	response.WriteJSON(w, http.StatusCreated, dtos.NewTournamentTeamResponse(team, nil))
}

func (h *Handler) HandleUpdateTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["teamId"]
	if teamID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Team ID is required", "teamId")
		return
	}
	var req dtos.UpdateTournamentTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	team, err := h.app.Calcutta.GetTournamentTeam(r.Context(), teamID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if req.Wins != nil {
		team.Wins = *req.Wins
	}
	if req.Byes != nil {
		team.Byes = *req.Byes
	}
	if req.IsEliminated != nil {
		team.IsEliminated = *req.IsEliminated
	}

	if err := team.ValidateDefault(); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", err.Error(), "")
		return
	}

	err = h.app.Tournament.UpdateTournamentTeam(r.Context(), team)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	updatedTeam, err := h.app.Tournament.GetTeams(r.Context(), team.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
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
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "Failed to retrieve updated team", "")
		return
	}

	school, err := h.app.School.GetByID(r.Context(), responseTeam.SchoolID)
	if err != nil {
		slog.Error("get_school_failed", "team_id", responseTeam.ID, "error", err)
		response.WriteJSON(w, http.StatusOK, dtos.NewTournamentTeamResponse(responseTeam, nil))
		return
	}
	response.WriteJSON(w, http.StatusOK, dtos.NewTournamentTeamResponse(responseTeam, school))
}

func (h *Handler) HandleReplaceTeams(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	var req dtos.ReplaceTeamsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.WriteJSON(w, http.StatusBadRequest, dtos.BracketValidationErrorResponse{
			Code:   "validation_error",
			Errors: errs,
		})
		return
	}

	inputs := make([]apptournament.ReplaceTeamsInput, 0, len(req.Teams))
	for _, t := range req.Teams {
		inputs = append(inputs, apptournament.ReplaceTeamsInput{
			SchoolID: t.SchoolID,
			Seed:     t.Seed,
			Region:   t.Region,
		})
	}

	teams, err := h.app.Tournament.ReplaceTeams(r.Context(), tournamentID, inputs)
	if err != nil {
		var bve *apptournament.BracketValidationError
		if errors.As(err, &bve) {
			response.WriteJSON(w, http.StatusBadRequest, dtos.BracketValidationErrorResponse{
				Code:   "bracket_validation_error",
				Errors: bve.Errors,
			})
			return
		}
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	resp := make([]*dtos.TournamentTeamResponse, 0, len(teams))
	for _, team := range teams {
		resp = append(resp, dtos.NewTournamentTeamResponse(team, team.School))
	}
	response.WriteJSON(w, http.StatusOK, resp)
}
