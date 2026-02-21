package lab

import (
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httputil"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleListEntries handles GET /api/lab/entries
func (h *Handler) HandleListEntries(w http.ResponseWriter, r *http.Request) {
	if h.app == nil || h.app.Lab == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	investmentModelID := strings.TrimSpace(r.URL.Query().Get("investment_model_id"))
	calcuttaID := strings.TrimSpace(r.URL.Query().Get("calcutta_id"))
	startingStateKey := strings.TrimSpace(r.URL.Query().Get("starting_state_key"))

	filter := models.LabListEntriesFilter{}
	if investmentModelID != "" {
		filter.InvestmentModelID = &investmentModelID
	}
	if calcuttaID != "" {
		filter.CalcuttaID = &calcuttaID
	}
	if startingStateKey != "" {
		filter.StartingStateKey = &startingStateKey
	}

	page := models.LabPagination{
		Limit:  httputil.GetQueryInt(r, "limit", 50),
		Offset: httputil.GetQueryInt(r, "offset", 0),
	}

	items, err := h.app.Lab.ListEntries(r.Context(), filter, page)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, listEntriesResponse{Items: items})
}

// HandleGetEntry handles GET /api/lab/entries/:id
// Returns enriched entry data with team names, seeds, regions, and naive allocation.
func (h *Handler) HandleGetEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := strings.TrimSpace(vars["id"])
	if id == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}
	if h.app == nil || h.app.Lab == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	entry, err := h.app.Lab.GetEntryEnriched(r.Context(), id)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, entry)
}

// HandleGetEntryByModelAndCalcutta handles GET /api/lab/models/:modelName/calcutta/:calcuttaId/entry
// Returns enriched entry data for the model/calcutta pair.
func (h *Handler) HandleGetEntryByModelAndCalcutta(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelName := strings.TrimSpace(vars["modelName"])
	calcuttaID := strings.TrimSpace(vars["calcuttaId"])

	if modelName == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "modelName is required", "modelName")
		return
	}
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "calcuttaId is required", "calcuttaId")
		return
	}
	if _, err := uuid.Parse(calcuttaID); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "calcuttaId must be a valid UUID", "calcuttaId")
		return
	}
	if h.app == nil || h.app.Lab == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	startingStateKey := strings.TrimSpace(r.URL.Query().Get("starting_state_key"))

	entry, err := h.app.Lab.GetEntryEnrichedByModelAndCalcutta(r.Context(), modelName, calcuttaID, startingStateKey)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, entry)
}

// Response types

type listEntriesResponse struct {
	Items []models.LabEntryDetail `json:"items"`
}
