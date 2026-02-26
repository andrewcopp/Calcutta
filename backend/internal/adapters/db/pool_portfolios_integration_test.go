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

func TestThatCreatePortfolioRejectsDuplicateUserInSamePool(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a portfolio already exists for this user in this pool
	base := mustSeedBase(t, ctx)
	_ = mustSeedPortfolio(t, ctx, base.poolRepo, base.pool.ID, base.user.ID)

	// WHEN creating a second portfolio for the same user in the same pool
	duplicate := &models.Portfolio{
		Name:   "Duplicate Portfolio",
		UserID: &base.user.ID,
		PoolID: base.pool.ID,
	}
	err := base.poolRepo.CreatePortfolio(ctx, duplicate)

	// THEN the error is an AlreadyExistsError
	var alreadyExists *apperrors.AlreadyExistsError
	if !errors.As(err, &alreadyExists) {
		t.Errorf("expected *apperrors.AlreadyExistsError, got %T: %v", err, err)
	}
}

func TestThatCreatePortfolioAllowsSameUserInDifferentPools(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a portfolio exists for a user in one pool
	base := mustSeedBase(t, ctx)
	_ = mustSeedPortfolio(t, ctx, base.poolRepo, base.pool.ID, base.user.ID)

	// AND a second pool exists
	secondPool := &models.Pool{
		TournamentID:         base.tournament.ID,
		OwnerID:              base.user.ID,
		CreatedBy:            base.user.ID,
		Name:                 "Second Pool",
		BudgetCredits:        100,
		MinTeams:             1,
		MaxTeams:             10,
		MaxInvestmentCredits: 50,
	}
	if err := base.poolRepo.Create(ctx, secondPool); err != nil {
		t.Fatalf("creating second pool: %v", err)
	}

	// WHEN creating a portfolio for the same user in the second pool
	portfolio := &models.Portfolio{
		Name:   "Second Pool Portfolio",
		UserID: &base.user.ID,
		PoolID: secondPool.ID,
	}
	err := base.poolRepo.CreatePortfolio(ctx, portfolio)

	// THEN no error occurs
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestThatCreatePortfolioAllowsNullUserID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool exists
	base := mustSeedBase(t, ctx)

	// WHEN creating a portfolio with nil UserID
	portfolio := &models.Portfolio{
		Name:   "Anonymous Portfolio",
		UserID: nil,
		PoolID: base.pool.ID,
	}
	err := base.poolRepo.CreatePortfolio(ctx, portfolio)

	// THEN no error occurs
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestThatCreatePortfolioAllowsMultipleNullUserIDsInSamePool(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool with one null-user portfolio already exists
	base := mustSeedBase(t, ctx)
	first := &models.Portfolio{
		Name:   "Anonymous Portfolio 1",
		UserID: nil,
		PoolID: base.pool.ID,
	}
	if err := base.poolRepo.CreatePortfolio(ctx, first); err != nil {
		t.Fatalf("creating first null-user portfolio: %v", err)
	}

	// WHEN creating another null-user portfolio in the same pool
	second := &models.Portfolio{
		Name:   "Anonymous Portfolio 2",
		UserID: nil,
		PoolID: base.pool.ID,
	}
	err := base.poolRepo.CreatePortfolio(ctx, second)

	// THEN no error occurs (NULL != NULL in partial index)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestThatGetPortfoliosReturnsPortfoliosOrderedByReturnsDescending(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool with two portfolios (no investments, so both have 0 returns)
	base := mustSeedBase(t, ctx)
	secondUser := mustSeedSecondUser(t, ctx, base.userRepo)
	portfolio1 := mustSeedPortfolio(t, ctx, base.poolRepo, base.pool.ID, base.user.ID)
	portfolio2 := mustSeedPortfolio(t, ctx, base.poolRepo, base.pool.ID, secondUser.ID)

	// WHEN retrieving portfolios
	portfolios, _, err := base.poolRepo.GetPortfolios(ctx, base.pool.ID)
	if err != nil {
		t.Fatalf("getting portfolios: %v", err)
	}

	// THEN both portfolios are returned
	if len(portfolios) != 2 {
		t.Errorf("expected 2 portfolios, got %d", len(portfolios))
	}

	// AND portfolios are in descending order by created_at (both have 0 returns)
	_ = portfolio1
	_ = portfolio2
	ids := make([]string, 0, len(portfolios))
	for _, p := range portfolios {
		ids = append(ids, p.ID)
	}
	if len(ids) == 2 && ids[0] != portfolio2.ID {
		t.Errorf("expected portfolio2 (%s) first (created later), got %s", portfolio2.ID, ids[0])
	}
}

func TestThatGetPortfolioReturnsNotFoundForMissingID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool repository
	base := mustSeedBase(t, ctx)

	// WHEN getting a portfolio that does not exist
	_, err := base.poolRepo.GetPortfolio(ctx, uuid.New().String())

	// THEN the error is a NotFoundError
	var notFound *apperrors.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected *apperrors.NotFoundError, got %T: %v", err, err)
	}
}
