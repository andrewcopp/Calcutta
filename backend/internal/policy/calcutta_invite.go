package policy

import (
	"context"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func CanInviteToCalcutta(
	ctx context.Context,
	authz AuthorizationChecker,
	userID string,
	calcutta *models.Calcutta,
) (Decision, error) {
	if userID == "" {
		return Decision{Allowed: false, Status: http.StatusUnauthorized, Code: "unauthorized", Message: "Authentication required"}, nil
	}
	if calcutta == nil {
		return Decision{Allowed: false, Status: http.StatusBadRequest, Code: "calcutta_missing", Message: "Calcutta not found"}, nil
	}

	isAdmin := false
	if authz != nil {
		ok, err := authz.HasPermission(ctx, userID, "global", "", permissionAdminOverride)
		if err != nil {
			return Decision{}, err
		}
		if ok {
			isAdmin = true
		}
	}

	if calcutta.OwnerID != userID && !isAdmin {
		return Decision{Allowed: false, IsAdmin: isAdmin, Status: http.StatusForbidden, Code: "forbidden", Message: "Only the commissioner can invite users"}, nil
	}

	return Decision{Allowed: true, IsAdmin: isAdmin}, nil
}
