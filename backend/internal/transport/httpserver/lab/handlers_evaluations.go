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

// HandleListEvaluations handles GET /api/lab/evaluations
func (h *Handler) HandleListEvaluations(w http.ResponseWriter, r *http.Request) {
	if h.app == nil || h.app.Lab == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	entryID := strings.TrimSpace(r.URL.Query().Get("entry_id"))
	investmentModelID := strings.TrimSpace(r.URL.Query().Get("investment_model_id"))
	calcuttaID := strings.TrimSpace(r.URL.Query().Get("calcutta_id"))

	filter := models.LabListEvaluationsFilter{}
	if entryID != "" {
		filter.EntryID = &entryID
	}
	if investmentModelID != "" {
		filter.InvestmentModelID = &investmentModelID
	}
	if calcuttaID != "" {
		filter.CalcuttaID = &calcuttaID
	}

	page := models.LabPagination{
		Limit:  httputil.GetQueryInt(r, "limit", 50),
		Offset: httputil.GetQueryInt(r, "offset", 0),
	}

	items, err := h.app.Lab.ListEvaluations(r.Context(), filter, page)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, listEvaluationsResponse{Items: items})
}

// HandleGetEvaluation handles GET /api/lab/evaluations/:id
func (h *Handler) HandleGetEvaluation(w http.ResponseWriter, r *http.Request) {
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

	evaluation, err := h.app.Lab.GetEvaluation(r.Context(), id)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, evaluation)
}

// HandleGetEvaluationEntryResults handles GET /api/lab/evaluations/:id/entries
func (h *Handler) HandleGetEvaluationEntryResults(w http.ResponseWriter, r *http.Request) {
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

	results, err := h.app.Lab.GetEvaluationEntryResults(r.Context(), id)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, evaluationEntryResultsResponse{Items: results})
}

// HandleGetEvaluationEntryProfile handles GET /api/lab/entry-results/:id
func (h *Handler) HandleGetEvaluationEntryProfile(w http.ResponseWriter, r *http.Request) {
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

	profile, err := h.app.Lab.GetEvaluationEntryProfile(r.Context(), id)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, profile)
}

// HandleGetEvaluationSummary handles GET /api/lab/evaluations/:id/summary
func (h *Handler) HandleGetEvaluationSummary(w http.ResponseWriter, r *http.Request) {
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

	summary, err := h.app.Lab.GetEvaluationSummary(r.Context(), id)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, summary)
}

// Response types

type listEvaluationsResponse struct {
	Items []models.LabEvaluationDetail `json:"items"`
}

type evaluationEntryResultsResponse struct {
	Items []models.LabEvaluationEntryResult `json:"items"`
}
