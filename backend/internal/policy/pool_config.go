package policy

import (
	"context"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// CanManagePool checks if a user can modify pool settings, payouts, etc.
func CanManagePool(
	ctx context.Context,
	authz AuthorizationChecker,
	userID string,
	pool *models.Pool,
) (Decision, error) {
	if userID == "" {
		return Decision{Allowed: false, Status: http.StatusUnauthorized, Code: "unauthorized", Message: "Authentication required"}, nil
	}
	if pool == nil {
		return Decision{Allowed: false, Status: http.StatusBadRequest, Code: "pool_missing", Message: "Pool not found"}, nil
	}

	isAdmin, err := isPoolAdminOrOwner(ctx, authz, userID, pool)
	if err != nil {
		return Decision{}, err
	}

	if !isAdmin {
		return Decision{Allowed: false, IsAdmin: false, Status: http.StatusForbidden, Code: "forbidden", Message: "Insufficient permissions"}, nil
	}

	return Decision{Allowed: true, IsAdmin: isAdmin}, nil
}
