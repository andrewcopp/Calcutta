package db

import (
	"context"
	"errors"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const grantGlobalAdminSQL = `
INSERT INTO core.grants (user_id, scope_type, scope_id, label_id)
SELECT $1, 'global', NULL, l.id
FROM core.labels l
WHERE l.key = 'site_admin'
  AND l.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM core.grants g
    JOIN core.labels l2 ON g.label_id = l2.id
    WHERE g.user_id = $1
      AND g.scope_type = 'global'
      AND g.revoked_at IS NULL
      AND l2.key = 'site_admin'
      AND l2.deleted_at IS NULL
  );
`

type AuthorizationRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewAuthorizationRepository(pool *pgxpool.Pool) *AuthorizationRepository {
	return &AuthorizationRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *AuthorizationRepository) GrantGlobalAdmin(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, grantGlobalAdminSQL, userID)
	return err
}

func (r *AuthorizationRepository) HasPermission(ctx context.Context, userID, scopeType, scopeID, permissionKey string) (bool, error) {
	_, err := r.q.HasPermission(ctx, sqlc.HasPermissionParams{
		UserID:    userID,
		ScopeType: scopeType,
		Column3:   scopeID,
		Key:       permissionKey,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GrantLabel grants a label to a user with the given scope.
func (r *AuthorizationRepository) GrantLabel(ctx context.Context, userID, labelKey, scopeType, scopeID string) error {
	query := `
		INSERT INTO core.grants (user_id, scope_type, scope_id, label_id)
		SELECT $1, $2, $3::uuid, l.id
		FROM core.labels l
		WHERE l.key = $4
		  AND l.deleted_at IS NULL
		  AND NOT EXISTS (
			SELECT 1
			FROM core.grants g
			WHERE g.user_id = $1
			  AND g.scope_type = $2
			  AND g.scope_id = $3::uuid
			  AND g.label_id = l.id
			  AND g.revoked_at IS NULL
		  )
	`
	_, err := r.pool.Exec(ctx, query, userID, scopeType, scopeID, labelKey)
	return err
}

// RevokeGrant revokes all grants for a user+label+scope combination.
func (r *AuthorizationRepository) RevokeGrant(ctx context.Context, userID, labelKey, scopeType, scopeID string) error {
	query := `
		UPDATE core.grants g
		SET revoked_at = NOW()
		FROM core.labels l
		WHERE g.label_id = l.id
		  AND g.user_id = $1
		  AND g.scope_type = $2
		  AND g.scope_id = $3::uuid
		  AND l.key = $4
		  AND l.deleted_at IS NULL
		  AND g.revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, userID, scopeType, scopeID, labelKey)
	return err
}

// ListUserGlobalPermissions returns all permission keys a user has via global grants.
func (r *AuthorizationRepository) ListUserGlobalPermissions(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT DISTINCT COALESCE(p_direct.key, p_label.key) AS permission_key
		FROM core.grants g
		LEFT JOIN core.permissions p_direct ON g.permission_id = p_direct.id AND p_direct.deleted_at IS NULL
		LEFT JOIN core.labels l ON g.label_id = l.id AND l.deleted_at IS NULL
		LEFT JOIN core.label_permissions lp ON lp.label_id = l.id AND lp.deleted_at IS NULL
		LEFT JOIN core.permissions p_label ON lp.permission_id = p_label.id AND p_label.deleted_at IS NULL
		WHERE g.user_id = $1
		  AND g.deleted_at IS NULL
		  AND g.revoked_at IS NULL
		  AND (g.expires_at IS NULL OR g.expires_at > NOW())
		  AND g.scope_type = 'global'
		  AND (p_direct.key IS NOT NULL OR p_label.key IS NOT NULL)
		ORDER BY permission_key
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		permissions = append(permissions, key)
	}
	return permissions, rows.Err()
}

// ListUserGlobalLabels returns all label keys a user has via active global grants.
func (r *AuthorizationRepository) ListUserGlobalLabels(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT DISTINCT l.key
		FROM core.grants g
		JOIN core.labels l ON g.label_id = l.id AND l.deleted_at IS NULL
		WHERE g.user_id = $1
		  AND g.deleted_at IS NULL
		  AND g.revoked_at IS NULL
		  AND (g.expires_at IS NULL OR g.expires_at > NOW())
		  AND g.scope_type = 'global'
		ORDER BY l.key
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		labels = append(labels, key)
	}
	return labels, rows.Err()
}

// GrantGlobalLabel grants a global label to a user.
func (r *AuthorizationRepository) GrantGlobalLabel(ctx context.Context, userID, labelKey string) error {
	query := `
		INSERT INTO core.grants (user_id, scope_type, scope_id, label_id)
		SELECT $1, 'global', NULL, l.id
		FROM core.labels l
		WHERE l.key = $2
		  AND l.deleted_at IS NULL
		  AND NOT EXISTS (
			SELECT 1
			FROM core.grants g
			JOIN core.labels l2 ON g.label_id = l2.id
			WHERE g.user_id = $1
			  AND g.scope_type = 'global'
			  AND g.revoked_at IS NULL
			  AND l2.key = $2
			  AND l2.deleted_at IS NULL
		  )
	`
	_, err := r.pool.Exec(ctx, query, userID, labelKey)
	return err
}

// RevokeGlobalLabel revokes a global label from a user.
func (r *AuthorizationRepository) RevokeGlobalLabel(ctx context.Context, userID, labelKey string) error {
	query := `
		UPDATE core.grants g
		SET revoked_at = NOW()
		FROM core.labels l
		WHERE g.label_id = l.id
		  AND g.user_id = $1
		  AND g.scope_type = 'global'
		  AND g.scope_id IS NULL
		  AND l.key = $2
		  AND l.deleted_at IS NULL
		  AND g.revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, userID, labelKey)
	return err
}

// ListGrantsByScope returns user IDs with a given label for a scope.
func (r *AuthorizationRepository) ListGrantsByScope(ctx context.Context, labelKey, scopeType, scopeID string) ([]string, error) {
	query := `
		SELECT g.user_id::text
		FROM core.grants g
		JOIN core.labels l ON g.label_id = l.id
		WHERE g.scope_type = $1
		  AND g.scope_id = $2::uuid
		  AND l.key = $3
		  AND l.deleted_at IS NULL
		  AND g.revoked_at IS NULL
		  AND (g.expires_at IS NULL OR g.expires_at > NOW())
	`
	rows, err := r.pool.Query(ctx, query, scopeType, scopeID, labelKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, uid)
	}
	return userIDs, rows.Err()
}
