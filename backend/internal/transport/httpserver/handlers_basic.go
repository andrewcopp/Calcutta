package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
)

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "healthy", "message": "API is running"})
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		response.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "message": "database pool not initialized"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.pool.Ping(ctx); err != nil {
		response.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "message": "database not reachable"})
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) schoolsHandler(w http.ResponseWriter, r *http.Request) {
	schools, err := s.app.School.List(r.Context())
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": dtos.NewSchoolListResponse(schools)})
}
