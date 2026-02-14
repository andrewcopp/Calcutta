package lab

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
	applab "github.com/andrewcopp/Calcutta/backend/internal/app/lab"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Handler handles lab-related HTTP requests.
type Handler struct {
	app        *app.App
	authUserID func(context.Context) string
}

// NewHandler creates a new lab handler.
func NewHandler(a *app.App) *Handler {
	return &Handler{app: a}
}

// NewHandlerWithAuthUserID creates a new lab handler with auth user ID function.
func NewHandlerWithAuthUserID(a *app.App, authUserID func(context.Context) string) *Handler {
	return &Handler{app: a, authUserID: authUserID}
}

func getQueryInt(r *http.Request, name string, defaultValue int) int {
	v := strings.TrimSpace(r.URL.Query().Get(name))
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// HandleListModels handles GET /api/lab/models
func (h *Handler) HandleListModels(w http.ResponseWriter, r *http.Request) {
	if h.app == nil || h.app.Lab == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	kind := strings.TrimSpace(r.URL.Query().Get("kind"))

	filter := applab.ListModelsFilter{}
	if kind != "" {
		filter.Kind = &kind
	}

	page := applab.Pagination{
		Limit:  getQueryInt(r, "limit", 50),
		Offset: getQueryInt(r, "offset", 0),
	}

	items, err := h.app.Lab.ListInvestmentModels(r.Context(), filter, page)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, listModelsResponse{Items: items})
}

// HandleGetModel handles GET /api/lab/models/:id
func (h *Handler) HandleGetModel(w http.ResponseWriter, r *http.Request) {
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

	model, err := h.app.Lab.GetInvestmentModel(r.Context(), id)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, model)
}

// HandleGetLeaderboard handles GET /api/lab/models/leaderboard
func (h *Handler) HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	if h.app == nil || h.app.Lab == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	items, err := h.app.Lab.GetModelLeaderboard(r.Context())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, leaderboardResponse{Items: items})
}

// HandleListEntries handles GET /api/lab/entries
func (h *Handler) HandleListEntries(w http.ResponseWriter, r *http.Request) {
	if h.app == nil || h.app.Lab == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	investmentModelID := strings.TrimSpace(r.URL.Query().Get("investment_model_id"))
	calcuttaID := strings.TrimSpace(r.URL.Query().Get("calcutta_id"))
	startingStateKey := strings.TrimSpace(r.URL.Query().Get("starting_state_key"))

	filter := applab.ListEntriesFilter{}
	if investmentModelID != "" {
		filter.InvestmentModelID = &investmentModelID
	}
	if calcuttaID != "" {
		filter.CalcuttaID = &calcuttaID
	}
	if startingStateKey != "" {
		filter.StartingStateKey = &startingStateKey
	}

	page := applab.Pagination{
		Limit:  getQueryInt(r, "limit", 50),
		Offset: getQueryInt(r, "offset", 0),
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

// HandleListEvaluations handles GET /api/lab/evaluations
func (h *Handler) HandleListEvaluations(w http.ResponseWriter, r *http.Request) {
	if h.app == nil || h.app.Lab == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	entryID := strings.TrimSpace(r.URL.Query().Get("entry_id"))
	investmentModelID := strings.TrimSpace(r.URL.Query().Get("investment_model_id"))
	calcuttaID := strings.TrimSpace(r.URL.Query().Get("calcutta_id"))

	filter := applab.ListEvaluationsFilter{}
	if entryID != "" {
		filter.EntryID = &entryID
	}
	if investmentModelID != "" {
		filter.InvestmentModelID = &investmentModelID
	}
	if calcuttaID != "" {
		filter.CalcuttaID = &calcuttaID
	}

	page := applab.Pagination{
		Limit:  getQueryInt(r, "limit", 50),
		Offset: getQueryInt(r, "offset", 0),
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

// HandleGenerateEntries handles POST /api/lab/models/:id/generate-entries
func (h *Handler) HandleGenerateEntries(w http.ResponseWriter, r *http.Request) {
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

	// Verify model exists
	_, err := h.app.Lab.GetInvestmentModel(r.Context(), id)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	// Parse optional request body
	var req applab.GenerateEntriesRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httperr.Write(w, r, http.StatusBadRequest, "validation_error", "invalid request body", "")
			return
		}
	}

	result, err := h.app.Lab.GenerateEntries(r.Context(), id, req)
	if err != nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", err.Error(), "")
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}

// HandleStartPipeline handles POST /api/lab/models/:id/pipeline/start
func (h *Handler) HandleStartPipeline(w http.ResponseWriter, r *http.Request) {
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

	// Verify model exists
	_, err := h.app.Lab.GetInvestmentModel(r.Context(), id)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	// Parse optional request body
	var req applab.StartPipelineRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httperr.Write(w, r, http.StatusBadRequest, "validation_error", "invalid request body", "")
			return
		}
	}

	result, err := h.app.Lab.StartPipeline(r.Context(), id, req)
	if err != nil {
		switch err.(type) {
		case *applab.PipelineAlreadyRunningError:
			httperr.Write(w, r, http.StatusConflict, "pipeline_running", err.Error(), "")
		case *applab.NoCalcuttasAvailableError:
			httperr.Write(w, r, http.StatusBadRequest, "no_calcuttas", err.Error(), "")
		case *applab.PipelineNotAvailableError:
			httperr.Write(w, r, http.StatusServiceUnavailable, "pipeline_unavailable", err.Error(), "")
		default:
			httperr.WriteFromErr(w, r, err, h.authUserID)
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}

// HandleGetModelPipelineProgress handles GET /api/lab/models/:id/pipeline/progress
func (h *Handler) HandleGetModelPipelineProgress(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.app.Lab.GetModelPipelineProgress(r.Context(), id)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}

// HandleGetPipelineRun handles GET /api/lab/pipeline-runs/:id
func (h *Handler) HandleGetPipelineRun(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.app.Lab.GetPipelineProgress(r.Context(), id)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}

// HandleCancelPipeline handles POST /api/lab/pipeline-runs/:id/cancel
func (h *Handler) HandleCancelPipeline(w http.ResponseWriter, r *http.Request) {
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

	err := h.app.Lab.CancelPipeline(r.Context(), id)
	if err != nil {
		switch err.(type) {
		case *applab.PipelineNotCancellableError:
			httperr.Write(w, r, http.StatusConflict, "not_cancellable", err.Error(), "")
		default:
			httperr.WriteFromErr(w, r, err, h.authUserID)
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// Response types

type listModelsResponse struct {
	Items []applab.InvestmentModel `json:"items"`
}

type leaderboardResponse struct {
	Items []applab.LeaderboardEntry `json:"items"`
}

type listEntriesResponse struct {
	Items []applab.EntryDetail `json:"items"`
}

type listEvaluationsResponse struct {
	Items []applab.EvaluationDetail `json:"items"`
}

type evaluationEntryResultsResponse struct {
	Items []applab.EvaluationEntryResult `json:"items"`
}
