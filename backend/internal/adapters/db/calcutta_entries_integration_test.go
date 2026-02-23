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

func TestThatCreateEntryRejectsDuplicateUserInSameCalcutta(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN an entry already exists for this user in this calcutta
	base := mustSeedBase(t, ctx)
	_ = mustSeedEntry(t, ctx, base.calcuttaRepo, base.calcutta.ID, base.user.ID)

	// WHEN creating a second entry for the same user in the same calcutta
	duplicate := &models.CalcuttaEntry{
		Name:       "Duplicate Entry",
		UserID:     &base.user.ID,
		CalcuttaID: base.calcutta.ID,
	}
	err := base.calcuttaRepo.CreateEntry(ctx, duplicate)

	// THEN the error is an AlreadyExistsError
	var alreadyExists *apperrors.AlreadyExistsError
	if !errors.As(err, &alreadyExists) {
		t.Errorf("expected *apperrors.AlreadyExistsError, got %T: %v", err, err)
	}
}

func TestThatCreateEntryAllowsSameUserInDifferentCalcuttas(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN an entry exists for a user in one calcutta
	base := mustSeedBase(t, ctx)
	_ = mustSeedEntry(t, ctx, base.calcuttaRepo, base.calcutta.ID, base.user.ID)

	// AND a second calcutta exists
	secondCalcutta := &models.Calcutta{
		TournamentID: base.tournament.ID,
		OwnerID:      base.user.ID,
		CreatedBy:    base.user.ID,
		Name:         "Second Pool",
		BudgetPoints: 100,
		MinTeams:     1,
		MaxTeams:     10,
		MaxBidPoints: 50,
	}
	if err := base.calcuttaRepo.Create(ctx, secondCalcutta); err != nil {
		t.Fatalf("creating second calcutta: %v", err)
	}

	// WHEN creating an entry for the same user in the second calcutta
	entry := &models.CalcuttaEntry{
		Name:       "Second Pool Entry",
		UserID:     &base.user.ID,
		CalcuttaID: secondCalcutta.ID,
	}
	err := base.calcuttaRepo.CreateEntry(ctx, entry)

	// THEN no error occurs
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestThatCreateEntryAllowsNullUserID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta exists
	base := mustSeedBase(t, ctx)

	// WHEN creating an entry with nil UserID
	entry := &models.CalcuttaEntry{
		Name:       "Anonymous Entry",
		UserID:     nil,
		CalcuttaID: base.calcutta.ID,
	}
	err := base.calcuttaRepo.CreateEntry(ctx, entry)

	// THEN no error occurs
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestThatCreateEntryAllowsMultipleNullUserIDsInSameCalcutta(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta with one null-user entry already exists
	base := mustSeedBase(t, ctx)
	first := &models.CalcuttaEntry{
		Name:       "Anonymous Entry 1",
		UserID:     nil,
		CalcuttaID: base.calcutta.ID,
	}
	if err := base.calcuttaRepo.CreateEntry(ctx, first); err != nil {
		t.Fatalf("creating first null-user entry: %v", err)
	}

	// WHEN creating another null-user entry in the same calcutta
	second := &models.CalcuttaEntry{
		Name:       "Anonymous Entry 2",
		UserID:     nil,
		CalcuttaID: base.calcutta.ID,
	}
	err := base.calcuttaRepo.CreateEntry(ctx, second)

	// THEN no error occurs (NULL != NULL in partial index)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestThatGetEntriesReturnsEntriesOrderedByPointsDescending(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta with two entries (no teams, so both have 0 points)
	base := mustSeedBase(t, ctx)
	secondUser := mustSeedSecondUser(t, ctx, base.userRepo)
	entry1 := mustSeedEntry(t, ctx, base.calcuttaRepo, base.calcutta.ID, base.user.ID)
	entry2 := mustSeedEntry(t, ctx, base.calcuttaRepo, base.calcutta.ID, secondUser.ID)

	// WHEN retrieving entries
	entries, _, err := base.calcuttaRepo.GetEntries(ctx, base.calcutta.ID)
	if err != nil {
		t.Fatalf("getting entries: %v", err)
	}

	// THEN both entries are returned
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	// AND entries are in descending order by created_at (both have 0 points)
	_ = entry1
	_ = entry2
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, e.ID)
	}
	if len(ids) == 2 && ids[0] != entry2.ID {
		t.Errorf("expected entry2 (%s) first (created later), got %s", entry2.ID, ids[0])
	}
}

func TestThatGetEntryReturnsNotFoundForMissingID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta repository
	base := mustSeedBase(t, ctx)

	// WHEN getting an entry that does not exist
	_, err := base.calcuttaRepo.GetEntry(ctx, uuid.New().String())

	// THEN the error is a NotFoundError
	var notFound *apperrors.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected *apperrors.NotFoundError, got %T: %v", err, err)
	}
}
