package httpserver

import (
 	"context"
	"net/http"
 	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
)

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy", "message": "API is running"})
}

 func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
 	if s.pool == nil {
 		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "message": "database pool not initialized"})
 		return
 	}

 	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
 	defer cancel()
 	if err := s.pool.Ping(ctx); err != nil {
 		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "message": "database not reachable"})
 		return
 	}

 	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
 }

func (s *Server) schoolsHandler(w http.ResponseWriter, r *http.Request) {
	schools, err := s.app.School.List(r.Context())
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewSchoolListResponse(schools))
}
