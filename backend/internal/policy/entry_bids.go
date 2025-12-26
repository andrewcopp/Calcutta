package policy

import (
	"context"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
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

	authorized := false
	if entry.UserID != nil && *entry.UserID == userID {
		authorized = true
	}
	if calcutta.OwnerID == userID {
		authorized = true
	}
	if isAdmin {
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
