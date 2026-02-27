package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
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
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	userIDs, err := s.authzRepo.ListGrantsByScope(r.Context(), "tournament_admin", "tournament", tournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	users, err := s.userRepo.GetByIDs(r.Context(), userIDs)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	moderators := make([]moderatorResponse, 0, len(users))
	for _, user := range users {
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

	response.WriteJSON(w, http.StatusOK, map[string]any{"items": moderators})
}

func (s *Server) grantTournamentModeratorHandler(w http.ResponseWriter, r *http.Request) {
	tournamentID := mux.Vars(r)["id"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	var req grantModeratorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if req.Email == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "email is required", "email")
		return
	}

	user, err := s.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "No user found with that email", "email")
		return
	}

	if err := s.authzRepo.GrantRole(r.Context(), user.ID, "tournament_admin", "tournament", tournamentID); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	response.WriteJSON(w, http.StatusCreated, moderatorResponse{
		ID:        user.ID,
		Email:     email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})
}

func (s *Server) revokeTournamentModeratorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	userID := vars["userId"]
	if tournamentID == "" || userID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Tournament ID and User ID are required", "")
		return
	}

	if err := s.authzRepo.RevokeGrant(r.Context(), userID, "tournament_admin", "tournament", tournamentID); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}
