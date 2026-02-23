//go:build integration

package db_test

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
)

func TestThatCreatedRoundCanBeRetrievedByCalcuttaID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta exists
	base := mustSeedBase(t, ctx)

	// WHEN creating a round and then retrieving rounds
	round := &models.CalcuttaRound{
		CalcuttaID: base.calcutta.ID,
		Round:      1,
		Points:     10,
	}
	if err := base.calcuttaRepo.CreateRound(ctx, round); err != nil {
		t.Fatalf("creating round: %v", err)
	}

	rounds, err := base.calcuttaRepo.GetRounds(ctx, base.calcutta.ID)
	if err != nil {
		t.Fatalf("getting rounds: %v", err)
	}

	// THEN exactly one round is returned with the correct points
	if len(rounds) != 1 {
		t.Fatalf("expected 1 round, got %d", len(rounds))
	}
	if rounds[0].Points != 10 {
		t.Errorf("expected points 10, got %d", rounds[0].Points)
	}
}

func TestThatCreateRoundRejectsNegativeWinIndex(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta exists
	base := mustSeedBase(t, ctx)

	// WHEN creating a round with a negative win index
	round := &models.CalcuttaRound{
		CalcuttaID: base.calcutta.ID,
		Round:      -1,
		Points:     10,
	}
	err := base.calcuttaRepo.CreateRound(ctx, round)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for negative win_index, got nil")
	}
}

func TestThatCreateRoundRejectsNegativePointsAwarded(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta exists
	base := mustSeedBase(t, ctx)

	// WHEN creating a round with negative points
	round := &models.CalcuttaRound{
		CalcuttaID: base.calcutta.ID,
		Round:      1,
		Points:     -5,
	}
	err := base.calcuttaRepo.CreateRound(ctx, round)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for negative points_awarded, got nil")
	}
}

func TestThatCreateRoundAllowsZeroPointsAwarded(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta exists
	base := mustSeedBase(t, ctx)

	// WHEN creating a round with zero points (bye round)
	round := &models.CalcuttaRound{
		CalcuttaID: base.calcutta.ID,
		Round:      0,
		Points:     0,
	}
	err := base.calcuttaRepo.CreateRound(ctx, round)

	// THEN no error occurs
	if err != nil {
		t.Errorf("expected no error for zero points_awarded, got %v", err)
	}
}
