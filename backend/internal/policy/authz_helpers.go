package policy

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// isCalcuttaAdminOrOwner checks if a user has calcutta management authority.
// Uses a single HasPermission call that resolves both global grants (site_admin)
// and calcutta-scoped grants (calcutta_admin). Falls back to owner_id check
// for backwards compatibility.
func isCalcuttaAdminOrOwner(ctx context.Context, authz AuthorizationChecker, userID string, calcutta *models.Calcutta) (bool, error) {
	if calcutta != nil && calcutta.OwnerID == userID {
		return true, nil
	}

	if authz != nil {
		ok, err := authz.HasPermission(ctx, userID, "calcutta", calcutta.ID, permissionAdminOverride)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	return false, nil
}
