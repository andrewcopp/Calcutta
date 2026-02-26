package httpserver

import (
	"context"
	"encoding/json"
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
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type adminUserListItem struct {
	ID               string     `json:"id"`
	Email            *string    `json:"email"`
	FirstName        string     `json:"firstName"`
	LastName         string     `json:"lastName"`
	Status           string     `json:"status"`
	InvitedAt        *time.Time `json:"invitedAt"`
	LastInviteSentAt *time.Time `json:"lastInviteSentAt"`
	InviteExpiresAt  *time.Time `json:"inviteExpiresAt"`
	InviteConsumedAt *time.Time `json:"inviteConsumedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	Roles            []string   `json:"roles"`
	Permissions      []string   `json:"permissions"`
}

type adminUsersListResponse struct {
	Items []adminUserListItem `json:"items"`
}

func (s *Server) registerAdminUsersRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/admin/users", s.requirePermission("admin.users.read", s.adminUsersListHandler)).Methods("GET")
	r.HandleFunc("/api/v1/admin/users/{id}/email", s.requirePermission("admin.users.write", s.adminUsersSetEmailHandler)).Methods("PATCH")
	r.HandleFunc("/api/v1/admin/users/{id}/invite", s.requirePermission("admin.users.write", s.adminUsersInviteHandler)).Methods("POST")
	r.HandleFunc("/api/v1/admin/users/{id}/invite/send", s.requirePermission("admin.users.write", s.adminUsersInviteSendHandler)).Methods("POST")
	r.HandleFunc("/api/v1/admin/users/{id}/reset-password", s.requirePermission("admin.users.write", s.adminResetPasswordHandler)).Methods("POST")
	r.HandleFunc("/api/v1/admin/users/{id}", s.requirePermission("admin.users.read", s.adminUserDetailHandler)).Methods("GET")
	r.HandleFunc("/api/v1/admin/users/{id}/roles", s.requirePermission("admin.users.write", s.adminGrantRoleHandler)).Methods("POST")
	r.HandleFunc("/api/v1/admin/users/{id}/roles/{roleKey}", s.requirePermission("admin.users.write", s.adminRevokeRoleHandler)).Methods("DELETE")
}

func (s *Server) adminUsersSetEmailHandler(w http.ResponseWriter, r *http.Request) {
	if s.userRepo == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "database not available", "")
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

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	email := strings.TrimSpace(req.Email)
	if email == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("email"), authUserID)
		return
	}
	if !strings.Contains(email, "@") {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("email", "invalid email format"), authUserID)
		return
	}

	user, err := s.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if user == nil {
		httperr.WriteFromErr(w, r, &apperrors.NotFoundError{Resource: "user", ID: userID}, authUserID)
		return
	}
	if user.Status != "stub" {
		httperr.WriteFromErr(w, r, &apperrors.InvalidArgumentError{Field: "status", Message: "can only set email on stub users"}, authUserID)
		return
	}

	existing, err := s.userRepo.GetByEmail(r.Context(), email)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if existing != nil {
		httperr.WriteFromErr(w, r, &apperrors.AlreadyExistsError{Resource: "user", Field: "email", Value: email}, authUserID)
		return
	}

	user.Email = &email
	user.Status = "invited"
	if err := s.userRepo.Update(r.Context(), user); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewUserResponse(user))
}

func (s *Server) adminUsersInviteHandler(w http.ResponseWriter, r *http.Request) {
	if s.userRepo == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "database not available", "")
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
	expiresAt := now.Add(7 * 24 * time.Hour)

	_, email, err := s.generateInviteToken(r.Context(), userID, now, expiresAt, false, nil)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.AdminInviteUserResponse{
		UserID:        userID,
		Email:         email,
		InviteExpires: expiresAt.Format(time.RFC3339),
		Status:        "requires_password_setup",
	})
}

func (s *Server) adminUsersInviteSendHandler(w http.ResponseWriter, r *http.Request) {
	if s.userRepo == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "database not available", "")
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
	if s.emailSender == nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_argument", "email sending is not configured", "")
		return
	}
	if strings.TrimSpace(s.cfg.InviteBaseURL) == "" {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_argument", "INVITE_BASE_URL is not configured", "")
		return
	}

	now := time.Now().UTC()
	expiresAt := now.Add(7 * 24 * time.Hour)

	var lastSent *time.Time
	inviteToken, email, err := s.generateInviteToken(r.Context(), userID, now, expiresAt, false, &lastSent)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if lastSent != nil && s.cfg.InviteResendMinSeconds > 0 {
		min := time.Duration(s.cfg.InviteResendMinSeconds) * time.Second
		if now.Sub(*lastSent) < min {
			retryAfter := int(min.Seconds())
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			httperr.Write(w, r, http.StatusTooManyRequests, "rate_limited", "Invite email recently sent; please try again soon", "")
			return
		}
	}

	inviteURL, err := buildInviteURL(s.cfg.InviteBaseURL, inviteToken)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	subject, body := buildInviteEmail(inviteURL, expiresAt)
	err = s.emailSender.Send(r.Context(), email, subject, body)
	if err != nil {
		requestctx.Logger(r.Context()).ErrorContext(r.Context(), "invite_email_send_failed", "event", "invite_email_send_failed", "user_id", userID, "email", email, "error", err.Error())
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to send invite email", "")
		return
	}

	if err := s.userRepo.UpdateLastInviteSentAt(r.Context(), userID, now); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	requestctx.Logger(r.Context()).InfoContext(r.Context(), "invite_email_sent", "event", "invite_email_sent", "user_id", userID, "email", email)
	response.WriteJSON(w, http.StatusOK, dtos.AdminInviteUserResponse{
		UserID:        userID,
		Email:         email,
		InviteExpires: expiresAt.Format(time.RFC3339),
		Status:        "requires_password_setup",
	})
}

func (s *Server) generateInviteToken(ctx context.Context, userID string, now time.Time, expiresAt time.Time, setLastSent bool, lastSentOut **time.Time) (string, string, error) {
	genTokenFn := func() (string, string, error) {
		raw, err := coreauth.NewInviteToken()
		if err != nil {
			return "", "", err
		}
		return raw, coreauth.HashInviteToken(raw), nil
	}

	result, err := s.userRepo.GenerateInviteToken(ctx, userID, now, expiresAt, setLastSent, genTokenFn)
	if err != nil {
		return "", "", err
	}

	if lastSentOut != nil {
		*lastSentOut = result.LastInviteSentAt
	}

	return result.Token, result.Email, nil
}

func buildInviteURL(base string, token string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		return "", fmt.Errorf("invite base url must include scheme")
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/invite"
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func buildInviteEmail(inviteURL string, expiresAt time.Time) (string, string) {
	subject := "You're invited to March Markets"
	body := "Hey!\n\n" +
		"You've been invited to March Markets.\n\n" +
		"Set your password here:\n" + inviteURL + "\n\n" +
		"This link expires at " + expiresAt.UTC().Format(time.RFC3339) + ".\n\n" +
		"- March Markets"
	return subject, body
}

func (s *Server) adminUsersListHandler(w http.ResponseWriter, r *http.Request) {
	if s.userRepo == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "database not available", "")
		return
	}

	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	if statusFilter != "" {
		switch statusFilter {
		case "active", "invited", "requires_password_setup", "stub":
		default:
			httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("status", "invalid status"), authUserID)
			return
		}
	}

	rows, err := s.userRepo.ListAdminUsers(r.Context(), statusFilter)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	items := make([]adminUserListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminUserListItem{
			ID:               row.ID,
			Email:            row.Email,
			FirstName:        row.FirstName,
			LastName:         row.LastName,
			Status:           row.Status,
			InvitedAt:        row.InvitedAt,
			LastInviteSentAt: row.LastInviteSentAt,
			InviteExpiresAt:  row.InviteExpiresAt,
			InviteConsumedAt: row.InviteConsumedAt,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
			Roles:            row.Roles,
			Permissions:      row.Permissions,
		})
	}

	response.WriteJSON(w, http.StatusOK, adminUsersListResponse{Items: items})
}
