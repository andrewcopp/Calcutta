//go:build integration

package db_test

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
)

func TestThatCreatedScoringRuleCanBeRetrievedByPoolID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool exists
	base := mustSeedBase(t, ctx)

	// WHEN creating a scoring rule and then retrieving scoring rules
	rule := &models.ScoringRule{
		PoolID:        base.pool.ID,
		WinIndex:      1,
		PointsAwarded: 10,
	}
	if err := base.poolRepo.CreateScoringRule(ctx, rule); err != nil {
		t.Fatalf("creating scoring rule: %v", err)
	}

	rules, err := base.poolRepo.GetScoringRules(ctx, base.pool.ID)
	if err != nil {
		t.Fatalf("getting scoring rules: %v", err)
	}

	// THEN exactly one scoring rule is returned with the correct points
	if len(rules) != 1 {
		t.Fatalf("expected 1 scoring rule, got %d", len(rules))
	}
	if rules[0].PointsAwarded != 10 {
		t.Errorf("expected points_awarded 10, got %d", rules[0].PointsAwarded)
	}
}

func TestThatCreateScoringRuleRejectsNegativeWinIndex(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool exists
	base := mustSeedBase(t, ctx)

	// WHEN creating a scoring rule with a negative win index
	rule := &models.ScoringRule{
		PoolID:        base.pool.ID,
		WinIndex:      -1,
		PointsAwarded: 10,
	}
	err := base.poolRepo.CreateScoringRule(ctx, rule)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for negative win_index, got nil")
	}
}

func TestThatCreateScoringRuleRejectsNegativePointsAwarded(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool exists
	base := mustSeedBase(t, ctx)

	// WHEN creating a scoring rule with negative points
	rule := &models.ScoringRule{
		PoolID:        base.pool.ID,
		WinIndex:      1,
		PointsAwarded: -5,
	}
	err := base.poolRepo.CreateScoringRule(ctx, rule)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for negative points_awarded, got nil")
	}
}

func TestThatCreateScoringRuleAllowsZeroPointsAwarded(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool exists
	base := mustSeedBase(t, ctx)

	// WHEN creating a scoring rule with zero points (bye round)
	rule := &models.ScoringRule{
		PoolID:        base.pool.ID,
		WinIndex:      0,
		PointsAwarded: 0,
	}
	err := base.poolRepo.CreateScoringRule(ctx, rule)

	// THEN no error occurs
	if err != nil {
		t.Errorf("expected no error for zero points_awarded, got %v", err)
	}
}
