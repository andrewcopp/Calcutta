//go:build integration

package db_test

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
)

func TestThatReplacePayoutsSoftDeletesOldAndInsertsNew(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool with initial payouts
	base := mustSeedBase(t, ctx)
	initial := []*models.PoolPayout{
		{Position: 1, AmountCents: 500},
		{Position: 2, AmountCents: 300},
	}
	if err := base.poolRepo.ReplacePayouts(ctx, base.pool.ID, initial); err != nil {
		t.Fatalf("creating initial payouts: %v", err)
	}

	// WHEN replacing with new payouts (different positions to avoid full-table UNIQUE conflict)
	// TODO: uq_core_payouts_pool_position is a full-table UNIQUE constraint, not a partial
	// index. Soft-deleted rows will conflict if we reuse the same positions. This should be
	// converted to a partial unique index: WHERE deleted_at IS NULL.
	replacement := []*models.PoolPayout{
		{Position: 3, AmountCents: 400},
		{Position: 4, AmountCents: 200},
		{Position: 5, AmountCents: 100},
	}
	if err := base.poolRepo.ReplacePayouts(ctx, base.pool.ID, replacement); err != nil {
		t.Fatalf("replacing payouts: %v", err)
	}

	// THEN only the replacement payouts are returned
	got, err := base.poolRepo.GetPayouts(ctx, base.pool.ID)
	if err != nil {
		t.Fatalf("getting payouts: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 payouts, got %d", len(got))
	}
}

func TestThatReplacePayoutsRejectsPositionZero(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool exists
	base := mustSeedBase(t, ctx)

	// WHEN replacing payouts with position zero
	payouts := []*models.PoolPayout{
		{Position: 0, AmountCents: 100},
	}
	err := base.poolRepo.ReplacePayouts(ctx, base.pool.ID, payouts)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for position zero, got nil")
	}
}

func TestThatReplacePayoutsRejectsNegativeAmountCents(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool exists
	base := mustSeedBase(t, ctx)

	// WHEN replacing payouts with negative amount
	payouts := []*models.PoolPayout{
		{Position: 1, AmountCents: -50},
	}
	err := base.poolRepo.ReplacePayouts(ctx, base.pool.ID, payouts)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for negative amount_cents, got nil")
	}
}

func TestThatReplacePayoutsAllowsZeroAmountCents(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool exists
	base := mustSeedBase(t, ctx)

	// WHEN replacing payouts with zero amount
	payouts := []*models.PoolPayout{
		{Position: 1, AmountCents: 0},
	}
	err := base.poolRepo.ReplacePayouts(ctx, base.pool.ID, payouts)

	// THEN no error occurs (CHECK is >= 0, not > 0)
	if err != nil {
		t.Errorf("expected no error for zero amount_cents, got %v", err)
	}
}

func TestThatReplacePayoutsRejectsDuplicatePositionsInSameBatch(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool exists
	base := mustSeedBase(t, ctx)

	// WHEN replacing payouts with two payouts having the same position
	payouts := []*models.PoolPayout{
		{Position: 1, AmountCents: 500},
		{Position: 1, AmountCents: 300},
	}
	err := base.poolRepo.ReplacePayouts(ctx, base.pool.ID, payouts)

	// THEN the error is non-nil (UNIQUE constraint violation)
	if err == nil {
		t.Error("expected error for duplicate positions in same batch, got nil")
	}
}
