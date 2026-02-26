package policy

import (
	"context"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// CanViewPool checks if a user can view a pool's dashboard and portfolios.
// Public and unlisted pools are viewable by anyone with the link.
// Private pools require ownership, participation, or a pool.read permission.
func CanViewPool(
	ctx context.Context,
	authz AuthorizationChecker,
	userID string,
	pool *models.Pool,
	participantUserIDs []string,
) (Decision, error) {
	if pool == nil {
		return Decision{Allowed: false, Status: http.StatusBadRequest, Code: "pool_missing", Message: "Pool not found"}, nil
	}

	if pool.Visibility == "public" || pool.Visibility == "unlisted" {
		return Decision{Allowed: true}, nil
	}

	// Private pool: require authentication
	if userID == "" {
		return Decision{Allowed: false, Status: http.StatusUnauthorized, Code: "unauthorized", Message: "Authentication required"}, nil
	}

	if pool.OwnerID == userID {
		return Decision{Allowed: true}, nil
	}

	for _, pid := range participantUserIDs {
		if pid == userID {
			return Decision{Allowed: true}, nil
		}
	}

	if authz != nil {
		ok, err := authz.HasPermission(ctx, userID, "pool", pool.ID, "pool.read")
		if err != nil {
			return Decision{}, err
		}
		if ok {
			return Decision{Allowed: true}, nil
		}
	}

	return Decision{Allowed: false, Status: http.StatusForbidden, Code: "forbidden", Message: "Insufficient permissions"}, nil
}
