package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) previewInviteHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}

	rawToken := r.URL.Query().Get("token")
	if rawToken == "" {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "token query parameter is required", "")
		return
	}

	now := time.Now().UTC()
	h := coreauth.HashInviteToken(rawToken)

	var firstName, calcuttaName, commissionerName string
	var tournamentStartingAt *time.Time

	err := s.pool.QueryRow(r.Context(), `
		SELECT
			u.first_name,
			c.name,
			comm.first_name || ' ' || comm.last_name,
			t.starting_at
		FROM core.users u
		JOIN core.calcutta_invitations ci ON ci.user_id = u.id AND ci.deleted_at IS NULL
		JOIN core.calcuttas c ON c.id = ci.calcutta_id AND c.deleted_at IS NULL
		JOIN core.users comm ON comm.id = ci.invited_by AND comm.deleted_at IS NULL
		LEFT JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		WHERE u.invite_token_hash = $1
		  AND u.deleted_at IS NULL
		  AND u.invite_consumed_at IS NULL
		  AND u.invite_expires_at IS NOT NULL
		  AND u.invite_expires_at > $2
		LIMIT 1
	`, h, now).Scan(&firstName, &calcuttaName, &commissionerName, &tournamentStartingAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httperr.Write(w, r, http.StatusNotFound, "not_found", "invite not found or expired", "")
			return
		}
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	resp := dtos.InvitePreviewResponse{
		FirstName:        firstName,
		CalcuttaName:     calcuttaName,
		CommissionerName: commissionerName,
	}
	if tournamentStartingAt != nil {
		ts := tournamentStartingAt.Format(time.RFC3339)
		resp.TournamentStartingAt = &ts
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) acceptInviteHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}
	if s.app == nil || s.app.Auth == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "auth service not available", "")
		return
	}

	var req dtos.AcceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	now := time.Now().UTC()
	h := coreauth.HashInviteToken(req.Token)

	tx, err := s.pool.Begin(r.Context())
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	defer tx.Rollback(r.Context())

	var userID string
	var email *string
	err = tx.QueryRow(r.Context(), `
		SELECT id::text, email
		FROM core.users
		WHERE invite_token_hash = $1
		  AND deleted_at IS NULL
		  AND invite_consumed_at IS NULL
		  AND invite_expires_at IS NOT NULL
		  AND invite_expires_at > $2
		FOR UPDATE
	`, h, now).Scan(&userID, &email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httperr.WriteFromErr(w, r, &apperrors.UnauthorizedError{Message: "invalid or expired invite token"}, authUserID)
			return
		}
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if email == nil || *email == "" {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_state", "User has no email set", "")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	hashStr := string(hash)

	_, err = tx.Exec(r.Context(), `
		UPDATE core.users
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
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	res, err := s.app.Auth.Login(r.Context(), *email, req.Password, r.UserAgent(), r.RemoteAddr, time.Now())
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	s.setRefreshCookie(w, res.RefreshToken, res.RefreshExpiresAt)
	response.WriteJSON(w, http.StatusOK, &dtos.AuthResponse{User: dtos.NewUserResponse(res.User), AccessToken: res.AccessToken})
}
