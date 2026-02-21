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

// HandleListModels handles GET /api/lab/models
func (h *Handler) HandleListModels(w http.ResponseWriter, r *http.Request) {
	if h.app == nil || h.app.Lab == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	kind := strings.TrimSpace(r.URL.Query().Get("kind"))

	filter := models.LabListModelsFilter{}
	if kind != "" {
		filter.Kind = &kind
	}

	page := models.LabPagination{
		Limit:  httputil.GetQueryInt(r, "limit", 50),
		Offset: httputil.GetQueryInt(r, "offset", 0),
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

// Response types

type listModelsResponse struct {
	Items []models.InvestmentModel `json:"items"`
}

type leaderboardResponse struct {
	Items []models.LabLeaderboardEntry `json:"items"`
}
