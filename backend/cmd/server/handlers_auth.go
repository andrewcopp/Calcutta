package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
	"github.com/andrewcopp/Calcutta/backend/pkg/services"
)

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req dtos.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	user, err := s.userService.Login(r.Context(), req.Email)
	if err != nil {
		var notFoundErr *services.NotFoundError
		if errors.As(err, &notFoundErr) {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "Invalid credentials", "")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewUserResponse(user))
}

func (s *Server) signupHandler(w http.ResponseWriter, r *http.Request) {
	var req dtos.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	user, err := s.userService.Signup(r.Context(), req.Email, req.FirstName, req.LastName)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, dtos.NewUserResponse(user))
}
