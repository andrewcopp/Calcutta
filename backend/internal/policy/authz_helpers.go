package policy

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// isPoolAdminOrOwner checks if a user has pool management authority.
// Uses a single HasPermission call that resolves both global grants (site_admin)
// and pool-scoped grants (pool_admin). Falls back to owner_id check
// for backwards compatibility.
func isPoolAdminOrOwner(ctx context.Context, authz AuthorizationChecker, userID string, pool *models.Pool) (bool, error) {
	if pool != nil && pool.OwnerID == userID {
		return true, nil
	}

	if authz != nil {
		ok, err := authz.HasPermission(ctx, userID, "pool", pool.ID, permissionAdminOverride)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	return false, nil
}
