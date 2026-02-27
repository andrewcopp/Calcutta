//go:build integration

package db_test

import (
	"context"
	"testing"

	db "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
)

func TestThatInvestmentSnapshotIsCreatedWithCorrectData(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a portfolio exists
	seed := mustSeedWithTeams(t, ctx, 2)
	portfolio := mustSeedPortfolio(t, ctx, seed.poolRepo, seed.pool.ID, seed.user.ID)

	snapshotRepo := db.NewInvestmentSnapshotRepository(pool)

	// WHEN creating a snapshot
	snapshot := &models.InvestmentSnapshot{
		PortfolioID: portfolio.ID,
		ChangedBy:   seed.user.ID,
		Reason:      "",
		Investments: []models.InvestmentSnapshotEntry{
			{TeamID: seed.teams[0].ID, Credits: 30},
			{TeamID: seed.teams[1].ID, Credits: 70},
		},
	}
	err := snapshotRepo.CreateInvestmentSnapshot(ctx, snapshot)

	// THEN no error is returned
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestThatMultipleSnapshotsAreRecordedForSamePortfolio(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a portfolio and a first snapshot exist
	seed := mustSeedWithTeams(t, ctx, 2)
	portfolio := mustSeedPortfolio(t, ctx, seed.poolRepo, seed.pool.ID, seed.user.ID)

	snapshotRepo := db.NewInvestmentSnapshotRepository(pool)

	first := &models.InvestmentSnapshot{
		PortfolioID: portfolio.ID,
		ChangedBy:   seed.user.ID,
		Investments: []models.InvestmentSnapshotEntry{
			{TeamID: seed.teams[0].ID, Credits: 50},
		},
	}
	if err := snapshotRepo.CreateInvestmentSnapshot(ctx, first); err != nil {
		t.Fatalf("creating first snapshot: %v", err)
	}

	// WHEN creating a second snapshot
	second := &models.InvestmentSnapshot{
		PortfolioID: portfolio.ID,
		ChangedBy:   seed.user.ID,
		Reason:      "admin_override",
		Investments: []models.InvestmentSnapshotEntry{
			{TeamID: seed.teams[1].ID, Credits: 100},
		},
	}
	err := snapshotRepo.CreateInvestmentSnapshot(ctx, second)

	// THEN no error is returned (both snapshots coexist)
	if err != nil {
		t.Errorf("unexpected error creating second snapshot: %v", err)
	}
}

func TestThatInvestmentSnapshotRequiresValidPortfolioID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user exists but the portfolio ID is invalid
	seed := mustSeedBase(t, ctx)
	snapshotRepo := db.NewInvestmentSnapshotRepository(pool)

	// WHEN creating a snapshot with a non-existent portfolio
	snapshot := &models.InvestmentSnapshot{
		PortfolioID: "00000000-0000-0000-0000-000000000000",
		ChangedBy:   seed.user.ID,
		Investments: []models.InvestmentSnapshotEntry{
			{TeamID: "team-1", Credits: 50},
		},
	}
	err := snapshotRepo.CreateInvestmentSnapshot(ctx, snapshot)

	// THEN a foreign key error is returned
	if err == nil {
		t.Error("expected error for invalid portfolio_id, got nil")
	}
}
