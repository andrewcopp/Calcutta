package policy

import (
	"context"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func CanCreateEntry(
	ctx context.Context,
	authz AuthorizationChecker,
	userID string,
	calcutta *models.Calcutta,
	tournament *models.Tournament,
	targetUserID *string,
	now time.Time,
) (Decision, error) {
	if userID == "" {
		return Decision{Allowed: false, Status: http.StatusUnauthorized, Code: "unauthorized", Message: "Authentication required"}, nil
	}
	if calcutta == nil {
		return Decision{Allowed: false, Status: http.StatusBadRequest, Code: "calcutta_missing", Message: "Calcutta not found"}, nil
	}
	if tournament == nil {
		return Decision{Allowed: false, Status: http.StatusBadRequest, Code: "tournament_missing", Message: "Tournament not found"}, nil
	}

	isAdmin, err := isCalcuttaAdminOrOwner(ctx, authz, userID, calcutta)
	if err != nil {
		return Decision{}, err
	}

	// Commissioner path: pool owner or admin can create entries for any user
	isCommissioner := isAdmin
	if targetUserID != nil && !isCommissioner {
		return Decision{Allowed: false, IsAdmin: isAdmin, Status: http.StatusForbidden, Code: "forbidden", Message: "Only the commissioner can create entries for other users"}, nil
	}

	if ok, reason := tournament.CanEditBids(now, isAdmin); !ok {
		code := "tournament_locked"
		if reason != "" {
			code = reason
		}
		return Decision{Allowed: false, IsAdmin: isAdmin, Status: http.StatusLocked, Code: code, Message: "Entries are locked"}, nil
	}

	return Decision{Allowed: true, IsAdmin: isAdmin}, nil
}
