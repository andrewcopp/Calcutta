package tournaments

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleGetPredictions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	batchID, found, err := h.app.Prediction.GetLatestBatchID(r.Context(), tournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !found {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "No predictions found for this tournament", "")
		return
	}

	teamValues, err := h.app.Prediction.GetTeamValues(r.Context(), batchID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	teams, err := h.app.Tournament.GetTeams(r.Context(), tournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewTournamentPredictionsResponse(tournamentID, batchID, teamValues, teams))
}
