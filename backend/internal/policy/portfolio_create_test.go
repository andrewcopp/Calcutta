package policy

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatUnauthenticatedUserCannotCreatePortfolio(t *testing.T) {
	// GIVEN no authenticated user
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	tournament := &models.Tournament{ID: "t1"}

	// WHEN checking create portfolio permission
	decision, err := CanCreatePortfolio(context.Background(), nil, "", pool, tournament, nil, time.Now())

	// THEN access is denied with unauthorized status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, decision.Status)
	}
}

func TestThatCreatePortfolioDeniesWhenPoolIsNil(t *testing.T) {
	// GIVEN a nil pool
	tournament := &models.Tournament{ID: "t1"}

	// WHEN checking create portfolio permission
	decision, err := CanCreatePortfolio(context.Background(), nil, "user1", nil, tournament, nil, time.Now())

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatCreatePortfolioDeniesWhenTournamentIsNil(t *testing.T) {
	// GIVEN a nil tournament
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}

	// WHEN checking create portfolio permission
	decision, err := CanCreatePortfolio(context.Background(), nil, "user1", pool, nil, nil, time.Now())

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatNonCommissionerCannotCreatePortfolioForAnotherUser(t *testing.T) {
	// GIVEN a non-commissioner user trying to create a portfolio for another user
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}
	targetUserID := "other-user"

	// WHEN checking create portfolio permission for another user
	decision, err := CanCreatePortfolio(context.Background(), authz, "regular-user", pool, tournament, &targetUserID, time.Now())

	// THEN access is denied with forbidden status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, decision.Status)
	}
}

func TestThatNonAdminCannotCreatePortfolioAfterTournamentStarts(t *testing.T) {
	// GIVEN a tournament that has already started and a non-admin user
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	startingAt := time.Now().Add(-24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking create portfolio permission
	decision, err := CanCreatePortfolio(context.Background(), authz, "regular-user", pool, tournament, nil, time.Now())

	// THEN access is denied with locked status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusLocked {
		t.Fatalf("expected status %d, got %d", http.StatusLocked, decision.Status)
	}
}

func TestThatUserCanCreatePortfolioForSelf(t *testing.T) {
	// GIVEN a regular user creating a portfolio for themselves before the tournament starts
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking create portfolio permission with no target user
	decision, err := CanCreatePortfolio(context.Background(), authz, "regular-user", pool, tournament, nil, time.Now())

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected user to be able to create portfolio for self")
	}
}

func TestThatCommissionerCanCreatePortfolioForAnotherUser(t *testing.T) {
	// GIVEN the pool owner creating a portfolio for another user before the tournament starts
	pool := &models.Pool{ID: "p1", OwnerID: "owner"}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	targetUserID := "other-user"

	// WHEN checking create portfolio permission for another user as the commissioner
	decision, err := CanCreatePortfolio(context.Background(), nil, "owner", pool, tournament, &targetUserID, time.Now())

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected commissioner to be able to create portfolio for another user")
	}
}
