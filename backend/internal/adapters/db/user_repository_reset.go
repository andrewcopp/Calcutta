package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ResetTokenResult holds the output of a successful reset token generation.
type ResetTokenResult struct {
	Token          string
	Email          string
	LastResetSentAt *time.Time
}

// GenerateResetToken runs a transactional flow that locks the user row,
// validates eligibility, generates a reset token, and persists the token hash.
func (r *UserRepository) GenerateResetToken(
	ctx context.Context,
	userID string,
	now, expiresAt time.Time,
	genTokenFn func() (raw string, hash string, err error),
) (*ResetTokenResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning reset token transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var email *string
	var status string
	var lastResetSentAt pgtype.Timestamptz
	err = tx.QueryRow(ctx, `
		SELECT email, status, last_reset_sent_at
		FROM core.users
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, userID).Scan(&email, &status, &lastResetSentAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "user", ID: userID}
		}
		return nil, fmt.Errorf("locking user row for reset token %s: %w", userID, err)
	}

	if email == nil || strings.TrimSpace(*email) == "" {
		return nil, &apperrors.InvalidArgumentError{
			Field:   "email",
			Message: "user has no email set",
		}
	}
	if status != "active" {
		return nil, &apperrors.InvalidArgumentError{
			Field:   "status",
			Message: "can only reset password for active users",
		}
	}

	result := &ResetTokenResult{Email: *email}
	if lastResetSentAt.Valid {
		t := lastResetSentAt.Time.UTC()
		result.LastResetSentAt = &t
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		raw, hash, genErr := genTokenFn()
		if genErr != nil {
			return nil, genErr
		}

		ct, execErr := tx.Exec(ctx, `
			UPDATE core.users
			SET
			  reset_token_hash = $2,
			  reset_expires_at = $3,
			  reset_consumed_at = NULL,
			  last_reset_sent_at = $4,
			  updated_at = $4
			WHERE id = $1 AND deleted_at IS NULL
		`, userID, hash, expiresAt, now)
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
		return nil, fmt.Errorf("failed to generate reset token after 3 attempts: %w", lastErr)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing reset token transaction: %w", err)
	}
	return result, nil
}

// GenerateResetTokenByEmail looks up a user by email and generates a reset token.
// Returns (nil, nil) if no matching active user is found (prevents email enumeration).
func (r *UserRepository) GenerateResetTokenByEmail(
	ctx context.Context,
	email string,
	now, expiresAt time.Time,
	genTokenFn func() (raw string, hash string, err error),
) (*ResetTokenResult, error) {
	var userID string
	err := r.pool.QueryRow(ctx, `
		SELECT id::text
		FROM core.users
		WHERE email = $1 AND status = 'active' AND deleted_at IS NULL
	`, strings.TrimSpace(email)).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("looking up user by email for reset: %w", err)
	}

	return r.GenerateResetToken(ctx, userID, now, expiresAt, genTokenFn)
}
