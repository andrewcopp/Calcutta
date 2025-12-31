package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
)

func refreshCookieSettingsFromEnv() (secure bool, sameSite http.SameSite) {
	env := os.Getenv("NODE_ENV")
	if env == "" {
		env = "development"
	}

	secure = env != "development"
	if secure {
		sameSite = http.SameSiteNoneMode
	} else {
		sameSite = http.SameSiteLaxMode
	}

	if v := strings.TrimSpace(os.Getenv("COOKIE_SECURE")); v != "" {
		if strings.EqualFold(v, "true") || v == "1" || strings.EqualFold(v, "yes") {
			secure = true
		}
		if strings.EqualFold(v, "false") || v == "0" || strings.EqualFold(v, "no") {
			secure = false
		}
	}

	switch strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SAMESITE"))) {
	case "none":
		sameSite = http.SameSiteNoneMode
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "strict":
		sameSite = http.SameSiteStrictMode
	}

	if sameSite == http.SameSiteNoneMode && !secure {
		sameSite = http.SameSiteLaxMode
	}

	return secure, sameSite
}

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

	res, err := s.app.Auth.Login(r.Context(), req.Email, req.Password, r.UserAgent(), r.RemoteAddr, time.Now())
	if err != nil {
		var unauthorizedErr *apperrors.UnauthorizedError
		if errors.As(err, &unauthorizedErr) {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "Invalid credentials", "")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	setRefreshCookie(w, r, res.RefreshToken, res.RefreshExpiresAt)
	writeJSON(w, http.StatusOK, &dtos.AuthResponse{User: dtos.NewUserResponse(res.User), AccessToken: res.AccessToken})
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

	res, err := s.app.Auth.Signup(r.Context(), req.Email, req.FirstName, req.LastName, req.Password, r.UserAgent(), r.RemoteAddr, time.Now())
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	setRefreshCookie(w, r, res.RefreshToken, res.RefreshExpiresAt)
	writeJSON(w, http.StatusCreated, &dtos.AuthResponse{User: dtos.NewUserResponse(res.User), AccessToken: res.AccessToken})
}

func (s *Server) refreshHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("refresh_token")
	if err != nil || c.Value == "" {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "Refresh token missing", "")
		return
	}

	res, err := s.app.Auth.Refresh(r.Context(), c.Value, time.Now())
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	setRefreshCookie(w, r, res.RefreshToken, res.RefreshExpiresAt)
	writeJSON(w, http.StatusOK, &dtos.AuthResponse{User: dtos.NewUserResponse(res.User), AccessToken: res.AccessToken})
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("refresh_token")
	if err == nil && c.Value != "" {
		_ = s.app.Auth.Logout(r.Context(), c.Value)
	}
	clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func setRefreshCookie(w http.ResponseWriter, r *http.Request, refreshToken string, expiresAt time.Time) {
	secure, sameSite := refreshCookieSettingsFromEnv()
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/auth",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func clearRefreshCookie(w http.ResponseWriter) {
	secure, sameSite := refreshCookieSettingsFromEnv()
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}
