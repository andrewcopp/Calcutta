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

func (s *Server) listCalcuttaCoManagersHandler(w http.ResponseWriter, r *http.Request) {
	calcuttaID := mux.Vars(r)["id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "id")
		return
	}

	userIDs, err := s.authzRepo.ListGrantsByScope(r.Context(), "calcutta_admin", "calcutta", calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	coManagers := make([]coManagerResponse, 0, len(userIDs))
	for _, uid := range userIDs {
		user, err := s.userRepo.GetByID(r.Context(), uid)
		if err != nil {
			httperr.WriteFromErr(w, r, err, authUserID)
			return
		}
		if user == nil {
			continue
		}
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

func (s *Server) grantCalcuttaCoManagerHandler(w http.ResponseWriter, r *http.Request) {
	calcuttaID := mux.Vars(r)["id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "id")
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

	if err := s.authzRepo.GrantRole(r.Context(), user.ID, "calcutta_admin", "calcutta", calcuttaID); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	response.WriteJSON(w, http.StatusCreated, map[string]any{
		"coManager": coManagerResponse{
			ID:        user.ID,
			Email:     email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	})
}

func (s *Server) revokeCalcuttaCoManagerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	targetUserID := vars["userId"]
	if calcuttaID == "" || targetUserID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID and User ID are required", "")
		return
	}

	// Prevent revoking the calcutta owner
	calcutta, err := s.app.Calcutta.GetCalcuttaByID(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if calcutta.OwnerID == targetUserID {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Cannot revoke the pool owner's admin access", "")
		return
	}

	if err := s.authzRepo.RevokeGrant(r.Context(), targetUserID, "calcutta_admin", "calcutta", calcuttaID); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}
