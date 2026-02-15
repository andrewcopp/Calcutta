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
WHERE l.key = 'global_admin'
  AND l.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM core.grants g
    JOIN core.labels l2 ON g.label_id = l2.id
    WHERE g.user_id = $1
      AND g.scope_type = 'global'
      AND g.revoked_at IS NULL
      AND l2.key = 'global_admin'
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
