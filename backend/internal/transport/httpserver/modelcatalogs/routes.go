package modelcatalogs

import (
	"net/http"

	appmodelcatalogs "github.com/andrewcopp/Calcutta/backend/internal/app/model_catalogs"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

type handler struct {
	svc *appmodelcatalogs.Service
}

type listModelCatalogResponse struct {
	Items []appmodelcatalogs.ModelDescriptor `json:"items"`
	Count int                                `json:"count"`
}

func RegisterRoutes(r *mux.Router, requirePermission func(string, http.HandlerFunc) http.HandlerFunc, svc *appmodelcatalogs.Service) {
	h := &handler{svc: svc}

	r.HandleFunc(
		"/api/advancement-models",
		requirePermission("analytics.suites.read", h.handleListAdvancementModels),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/market-share-models",
		requirePermission("analytics.suites.read", h.handleListMarketShareModels),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/entry-optimizers",
		requirePermission("analytics.suites.read", h.handleListEntryOptimizers),
	).Methods("GET", "OPTIONS")
}

func (h *handler) handleListAdvancementModels(w http.ResponseWriter, r *http.Request) {
	items := h.svc.ListAdvancementModels()
	response.WriteJSON(w, http.StatusOK, listModelCatalogResponse{Items: items, Count: len(items)})
}

func (h *handler) handleListMarketShareModels(w http.ResponseWriter, r *http.Request) {
	items := h.svc.ListMarketShareModels()
	response.WriteJSON(w, http.StatusOK, listModelCatalogResponse{Items: items, Count: len(items)})
}

func (h *handler) handleListEntryOptimizers(w http.ResponseWriter, r *http.Request) {
	items := h.svc.ListEntryOptimizers()
	response.WriteJSON(w, http.StatusOK, listModelCatalogResponse{Items: items, Count: len(items)})
}
