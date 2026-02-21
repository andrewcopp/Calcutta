package tournaments

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

// Team handlers are in handlers_teams.go

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

	winningTeams, err := h.app.Tournament.ListWinningTeams(r.Context())
	if err != nil {
		slog.Error("list_winning_teams_failed", "error", err)
		winningTeams = make(map[string]*models.TournamentTeam)
	}

	schools, err := h.app.School.List(r.Context())
	if err != nil {
		slog.Error("list_schools_failed", "error", err)
	}
	schoolByID := make(map[string]string, len(schools))
	for _, s := range schools {
		schoolByID[s.ID] = s.Name
	}

	resp := make([]*dtos.TournamentResponse, 0, len(tournaments))
	for _, tournament := range tournaments {
		tournament := tournament
		winnerName := ""

		if team, ok := winningTeams[tournament.ID]; ok && team != nil {
			winnerName = schoolByID[team.SchoolID]
		}

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
		slog.Error("get_winning_team_failed", "tournament_id", tournament.ID, "error", err)
	}

	winnerName := ""
	if team != nil {
		school, err := h.app.School.GetByID(r.Context(), team.SchoolID)
		if err != nil {
			slog.Error("get_school_failed", "team_id", team.ID, "error", err)
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

	tournament, err := h.app.Tournament.Create(r.Context(), req.Competition, req.Year, req.Rounds)
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

	if req.StartingAt.Present {
		if err := h.app.Tournament.UpdateStartingAt(r.Context(), tournamentID, req.StartingAt.Value); err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
	}

	if req.FinalFourTopLeft != nil || req.FinalFourBottomLeft != nil || req.FinalFourTopRight != nil || req.FinalFourBottomRight != nil {
		// Fetch current tournament to fill in any unspecified fields
		current, err := h.app.Tournament.GetByID(r.Context(), tournamentID)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
		if current == nil {
			httperr.Write(w, r, http.StatusNotFound, "not_found", "Tournament not found", "id")
			return
		}
		tl := current.FinalFourTopLeft
		bl := current.FinalFourBottomLeft
		tr := current.FinalFourTopRight
		br := current.FinalFourBottomRight
		if req.FinalFourTopLeft != nil {
			tl = *req.FinalFourTopLeft
		}
		if req.FinalFourBottomLeft != nil {
			bl = *req.FinalFourBottomLeft
		}
		if req.FinalFourTopRight != nil {
			tr = *req.FinalFourTopRight
		}
		if req.FinalFourBottomRight != nil {
			br = *req.FinalFourBottomRight
		}
		if err := h.app.Tournament.UpdateFinalFour(r.Context(), tournamentID, tl, bl, tr, br); err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
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
		slog.Error("get_winning_team_failed", "tournament_id", tournament.ID, "error", err)
	}

	winnerName := ""
	if team != nil {
		school, err := h.app.School.GetByID(r.Context(), team.SchoolID)
		if err != nil {
			slog.Error("get_school_failed", "team_id", team.ID, "error", err)
		} else if school != nil {
			winnerName = school.Name
		}
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewTournamentResponse(tournament, winnerName))
}

func (h *Handler) HandleListCompetitions(w http.ResponseWriter, r *http.Request) {
	competitions, err := h.app.Tournament.ListCompetitions(r.Context())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	resp := make([]dtos.CompetitionResponse, 0, len(competitions))
	for _, c := range competitions {
		resp = append(resp, dtos.CompetitionResponse{ID: c.ID, Name: c.Name})
	}
	response.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleListSeasons(w http.ResponseWriter, r *http.Request) {
	seasons, err := h.app.Tournament.ListSeasons(r.Context())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	resp := make([]dtos.SeasonResponse, 0, len(seasons))
	for _, s := range seasons {
		resp = append(resp, dtos.SeasonResponse{ID: s.ID, Year: s.Year})
	}
	response.WriteJSON(w, http.StatusOK, resp)
}

