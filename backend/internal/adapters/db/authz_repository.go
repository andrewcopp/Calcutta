package db

import (
	"context"
	"errors"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const grantGlobalAdminSQL = `
INSERT INTO core.grants (user_id, scope_type, scope_id, role_id)
SELECT $1, 'global', NULL, r.id
FROM core.roles r
WHERE r.key = 'site_admin'
  AND r.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM core.grants g
    JOIN core.roles r2 ON g.role_id = r2.id
    WHERE g.user_id = $1
      AND g.scope_type = 'global'
      AND g.revoked_at IS NULL
      AND r2.key = 'site_admin'
      AND r2.deleted_at IS NULL
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

// GrantRole grants a role to a user with the given scope.
func (r *AuthorizationRepository) GrantRole(ctx context.Context, userID, roleKey, scopeType, scopeID string) error {
	query := `
		INSERT INTO core.grants (user_id, scope_type, scope_id, role_id)
		SELECT $1, $2, $3::uuid, r.id
		FROM core.roles r
		WHERE r.key = $4
		  AND r.deleted_at IS NULL
		  AND NOT EXISTS (
			SELECT 1
			FROM core.grants g
			WHERE g.user_id = $1
			  AND g.scope_type = $2
			  AND g.scope_id = $3::uuid
			  AND g.role_id = r.id
			  AND g.revoked_at IS NULL
		  )
	`
	_, err := r.pool.Exec(ctx, query, userID, scopeType, scopeID, roleKey)
	return err
}

// RevokeGrant revokes all grants for a user+role+scope combination.
func (r *AuthorizationRepository) RevokeGrant(ctx context.Context, userID, roleKey, scopeType, scopeID string) error {
	query := `
		UPDATE core.grants g
		SET revoked_at = NOW()
		FROM core.roles r
		WHERE g.role_id = r.id
		  AND g.user_id = $1
		  AND g.scope_type = $2
		  AND g.scope_id = $3::uuid
		  AND r.key = $4
		  AND r.deleted_at IS NULL
		  AND g.revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, userID, scopeType, scopeID, roleKey)
	return err
}

// ListUserGlobalPermissions returns all permission keys a user has via global grants.
func (r *AuthorizationRepository) ListUserGlobalPermissions(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT DISTINCT COALESCE(p_direct.key, p_role.key) AS permission_key
		FROM core.grants g
		LEFT JOIN core.permissions p_direct ON g.permission_id = p_direct.id AND p_direct.deleted_at IS NULL
		LEFT JOIN core.roles r ON g.role_id = r.id AND r.deleted_at IS NULL
		LEFT JOIN core.role_permissions rp ON rp.role_id = r.id AND rp.deleted_at IS NULL
		LEFT JOIN core.permissions p_role ON rp.permission_id = p_role.id AND p_role.deleted_at IS NULL
		WHERE g.user_id = $1
		  AND g.deleted_at IS NULL
		  AND g.revoked_at IS NULL
		  AND (g.expires_at IS NULL OR g.expires_at > NOW())
		  AND g.scope_type = 'global'
		  AND (p_direct.key IS NOT NULL OR p_role.key IS NOT NULL)
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

// ListUserGlobalRoles returns all role keys a user has via active global grants.
func (r *AuthorizationRepository) ListUserGlobalRoles(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT DISTINCT r.key
		FROM core.grants g
		JOIN core.roles r ON g.role_id = r.id AND r.deleted_at IS NULL
		WHERE g.user_id = $1
		  AND g.deleted_at IS NULL
		  AND g.revoked_at IS NULL
		  AND (g.expires_at IS NULL OR g.expires_at > NOW())
		  AND g.scope_type = 'global'
		ORDER BY r.key
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		roles = append(roles, key)
	}
	return roles, rows.Err()
}

// GrantGlobalRole grants a global role to a user.
func (r *AuthorizationRepository) GrantGlobalRole(ctx context.Context, userID, roleKey string) error {
	query := `
		INSERT INTO core.grants (user_id, scope_type, scope_id, role_id)
		SELECT $1, 'global', NULL, r.id
		FROM core.roles r
		WHERE r.key = $2
		  AND r.deleted_at IS NULL
		  AND NOT EXISTS (
			SELECT 1
			FROM core.grants g
			JOIN core.roles r2 ON g.role_id = r2.id
			WHERE g.user_id = $1
			  AND g.scope_type = 'global'
			  AND g.revoked_at IS NULL
			  AND r2.key = $2
			  AND r2.deleted_at IS NULL
		  )
	`
	_, err := r.pool.Exec(ctx, query, userID, roleKey)
	return err
}

// RevokeGlobalRole revokes a global role from a user.
func (r *AuthorizationRepository) RevokeGlobalRole(ctx context.Context, userID, roleKey string) error {
	query := `
		UPDATE core.grants g
		SET revoked_at = NOW()
		FROM core.roles r
		WHERE g.role_id = r.id
		  AND g.user_id = $1
		  AND g.scope_type = 'global'
		  AND g.scope_id IS NULL
		  AND r.key = $2
		  AND r.deleted_at IS NULL
		  AND g.revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, userID, roleKey)
	return err
}

// RoleGrantRow represents a role grant with scope info and display name.
type RoleGrantRow struct {
	Key       string
	ScopeType string
	ScopeID   *string
	ScopeName *string
}

// ListUserRolesWithScope returns all active role grants for a user with scope details.
func (r *AuthorizationRepository) ListUserRolesWithScope(ctx context.Context, userID string) ([]RoleGrantRow, error) {
	query := `
		SELECT DISTINCT r.key,
			g.scope_type,
			g.scope_id::text,
			COALESCE(c.name, comp.name || ' ' || s.year) AS scope_name
		FROM core.grants g
		JOIN core.roles r ON g.role_id = r.id AND r.deleted_at IS NULL
		LEFT JOIN core.calcuttas c ON g.scope_type = 'calcutta' AND g.scope_id = c.id
		LEFT JOIN core.tournaments t ON g.scope_type = 'tournament' AND g.scope_id = t.id
		LEFT JOIN core.competitions comp ON t.competition_id = comp.id
		LEFT JOIN core.seasons s ON t.season_id = s.id
		WHERE g.user_id = $1
		  AND g.deleted_at IS NULL
		  AND g.revoked_at IS NULL
		  AND (g.expires_at IS NULL OR g.expires_at > NOW())
		ORDER BY r.key, g.scope_type
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []RoleGrantRow
	for rows.Next() {
		var row RoleGrantRow
		if err := rows.Scan(&row.Key, &row.ScopeType, &row.ScopeID, &row.ScopeName); err != nil {
			return nil, err
		}
		grants = append(grants, row)
	}
	return grants, rows.Err()
}

// ListGrantsByScope returns user IDs with a given role for a scope.
func (r *AuthorizationRepository) ListGrantsByScope(ctx context.Context, roleKey, scopeType, scopeID string) ([]string, error) {
	query := `
		SELECT g.user_id::text
		FROM core.grants g
		JOIN core.roles r ON g.role_id = r.id
		WHERE g.scope_type = $1
		  AND g.scope_id = $2::uuid
		  AND r.key = $3
		  AND r.deleted_at IS NULL
		  AND g.revoked_at IS NULL
		  AND (g.expires_at IS NULL OR g.expires_at > NOW())
	`
	rows, err := r.pool.Query(ctx, query, scopeType, scopeID, roleKey)
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
