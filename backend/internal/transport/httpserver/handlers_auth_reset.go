package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/requestctx"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if s.userRepo == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "database not available", "")
		return
	}
	if s.emailSender == nil {
		// Silently succeed to avoid leaking configuration details.
		response.WriteJSON(w, http.StatusOK, map[string]string{
			"message": "If that email is registered, a reset link has been sent.",
		})
		return
	}
	if strings.TrimSpace(s.cfg.InviteBaseURL) == "" {
		response.WriteJSON(w, http.StatusOK, map[string]string{
			"message": "If that email is registered, a reset link has been sent.",
		})
		return
	}

	var req dtos.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	now := time.Now().UTC()
	expiresAt := now.Add(30 * time.Minute)

	genTokenFn := func() (string, string, error) {
		raw, err := coreauth.NewResetToken()
		if err != nil {
			return "", "", err
		}
		return raw, coreauth.HashResetToken(raw), nil
	}

	result, err := s.userRepo.GenerateResetTokenByEmail(r.Context(), req.Email, now, expiresAt, genTokenFn)
	if err != nil {
		requestctx.Logger(r.Context()).ErrorContext(r.Context(), "forgot_password_generate_failed",
			"event", "forgot_password_generate_failed", "email", req.Email, "error", err.Error())
		// Fall through to generic response.
	}

	if result != nil {
		resetURL, urlErr := buildResetURL(s.cfg.InviteBaseURL, result.Token)
		if urlErr != nil {
			requestctx.Logger(r.Context()).ErrorContext(r.Context(), "forgot_password_url_build_failed",
				"event", "forgot_password_url_build_failed", "error", urlErr.Error())
		} else {
			subject, body := buildResetEmail(resetURL, expiresAt)
			sendErr := s.emailSender.Send(r.Context(), result.Email, subject, body)
			if sendErr != nil {
				requestctx.Logger(r.Context()).ErrorContext(r.Context(), "forgot_password_email_send_failed",
					"event", "forgot_password_email_send_failed", "email", result.Email, "error", sendErr.Error())
			} else {
				requestctx.Logger(r.Context()).InfoContext(r.Context(), "reset_email_sent",
					"event", "reset_email_sent", "email", result.Email)
			}
		}
	}

	// Always return the same response regardless of whether the email exists.
	response.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "If that email is registered, a reset link has been sent.",
	})
}

func (s *Server) resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}
	if s.app == nil || s.app.Auth == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "auth service not available", "")
		return
	}

	var req dtos.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	now := time.Now().UTC()
	h := coreauth.HashResetToken(req.Token)

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
		WHERE reset_token_hash = $1
		  AND deleted_at IS NULL
		  AND reset_consumed_at IS NULL
		  AND reset_expires_at IS NOT NULL
		  AND reset_expires_at > $2
		FOR UPDATE
	`, h, now).Scan(&userID, &email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httperr.WriteFromErr(w, r, &apperrors.UnauthorizedError{Message: "invalid or expired reset token"}, authUserID)
			return
		}
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if email == nil || *email == "" {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_state", "User has no email set", "")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), coreauth.BcryptCost)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	hashStr := string(hash)

	_, err = tx.Exec(r.Context(), `
		UPDATE core.users
		SET
		  password_hash = $2,
		  reset_consumed_at = $3,
		  reset_token_hash = NULL,
		  reset_expires_at = NULL,
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

	// Revoke all existing sessions (best-effort, outside transaction).
	if err := s.authRepo.RevokeAllSessionsForUser(r.Context(), userID); err != nil {
		requestctx.Logger(r.Context()).ErrorContext(r.Context(), "reset_password_revoke_sessions_failed",
			"event", "reset_password_revoke_sessions_failed", "user_id", userID, "error", err.Error())
	}

	// Auto-login.
	res, err := s.app.Auth.Login(r.Context(), *email, req.Password, r.UserAgent(), r.RemoteAddr, time.Now())
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	s.setRefreshCookie(w, res.RefreshToken, res.RefreshExpiresAt)
	response.WriteJSON(w, http.StatusOK, &dtos.AuthResponse{User: dtos.NewUserResponse(res.User), AccessToken: res.AccessToken})
}

func buildResetURL(base string, token string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		return "", fmt.Errorf("invite base url must include scheme")
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/reset-password"
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func buildResetEmail(resetURL string, expiresAt time.Time) (string, string) {
	subject := "Reset your March Markets password"
	body := "Hey!\n\n" +
		"We received a request to reset your password.\n\n" +
		"Reset your password here:\n" + resetURL + "\n\n" +
		"This link expires at " + expiresAt.UTC().Format(time.RFC3339) + ".\n\n" +
		"If you didn't request this, you can ignore this email.\n\n" +
		"- March Markets"
	return subject, body
}
