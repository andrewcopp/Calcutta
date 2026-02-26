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

func TestThatCreatedPortfolioCanBeRetrievedByID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user, tournament, and pool exist
	userRepo := db.NewUserRepository(pool)
	tournamentRepo := db.NewTournamentRepository(pool)
	poolRepo := db.NewPoolRepository(pool)

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

	p := &models.Pool{
		TournamentID:         tournament.ID,
		OwnerID:              user.ID,
		CreatedBy:            user.ID,
		Name:                 "Test Pool",
		BudgetCredits:        100,
		MinTeams:             3,
		MaxTeams:             10,
		MaxInvestmentCredits: 50,
	}
	if err := poolRepo.Create(ctx, p); err != nil {
		t.Fatalf("creating pool: %v", err)
	}

	portfolio := &models.Portfolio{
		Name:   "Test Portfolio",
		UserID: &user.ID,
		PoolID: p.ID,
	}

	// WHEN creating and then retrieving the portfolio
	if err := poolRepo.CreatePortfolio(ctx, portfolio); err != nil {
		t.Fatalf("creating portfolio: %v", err)
	}

	got, err := poolRepo.GetPortfolio(ctx, portfolio.ID)
	if err != nil {
		t.Fatalf("getting portfolio: %v", err)
	}

	// THEN the retrieved portfolio has the correct name
	if got.Name != "Test Portfolio" {
		t.Errorf("expected portfolio name %q, got %q", "Test Portfolio", got.Name)
	}
}

func TestThatCreatePoolAssignsIDAndTimestamps(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	userRepo := db.NewUserRepository(pool)
	tournamentRepo := db.NewTournamentRepository(pool)
	poolRepo := db.NewPoolRepository(pool)

	user := &models.User{ID: uuid.New().String(), FirstName: "Test", LastName: "User", Status: "active"}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("creating user: %v", err)
	}
	tournament := &models.Tournament{ID: uuid.New().String(), Rounds: 6}
	if err := tournamentRepo.Create(ctx, tournament, "NCAA Tournament", 2026); err != nil {
		t.Fatalf("creating tournament: %v", err)
	}

	// WHEN creating a pool
	p := &models.Pool{
		TournamentID:         tournament.ID,
		OwnerID:              user.ID,
		CreatedBy:            user.ID,
		Name:                 "Test Pool",
		BudgetCredits:        100,
		MinTeams:             3,
		MaxTeams:             10,
		MaxInvestmentCredits: 50,
	}
	if err := poolRepo.Create(ctx, p); err != nil {
		t.Fatalf("creating pool: %v", err)
	}

	// THEN the pool has a non-empty ID and non-zero timestamps
	if p.ID == "" {
		t.Error("expected non-empty ID")
	}
	if p.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestThatCreatePoolRejectsBudgetCreditsOfZero(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	base := mustSeedBase(t, ctx)

	// WHEN creating a pool with zero budget credits
	p := &models.Pool{
		TournamentID:         base.tournament.ID,
		OwnerID:              base.user.ID,
		CreatedBy:            base.user.ID,
		Name:                 "Bad Pool",
		BudgetCredits:        0,
		MinTeams:             1,
		MaxTeams:             10,
		MaxInvestmentCredits: 0,
	}
	err := base.poolRepo.Create(ctx, p)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for zero budget credits, got nil")
	}
}

func TestThatCreatePoolRejectsMaxInvestmentCreditsExceedingBudget(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	base := mustSeedBase(t, ctx)

	// WHEN creating a pool where max_investment_credits > budget_credits
	p := &models.Pool{
		TournamentID:         base.tournament.ID,
		OwnerID:              base.user.ID,
		CreatedBy:            base.user.ID,
		Name:                 "Bad Pool",
		BudgetCredits:        50,
		MinTeams:             1,
		MaxTeams:             10,
		MaxInvestmentCredits: 100,
	}
	err := base.poolRepo.Create(ctx, p)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for max_investment_credits > budget_credits, got nil")
	}
}

func TestThatCreatePoolRejectsMinTeamsOfZero(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	base := mustSeedBase(t, ctx)

	// WHEN creating a pool with zero min teams
	p := &models.Pool{
		TournamentID:         base.tournament.ID,
		OwnerID:              base.user.ID,
		CreatedBy:            base.user.ID,
		Name:                 "Bad Pool",
		BudgetCredits:        100,
		MinTeams:             0,
		MaxTeams:             10,
		MaxInvestmentCredits: 50,
	}
	err := base.poolRepo.Create(ctx, p)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for zero min_teams, got nil")
	}
}

func TestThatCreatePoolRejectsMaxTeamsLessThanMinTeams(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	base := mustSeedBase(t, ctx)

	// WHEN creating a pool where max_teams < min_teams
	p := &models.Pool{
		TournamentID:         base.tournament.ID,
		OwnerID:              base.user.ID,
		CreatedBy:            base.user.ID,
		Name:                 "Bad Pool",
		BudgetCredits:        100,
		MinTeams:             5,
		MaxTeams:             3,
		MaxInvestmentCredits: 50,
	}
	err := base.poolRepo.Create(ctx, p)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for max_teams < min_teams, got nil")
	}
}

func TestThatCreatePoolRejectsInvalidVisibility(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	base := mustSeedBase(t, ctx)

	// WHEN creating a pool with invalid visibility
	p := &models.Pool{
		TournamentID:         base.tournament.ID,
		OwnerID:              base.user.ID,
		CreatedBy:            base.user.ID,
		Name:                 "Bad Pool",
		BudgetCredits:        100,
		MinTeams:             1,
		MaxTeams:             10,
		MaxInvestmentCredits: 50,
		Visibility:           "invalid",
	}
	err := base.poolRepo.Create(ctx, p)

	// THEN the error is non-nil (CHECK constraint violation)
	if err == nil {
		t.Error("expected error for invalid visibility, got nil")
	}
}

func TestThatCreatePoolDefaultsVisibilityToPrivate(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a user and tournament exist
	base := mustSeedBase(t, ctx)

	// WHEN creating a pool with empty visibility
	p := &models.Pool{
		TournamentID:         base.tournament.ID,
		OwnerID:              base.user.ID,
		CreatedBy:            base.user.ID,
		Name:                 "Default Vis Pool",
		BudgetCredits:        100,
		MinTeams:             1,
		MaxTeams:             10,
		MaxInvestmentCredits: 50,
		Visibility:           "",
	}
	if err := base.poolRepo.Create(ctx, p); err != nil {
		t.Fatalf("creating pool: %v", err)
	}

	// THEN the visibility is set to "private"
	got, err := base.poolRepo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("getting pool: %v", err)
	}
	if got.Visibility != "private" {
		t.Errorf("expected visibility %q, got %q", "private", got.Visibility)
	}
}

func TestThatUpdatePoolReturnsNotFoundForMissingID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool repository
	poolRepo := db.NewPoolRepository(pool)

	// WHEN updating a pool that does not exist
	p := &models.Pool{
		ID:                   uuid.New().String(),
		TournamentID:         uuid.New().String(),
		OwnerID:              uuid.New().String(),
		Name:                 "Does Not Exist",
		BudgetCredits:        100,
		MinTeams:             1,
		MaxTeams:             10,
		MaxInvestmentCredits: 50,
		Visibility:           "private",
	}
	err := poolRepo.Update(ctx, p)

	// THEN the error is a NotFoundError
	var notFound *apperrors.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected *apperrors.NotFoundError, got %T: %v", err, err)
	}
}

func TestThatUpdatePoolModifiesName(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool exists
	base := mustSeedBase(t, ctx)

	// WHEN updating the name
	base.pool.Name = "Updated Pool Name"
	if err := base.poolRepo.Update(ctx, base.pool); err != nil {
		t.Fatalf("updating pool: %v", err)
	}

	// THEN the retrieved pool has the new name
	got, err := base.poolRepo.GetByID(ctx, base.pool.ID)
	if err != nil {
		t.Fatalf("getting pool: %v", err)
	}
	if got.Name != "Updated Pool Name" {
		t.Errorf("expected name %q, got %q", "Updated Pool Name", got.Name)
	}
}
