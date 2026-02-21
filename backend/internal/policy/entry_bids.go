package policy

import (
	"context"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type AuthorizationChecker interface {
	HasPermission(ctx context.Context, userID string, scope string, scopeID string, permission string) (bool, error)
}

type Decision struct {
	Allowed bool
	IsAdmin bool
	Status  int
	Code    string
	Message string
}

const (
	permissionAdminOverride = "calcutta.config.write"
)

func CanEditEntryBids(
	ctx context.Context,
	authz AuthorizationChecker,
	userID string,
	entry *models.CalcuttaEntry,
	calcutta *models.Calcutta,
	tournament *models.Tournament,
	now time.Time,
) (Decision, error) {
	if userID == "" {
		return Decision{Allowed: false, Status: http.StatusUnauthorized, Code: "unauthorized", Message: "Authentication required"}, nil
	}
	if entry == nil {
		return Decision{Allowed: false, Status: http.StatusBadRequest, Code: "entry_missing", Message: "Entry not found"}, nil
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

	authorized := isAdmin
	if entry.UserID != nil && *entry.UserID == userID {
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
		return Decision{Allowed: false, IsAdmin: isAdmin, Status: http.StatusLocked, Code: code, Message: "Bids are locked"}, nil
	}

	return Decision{Allowed: true, IsAdmin: isAdmin}, nil
}

func CanViewEntryData(
	ctx context.Context,
	authz AuthorizationChecker,
	userID string,
	entry *models.CalcuttaEntry,
	calcutta *models.Calcutta,
) (Decision, error) {
	if userID == "" {
		return Decision{Allowed: false, Status: http.StatusUnauthorized, Code: "unauthorized", Message: "Authentication required"}, nil
	}
	if entry == nil {
		return Decision{Allowed: false, Status: http.StatusBadRequest, Code: "entry_missing", Message: "Entry not found"}, nil
	}
	if calcutta == nil {
		return Decision{Allowed: false, Status: http.StatusBadRequest, Code: "calcutta_missing", Message: "Calcutta not found"}, nil
	}

	isAdmin, err := isCalcuttaAdminOrOwner(ctx, authz, userID, calcutta)
	if err != nil {
		return Decision{}, err
	}

	authorized := IsEntryOwnerOrCalcuttaOwner(userID, entry, calcutta) || isAdmin
	if !authorized {
		return Decision{Allowed: false, IsAdmin: isAdmin, Status: http.StatusForbidden, Code: "forbidden", Message: "Insufficient permissions"}, nil
	}

	return Decision{Allowed: true, IsAdmin: isAdmin}, nil
}

// IsBiddingPhaseViewAllowed checks whether a user may view another entry's
// bid-sensitive data (teams, portfolios, portfolio-teams) while bidding is
// still open. Pure function — no interfaces, no context.
func IsBiddingPhaseViewAllowed(userID string, entry *models.CalcuttaEntry, tournament *models.Tournament, now time.Time, isAdmin bool) bool {
	if tournament.HasStarted(now) {
		return true
	}
	if isAdmin {
		return true
	}
	if entry != nil && entry.UserID != nil && *entry.UserID == userID {
		return true
	}
	return false
}

// IsEntryOwnerOrCalcuttaOwner is a pure function that checks if a user owns the entry or the calcutta.
// This can be tested without mocking authorization.
func IsEntryOwnerOrCalcuttaOwner(userID string, entry *models.CalcuttaEntry, calcutta *models.Calcutta) bool {
	if entry != nil && entry.UserID != nil && *entry.UserID == userID {
		return true
	}
	if calcutta != nil && calcutta.OwnerID == userID {
		return true
	}
	return false
}
