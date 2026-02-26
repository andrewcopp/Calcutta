package policy

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatIsPortfolioOwnerOrPoolOwnerReturnsFalseWhenPortfolioNotOwned(t *testing.T) {
	// GIVEN a user who does not own the portfolio or pool
	userID := "u1"
	otherUserID := "u2"
	portfolio := &models.Portfolio{UserID: &otherUserID}
	pool := &models.Pool{OwnerID: "owner"}

	// WHEN checking ownership
	result := IsPortfolioOwnerOrPoolOwner(userID, portfolio, pool)

	// THEN the result is false
	if result {
		t.Fatalf("expected false")
	}
}

func TestThatIsPortfolioOwnerOrPoolOwnerReturnsTrueWhenPortfolioOwned(t *testing.T) {
	// GIVEN a user who owns the portfolio
	userID := "u1"
	portfolio := &models.Portfolio{UserID: &userID}
	pool := &models.Pool{OwnerID: "owner"}

	// WHEN checking ownership
	result := IsPortfolioOwnerOrPoolOwner(userID, portfolio, pool)

	// THEN the result is true
	if !result {
		t.Fatalf("expected true")
	}
}

func TestThatIsPortfolioOwnerOrPoolOwnerReturnsTrueWhenPoolOwner(t *testing.T) {
	// GIVEN a user who owns the pool
	userID := "u1"
	portfolio := &models.Portfolio{UserID: nil}
	pool := &models.Pool{OwnerID: userID}

	// WHEN checking ownership
	result := IsPortfolioOwnerOrPoolOwner(userID, portfolio, pool)

	// THEN the result is true
	if !result {
		t.Fatalf("expected true")
	}
}

func TestThatIsPortfolioOwnerOrPoolOwnerReturnsFalseForNilPortfolio(t *testing.T) {
	// GIVEN a nil portfolio
	userID := "u1"
	pool := &models.Pool{OwnerID: "owner"}

	// WHEN checking ownership
	result := IsPortfolioOwnerOrPoolOwner(userID, nil, pool)

	// THEN the result is false
	if result {
		t.Fatalf("expected false")
	}
}

func TestThatIsPortfolioOwnerOrPoolOwnerReturnsFalseForNilPool(t *testing.T) {
	// GIVEN a nil pool and portfolio with different owner
	userID := "u1"
	otherUserID := "u2"
	portfolio := &models.Portfolio{UserID: &otherUserID}

	// WHEN checking ownership
	result := IsPortfolioOwnerOrPoolOwner(userID, portfolio, nil)

	// THEN the result is false
	if result {
		t.Fatalf("expected false")
	}
}
