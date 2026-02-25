package httpserver

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/requestctx"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (s *Server) adminResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if s.userRepo == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "database not available", "")
		return
	}
	if s.emailSender == nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_argument", "email sending is not configured", "")
		return
	}
	if strings.TrimSpace(s.cfg.InviteBaseURL) == "" {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_argument", "INVITE_BASE_URL is not configured", "")
		return
	}

	userID := strings.TrimSpace(mux.Vars(r)["id"])
	if userID == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("id"), authUserID)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("id", "invalid uuid"), authUserID)
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

	result, err := s.userRepo.GenerateResetToken(r.Context(), userID, now, expiresAt, genTokenFn)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	// Rate limit check.
	if result.LastResetSentAt != nil && s.cfg.InviteResendMinSeconds > 0 {
		min := time.Duration(s.cfg.InviteResendMinSeconds) * time.Second
		if now.Sub(*result.LastResetSentAt) < min {
			retryAfter := int(min.Seconds())
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			httperr.Write(w, r, http.StatusTooManyRequests, "rate_limited", "Reset email recently sent; please try again soon", "")
			return
		}
	}

	resetURL, err := buildResetURL(s.cfg.InviteBaseURL, result.Token)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	subject, body := buildResetEmail(resetURL, expiresAt)
	err = s.emailSender.Send(r.Context(), result.Email, subject, body)
	if err != nil {
		requestctx.Logger(r.Context()).ErrorContext(r.Context(), "admin_reset_email_send_failed",
			"event", "admin_reset_email_send_failed", "user_id", userID, "email", result.Email, "error", err.Error())
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to send reset email", "")
		return
	}

	requestctx.Logger(r.Context()).InfoContext(r.Context(), "admin_reset_email_sent",
		"event", "admin_reset_email_sent", "user_id", userID, "email", result.Email)
	response.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset email sent.",
	})
}
