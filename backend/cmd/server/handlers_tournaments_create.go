package main

import (
	"encoding/json"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
)

func (s *Server) createTournamentHandler(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateTournamentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	// Create tournament
	tournament, err := s.tournamentService.CreateTournament(r.Context(), req.Name, req.Rounds)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, dtos.NewTournamentResponse(tournament, ""))
}
