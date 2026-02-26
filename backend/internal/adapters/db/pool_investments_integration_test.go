//go:build integration

package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
	"github.com/google/uuid"
)

func TestThatReplaceInvestmentsSoftDeletesOldAndInsertsNew(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a portfolio with initial investments
	seed := mustSeedWithTeams(t, ctx, 3)
	portfolio := mustSeedPortfolio(t, ctx, seed.poolRepo, seed.pool.ID, seed.user.ID)

	initial := []*models.Investment{
		{TeamID: seed.teams[0].ID, Credits: 20},
		{TeamID: seed.teams[1].ID, Credits: 30},
	}
	if err := seed.poolRepo.ReplaceInvestments(ctx, portfolio.ID, initial); err != nil {
		t.Fatalf("creating initial investments: %v", err)
	}

	// WHEN replacing with different investments
	replacement := []*models.Investment{
		{TeamID: seed.teams[2].ID, Credits: 50},
	}
	if err := seed.poolRepo.ReplaceInvestments(ctx, portfolio.ID, replacement); err != nil {
		t.Fatalf("replacing investments: %v", err)
	}

	// THEN only the replacement investment is returned
	got, err := seed.poolRepo.GetInvestments(ctx, portfolio.ID)
	if err != nil {
		t.Fatalf("getting investments: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 investment, got %d", len(got))
	}
}

func TestThatReplaceInvestmentsRejectsZeroCredits(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a portfolio and a team exist
	seed := mustSeedWithTeams(t, ctx, 1)
	portfolio := mustSeedPortfolio(t, ctx, seed.poolRepo, seed.pool.ID, seed.user.ID)

	// WHEN replacing with zero credits
	investments := []*models.Investment{
		{TeamID: seed.teams[0].ID, Credits: 0},
	}
	err := seed.poolRepo.ReplaceInvestments(ctx, portfolio.ID, investments)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for zero credits, got nil")
	}
}

func TestThatReplaceInvestmentsRejectsDuplicateTeamInSamePortfolio(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a portfolio and a team exist
	seed := mustSeedWithTeams(t, ctx, 1)
	portfolio := mustSeedPortfolio(t, ctx, seed.poolRepo, seed.pool.ID, seed.user.ID)

	// WHEN replacing with the same team twice
	investments := []*models.Investment{
		{TeamID: seed.teams[0].ID, Credits: 20},
		{TeamID: seed.teams[0].ID, Credits: 30},
	}
	err := seed.poolRepo.ReplaceInvestments(ctx, portfolio.ID, investments)

	// THEN the error is non-nil (UNIQUE partial index violation)
	if err == nil {
		t.Error("expected error for duplicate team in same portfolio, got nil")
	}
}

func TestThatReplaceInvestmentsReturnsNotFoundForMissingPortfolio(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool repository
	seed := mustSeedWithTeams(t, ctx, 1)

	// WHEN replacing investments for a non-existent portfolio
	investments := []*models.Investment{
		{TeamID: seed.teams[0].ID, Credits: 10},
	}
	err := seed.poolRepo.ReplaceInvestments(ctx, uuid.New().String(), investments)

	// THEN the error is a NotFoundError
	var notFound *apperrors.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected *apperrors.NotFoundError, got %T: %v", err, err)
	}
}

func TestThatGetInvestmentsJoinsTeamAndSchoolData(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a portfolio with one investment
	seed := mustSeedWithTeams(t, ctx, 1)
	portfolio := mustSeedPortfolio(t, ctx, seed.poolRepo, seed.pool.ID, seed.user.ID)

	investments := []*models.Investment{
		{TeamID: seed.teams[0].ID, Credits: 25},
	}
	if err := seed.poolRepo.ReplaceInvestments(ctx, portfolio.ID, investments); err != nil {
		t.Fatalf("replacing investments: %v", err)
	}

	// WHEN retrieving investments
	got, err := seed.poolRepo.GetInvestments(ctx, portfolio.ID)
	if err != nil {
		t.Fatalf("getting investments: %v", err)
	}

	// THEN the investment includes joined TournamentTeam with School data
	if len(got) != 1 {
		t.Fatalf("expected 1 investment, got %d", len(got))
	}
	if got[0].Team == nil {
		t.Fatal("expected non-nil Team")
	}
	if got[0].Team.School == nil {
		t.Fatal("expected non-nil School on Team")
	}
	if got[0].Team.School.Name == "" {
		t.Error("expected non-empty school name")
	}
}

func TestThatGetInvestmentsByPortfolioIDsGroupsByPortfolioID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN two portfolios, each with different investments
	seed := mustSeedWithTeams(t, ctx, 3)
	secondUser := mustSeedSecondUser(t, ctx, seed.userRepo)
	portfolio1 := mustSeedPortfolio(t, ctx, seed.poolRepo, seed.pool.ID, seed.user.ID)
	portfolio2 := mustSeedPortfolio(t, ctx, seed.poolRepo, seed.pool.ID, secondUser.ID)

	investments1 := []*models.Investment{
		{TeamID: seed.teams[0].ID, Credits: 20},
		{TeamID: seed.teams[1].ID, Credits: 30},
	}
	if err := seed.poolRepo.ReplaceInvestments(ctx, portfolio1.ID, investments1); err != nil {
		t.Fatalf("replacing portfolio1 investments: %v", err)
	}
	investments2 := []*models.Investment{
		{TeamID: seed.teams[2].ID, Credits: 50},
	}
	if err := seed.poolRepo.ReplaceInvestments(ctx, portfolio2.ID, investments2); err != nil {
		t.Fatalf("replacing portfolio2 investments: %v", err)
	}

	// WHEN retrieving investments by portfolio IDs
	got, err := seed.poolRepo.GetInvestmentsByPortfolioIDs(ctx, []string{portfolio1.ID, portfolio2.ID})
	if err != nil {
		t.Fatalf("getting investments by IDs: %v", err)
	}

	// THEN the map has correct groupings
	if len(got[portfolio1.ID]) != 2 {
		t.Errorf("expected 2 investments for portfolio1, got %d", len(got[portfolio1.ID]))
	}
	if len(got[portfolio2.ID]) != 1 {
		t.Errorf("expected 1 investment for portfolio2, got %d", len(got[portfolio2.ID]))
	}
}

func TestThatGetInvestmentsByPortfolioIDsReturnsEmptyMapForEmptyInput(t *testing.T) {
	ctx := context.Background()

	// GIVEN a pool repository
	seed := mustSeedBase(t, ctx)
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// WHEN passing an empty slice of portfolio IDs
	got, err := seed.poolRepo.GetInvestmentsByPortfolioIDs(ctx, []string{})

	// THEN an empty map is returned with no error
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %d entries", len(got))
	}
}
