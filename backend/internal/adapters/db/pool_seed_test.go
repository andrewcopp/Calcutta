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
	poolRepo       *db.PoolRepository
	user           *models.User
	tournament     *models.Tournament
	pool           *models.Pool
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

	return seedBase{
		userRepo:       userRepo,
		tournamentRepo: tournamentRepo,
		schoolRepo:     schoolRepo,
		poolRepo:       poolRepo,
		user:           user,
		tournament:     tournament,
		pool:           p,
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

func mustSeedPortfolio(t *testing.T, ctx context.Context, repo *db.PoolRepository, poolID, userID string) *models.Portfolio {
	t.Helper()

	portfolio := &models.Portfolio{
		Name:   "Portfolio " + uuid.New().String()[:8],
		UserID: &userID,
		PoolID: poolID,
	}
	if err := repo.CreatePortfolio(ctx, portfolio); err != nil {
		t.Fatalf("creating portfolio: %v", err)
	}
	return portfolio
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
