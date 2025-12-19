package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
)

func (s *Server) calcuttasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		s.createCalcuttaHandler(w, r)
		return
	}

	log.Printf("Handling GET request to /api/calcuttas")
	calcuttas, err := s.calcuttaService.GetAllCalcuttas(r.Context())
	if err != nil {
		log.Printf("Error getting all calcuttas: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}
	log.Printf("Successfully retrieved %d calcuttas", len(calcuttas))

	response := dtos.NewCalcuttaListResponse(calcuttas)
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) createCalcuttaHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling POST request to /api/calcuttas")

	var req dtos.CreateCalcuttaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if err := req.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	calcutta := req.ToModel()
	if err := s.calcuttaService.CreateCalcuttaWithRounds(r.Context(), calcutta); err != nil {
		log.Printf("Error creating calcutta with rounds: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}
	log.Printf("Successfully created calcutta %s with rounds", calcutta.ID)
	writeJSON(w, http.StatusCreated, dtos.NewCalcuttaResponse(calcutta))
}
