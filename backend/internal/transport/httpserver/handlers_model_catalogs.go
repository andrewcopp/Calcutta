package httpserver

import (
	modelcatalogs "github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/modelcatalogs"
	"github.com/gorilla/mux"
)

func (s *Server) registerModelCatalogRoutes(r *mux.Router) {
	modelcatalogs.RegisterRoutes(r, s.requirePermission, s.app.ModelCatalogs)
}
