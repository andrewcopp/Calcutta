package httpserver

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
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

	bracket, err := s.app.Bracket.GetBracket(r.Context(), tournamentID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

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
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	bracket, err := s.app.Bracket.SelectWinner(r.Context(), tournamentID, gameID, req.WinnerTeamID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

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

	bracket, err := s.app.Bracket.UnselectWinner(r.Context(), tournamentID, gameID)
	if err != nil {
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

	err := s.app.Bracket.ValidateBracketSetup(r.Context(), tournamentID)
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
