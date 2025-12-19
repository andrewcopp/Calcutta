package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
	"github.com/gorilla/mux"
)

func (s *Server) getBracketHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("PANIC in getBracketHandler: %v", rec)
			writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error building bracket", "")
		}
	}()

	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	log.Printf("Building bracket for tournament: %s", tournamentID)
	bracket, err := s.bracketService.GetBracket(r.Context(), tournamentID)
	if err != nil {
		log.Printf("Error getting bracket for tournament %s: %v", tournamentID, err)
		writeErrorFromErr(w, r, err)
		return
	}

	log.Printf("Successfully built bracket with %d games", len(bracket.Games))
	writeJSON(w, http.StatusOK, dtos.NewBracketResponse(bracket))
}

func (s *Server) selectWinnerHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("PANIC in selectWinnerHandler: %v", rec)
			writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error selecting winner", "")
		}
	}()

	vars := mux.Vars(r)
	tournamentID := vars["tournamentId"]
	gameID := vars["gameId"]

	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "tournamentId")
		return
	}
	if gameID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Game ID is required", "gameId")
		return
	}

	var req dtos.SelectWinnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if err := req.Validate(); err != nil {
		log.Printf("Request validation failed: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	log.Printf("Selecting winner for tournament %s, game %s, winner team %s", tournamentID, gameID, req.WinnerTeamID)
	bracket, err := s.bracketService.SelectWinner(r.Context(), tournamentID, gameID, req.WinnerTeamID)
	if err != nil {
		log.Printf("Error selecting winner for tournament %s, game %s: %v", tournamentID, gameID, err)
		writeErrorFromErr(w, r, err)
		return
	}

	log.Printf("Successfully selected winner, returning bracket with %d games", len(bracket.Games))
	writeJSON(w, http.StatusOK, dtos.NewBracketResponse(bracket))
}

func (s *Server) unselectWinnerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["tournamentId"]
	gameID := vars["gameId"]

	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "tournamentId")
		return
	}
	if gameID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Game ID is required", "gameId")
		return
	}

	bracket, err := s.bracketService.UnselectWinner(r.Context(), tournamentID, gameID)
	if err != nil {
		log.Printf("Error unselecting winner: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, dtos.NewBracketResponse(bracket))
}

func (s *Server) validateBracketSetupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	err := s.bracketService.ValidateBracketSetup(r.Context(), tournamentID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid":  false,
			"errors": []string{err.Error()},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":  true,
		"errors": []string{},
	})
}
