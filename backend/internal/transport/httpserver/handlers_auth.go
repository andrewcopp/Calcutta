package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
)

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req dtos.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	res, err := s.app.Auth.Login(r.Context(), req.Email, req.Password, r.UserAgent(), r.RemoteAddr, time.Now())
	if err != nil {
		var unauthorizedErr *apperrors.UnauthorizedError
		if errors.As(err, &unauthorizedErr) {
			httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Invalid credentials", "")
			return
		}
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	s.setRefreshCookie(w, res.RefreshToken, res.RefreshExpiresAt)
	response.WriteJSON(w, http.StatusOK, &dtos.AuthResponse{User: dtos.NewUserResponse(res.User), AccessToken: res.AccessToken})
}

func (s *Server) refreshHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("refresh_token")
	if err != nil || c.Value == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Refresh token missing", "")
		return
	}

	res, err := s.app.Auth.Refresh(r.Context(), c.Value, time.Now())
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	s.setRefreshCookie(w, res.RefreshToken, res.RefreshExpiresAt)
	response.WriteJSON(w, http.StatusOK, &dtos.AuthResponse{User: dtos.NewUserResponse(res.User), AccessToken: res.AccessToken})
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("refresh_token")
	if err == nil && c.Value != "" {
		_ = s.app.Auth.Logout(r.Context(), c.Value)
	}
	s.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) setRefreshCookie(w http.ResponseWriter, refreshToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/auth",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: s.cookieSameSite,
	})
}

func (s *Server) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: s.cookieSameSite,
	})
}

func (s *Server) mePermissionsHandler(w http.ResponseWriter, r *http.Request) {
	userID := authUserID(r.Context())
	if userID == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	permissions, err := s.authzRepo.ListUserGlobalPermissions(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if permissions == nil {
		permissions = []string{}
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{"permissions": permissions})
}
