package tournaments

import (
	"encoding/json"
	"net/http"

	apptournament "github.com/andrewcopp/Calcutta/backend/internal/app/tournament"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleUpdateKenPomStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	var req dtos.UpdateKenPomStatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	inputs := make([]apptournament.KenPomUpdateInput, 0, len(req.Stats))
	for _, s := range req.Stats {
		inputs = append(inputs, apptournament.KenPomUpdateInput{
			TeamID: s.TeamID,
			NetRtg: s.NetRtg,
			ORtg:   s.ORtg,
			DRtg:   s.DRtg,
			AdjT:   s.AdjT,
		})
	}

	if err := h.app.Tournament.UpdateKenPomStats(r.Context(), tournamentID, inputs); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
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
