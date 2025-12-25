package db

import (
	"context"
	"errors"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthorizationRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewAuthorizationRepository(pool *pgxpool.Pool) *AuthorizationRepository {
	return &AuthorizationRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *AuthorizationRepository) EnsureBootstrapAdmin(ctx context.Context, userID string) error {
	return r.q.EnsureBootstrapGlobalAdmin(ctx, userID)
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
