package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type moderatorResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type grantModeratorRequest struct {
	Email string `json:"email"`
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

	moderators := make([]moderatorResponse, 0, len(userIDs))
	for _, uid := range userIDs {
		user, err := s.userRepo.GetByID(r.Context(), uid)
		if err != nil {
			continue
		}
		email := ""
		if user.Email != nil {
			email = *user.Email
		}
		moderators = append(moderators, moderatorResponse{
			ID:        user.ID,
			Email:     email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"moderators": moderators})
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
	if req.Email == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "email is required", "email")
		return
	}

	user, err := s.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "No user found with that email", "email")
		return
	}

	if err := s.authzRepo.GrantLabel(r.Context(), user.ID, "tournament_operator", "tournament", tournamentID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"moderator": moderatorResponse{
			ID:        user.ID,
			Email:     email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	})
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
