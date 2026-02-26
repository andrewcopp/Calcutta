package policy

import (
	"context"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func CanEditPortfolioInvestments(
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
		return Decision{Allowed: false, IsAdmin: isAdmin, Status: http.StatusLocked, Code: code, Message: "Investments are locked"}, nil
	}

	return Decision{Allowed: true, IsAdmin: isAdmin}, nil
}

func CanViewPortfolioData(
	ctx context.Context,
	authz AuthorizationChecker,
	userID string,
	portfolio *models.Portfolio,
	pool *models.Pool,
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

	isAdmin, err := isPoolAdminOrOwner(ctx, authz, userID, pool)
	if err != nil {
		return Decision{}, err
	}

	authorized := IsPortfolioOwnerOrPoolOwner(userID, portfolio, pool) || isAdmin
	if !authorized {
		return Decision{Allowed: false, IsAdmin: isAdmin, Status: http.StatusForbidden, Code: "forbidden", Message: "Insufficient permissions"}, nil
	}

	return Decision{Allowed: true, IsAdmin: isAdmin}, nil
}

// IsBiddingPhaseViewAllowed checks whether a user may view another portfolio's
// investment-sensitive data (investments, ownership summaries, ownership details) while bidding is
// still open. Pure function -- no interfaces, no context.
func IsBiddingPhaseViewAllowed(userID string, portfolio *models.Portfolio, tournament *models.Tournament, now time.Time, isAdmin bool) bool {
	if tournament.HasStarted(now) {
		return true
	}
	if isAdmin {
		return true
	}
	if portfolio != nil && portfolio.UserID != nil && *portfolio.UserID == userID {
		return true
	}
	return false
}

// IsPortfolioOwnerOrPoolOwner is a pure function that checks if a user owns the portfolio or the pool.
// This can be tested without mocking authorization.
func IsPortfolioOwnerOrPoolOwner(userID string, portfolio *models.Portfolio, pool *models.Pool) bool {
	if portfolio != nil && portfolio.UserID != nil && *portfolio.UserID == userID {
		return true
	}
	if pool != nil && pool.OwnerID == userID {
		return true
	}
	return false
}
