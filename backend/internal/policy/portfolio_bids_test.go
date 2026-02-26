package policy

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// --- CanEditPortfolioInvestments tests ---

func TestThatUnauthenticatedUserCannotEditPortfolioInvestments(t *testing.T) {
	// GIVEN no authenticated user
	userID := "u1"
	portfolio := &models.Portfolio{ID: "p1", UserID: &userID}
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}
	tournament := &models.Tournament{ID: "t1"}

	// WHEN checking edit investments permission
	decision, err := CanEditPortfolioInvestments(context.Background(), nil, "", portfolio, pool, tournament, time.Now())

	// THEN access is denied with unauthorized status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, decision.Status)
	}
}

func TestThatEditPortfolioInvestmentsDeniesWhenPortfolioIsNil(t *testing.T) {
	// GIVEN a nil portfolio
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}
	tournament := &models.Tournament{ID: "t1"}

	// WHEN checking edit investments permission
	decision, err := CanEditPortfolioInvestments(context.Background(), nil, "user1", nil, pool, tournament, time.Now())

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatEditPortfolioInvestmentsDeniesWhenPoolIsNil(t *testing.T) {
	// GIVEN a nil pool
	userID := "u1"
	portfolio := &models.Portfolio{ID: "p1", UserID: &userID}
	tournament := &models.Tournament{ID: "t1"}

	// WHEN checking edit investments permission
	decision, err := CanEditPortfolioInvestments(context.Background(), nil, "u1", portfolio, nil, tournament, time.Now())

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatEditPortfolioInvestmentsDeniesWhenTournamentIsNil(t *testing.T) {
	// GIVEN a nil tournament
	userID := "u1"
	portfolio := &models.Portfolio{ID: "p1", UserID: &userID}
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}

	// WHEN checking edit investments permission
	decision, err := CanEditPortfolioInvestments(context.Background(), nil, "u1", portfolio, pool, nil, time.Now())

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatNonOwnerNonAdminCannotEditPortfolioInvestments(t *testing.T) {
	// GIVEN a user who is neither the portfolio owner nor an admin
	otherUserID := "other-user"
	portfolio := &models.Portfolio{ID: "p1", UserID: &otherUserID}
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking edit investments permission
	decision, err := CanEditPortfolioInvestments(context.Background(), authz, "stranger", portfolio, pool, tournament, time.Now())

	// THEN access is denied with forbidden status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, decision.Status)
	}
}

func TestThatPortfolioOwnerCannotEditInvestmentsAfterTournamentStarts(t *testing.T) {
	// GIVEN the portfolio owner and a tournament that has already started
	userID := "u1"
	portfolio := &models.Portfolio{ID: "p1", UserID: &userID}
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}
	startingAt := time.Now().Add(-24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking edit investments permission
	decision, err := CanEditPortfolioInvestments(context.Background(), authz, "u1", portfolio, pool, tournament, time.Now())

	// THEN access is denied with locked status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusLocked {
		t.Fatalf("expected status %d, got %d", http.StatusLocked, decision.Status)
	}
}

func TestThatPortfolioOwnerCanEditInvestmentsBeforeTournamentStarts(t *testing.T) {
	// GIVEN the portfolio owner and a tournament that has not started
	userID := "u1"
	portfolio := &models.Portfolio{ID: "p1", UserID: &userID}
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}
	startingAt := time.Now().Add(24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking edit investments permission
	decision, err := CanEditPortfolioInvestments(context.Background(), authz, "u1", portfolio, pool, tournament, time.Now())

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected portfolio owner to be able to edit investments before tournament starts")
	}
}

func TestThatAdminCanEditInvestmentsAfterTournamentStarts(t *testing.T) {
	// GIVEN an admin user and a tournament that has already started
	otherUserID := "other-user"
	portfolio := &models.Portfolio{ID: "p1", UserID: &otherUserID}
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}
	startingAt := time.Now().Add(-24 * time.Hour)
	tournament := &models.Tournament{ID: "t1", StartingAt: &startingAt}

	// WHEN the pool owner checks edit investments permission
	decision, err := CanEditPortfolioInvestments(context.Background(), nil, "owner", portfolio, pool, tournament, time.Now())

	// THEN access is allowed because admins bypass the tournament lock
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected admin to be able to edit investments even after tournament starts")
	}
}

// --- CanViewPortfolioData tests ---

func TestThatUnauthenticatedUserCannotViewPortfolioData(t *testing.T) {
	// GIVEN no authenticated user
	userID := "u1"
	portfolio := &models.Portfolio{ID: "p1", UserID: &userID}
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}

	// WHEN checking view portfolio data permission
	decision, err := CanViewPortfolioData(context.Background(), nil, "", portfolio, pool)

	// THEN access is denied with unauthorized status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, decision.Status)
	}
}

func TestThatViewPortfolioDataDeniesWhenPortfolioIsNil(t *testing.T) {
	// GIVEN a nil portfolio
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}

	// WHEN checking view portfolio data permission
	decision, err := CanViewPortfolioData(context.Background(), nil, "user1", nil, pool)

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatViewPortfolioDataDeniesWhenPoolIsNil(t *testing.T) {
	// GIVEN a nil pool
	userID := "u1"
	portfolio := &models.Portfolio{ID: "p1", UserID: &userID}

	// WHEN checking view portfolio data permission
	decision, err := CanViewPortfolioData(context.Background(), nil, "u1", portfolio, nil)

	// THEN access is denied with bad request status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, decision.Status)
	}
}

func TestThatNonOwnerNonAdminCannotViewPortfolioData(t *testing.T) {
	// GIVEN a user who is neither the portfolio owner, pool owner, nor an admin
	otherUserID := "other-user"
	portfolio := &models.Portfolio{ID: "p1", UserID: &otherUserID}
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking view portfolio data permission
	decision, err := CanViewPortfolioData(context.Background(), authz, "stranger", portfolio, pool)

	// THEN access is denied with forbidden status
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Status != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, decision.Status)
	}
}

func TestThatPortfolioOwnerCanViewPortfolioData(t *testing.T) {
	// GIVEN the portfolio owner
	userID := "u1"
	portfolio := &models.Portfolio{ID: "p1", UserID: &userID}
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: false}

	// WHEN checking view portfolio data permission
	decision, err := CanViewPortfolioData(context.Background(), authz, "u1", portfolio, pool)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected portfolio owner to be able to view portfolio data")
	}
}

func TestThatPoolOwnerCanViewPortfolioData(t *testing.T) {
	// GIVEN the pool owner viewing another user's portfolio
	otherUserID := "other-user"
	portfolio := &models.Portfolio{ID: "p1", UserID: &otherUserID}
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}

	// WHEN checking view portfolio data permission as the pool owner
	decision, err := CanViewPortfolioData(context.Background(), nil, "owner", portfolio, pool)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected pool owner to be able to view any portfolio data")
	}
}

func TestThatAdminCanViewPortfolioData(t *testing.T) {
	// GIVEN an admin user viewing another user's portfolio
	otherUserID := "other-user"
	portfolio := &models.Portfolio{ID: "p1", UserID: &otherUserID}
	pool := &models.Pool{ID: "pool1", OwnerID: "owner"}
	authz := &mockAuthzChecker{result: true}

	// WHEN checking view portfolio data permission as an admin
	decision, err := CanViewPortfolioData(context.Background(), authz, "admin-user", portfolio, pool)

	// THEN access is allowed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("expected admin to be able to view any portfolio data")
	}
}
