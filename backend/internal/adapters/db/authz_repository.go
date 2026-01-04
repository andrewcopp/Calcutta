package db

import (
	"context"
	"errors"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const grantGlobalAdminSQL = `
INSERT INTO grants (user_id, scope_type, scope_id, label_id)
SELECT $1, 'global', NULL, l.id
FROM labels l
WHERE l.key = 'global_admin'
  AND l.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM grants g
    JOIN labels l2 ON g.label_id = l2.id
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
