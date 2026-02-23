//go:build integration

package db_test

import (
	"context"
	"errors"
	"os"
	"testing"

	db "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	p, cleanup, err := testutil.StartPostgresContainer(ctx)
	if err != nil {
		panic("starting test container: " + err.Error())
	}
	pool = p

	code := m.Run()
	cleanup()
	os.Exit(code)
}

func TestThatCreatedEntryCanBeRetrievedByID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user, tournament, and calcutta exist
	userRepo := db.NewUserRepository(pool)
	tournamentRepo := db.NewTournamentRepository(pool)
	calcuttaRepo := db.NewCalcuttaRepository(pool)

	user := &models.User{
		ID:        uuid.New().String(),
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("creating user: %v", err)
	}

	tournament := &models.Tournament{
		ID:     uuid.New().String(),
		Rounds: 6,
	}
	if err := tournamentRepo.Create(ctx, tournament, "NCAA Tournament", 2026); err != nil {
		t.Fatalf("creating tournament: %v", err)
	}

	calcutta := &models.Calcutta{
		TournamentID: tournament.ID,
		OwnerID:      user.ID,
		CreatedBy:    user.ID,
		Name:         "Test Pool",
		BudgetPoints: 100,
		MinTeams:     3,
		MaxTeams:     10,
		MaxBidPoints: 50,
	}
	if err := calcuttaRepo.Create(ctx, calcutta); err != nil {
		t.Fatalf("creating calcutta: %v", err)
	}

	entry := &models.CalcuttaEntry{
		Name:       "Test Entry",
		UserID:     &user.ID,
		CalcuttaID: calcutta.ID,
	}

	// WHEN creating and then retrieving the entry
	if err := calcuttaRepo.CreateEntry(ctx, entry); err != nil {
		t.Fatalf("creating entry: %v", err)
	}

	got, err := calcuttaRepo.GetEntry(ctx, entry.ID)
	if err != nil {
		t.Fatalf("getting entry: %v", err)
	}

	// THEN the retrieved entry has the correct name
	if got.Name != "Test Entry" {
		t.Errorf("expected entry name %q, got %q", "Test Entry", got.Name)
	}
}

func TestThatCreateCalcuttaAssignsIDAndTimestamps(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	userRepo := db.NewUserRepository(pool)
	tournamentRepo := db.NewTournamentRepository(pool)
	calcuttaRepo := db.NewCalcuttaRepository(pool)

	user := &models.User{ID: uuid.New().String(), FirstName: "Test", LastName: "User", Status: "active"}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("creating user: %v", err)
	}
	tournament := &models.Tournament{ID: uuid.New().String(), Rounds: 6}
	if err := tournamentRepo.Create(ctx, tournament, "NCAA Tournament", 2026); err != nil {
		t.Fatalf("creating tournament: %v", err)
	}

	// WHEN creating a calcutta
	calcutta := &models.Calcutta{
		TournamentID: tournament.ID,
		OwnerID:      user.ID,
		CreatedBy:    user.ID,
		Name:         "Test Pool",
		BudgetPoints: 100,
		MinTeams:     3,
		MaxTeams:     10,
		MaxBidPoints: 50,
	}
	if err := calcuttaRepo.Create(ctx, calcutta); err != nil {
		t.Fatalf("creating calcutta: %v", err)
	}

	// THEN the calcutta has a non-empty ID and non-zero timestamps
	if calcutta.ID == "" {
		t.Error("expected non-empty ID")
	}
	if calcutta.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if calcutta.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestThatCreateCalcuttaRejectsBudgetPointsOfZero(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	base := mustSeedBase(t, ctx)

	// WHEN creating a calcutta with zero budget points
	calcutta := &models.Calcutta{
		TournamentID: base.tournament.ID,
		OwnerID:      base.user.ID,
		CreatedBy:    base.user.ID,
		Name:         "Bad Pool",
		BudgetPoints: 0,
		MinTeams:     1,
		MaxTeams:     10,
		MaxBidPoints: 0,
	}
	err := base.calcuttaRepo.Create(ctx, calcutta)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for zero budget points, got nil")
	}
}

func TestThatCreateCalcuttaRejectsMaxBidExceedingBudget(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	base := mustSeedBase(t, ctx)

	// WHEN creating a calcutta where max_bid_points > budget_points
	calcutta := &models.Calcutta{
		TournamentID: base.tournament.ID,
		OwnerID:      base.user.ID,
		CreatedBy:    base.user.ID,
		Name:         "Bad Pool",
		BudgetPoints: 50,
		MinTeams:     1,
		MaxTeams:     10,
		MaxBidPoints: 100,
	}
	err := base.calcuttaRepo.Create(ctx, calcutta)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for max_bid_points > budget_points, got nil")
	}
}

func TestThatCreateCalcuttaRejectsMinTeamsOfZero(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	base := mustSeedBase(t, ctx)

	// WHEN creating a calcutta with zero min teams
	calcutta := &models.Calcutta{
		TournamentID: base.tournament.ID,
		OwnerID:      base.user.ID,
		CreatedBy:    base.user.ID,
		Name:         "Bad Pool",
		BudgetPoints: 100,
		MinTeams:     0,
		MaxTeams:     10,
		MaxBidPoints: 50,
	}
	err := base.calcuttaRepo.Create(ctx, calcutta)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for zero min_teams, got nil")
	}
}

func TestThatCreateCalcuttaRejectsMaxTeamsLessThanMinTeams(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	base := mustSeedBase(t, ctx)

	// WHEN creating a calcutta where max_teams < min_teams
	calcutta := &models.Calcutta{
		TournamentID: base.tournament.ID,
		OwnerID:      base.user.ID,
		CreatedBy:    base.user.ID,
		Name:         "Bad Pool",
		BudgetPoints: 100,
		MinTeams:     5,
		MaxTeams:     3,
		MaxBidPoints: 50,
	}
	err := base.calcuttaRepo.Create(ctx, calcutta)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for max_teams < min_teams, got nil")
	}
}

func TestThatCreateCalcuttaRejectsInvalidVisibility(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	base := mustSeedBase(t, ctx)

	// WHEN creating a calcutta with invalid visibility
	calcutta := &models.Calcutta{
		TournamentID: base.tournament.ID,
		OwnerID:      base.user.ID,
		CreatedBy:    base.user.ID,
		Name:         "Bad Pool",
		BudgetPoints: 100,
		MinTeams:     1,
		MaxTeams:     10,
		MaxBidPoints: 50,
		Visibility:   "invalid",
	}
	err := base.calcuttaRepo.Create(ctx, calcutta)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for invalid visibility, got nil")
	}
}

func TestThatCreateCalcuttaDefaultsVisibilityToPrivate(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	base := mustSeedBase(t, ctx)

	// WHEN creating a calcutta with empty visibility
	calcutta := &models.Calcutta{
		TournamentID: base.tournament.ID,
		OwnerID:      base.user.ID,
		CreatedBy:    base.user.ID,
		Name:         "Default Vis Pool",
		BudgetPoints: 100,
		MinTeams:     1,
		MaxTeams:     10,
		MaxBidPoints: 50,
		Visibility:   "",
	}
	if err := base.calcuttaRepo.Create(ctx, calcutta); err != nil {
		t.Fatalf("creating calcutta: %v", err)
	}

	// THEN the visibility is set to "private"
	got, err := base.calcuttaRepo.GetByID(ctx, calcutta.ID)
	if err != nil {
		t.Fatalf("getting calcutta: %v", err)
	}
	if got.Visibility != "private" {
		t.Errorf("expected visibility %q, got %q", "private", got.Visibility)
	}
}

func TestThatUpdateCalcuttaReturnsNotFoundForMissingID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta repository
	calcuttaRepo := db.NewCalcuttaRepository(pool)

	// WHEN updating a calcutta that does not exist
	calcutta := &models.Calcutta{
		ID:           uuid.New().String(),
		TournamentID: uuid.New().String(),
		OwnerID:      uuid.New().String(),
		Name:         "Does Not Exist",
		BudgetPoints: 100,
		MinTeams:     1,
		MaxTeams:     10,
		MaxBidPoints: 50,
		Visibility:   "private",
	}
	err := calcuttaRepo.Update(ctx, calcutta)

	// THEN the error is a NotFoundError
	var notFound *apperrors.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected *apperrors.NotFoundError, got %T: %v", err, err)
	}
}

func TestThatUpdateCalcuttaModifiesName(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta exists
	base := mustSeedBase(t, ctx)

	// WHEN updating the name
	base.calcutta.Name = "Updated Pool Name"
	if err := base.calcuttaRepo.Update(ctx, base.calcutta); err != nil {
		t.Fatalf("updating calcutta: %v", err)
	}

	// THEN the retrieved calcutta has the new name
	got, err := base.calcuttaRepo.GetByID(ctx, base.calcutta.ID)
	if err != nil {
		t.Fatalf("getting calcutta: %v", err)
	}
	if got.Name != "Updated Pool Name" {
		t.Errorf("expected name %q, got %q", "Updated Pool Name", got.Name)
	}
}
