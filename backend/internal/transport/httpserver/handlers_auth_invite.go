package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) acceptInviteHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}
	if s.app == nil || s.app.Auth == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "auth service not available", "")
		return
	}

	var req dtos.AcceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	now := time.Now().UTC()
	h := coreauth.HashInviteToken(req.Token)

	tx, err := s.pool.Begin(r.Context())
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer tx.Rollback(r.Context())

	var userID string
	var email string
	err = tx.QueryRow(r.Context(), `
		SELECT id::text, email
		FROM users
		WHERE invite_token_hash = $1
		  AND deleted_at IS NULL
		  AND invite_consumed_at IS NULL
		  AND invite_expires_at IS NOT NULL
		  AND invite_expires_at > $2
		FOR UPDATE
	`, h, now).Scan(&userID, &email)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeErrorFromErr(w, r, &apperrors.UnauthorizedError{Message: "invalid or expired invite token"})
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	hashStr := string(hash)

	_, err = tx.Exec(r.Context(), `
		UPDATE users
		SET
		  password_hash = $2,
		  status = 'active',
		  invite_consumed_at = $3,
		  invite_token_hash = NULL,
		  invite_expires_at = NULL,
		  updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`, userID, hashStr, now)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	res, err := s.app.Auth.Login(r.Context(), email, req.Password, r.UserAgent(), r.RemoteAddr, time.Now())
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	setRefreshCookie(w, r, res.RefreshToken, res.RefreshExpiresAt)
	writeJSON(w, http.StatusOK, &dtos.AuthResponse{User: dtos.NewUserResponse(res.User), AccessToken: res.AccessToken})
}
