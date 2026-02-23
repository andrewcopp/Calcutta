//go:build integration

package db_test

import (
	"context"
	"os"
	"testing"

	db "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
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
