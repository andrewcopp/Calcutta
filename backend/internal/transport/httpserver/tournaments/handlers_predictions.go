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
	tournamentID := vars["tournamentId"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "tournamentId")
		return
	}

	batchID := r.URL.Query().Get("batchId")
	var throughRound int
	if batchID == "" {
		batch, found, err := h.app.Prediction.GetLatestBatch(r.Context(), tournamentID)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
		if !found {
			httperr.Write(w, r, http.StatusNotFound, "not_found", "No predictions found for this tournament", "")
			return
		}
		batchID = batch.ID
		throughRound = batch.ThroughRound
	} else {
		batch, err := h.app.Prediction.GetBatchSummary(r.Context(), batchID)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
		throughRound = batch.ThroughRound
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

	response.WriteJSON(w, http.StatusOK, dtos.NewTournamentPredictionsResponse(tournamentID, batchID, throughRound, teamValues, teams))
}

func (h *Handler) HandleListPredictionBatches(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["tournamentId"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "tournamentId")
		return
	}

	batches, err := h.app.Prediction.ListBatches(r.Context(), tournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{"items": dtos.NewPredictionBatchListResponse(batches)})
}
