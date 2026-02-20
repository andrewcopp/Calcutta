package policy

import (
	"context"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// CanManageCalcutta checks if a user can modify calcutta settings, payouts, etc.
func CanManageCalcutta(
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

	isAdmin, err := isCalcuttaAdminOrOwner(ctx, authz, userID, calcutta)
	if err != nil {
		return Decision{}, err
	}

	if !isAdmin {
		return Decision{Allowed: false, IsAdmin: false, Status: http.StatusForbidden, Code: "forbidden", Message: "Insufficient permissions"}, nil
	}

	return Decision{Allowed: true, IsAdmin: isAdmin}, nil
}
