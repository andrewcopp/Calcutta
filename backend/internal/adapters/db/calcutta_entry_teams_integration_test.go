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

func TestThatReplaceEntryTeamsSoftDeletesOldAndInsertsNew(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN an entry with initial teams
	seed := mustSeedWithTeams(t, ctx, 3)
	entry := mustSeedEntry(t, ctx, seed.calcuttaRepo, seed.calcutta.ID, seed.user.ID)

	initial := []*models.CalcuttaEntryTeam{
		{TeamID: seed.teams[0].ID, BidPoints: 20},
		{TeamID: seed.teams[1].ID, BidPoints: 30},
	}
	if err := seed.calcuttaRepo.ReplaceEntryTeams(ctx, entry.ID, initial); err != nil {
		t.Fatalf("creating initial entry teams: %v", err)
	}

	// WHEN replacing with different teams
	replacement := []*models.CalcuttaEntryTeam{
		{TeamID: seed.teams[2].ID, BidPoints: 50},
	}
	if err := seed.calcuttaRepo.ReplaceEntryTeams(ctx, entry.ID, replacement); err != nil {
		t.Fatalf("replacing entry teams: %v", err)
	}

	// THEN only the replacement team is returned
	got, err := seed.calcuttaRepo.GetEntryTeams(ctx, entry.ID)
	if err != nil {
		t.Fatalf("getting entry teams: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 entry team, got %d", len(got))
	}
}

func TestThatReplaceEntryTeamsRejectsZeroBidPoints(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN an entry and a team exist
	seed := mustSeedWithTeams(t, ctx, 1)
	entry := mustSeedEntry(t, ctx, seed.calcuttaRepo, seed.calcutta.ID, seed.user.ID)

	// WHEN replacing with zero bid points
	teams := []*models.CalcuttaEntryTeam{
		{TeamID: seed.teams[0].ID, BidPoints: 0},
	}
	err := seed.calcuttaRepo.ReplaceEntryTeams(ctx, entry.ID, teams)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for zero bid_points, got nil")
	}
}

func TestThatReplaceEntryTeamsRejectsDuplicateTeamInSameEntry(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN an entry and a team exist
	seed := mustSeedWithTeams(t, ctx, 1)
	entry := mustSeedEntry(t, ctx, seed.calcuttaRepo, seed.calcutta.ID, seed.user.ID)

	// WHEN replacing with the same team twice
	teams := []*models.CalcuttaEntryTeam{
		{TeamID: seed.teams[0].ID, BidPoints: 20},
		{TeamID: seed.teams[0].ID, BidPoints: 30},
	}
	err := seed.calcuttaRepo.ReplaceEntryTeams(ctx, entry.ID, teams)

	// THEN the error is non-nil (UNIQUE partial index violation)
	if err == nil {
		t.Error("expected error for duplicate team in same entry, got nil")
	}
}

func TestThatReplaceEntryTeamsReturnsNotFoundForMissingEntry(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta repository
	seed := mustSeedWithTeams(t, ctx, 1)

	// WHEN replacing entry teams for a non-existent entry
	teams := []*models.CalcuttaEntryTeam{
		{TeamID: seed.teams[0].ID, BidPoints: 10},
	}
	err := seed.calcuttaRepo.ReplaceEntryTeams(ctx, uuid.New().String(), teams)

	// THEN the error is a NotFoundError
	var notFound *apperrors.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected *apperrors.NotFoundError, got %T: %v", err, err)
	}
}

func TestThatGetEntryTeamsJoinsTeamAndSchoolData(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN an entry with one team
	seed := mustSeedWithTeams(t, ctx, 1)
	entry := mustSeedEntry(t, ctx, seed.calcuttaRepo, seed.calcutta.ID, seed.user.ID)

	teams := []*models.CalcuttaEntryTeam{
		{TeamID: seed.teams[0].ID, BidPoints: 25},
	}
	if err := seed.calcuttaRepo.ReplaceEntryTeams(ctx, entry.ID, teams); err != nil {
		t.Fatalf("replacing entry teams: %v", err)
	}

	// WHEN retrieving entry teams
	got, err := seed.calcuttaRepo.GetEntryTeams(ctx, entry.ID)
	if err != nil {
		t.Fatalf("getting entry teams: %v", err)
	}

	// THEN the team includes joined TournamentTeam with School data
	if len(got) != 1 {
		t.Fatalf("expected 1 entry team, got %d", len(got))
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

func TestThatGetEntryTeamsByEntryIDsGroupsByEntryID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN two entries, each with different teams
	seed := mustSeedWithTeams(t, ctx, 3)
	secondUser := mustSeedSecondUser(t, ctx, seed.userRepo)
	entry1 := mustSeedEntry(t, ctx, seed.calcuttaRepo, seed.calcutta.ID, seed.user.ID)
	entry2 := mustSeedEntry(t, ctx, seed.calcuttaRepo, seed.calcutta.ID, secondUser.ID)

	teams1 := []*models.CalcuttaEntryTeam{
		{TeamID: seed.teams[0].ID, BidPoints: 20},
		{TeamID: seed.teams[1].ID, BidPoints: 30},
	}
	if err := seed.calcuttaRepo.ReplaceEntryTeams(ctx, entry1.ID, teams1); err != nil {
		t.Fatalf("replacing entry1 teams: %v", err)
	}
	teams2 := []*models.CalcuttaEntryTeam{
		{TeamID: seed.teams[2].ID, BidPoints: 50},
	}
	if err := seed.calcuttaRepo.ReplaceEntryTeams(ctx, entry2.ID, teams2); err != nil {
		t.Fatalf("replacing entry2 teams: %v", err)
	}

	// WHEN retrieving teams by entry IDs
	got, err := seed.calcuttaRepo.GetEntryTeamsByEntryIDs(ctx, []string{entry1.ID, entry2.ID})
	if err != nil {
		t.Fatalf("getting entry teams by IDs: %v", err)
	}

	// THEN the map has correct groupings
	if len(got[entry1.ID]) != 2 {
		t.Errorf("expected 2 teams for entry1, got %d", len(got[entry1.ID]))
	}
	if len(got[entry2.ID]) != 1 {
		t.Errorf("expected 1 team for entry2, got %d", len(got[entry2.ID]))
	}
}

func TestThatGetEntryTeamsByEntryIDsReturnsEmptyMapForEmptyInput(t *testing.T) {
	ctx := context.Background()

	// GIVEN a calcutta repository
	seed := mustSeedBase(t, ctx)
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// WHEN passing an empty slice of entry IDs
	got, err := seed.calcuttaRepo.GetEntryTeamsByEntryIDs(ctx, []string{})

	// THEN an empty map is returned with no error
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %d entries", len(got))
	}
}
