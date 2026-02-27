package policy

import (
	"context"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func CanDeletePortfolio(
	ctx context.Context,
	authz AuthorizationChecker,
	userID string,
	portfolio *models.Portfolio,
	pool *models.Pool,
	tournament *models.Tournament,
	now time.Time,
) (Decision, error) {
	if userID == "" {
		return Decision{Allowed: false, Status: http.StatusUnauthorized, Code: "unauthorized", Message: "Authentication required"}, nil
	}
	if portfolio == nil {
		return Decision{Allowed: false, Status: http.StatusBadRequest, Code: "portfolio_missing", Message: "Portfolio not found"}, nil
	}
	if pool == nil {
		return Decision{Allowed: false, Status: http.StatusBadRequest, Code: "pool_missing", Message: "Pool not found"}, nil
	}
	if tournament == nil {
		return Decision{Allowed: false, Status: http.StatusBadRequest, Code: "tournament_missing", Message: "Tournament not found"}, nil
	}

	isAdmin, err := isPoolAdminOrOwner(ctx, authz, userID, pool)
	if err != nil {
		return Decision{}, err
	}

	authorized := isAdmin
	if portfolio.UserID != nil && *portfolio.UserID == userID {
		authorized = true
	}
	if !authorized {
		return Decision{Allowed: false, IsAdmin: isAdmin, Status: http.StatusForbidden, Code: "forbidden", Message: "Insufficient permissions"}, nil
	}

	if ok, reason := tournament.CanEditBids(now, isAdmin); !ok {
		code := "tournament_locked"
		if reason != "" {
			code = reason
		}
		return Decision{Allowed: false, IsAdmin: isAdmin, Status: http.StatusLocked, Code: code, Message: "Portfolio cannot be deleted after tournament starts"}, nil
	}

	return Decision{Allowed: true, IsAdmin: isAdmin}, nil
}
