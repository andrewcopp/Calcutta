package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

func (s *Server) getBracketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["tournamentId"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "tournamentId")
		return
	}

	bracket, err := s.app.Bracket.GetBracket(r.Context(), tournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewBracketResponse(bracket))
}

func (s *Server) selectWinnerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["tournamentId"]
	gameID := vars["gameId"]

	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "tournamentId")
		return
	}
	if gameID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Game ID is required", "gameId")
		return
	}

	var req dtos.SelectWinnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	bracket, err := s.app.Bracket.SelectWinner(r.Context(), tournamentID, gameID, req.WinnerTeamID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewBracketResponse(bracket))
}

func (s *Server) unselectWinnerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["tournamentId"]
	gameID := vars["gameId"]

	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "tournamentId")
		return
	}
	if gameID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Game ID is required", "gameId")
		return
	}

	bracket, err := s.app.Bracket.UnselectWinner(r.Context(), tournamentID, gameID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewBracketResponse(bracket))
}

func (s *Server) validateBracketSetupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["tournamentId"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "tournamentId")
		return
	}

	err := s.app.Bracket.ValidateBracketSetup(r.Context(), tournamentID)
	if err != nil {
		response.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"valid":  false,
			"errors": []string{err.Error()},
		})
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"valid":  true,
		"errors": []string{},
	})
}
