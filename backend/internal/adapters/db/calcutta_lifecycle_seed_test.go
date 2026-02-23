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
// a 4-team, 2-round tournament with 3 entries and bids.
type lifecycleFixture struct {
	calcuttaRepo   *db.CalcuttaRepository
	tournamentRepo *db.TournamentRepository

	commissioner *models.User
	alice        *models.User
	bob          *models.User
	charlie      *models.User

	tournament *models.Tournament
	calcutta   *models.Calcutta

	teamA *models.TournamentTeam // 1-seed
	teamB *models.TournamentTeam // 2-seed
	teamC *models.TournamentTeam // 3-seed
	teamD *models.TournamentTeam // 4-seed

	aliceEntry   *models.CalcuttaEntry
	bobEntry     *models.CalcuttaEntry
	charlieEntry *models.CalcuttaEntry
}

// mustSeedLifecycleTournament creates the full lifecycle fixture:
//
//	4 teams, 2 rounds, 3 entries with bids, scoring rules, and payouts.
//
// Scoring rules:
//
//	win_index 0 = 0 pts (baseline)
//	win_index 1 = 10 pts (semifinal win)
//	win_index 2 = 25 pts (final win)
//
// Bids:
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
	calcuttaRepo := db.NewCalcuttaRepository(pool)

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

	// --- Calcutta ---
	calcutta := &models.Calcutta{
		TournamentID: tournament.ID,
		OwnerID:      commissioner.ID,
		CreatedBy:    commissioner.ID,
		Name:         "Lifecycle Test Pool",
		BudgetPoints: 100,
		MinTeams:     1,
		MaxTeams:     4,
		MaxBidPoints: 50,
	}
	if err := calcuttaRepo.Create(ctx, calcutta); err != nil {
		t.Fatalf("creating calcutta: %v", err)
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
		round := &models.CalcuttaRound{
			CalcuttaID: calcutta.ID,
			Round:      sr.winIndex,
			Points:     sr.points,
		}
		if err := calcuttaRepo.CreateRound(ctx, round); err != nil {
			t.Fatalf("creating scoring rule (win_index=%d): %v", sr.winIndex, err)
		}
	}

	// --- Entries ---
	aliceEntry := mustCreateEntry(t, ctx, calcuttaRepo, calcutta.ID, alice.ID, "Alice's Entry")
	bobEntry := mustCreateEntry(t, ctx, calcuttaRepo, calcutta.ID, bob.ID, "Bob's Entry")
	charlieEntry := mustCreateEntry(t, ctx, calcuttaRepo, calcutta.ID, charlie.ID, "Charlie's Entry")

	// --- Bids ---
	// Alice: A=40, B=30, C=20, D=10
	mustSetBids(t, ctx, calcuttaRepo, aliceEntry.ID, []bidSpec{
		{teamA.ID, 40},
		{teamB.ID, 30},
		{teamC.ID, 20},
		{teamD.ID, 10},
	})

	// Bob: A=10, B=20, C=30, D=40
	mustSetBids(t, ctx, calcuttaRepo, bobEntry.ID, []bidSpec{
		{teamA.ID, 10},
		{teamB.ID, 20},
		{teamC.ID, 30},
		{teamD.ID, 40},
	})

	// Charlie: B=50, C=50
	mustSetBids(t, ctx, calcuttaRepo, charlieEntry.ID, []bidSpec{
		{teamB.ID, 50},
		{teamC.ID, 50},
	})

	// --- Payouts ---
	payouts := []*models.CalcuttaPayout{
		{Position: 1, AmountCents: 500},
		{Position: 2, AmountCents: 300},
	}
	if err := calcuttaRepo.ReplacePayouts(ctx, calcutta.ID, payouts); err != nil {
		t.Fatalf("creating payouts: %v", err)
	}

	return lifecycleFixture{
		calcuttaRepo:   calcuttaRepo,
		tournamentRepo: tournamentRepo,
		commissioner:   commissioner,
		alice:          alice,
		bob:            bob,
		charlie:        charlie,
		tournament:     tournament,
		calcutta:       calcutta,
		teamA:          teamA,
		teamB:          teamB,
		teamC:          teamC,
		teamD:          teamD,
		aliceEntry:     aliceEntry,
		bobEntry:       bobEntry,
		charlieEntry:   charlieEntry,
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

func mustCreateEntry(t *testing.T, ctx context.Context, repo *db.CalcuttaRepository, calcuttaID, userID, name string) *models.CalcuttaEntry {
	t.Helper()
	entry := &models.CalcuttaEntry{
		Name:       name,
		UserID:     &userID,
		CalcuttaID: calcuttaID,
	}
	if err := repo.CreateEntry(ctx, entry); err != nil {
		t.Fatalf("creating entry %s: %v", name, err)
	}
	return entry
}

type bidSpec struct {
	teamID string
	points int
}

func mustSetBids(t *testing.T, ctx context.Context, repo *db.CalcuttaRepository, entryID string, bids []bidSpec) {
	t.Helper()
	teams := make([]*models.CalcuttaEntryTeam, 0, len(bids))
	for _, b := range bids {
		teams = append(teams, &models.CalcuttaEntryTeam{
			TeamID:    b.teamID,
			BidPoints: b.points,
		})
	}
	if err := repo.ReplaceEntryTeams(ctx, entryID, teams); err != nil {
		t.Fatalf("setting bids for entry %s: %v", entryID, err)
	}
}
