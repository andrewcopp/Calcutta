package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
	"github.com/gorilla/mux"
)

func (s *Server) calcuttaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	log.Printf("Handling GET request to /api/calcuttas/{id}")

	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		log.Printf("Error: Calcutta ID is empty")
		http.Error(w, "Calcutta ID is required", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching calcutta with ID: %s", calcuttaID)
	calcutta, err := s.calcuttaService.GetCalcuttaByID(r.Context(), calcuttaID)
	if err != nil {
		if err.Error() == "calcutta not found" {
			log.Printf("Calcutta not found with ID: %s", calcuttaID)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		log.Printf("Error fetching calcutta: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved calcutta: %s", calcutta.ID)
	json.NewEncoder(w).Encode(dtos.NewCalcuttaResponse(calcutta))
}
