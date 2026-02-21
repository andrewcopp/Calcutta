package lab

import (
	"encoding/json"
	"net/http"
	"strings"

	applab "github.com/andrewcopp/Calcutta/backend/internal/app/lab"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

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
	var req models.LabStartPipelineRequest
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
