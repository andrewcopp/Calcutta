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

// lifecycleFixture holds everything needed for lifecycle integration tests:
// a 4-team, 2-round tournament with 3 portfolios and investments.
type lifecycleFixture struct {
	poolRepo       *db.PoolRepository
	tournamentRepo *db.TournamentRepository

	commissioner *models.User
	alice        *models.User
	bob          *models.User
	charlie      *models.User

	tournament *models.Tournament
	pool       *models.Pool

	teamA *models.TournamentTeam // 1-seed
	teamB *models.TournamentTeam // 2-seed
	teamC *models.TournamentTeam // 3-seed
	teamD *models.TournamentTeam // 4-seed

	alicePortfolio   *models.Portfolio
	bobPortfolio     *models.Portfolio
	charliePortfolio *models.Portfolio
}

// mustSeedLifecycleTournament creates the full lifecycle fixture:
//
//	4 teams, 2 rounds, 3 portfolios with investments, scoring rules, and payouts.
//
// Scoring rules:
//
//	win_index 0 = 0 pts (baseline)
//	win_index 1 = 10 pts (semifinal win)
//	win_index 2 = 25 pts (final win)
//
// Investments:
//
//	Alice:   A=40, B=30, C=20, D=10
//	Bob:     A=10, B=20, C=30, D=40
//	Charlie: B=50, C=50
//
// Payouts: 1st=500c, 2nd=300c
func mustSeedLifecycleTournament(t *testing.T, ctx context.Context) lifecycleFixture {
	t.Helper()

	userRepo := db.NewUserRepository(pool)
	tournamentRepo := db.NewTournamentRepository(pool)
	schoolRepo := db.NewSchoolRepository(pool)
	poolRepo := db.NewPoolRepository(pool)

	// --- Users ---
	commissioner := mustCreateUser(t, ctx, userRepo, "Commissioner", "Admin")
	alice := mustCreateUser(t, ctx, userRepo, "Alice", "Player")
	bob := mustCreateUser(t, ctx, userRepo, "Bob", "Player")
	charlie := mustCreateUser(t, ctx, userRepo, "Charlie", "Player")

	// --- Tournament (2 rounds) ---
	tournament := &models.Tournament{
		ID:     uuid.New().String(),
		Rounds: 2,
	}
	if err := tournamentRepo.Create(ctx, tournament, "Mini Tournament", 2026); err != nil {
		t.Fatalf("creating tournament: %v", err)
	}

	// --- Schools + Teams ---
	teamA := mustCreateTeam(t, ctx, schoolRepo, tournamentRepo, tournament.ID, "Alpha University", 1)
	teamB := mustCreateTeam(t, ctx, schoolRepo, tournamentRepo, tournament.ID, "Bravo College", 2)
	teamC := mustCreateTeam(t, ctx, schoolRepo, tournamentRepo, tournament.ID, "Charlie State", 3)
	teamD := mustCreateTeam(t, ctx, schoolRepo, tournamentRepo, tournament.ID, "Delta Tech", 4)

	// --- Pool ---
	p := &models.Pool{
		TournamentID:         tournament.ID,
		OwnerID:              commissioner.ID,
		CreatedBy:            commissioner.ID,
		Name:                 "Lifecycle Test Pool",
		BudgetCredits:        100,
		MinTeams:             1,
		MaxTeams:             4,
		MaxInvestmentCredits: 50,
	}
	if err := poolRepo.Create(ctx, p); err != nil {
		t.Fatalf("creating pool: %v", err)
	}

	// --- Scoring Rules ---
	scoringRules := []struct {
		winIndex int
		points   int
	}{
		{0, 0},
		{1, 10},
		{2, 25},
	}
	for _, sr := range scoringRules {
		rule := &models.ScoringRule{
			PoolID:        p.ID,
			WinIndex:      sr.winIndex,
			PointsAwarded: sr.points,
		}
		if err := poolRepo.CreateScoringRule(ctx, rule); err != nil {
			t.Fatalf("creating scoring rule (win_index=%d): %v", sr.winIndex, err)
		}
	}

	// --- Portfolios ---
	alicePortfolio := mustCreatePortfolio(t, ctx, poolRepo, p.ID, alice.ID, "Alice's Portfolio")
	bobPortfolio := mustCreatePortfolio(t, ctx, poolRepo, p.ID, bob.ID, "Bob's Portfolio")
	charliePortfolio := mustCreatePortfolio(t, ctx, poolRepo, p.ID, charlie.ID, "Charlie's Portfolio")

	// --- Investments ---
	// Alice: A=40, B=30, C=20, D=10
	mustSetInvestments(t, ctx, poolRepo, alicePortfolio.ID, []investmentSpec{
		{teamA.ID, 40},
		{teamB.ID, 30},
		{teamC.ID, 20},
		{teamD.ID, 10},
	})

	// Bob: A=10, B=20, C=30, D=40
	mustSetInvestments(t, ctx, poolRepo, bobPortfolio.ID, []investmentSpec{
		{teamA.ID, 10},
		{teamB.ID, 20},
		{teamC.ID, 30},
		{teamD.ID, 40},
	})

	// Charlie: B=50, C=50
	mustSetInvestments(t, ctx, poolRepo, charliePortfolio.ID, []investmentSpec{
		{teamB.ID, 50},
		{teamC.ID, 50},
	})

	// --- Payouts ---
	payouts := []*models.PoolPayout{
		{Position: 1, AmountCents: 500},
		{Position: 2, AmountCents: 300},
	}
	if err := poolRepo.ReplacePayouts(ctx, p.ID, payouts); err != nil {
		t.Fatalf("creating payouts: %v", err)
	}

	return lifecycleFixture{
		poolRepo:         poolRepo,
		tournamentRepo:   tournamentRepo,
		commissioner:     commissioner,
		alice:            alice,
		bob:              bob,
		charlie:          charlie,
		tournament:       tournament,
		pool:             p,
		teamA:            teamA,
		teamB:            teamB,
		teamC:            teamC,
		teamD:            teamD,
		alicePortfolio:   alicePortfolio,
		bobPortfolio:     bobPortfolio,
		charliePortfolio: charliePortfolio,
	}
}

// mustAdvanceToRound1 simulates semifinal results:
// Team A wins (1 win), Team B wins (1 win), Team C eliminated, Team D eliminated.
func mustAdvanceToRound1(t *testing.T, ctx context.Context, f *lifecycleFixture) {
	t.Helper()

	f.teamA.Wins = 1
	f.teamB.Wins = 1
	f.teamC.IsEliminated = true
	f.teamD.IsEliminated = true

	for _, team := range []*models.TournamentTeam{f.teamA, f.teamB, f.teamC, f.teamD} {
		if err := f.tournamentRepo.UpdateTournamentTeam(ctx, team); err != nil {
			t.Fatalf("updating team %s: %v", team.ID, err)
		}
	}
}

// mustAdvanceToFinal simulates the full tournament:
// Team A wins final (2 wins, champion), Team B eliminated (1 win, runner-up).
func mustAdvanceToFinal(t *testing.T, ctx context.Context, f *lifecycleFixture) {
	t.Helper()

	mustAdvanceToRound1(t, ctx, f)

	f.teamA.Wins = 2
	f.teamB.IsEliminated = true

	for _, team := range []*models.TournamentTeam{f.teamA, f.teamB} {
		if err := f.tournamentRepo.UpdateTournamentTeam(ctx, team); err != nil {
			t.Fatalf("updating team %s: %v", team.ID, err)
		}
	}
}

// --- helpers ---

func mustCreateUser(t *testing.T, ctx context.Context, repo *db.UserRepository, first, last string) *models.User {
	t.Helper()
	user := &models.User{
		ID:        uuid.New().String(),
		FirstName: first,
		LastName:  last,
		Status:    "active",
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("creating user %s %s: %v", first, last, err)
	}
	return user
}

func mustCreateTeam(t *testing.T, ctx context.Context, schoolRepo *db.SchoolRepository, tournamentRepo *db.TournamentRepository, tournamentID, schoolName string, seed int) *models.TournamentTeam {
	t.Helper()
	school := &models.School{
		ID:   uuid.New().String(),
		Name: fmt.Sprintf("%s %s", schoolName, uuid.New().String()[:8]),
	}
	if err := schoolRepo.Create(ctx, school); err != nil {
		t.Fatalf("creating school %s: %v", schoolName, err)
	}
	team := &models.TournamentTeam{
		ID:           uuid.New().String(),
		SchoolID:     school.ID,
		TournamentID: tournamentID,
		Seed:         seed,
		Region:       "East",
	}
	if err := tournamentRepo.CreateTeam(ctx, team); err != nil {
		t.Fatalf("creating team (seed %d): %v", seed, err)
	}
	return team
}

func mustCreatePortfolio(t *testing.T, ctx context.Context, repo *db.PoolRepository, poolID, userID, name string) *models.Portfolio {
	t.Helper()
	portfolio := &models.Portfolio{
		Name:   name,
		UserID: &userID,
		PoolID: poolID,
	}
	if err := repo.CreatePortfolio(ctx, portfolio); err != nil {
		t.Fatalf("creating portfolio %s: %v", name, err)
	}
	return portfolio
}

type investmentSpec struct {
	teamID  string
	credits int
}

func mustSetInvestments(t *testing.T, ctx context.Context, repo *db.PoolRepository, portfolioID string, specs []investmentSpec) {
	t.Helper()
	investments := make([]*models.Investment, 0, len(specs))
	for _, s := range specs {
		investments = append(investments, &models.Investment{
			TeamID:  s.teamID,
			Credits: s.credits,
		})
	}
	if err := repo.ReplaceInvestments(ctx, portfolioID, investments); err != nil {
		t.Fatalf("setting investments for portfolio %s: %v", portfolioID, err)
	}
}
