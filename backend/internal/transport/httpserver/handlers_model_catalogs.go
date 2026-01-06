package httpserver

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/features/model_catalogs"
	"github.com/gorilla/mux"
)

type listModelCatalogResponse struct {
	Items []model_catalogs.ModelDescriptor `json:"items"`
	Count int                              `json:"count"`
}

func (s *Server) registerModelCatalogRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/advancement-models",
		s.requirePermission("analytics.suites.read", s.handleListAdvancementModels),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/market-share-models",
		s.requirePermission("analytics.suites.read", s.handleListMarketShareModels),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/entry-optimizers",
		s.requirePermission("analytics.suites.read", s.handleListEntryOptimizers),
	).Methods("GET", "OPTIONS")
}

func (s *Server) handleListAdvancementModels(w http.ResponseWriter, r *http.Request) {
	items := s.app.ModelCatalogs.ListAdvancementModels()
	writeJSON(w, http.StatusOK, listModelCatalogResponse{Items: items, Count: len(items)})
}

func (s *Server) handleListMarketShareModels(w http.ResponseWriter, r *http.Request) {
	items := s.app.ModelCatalogs.ListMarketShareModels()
	writeJSON(w, http.StatusOK, listModelCatalogResponse{Items: items, Count: len(items)})
}

func (s *Server) handleListEntryOptimizers(w http.ResponseWriter, r *http.Request) {
	items := s.app.ModelCatalogs.ListEntryOptimizers()
	writeJSON(w, http.StatusOK, listModelCatalogResponse{Items: items, Count: len(items)})
}
