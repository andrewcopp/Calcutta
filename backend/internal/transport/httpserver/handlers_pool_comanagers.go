package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

type coManagerResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type grantCoManagerRequest struct {
	Email string `json:"email"`
}

func (s *Server) registerPoolCoManagerRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/pools/{id}/co-managers", s.requirePermissionWithScope("pool.config.write", "pool", "id", s.listPoolCoManagersHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/pools/{id}/co-managers", s.requirePermissionWithScope("pool.config.write", "pool", "id", s.grantPoolCoManagerHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/pools/{id}/co-managers/{userId}", s.requirePermissionWithScope("pool.config.write", "pool", "id", s.revokePoolCoManagerHandler)).Methods("DELETE", "OPTIONS")
}

func (s *Server) listPoolCoManagersHandler(w http.ResponseWriter, r *http.Request) {
	poolID := mux.Vars(r)["id"]
	if poolID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Pool ID is required", "id")
		return
	}

	userIDs, err := s.authzRepo.ListGrantsByScope(r.Context(), "pool_admin", "pool", poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	users, err := s.userRepo.GetByIDs(r.Context(), userIDs)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	coManagers := make([]coManagerResponse, 0, len(users))
	for _, user := range users {
		email := ""
		if user.Email != nil {
			email = *user.Email
		}
		coManagers = append(coManagers, coManagerResponse{
			ID:        user.ID,
			Email:     email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		})
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{"items": coManagers})
}

func (s *Server) grantPoolCoManagerHandler(w http.ResponseWriter, r *http.Request) {
	poolID := mux.Vars(r)["id"]
	if poolID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Pool ID is required", "id")
		return
	}

	var req grantCoManagerRequest
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

	if err := s.authzRepo.GrantRole(r.Context(), user.ID, "pool_admin", "pool", poolID); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	response.WriteJSON(w, http.StatusCreated, coManagerResponse{
		ID:        user.ID,
		Email:     email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})
}

func (s *Server) revokePoolCoManagerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	poolID := vars["id"]
	targetUserID := vars["userId"]
	if poolID == "" || targetUserID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Pool ID and User ID are required", "")
		return
	}

	// Prevent revoking the pool owner
	pool, err := s.app.Pool.GetPoolByID(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if pool.OwnerID == targetUserID {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Cannot revoke the pool owner's admin access", "")
		return
	}

	if err := s.authzRepo.RevokeGrant(r.Context(), targetUserID, "pool_admin", "pool", poolID); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}
