package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type adminUserListItem struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	Status           string     `json:"status"`
	InvitedAt        *time.Time `json:"invited_at"`
	LastInviteSentAt *time.Time `json:"last_invite_sent_at"`
	InviteExpiresAt  *time.Time `json:"invite_expires_at"`
	InviteConsumedAt *time.Time `json:"invite_consumed_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	Labels           []string   `json:"labels"`
	Permissions      []string   `json:"permissions"`
}

type adminUsersListResponse struct {
	Items []adminUserListItem `json:"items"`
}

func (s *Server) registerAdminUsersRoutes(r *mux.Router) {
	r.HandleFunc("/api/admin/users", s.requirePermission("admin.users.read", s.adminUsersListHandler)).Methods("GET")
	r.HandleFunc("/api/admin/users/{id}/invite", s.requirePermission("admin.users.write", s.adminUsersInviteHandler)).Methods("POST")
	r.HandleFunc("/api/admin/users/{id}/invite/send", s.requirePermission("admin.users.write", s.adminUsersInviteSendHandler)).Methods("POST")
	r.HandleFunc("/api/admin/users/{id}/invite/resend", s.requirePermission("admin.users.write", s.adminUsersInviteSendHandler)).Methods("POST")
}

func (s *Server) adminUsersInviteHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}

	userID := strings.TrimSpace(mux.Vars(r)["id"])
	if userID == "" {
		writeErrorFromErr(w, r, dtos.ErrFieldRequired("id"))
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		writeErrorFromErr(w, r, dtos.ErrFieldInvalid("id", "invalid uuid"))
		return
	}

	now := time.Now().UTC()
	expiresAt := now.Add(7 * 24 * time.Hour)

	inviteToken, email, err := s.generateInviteToken(r.Context(), userID, now, expiresAt, false, nil)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, dtos.AdminInviteUserResponse{
		UserID:        userID,
		Email:         email,
		InviteToken:   inviteToken,
		InviteExpires: expiresAt.Format(time.RFC3339),
		Status:        "requires_password_setup",
	})
}

func (s *Server) adminUsersInviteSendHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}

	userID := strings.TrimSpace(mux.Vars(r)["id"])
	if userID == "" {
		writeErrorFromErr(w, r, dtos.ErrFieldRequired("id"))
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		writeErrorFromErr(w, r, dtos.ErrFieldInvalid("id", "invalid uuid"))
		return
	}
	if s.emailSender == nil {
		writeError(w, r, http.StatusBadRequest, "invalid_argument", "email sending is not configured", "")
		return
	}
	if strings.TrimSpace(s.cfg.InviteBaseURL) == "" {
		writeError(w, r, http.StatusBadRequest, "invalid_argument", "INVITE_BASE_URL is not configured", "")
		return
	}

	now := time.Now().UTC()
	expiresAt := now.Add(7 * 24 * time.Hour)

	var lastSent *time.Time
	inviteToken, email, err := s.generateInviteToken(r.Context(), userID, now, expiresAt, false, &lastSent)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if lastSent != nil && s.cfg.InviteResendMinSeconds > 0 {
		min := time.Duration(s.cfg.InviteResendMinSeconds) * time.Second
		if now.Sub(*lastSent) < min {
			retryAfter := int(min.Seconds())
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			writeError(w, r, http.StatusTooManyRequests, "rate_limited", "Invite email recently sent; please try again soon", "")
			return
		}
	}

	inviteURL, err := buildInviteURL(s.cfg.InviteBaseURL, inviteToken)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	subject, body := buildInviteEmail(inviteURL, expiresAt)
	err = s.emailSender.Send(r.Context(), email, subject, body)
	if err != nil {
		requestLogger(r.Context()).ErrorContext(r.Context(), "invite_email_send_failed", "event", "invite_email_send_failed", "user_id", userID, "email", email, "error", err.Error())
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to send invite email", "")
		return
	}

	_, err = s.pool.Exec(r.Context(), `
		UPDATE users
		SET last_invite_sent_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`, userID, now)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	requestLogger(r.Context()).InfoContext(r.Context(), "invite_email_sent", "event", "invite_email_sent", "user_id", userID, "email", email)
	writeJSON(w, http.StatusOK, dtos.AdminInviteUserResponse{
		UserID:        userID,
		Email:         email,
		InviteToken:   inviteToken,
		InviteExpires: expiresAt.Format(time.RFC3339),
		Status:        "requires_password_setup",
	})
}

func (s *Server) generateInviteToken(ctx context.Context, userID string, now time.Time, expiresAt time.Time, setLastSent bool, lastSentOut **time.Time) (string, string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback(ctx)

	var email string
	var passwordHash *string
	var lastInviteSentAt pgtype.Timestamptz
	err = tx.QueryRow(ctx, `
		SELECT email, password_hash, last_invite_sent_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, userID).Scan(&email, &passwordHash, &lastInviteSentAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", "", &apperrors.NotFoundError{Resource: "user", ID: userID}
		}
		return "", "", err
	}
	if lastSentOut != nil {
		if lastInviteSentAt.Valid {
			t := lastInviteSentAt.Time.UTC()
			*lastSentOut = &t
		} else {
			*lastSentOut = nil
		}
	}
	if passwordHash != nil && strings.TrimSpace(*passwordHash) != "" {
		return "", "", &apperrors.InvalidArgumentError{Field: "id", Message: "user already has a password set"}
	}

	var inviteToken string
	var inviteHash string
	for i := 0; i < 3; i++ {
		created, err := coreauth.NewInviteToken()
		if err != nil {
			return "", "", err
		}
		createdHash := coreauth.HashInviteToken(created)

		var ct pgconn.CommandTag
		if setLastSent {
			ct, err = tx.Exec(ctx, `
				UPDATE users
				SET
				  status = 'requires_password_setup',
				  invite_token_hash = $2,
				  invite_expires_at = $3,
				  invite_consumed_at = NULL,
				  invited_at = COALESCE(invited_at, $4),
				  last_invite_sent_at = $4,
				  updated_at = $4
				WHERE id = $1 AND deleted_at IS NULL
			`, userID, createdHash, expiresAt, now)
		} else {
			ct, err = tx.Exec(ctx, `
				UPDATE users
				SET
				  status = 'requires_password_setup',
				  invite_token_hash = $2,
				  invite_expires_at = $3,
				  invite_consumed_at = NULL,
				  invited_at = COALESCE(invited_at, $4),
				  updated_at = $4
				WHERE id = $1 AND deleted_at IS NULL
			`, userID, createdHash, expiresAt, now)
		}
		if err != nil {
			continue
		}
		if ct.RowsAffected() == 0 {
			return "", "", &apperrors.NotFoundError{Resource: "user", ID: userID}
		}
		inviteToken = created
		inviteHash = createdHash
		break
	}
	if inviteToken == "" || inviteHash == "" {
		return "", "", fmt.Errorf("failed to generate invite token")
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", err
	}
	return inviteToken, email, nil
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
	if s.pool == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}

	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	if statusFilter != "" {
		switch statusFilter {
		case "active", "invited", "requires_password_setup":
		default:
			writeErrorFromErr(w, r, dtos.ErrFieldInvalid("status", "invalid status"))
			return
		}
	}

	rows, err := s.pool.Query(r.Context(), `
		WITH active_grants AS (
			SELECT *
			FROM grants g
			WHERE g.revoked_at IS NULL
			  AND g.scope_type = 'global'
			  AND (g.expires_at IS NULL OR g.expires_at > NOW())
		),
		user_labels AS (
			SELECT g.user_id, l.key
			FROM active_grants g
			JOIN labels l ON g.label_id = l.id
			WHERE l.deleted_at IS NULL
		),
		user_permissions AS (
			SELECT g.user_id, p.key
			FROM active_grants g
			JOIN permissions p ON g.permission_id = p.id
			WHERE p.deleted_at IS NULL
			UNION
			SELECT g.user_id, p2.key
			FROM active_grants g
			JOIN labels l ON g.label_id = l.id AND l.deleted_at IS NULL
			JOIN label_permissions lp ON lp.label_id = l.id
			JOIN permissions p2 ON lp.permission_id = p2.id AND p2.deleted_at IS NULL
		)
		SELECT
			u.id::text,
			u.email,
			u.first_name,
			u.last_name,
			u.status,
			u.invited_at,
			u.last_invite_sent_at,
			u.invite_expires_at,
			u.invite_consumed_at,
			u.created_at,
			u.updated_at,
			COALESCE(array_agg(DISTINCT ul.key) FILTER (WHERE ul.key IS NOT NULL), ARRAY[]::text[]) AS labels,
			COALESCE(array_agg(DISTINCT up.key) FILTER (WHERE up.key IS NOT NULL), ARRAY[]::text[]) AS permissions
		FROM users u
		LEFT JOIN user_labels ul ON ul.user_id = u.id
		LEFT JOIN user_permissions up ON up.user_id = u.id
		WHERE u.deleted_at IS NULL
		  AND ($1 = '' OR u.status = $1)
		GROUP BY u.id, u.email, u.first_name, u.last_name, u.status, u.invited_at, u.last_invite_sent_at, u.invite_expires_at, u.invite_consumed_at, u.created_at, u.updated_at
		ORDER BY u.created_at DESC
	`, statusFilter)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]adminUserListItem, 0)
	for rows.Next() {
		var it adminUserListItem
		var labels []string
		var perms []string

		var invitedAt pgtype.Timestamptz
		var lastInviteSentAt pgtype.Timestamptz
		var inviteExpiresAt pgtype.Timestamptz
		var inviteConsumedAt pgtype.Timestamptz

		if err := rows.Scan(
			&it.ID,
			&it.Email,
			&it.FirstName,
			&it.LastName,
			&it.Status,
			&invitedAt,
			&lastInviteSentAt,
			&inviteExpiresAt,
			&inviteConsumedAt,
			&it.CreatedAt,
			&it.UpdatedAt,
			&labels,
			&perms,
		); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		it.InvitedAt = timestamptzPtr(invitedAt)
		it.LastInviteSentAt = timestamptzPtr(lastInviteSentAt)
		it.InviteExpiresAt = timestamptzPtr(inviteExpiresAt)
		it.InviteConsumedAt = timestamptzPtr(inviteConsumedAt)

		sort.Strings(labels)
		sort.Strings(perms)
		it.Labels = labels
		it.Permissions = perms

		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, adminUsersListResponse{Items: items})
}

func timestamptzPtr(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	if t.IsZero() {
		return nil
	}
	ut := t.UTC()
	return &ut
}
