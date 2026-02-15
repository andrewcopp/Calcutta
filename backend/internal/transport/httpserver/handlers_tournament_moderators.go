package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type grantModeratorRequest struct {
	UserID string `json:"userId"`
}

func (s *Server) listTournamentModeratorsHandler(w http.ResponseWriter, r *http.Request) {
	tournamentID := mux.Vars(r)["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	userIDs, err := s.authzRepo.ListGrantsByScope(r.Context(), "tournament_operator", "tournament", tournamentID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"moderators": userIDs})
}

func (s *Server) grantTournamentModeratorHandler(w http.ResponseWriter, r *http.Request) {
	tournamentID := mux.Vars(r)["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	var req grantModeratorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if req.UserID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "userId is required", "userId")
		return
	}

	if err := s.authzRepo.GrantLabel(r.Context(), req.UserID, "tournament_operator", "tournament", tournamentID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "granted"})
}

func (s *Server) revokeTournamentModeratorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	userID := vars["userId"]
	if tournamentID == "" || userID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Tournament ID and User ID are required", "")
		return
	}

	if err := s.authzRepo.RevokeGrant(r.Context(), userID, "tournament_operator", "tournament", tournamentID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}
