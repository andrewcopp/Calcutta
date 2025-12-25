package httpserver

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
)

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy", "message": "API is running"})
}

func (s *Server) schoolsHandler(w http.ResponseWriter, r *http.Request) {
	schools, err := s.app.School.List(r.Context())
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewSchoolListResponse(schools))
}
