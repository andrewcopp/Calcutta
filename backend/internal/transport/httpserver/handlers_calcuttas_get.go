package httpserver

import (
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/gorilla/mux"
)

func (s *Server) calcuttaHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling GET request to /api/calcuttas/{id}")

	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		log.Printf("Error: Calcutta ID is empty")
		writeError(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "id")
		return
	}

	log.Printf("Fetching calcutta with ID: %s", calcuttaID)
	calcutta, err := s.app.Calcutta.GetCalcuttaByID(r.Context(), calcuttaID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	log.Printf("Successfully retrieved calcutta: %s", calcutta.ID)
	writeJSON(w, http.StatusOK, dtos.NewCalcuttaResponse(calcutta))
}
