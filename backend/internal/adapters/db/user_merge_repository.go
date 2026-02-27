package db

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.UserMergeRepository = (*UserMergeRepository)(nil)

type UserMergeRepository struct {
	pool *pgxpool.Pool
}

func NewUserMergeRepository(pool *pgxpool.Pool) *UserMergeRepository {
	return &UserMergeRepository{pool: pool}
}

func (r *UserMergeRepository) MergeUsers(ctx context.Context, sourceUserID, targetUserID, mergedBy string) (*models.UserMerge, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin merge transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock both user rows
	var sourceStatus string
	err = tx.QueryRow(ctx, `SELECT status FROM core.users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, sourceUserID).Scan(&sourceStatus)
	if err != nil {
		return nil, &apperrors.NotFoundError{Resource: "source user", ID: sourceUserID}
	}
	var targetStatus string
	err = tx.QueryRow(ctx, `SELECT status FROM core.users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, targetUserID).Scan(&targetStatus)
	if err != nil {
		return nil, &apperrors.NotFoundError{Resource: "target user", ID: targetUserID}
	}

	// Move portfolios (skip if target already has portfolio in the same pool)
	tag, err := tx.Exec(ctx, `
		UPDATE core.portfolios SET user_id = $2, updated_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
		  AND pool_id NOT IN (
			SELECT pool_id FROM core.portfolios WHERE user_id = $2 AND deleted_at IS NULL
		  )
	`, sourceUserID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("move portfolios: %w", err)
	}
	entriesMoved := int(tag.RowsAffected())

	// Move invitations (skip if target already invited to same pool)
	tag, err = tx.Exec(ctx, `
		UPDATE core.pool_invitations SET user_id = $2, updated_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
		  AND pool_id NOT IN (
			SELECT pool_id FROM core.pool_invitations WHERE user_id = $2 AND deleted_at IS NULL
		  )
	`, sourceUserID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("move invitations: %w", err)
	}
	invitationsMoved := int(tag.RowsAffected())

	// Move grants
	tag, err = tx.Exec(ctx, `
		UPDATE core.grants SET user_id = $2, updated_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
	`, sourceUserID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("move grants: %w", err)
	}
	grantsMoved := int(tag.RowsAffected())

	// Transfer pool ownership
	_, err = tx.Exec(ctx, `
		UPDATE core.pools SET owner_id = $2, updated_at = NOW()
		WHERE owner_id = $1 AND deleted_at IS NULL
	`, sourceUserID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("move pool ownership: %w", err)
	}

	// Soft-delete source user
	_, err = tx.Exec(ctx, `
		UPDATE core.users SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, sourceUserID)
	if err != nil {
		return nil, fmt.Errorf("soft-delete source user: %w", err)
	}

	// Record the merge
	merge := &models.UserMerge{
		SourceUserID:     sourceUserID,
		TargetUserID:     targetUserID,
		MergedBy:         mergedBy,
		EntriesMoved:     entriesMoved,
		InvitationsMoved: invitationsMoved,
		GrantsMoved:      grantsMoved,
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO core.user_merges (source_user_id, target_user_id, merged_by, entries_moved, invitations_moved, grants_moved)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, sourceUserID, targetUserID, mergedBy, entriesMoved, invitationsMoved, grantsMoved).Scan(&merge.ID, &merge.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert user_merge record: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit merge transaction: %w", err)
	}
	committed = true

	return merge, nil
}

func (r *UserMergeRepository) BatchMergeUsers(ctx context.Context, sourceUserIDs []string, targetUserID, mergedBy string) ([]*models.UserMerge, error) {
	if len(sourceUserIDs) == 0 {
		return nil, &apperrors.InvalidArgumentError{Field: "sourceUserIds", Message: "at least one source user is required"}
	}
	for _, id := range sourceUserIDs {
		if id == targetUserID {
			return nil, &apperrors.InvalidArgumentError{Field: "sourceUserIds", Message: "target user cannot be in source list"}
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin batch merge transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock target row first
	var targetStatus string
	err = tx.QueryRow(ctx, `SELECT status FROM core.users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, targetUserID).Scan(&targetStatus)
	if err != nil {
		return nil, &apperrors.NotFoundError{Resource: "target user", ID: targetUserID}
	}

	// Lock source rows in deterministic order to prevent deadlocks
	sorted := make([]string, len(sourceUserIDs))
	copy(sorted, sourceUserIDs)
	sort.Strings(sorted)

	for _, sid := range sorted {
		var s string
		err = tx.QueryRow(ctx, `SELECT status FROM core.users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, sid).Scan(&s)
		if err != nil {
			return nil, &apperrors.NotFoundError{Resource: "source user", ID: sid}
		}
	}

	var merges []*models.UserMerge
	for _, sourceID := range sorted {
		// Move portfolios
		tag, err := tx.Exec(ctx, `
			UPDATE core.portfolios SET user_id = $2, updated_at = NOW()
			WHERE user_id = $1 AND deleted_at IS NULL
			  AND pool_id NOT IN (
				SELECT pool_id FROM core.portfolios WHERE user_id = $2 AND deleted_at IS NULL
			  )
		`, sourceID, targetUserID)
		if err != nil {
			return nil, fmt.Errorf("move portfolios for %s: %w", sourceID, err)
		}
		entriesMoved := int(tag.RowsAffected())

		// Move invitations
		tag, err = tx.Exec(ctx, `
			UPDATE core.pool_invitations SET user_id = $2, updated_at = NOW()
			WHERE user_id = $1 AND deleted_at IS NULL
			  AND pool_id NOT IN (
				SELECT pool_id FROM core.pool_invitations WHERE user_id = $2 AND deleted_at IS NULL
			  )
		`, sourceID, targetUserID)
		if err != nil {
			return nil, fmt.Errorf("move invitations for %s: %w", sourceID, err)
		}
		invitationsMoved := int(tag.RowsAffected())

		// Move grants
		tag, err = tx.Exec(ctx, `
			UPDATE core.grants SET user_id = $2, updated_at = NOW()
			WHERE user_id = $1 AND deleted_at IS NULL
		`, sourceID, targetUserID)
		if err != nil {
			return nil, fmt.Errorf("move grants for %s: %w", sourceID, err)
		}
		grantsMoved := int(tag.RowsAffected())

		// Transfer pool ownership
		_, err = tx.Exec(ctx, `
			UPDATE core.pools SET owner_id = $2, updated_at = NOW()
			WHERE owner_id = $1 AND deleted_at IS NULL
		`, sourceID, targetUserID)
		if err != nil {
			return nil, fmt.Errorf("move pool ownership for %s: %w", sourceID, err)
		}

		// Soft-delete source user
		_, err = tx.Exec(ctx, `
			UPDATE core.users SET deleted_at = NOW(), updated_at = NOW()
			WHERE id = $1
		`, sourceID)
		if err != nil {
			return nil, fmt.Errorf("soft-delete source user %s: %w", sourceID, err)
		}

		// Record the merge
		merge := &models.UserMerge{
			SourceUserID:     sourceID,
			TargetUserID:     targetUserID,
			MergedBy:         mergedBy,
			EntriesMoved:     entriesMoved,
			InvitationsMoved: invitationsMoved,
			GrantsMoved:      grantsMoved,
		}
		err = tx.QueryRow(ctx, `
			INSERT INTO core.user_merges (source_user_id, target_user_id, merged_by, entries_moved, invitations_moved, grants_moved)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at
		`, sourceID, targetUserID, mergedBy, entriesMoved, invitationsMoved, grantsMoved).Scan(&merge.ID, &merge.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert user_merge record for %s: %w", sourceID, err)
		}
		merges = append(merges, merge)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit batch merge transaction: %w", err)
	}
	committed = true

	return merges, nil
}

func (r *UserMergeRepository) ListStubUsers(ctx context.Context) ([]*models.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, email, first_name, last_name, status, external_provider, created_at, updated_at
		FROM core.users
		WHERE status = 'stub' AND deleted_at IS NULL
		ORDER BY last_name, first_name
	`)
	if err != nil {
		return nil, fmt.Errorf("list stub users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.Status, &u.ExternalProvider, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan stub user: %w", err)
		}
		u.CreatedAt = createdAt
		u.UpdatedAt = updatedAt
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserMergeRepository) FindMergeCandidates(ctx context.Context, userID string) ([]*models.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u2.id, u2.email, u2.first_name, u2.last_name, u2.status, u2.external_provider, u2.created_at, u2.updated_at
		FROM core.users u1
		JOIN core.users u2 ON u2.first_name = u1.first_name AND u2.last_name = u1.last_name
		WHERE u1.id = $1 AND u2.id != $1 AND u2.deleted_at IS NULL
		ORDER BY u2.status, u2.last_name, u2.first_name
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("find merge candidates: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.Status, &u.ExternalProvider, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan merge candidate: %w", err)
		}
		u.CreatedAt = createdAt
		u.UpdatedAt = updatedAt
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserMergeRepository) ListMergeHistory(ctx context.Context, userID string) ([]*models.UserMerge, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, source_user_id, target_user_id, merged_by, entries_moved, invitations_moved, grants_moved, created_at
		FROM core.user_merges
		WHERE source_user_id = $1 OR target_user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list merge history: %w", err)
	}
	defer rows.Close()

	var merges []*models.UserMerge
	for rows.Next() {
		m := &models.UserMerge{}
		if err := rows.Scan(&m.ID, &m.SourceUserID, &m.TargetUserID, &m.MergedBy, &m.EntriesMoved, &m.InvitationsMoved, &m.GrantsMoved, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan merge history: %w", err)
		}
		merges = append(merges, m)
	}
	return merges, rows.Err()
}
