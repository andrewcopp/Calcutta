package main

import (
	"encoding/json"
	"net/http"
)

func (s *Server) createTournamentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var request struct {
		Name   string `json:"name"`
		Rounds int    `json:"rounds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if request.Rounds <= 0 {
		http.Error(w, "Rounds must be greater than 0", http.StatusBadRequest)
		return
	}

	// Create tournament
	tournament, err := s.tournamentService.CreateTournament(r.Context(), request.Name, request.Rounds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tournament)
}
