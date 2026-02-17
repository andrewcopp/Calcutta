package tournaments

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	app        *app.App
	authUserID func(context.Context) string
}

func NewHandlerWithAuthUserID(a *app.App, authUserID func(context.Context) string) *Handler {
	return &Handler{app: a, authUserID: authUserID}
}

func (h *Handler) HandleListTournaments(w http.ResponseWriter, r *http.Request) {
	tournaments, err := h.app.Tournament.List(r.Context())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	resp := make([]*dtos.TournamentResponse, 0, len(tournaments))
	for _, tournament := range tournaments {
		tournament := tournament
		team, err := h.app.Tournament.GetWinningTeam(r.Context(), tournament.ID)
		if err != nil {
			log.Printf("Error getting winning team for tournament %s: %v", tournament.ID, err)
			continue
		}

		winnerName := ""
		if team != nil {
			school, err := h.app.School.GetByID(r.Context(), team.SchoolID)
			if err != nil {
				log.Printf("Error getting school for team %s: %v", team.ID, err)
				continue
			}
			if school != nil {
				winnerName = school.Name
			}
		}

		log.Printf("Processing tournament: ID=%s, Name=%s", tournament.ID, tournament.Name)
		resp = append(resp, dtos.NewTournamentResponse(&tournament, winnerName))
	}
	response.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleGetTournament(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	tournament, err := h.app.Tournament.GetByID(r.Context(), tournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if tournament == nil {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "Tournament not found", "id")
		return
	}

	team, err := h.app.Tournament.GetWinningTeam(r.Context(), tournament.ID)
	if err != nil {
		log.Printf("Error getting winning team for tournament %s: %v", tournament.ID, err)
	}

	winnerName := ""
	if team != nil {
		school, err := h.app.School.GetByID(r.Context(), team.SchoolID)
		if err != nil {
			log.Printf("Error getting school for team %s: %v", team.ID, err)
		} else if school != nil {
			winnerName = school.Name
		}
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewTournamentResponse(tournament, winnerName))
}

func (h *Handler) HandleCreateTournament(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateTournamentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	tournament, err := h.app.Tournament.Create(r.Context(), req.Name, req.Rounds)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	response.WriteJSON(w, http.StatusCreated, dtos.NewTournamentResponse(tournament, ""))
}

func (h *Handler) HandleUpdateTournament(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	var req dtos.UpdateTournamentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	if err := h.app.Tournament.UpdateStartingAt(r.Context(), tournamentID, req.StartingAt.Value); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	tournament, err := h.app.Tournament.GetByID(r.Context(), tournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if tournament == nil {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "Tournament not found", "id")
		return
	}

	team, err := h.app.Tournament.GetWinningTeam(r.Context(), tournament.ID)
	if err != nil {
		log.Printf("Error getting winning team for tournament %s: %v", tournament.ID, err)
	}

	winnerName := ""
	if team != nil {
		school, err := h.app.School.GetByID(r.Context(), team.SchoolID)
		if err != nil {
			log.Printf("Error getting school for team %s: %v", team.ID, err)
		} else if school != nil {
			winnerName = school.Name
		}
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewTournamentResponse(tournament, winnerName))
}

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
		Eliminated:   false,
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
	if req.Eliminated != nil {
		team.Eliminated = *req.Eliminated
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
		log.Printf("Error getting school for team %s: %v", responseTeam.ID, err)
		response.WriteJSON(w, http.StatusOK, dtos.NewTournamentTeamResponse(responseTeam, nil))
		return
	}
	response.WriteJSON(w, http.StatusOK, dtos.NewTournamentTeamResponse(responseTeam, school))
}
