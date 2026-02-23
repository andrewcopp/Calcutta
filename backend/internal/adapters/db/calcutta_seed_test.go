//go:build integration

package db_test

import (
	"context"
	"fmt"
	"testing"

	db "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/google/uuid"
)

type seedBase struct {
	userRepo       *db.UserRepository
	tournamentRepo *db.TournamentRepository
	schoolRepo     *db.SchoolRepository
	calcuttaRepo   *db.CalcuttaRepository
	user           *models.User
	tournament     *models.Tournament
	calcutta       *models.Calcutta
}

type seedWithTeams struct {
	seedBase
	teams []*models.TournamentTeam
}

func mustSeedBase(t *testing.T, ctx context.Context) seedBase {
	t.Helper()

	userRepo := db.NewUserRepository(pool)
	tournamentRepo := db.NewTournamentRepository(pool)
	schoolRepo := db.NewSchoolRepository(pool)
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

	return seedBase{
		userRepo:       userRepo,
		tournamentRepo: tournamentRepo,
		schoolRepo:     schoolRepo,
		calcuttaRepo:   calcuttaRepo,
		user:           user,
		tournament:     tournament,
		calcutta:       calcutta,
	}
}

func mustSeedWithTeams(t *testing.T, ctx context.Context, n int) seedWithTeams {
	t.Helper()

	base := mustSeedBase(t, ctx)

	teams := make([]*models.TournamentTeam, 0, n)
	for i := 0; i < n; i++ {
		school := &models.School{
			ID:   uuid.New().String(),
			Name: fmt.Sprintf("School %d %s", i+1, uuid.New().String()[:8]),
		}
		if err := base.schoolRepo.Create(ctx, school); err != nil {
			t.Fatalf("creating school %d: %v", i+1, err)
		}

		team := &models.TournamentTeam{
			ID:           uuid.New().String(),
			SchoolID:     school.ID,
			TournamentID: base.tournament.ID,
			Seed:         (i % 16) + 1,
			Region:       "East",
		}
		if err := base.tournamentRepo.CreateTeam(ctx, team); err != nil {
			t.Fatalf("creating team %d: %v", i+1, err)
		}
		teams = append(teams, team)
	}

	return seedWithTeams{seedBase: base, teams: teams}
}

func mustSeedEntry(t *testing.T, ctx context.Context, repo *db.CalcuttaRepository, calcuttaID, userID string) *models.CalcuttaEntry {
	t.Helper()

	entry := &models.CalcuttaEntry{
		Name:       "Entry " + uuid.New().String()[:8],
		UserID:     &userID,
		CalcuttaID: calcuttaID,
	}
	if err := repo.CreateEntry(ctx, entry); err != nil {
		t.Fatalf("creating entry: %v", err)
	}
	return entry
}

func mustSeedSecondUser(t *testing.T, ctx context.Context, repo *db.UserRepository) *models.User {
	t.Helper()

	user := &models.User{
		ID:        uuid.New().String(),
		FirstName: "Second",
		LastName:  "User",
		Status:    "active",
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("creating second user: %v", err)
	}
	return user
}
