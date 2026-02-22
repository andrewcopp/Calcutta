package db

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// AdminUserRow holds the result of the admin user list query, including
// aggregated labels and permissions from the authorization system.
type AdminUserRow struct {
	ID               string
	Email            *string
	FirstName        string
	LastName         string
	Status           string
	InvitedAt        *time.Time
	LastInviteSentAt *time.Time
	InviteExpiresAt  *time.Time
	InviteConsumedAt *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Labels           []string
	Permissions      []string
}

// InviteTokenResult holds the output of a successful invite token generation.
type InviteTokenResult struct {
	Token            string
	Email            string
	LastInviteSentAt *time.Time
}

// UpdateLastInviteSentAt sets the last_invite_sent_at timestamp for a user.
func (r *UserRepository) UpdateLastInviteSentAt(ctx context.Context, userID string, now time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE core.users
		SET last_invite_sent_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`, userID, now)
	return err
}

// GenerateInviteToken runs a transactional flow that locks the user row,
// validates eligibility, generates an invite token (via the provided callback),
// and persists the token hash. The genTokenFn callback should return a raw
// token string and its hash; it may be retried up to 3 times on failure.
func (r *UserRepository) GenerateInviteToken(
	ctx context.Context,
	userID string,
	now, expiresAt time.Time,
	setLastSent bool,
	genTokenFn func() (raw string, hash string, err error),
) (*InviteTokenResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Lock and read the user row.
	var email *string
	var passwordHash *string
	var lastInviteSentAt pgtype.Timestamptz
	err = tx.QueryRow(ctx, `
		SELECT email, password_hash, last_invite_sent_at
		FROM core.users
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, userID).Scan(&email, &passwordHash, &lastInviteSentAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "user", ID: userID}
		}
		return nil, err
	}

	if email == nil || strings.TrimSpace(*email) == "" {
		return nil, &apperrors.InvalidArgumentError{
			Field:   "email",
			Message: "user has no email set; assign an email first",
		}
	}
	if passwordHash != nil && strings.TrimSpace(*passwordHash) != "" {
		return nil, &apperrors.InvalidArgumentError{
			Field:   "id",
			Message: "user already has a password set",
		}
	}

	result := &InviteTokenResult{Email: *email}
	if lastInviteSentAt.Valid {
		t := lastInviteSentAt.Time.UTC()
		result.LastInviteSentAt = &t
	}

	// Try up to 3 times to generate a unique token.
	var lastErr error
	for i := 0; i < 3; i++ {
		raw, hash, genErr := genTokenFn()
		if genErr != nil {
			return nil, genErr
		}

		ct, execErr := tx.Exec(ctx, buildInviteUpdateSQL(setLastSent),
			userID, hash, expiresAt, now)
		if execErr != nil {
			lastErr = execErr
			continue
		}
		if ct.RowsAffected() == 0 {
			return nil, &apperrors.NotFoundError{Resource: "user", ID: userID}
		}
		result.Token = raw
		break
	}
	if result.Token == "" {
		return nil, fmt.Errorf("failed to generate invite token after 3 attempts: %w", lastErr)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func buildInviteUpdateSQL(setLastSent bool) string {
	if setLastSent {
		return `
			UPDATE core.users
			SET
			  status = 'requires_password_setup',
			  invite_token_hash = $2,
			  invite_expires_at = $3,
			  invite_consumed_at = NULL,
			  invited_at = COALESCE(invited_at, $4),
			  last_invite_sent_at = $4,
			  updated_at = $4
			WHERE id = $1 AND deleted_at IS NULL`
	}
	return `
		UPDATE core.users
		SET
		  status = 'requires_password_setup',
		  invite_token_hash = $2,
		  invite_expires_at = $3,
		  invite_consumed_at = NULL,
		  invited_at = COALESCE(invited_at, $4),
		  updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL`
}

// GetAdminUserByID returns a single user with their aggregated labels and
// permissions, using the same CTE logic as ListAdminUsers.
func (r *UserRepository) GetAdminUserByID(ctx context.Context, userID string) (*AdminUserRow, error) {
	row := r.pool.QueryRow(ctx, `
		WITH active_grants AS (
			SELECT g.id, g.user_id, g.scope_type, g.scope_id, g.label_id, g.permission_id, g.created_at, g.updated_at, g.expires_at, g.revoked_at, g.deleted_at
			FROM core.grants g
			WHERE g.revoked_at IS NULL
			  AND g.scope_type = 'global'
			  AND (g.expires_at IS NULL OR g.expires_at > NOW())
		),
		user_labels AS (
			SELECT g.user_id, l.key
			FROM active_grants g
			JOIN core.labels l ON g.label_id = l.id
			WHERE l.deleted_at IS NULL
		),
		user_permissions AS (
			SELECT g.user_id, p.key
			FROM active_grants g
			JOIN core.permissions p ON g.permission_id = p.id
			WHERE p.deleted_at IS NULL
			UNION
			SELECT g.user_id, p2.key
			FROM active_grants g
			JOIN core.labels l ON g.label_id = l.id AND l.deleted_at IS NULL
			JOIN core.label_permissions lp ON lp.label_id = l.id
			JOIN core.permissions p2 ON lp.permission_id = p2.id AND p2.deleted_at IS NULL
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
		FROM core.users u
		LEFT JOIN user_labels ul ON ul.user_id = u.id
		LEFT JOIN user_permissions up ON up.user_id = u.id
		WHERE u.deleted_at IS NULL
		  AND u.id = $1
		GROUP BY u.id, u.email, u.first_name, u.last_name, u.status, u.invited_at, u.last_invite_sent_at, u.invite_expires_at, u.invite_consumed_at, u.created_at, u.updated_at
	`, userID)

	var item AdminUserRow
	var labels []string
	var perms []string
	var invitedAt pgtype.Timestamptz
	var lastInviteSentAt pgtype.Timestamptz
	var inviteExpiresAt pgtype.Timestamptz
	var inviteConsumedAt pgtype.Timestamptz

	if err := row.Scan(
		&item.ID,
		&item.Email,
		&item.FirstName,
		&item.LastName,
		&item.Status,
		&invitedAt,
		&lastInviteSentAt,
		&inviteExpiresAt,
		&inviteConsumedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&labels,
		&perms,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	item.InvitedAt = TimestamptzToPtrTimeUTC(invitedAt)
	item.LastInviteSentAt = TimestamptzToPtrTimeUTC(lastInviteSentAt)
	item.InviteExpiresAt = TimestamptzToPtrTimeUTC(inviteExpiresAt)
	item.InviteConsumedAt = TimestamptzToPtrTimeUTC(inviteConsumedAt)

	sort.Strings(labels)
	sort.Strings(perms)
	item.Labels = labels
	item.Permissions = perms

	return &item, nil
}

// ListAdminUsers returns all non-deleted users with their aggregated labels
// and permissions, optionally filtered by status. Results are ordered by
// created_at descending.
func (r *UserRepository) ListAdminUsers(ctx context.Context, statusFilter string) ([]AdminUserRow, error) {
	rows, err := r.pool.Query(ctx, `
		WITH active_grants AS (
			SELECT g.id, g.user_id, g.scope_type, g.scope_id, g.label_id, g.permission_id, g.created_at, g.updated_at, g.expires_at, g.revoked_at, g.deleted_at
			FROM core.grants g
			WHERE g.revoked_at IS NULL
			  AND g.scope_type = 'global'
			  AND (g.expires_at IS NULL OR g.expires_at > NOW())
		),
		user_labels AS (
			SELECT g.user_id, l.key
			FROM active_grants g
			JOIN core.labels l ON g.label_id = l.id
			WHERE l.deleted_at IS NULL
		),
		user_permissions AS (
			SELECT g.user_id, p.key
			FROM active_grants g
			JOIN core.permissions p ON g.permission_id = p.id
			WHERE p.deleted_at IS NULL
			UNION
			SELECT g.user_id, p2.key
			FROM active_grants g
			JOIN core.labels l ON g.label_id = l.id AND l.deleted_at IS NULL
			JOIN core.label_permissions lp ON lp.label_id = l.id
			JOIN core.permissions p2 ON lp.permission_id = p2.id AND p2.deleted_at IS NULL
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
		FROM core.users u
		LEFT JOIN user_labels ul ON ul.user_id = u.id
		LEFT JOIN user_permissions up ON up.user_id = u.id
		WHERE u.deleted_at IS NULL
		  AND ($1 = '' OR u.status = $1)
		GROUP BY u.id, u.email, u.first_name, u.last_name, u.status, u.invited_at, u.last_invite_sent_at, u.invite_expires_at, u.invite_consumed_at, u.created_at, u.updated_at
		ORDER BY u.created_at DESC
	`, statusFilter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AdminUserRow
	for rows.Next() {
		var item AdminUserRow
		var labels []string
		var perms []string

		var invitedAt pgtype.Timestamptz
		var lastInviteSentAt pgtype.Timestamptz
		var inviteExpiresAt pgtype.Timestamptz
		var inviteConsumedAt pgtype.Timestamptz

		if err := rows.Scan(
			&item.ID,
			&item.Email,
			&item.FirstName,
			&item.LastName,
			&item.Status,
			&invitedAt,
			&lastInviteSentAt,
			&inviteExpiresAt,
			&inviteConsumedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
			&labels,
			&perms,
		); err != nil {
			return nil, err
		}

		item.InvitedAt = TimestamptzToPtrTimeUTC(invitedAt)
		item.LastInviteSentAt = TimestamptzToPtrTimeUTC(lastInviteSentAt)
		item.InviteExpiresAt = TimestamptzToPtrTimeUTC(inviteExpiresAt)
		item.InviteConsumedAt = TimestamptzToPtrTimeUTC(inviteConsumedAt)

		sort.Strings(labels)
		sort.Strings(perms)
		item.Labels = labels
		item.Permissions = perms

		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
