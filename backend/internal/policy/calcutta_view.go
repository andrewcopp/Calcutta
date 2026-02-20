package policy

import (
	"context"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// CanViewCalcutta checks if a user can view a calcutta's dashboard and entries.
// Public and unlisted calcuttas are viewable by anyone with the link.
// Private calcuttas require ownership, participation, or a calcutta.read permission.
func CanViewCalcutta(
	ctx context.Context,
	authz AuthorizationChecker,
	userID string,
	calcutta *models.Calcutta,
	participantUserIDs []string,
) (Decision, error) {
	if calcutta == nil {
		return Decision{Allowed: false, Status: http.StatusBadRequest, Code: "calcutta_missing", Message: "Calcutta not found"}, nil
	}

	if calcutta.Visibility == "public" || calcutta.Visibility == "unlisted" {
		return Decision{Allowed: true}, nil
	}

	// Private calcutta: require authentication
	if userID == "" {
		return Decision{Allowed: false, Status: http.StatusUnauthorized, Code: "unauthorized", Message: "Authentication required"}, nil
	}

	if calcutta.OwnerID == userID {
		return Decision{Allowed: true}, nil
	}

	for _, pid := range participantUserIDs {
		if pid == userID {
			return Decision{Allowed: true}, nil
		}
	}

	if authz != nil {
		ok, err := authz.HasPermission(ctx, userID, "calcutta", calcutta.ID, "calcutta.read")
		if err != nil {
			return Decision{}, err
		}
		if ok {
			return Decision{Allowed: true}, nil
		}
	}

	return Decision{Allowed: false, Status: http.StatusForbidden, Code: "forbidden", Message: "Insufficient permissions"}, nil
}
