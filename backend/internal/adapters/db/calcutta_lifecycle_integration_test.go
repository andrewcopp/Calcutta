//go:build integration

package db_test

import (
	"context"
	"math"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
)

const floatEpsilon = 0.001

func assertFloat(t *testing.T, got, want float64, msg string) {
	t.Helper()
	if math.Abs(got-want) > floatEpsilon {
		t.Errorf("%s: got %.4f, want %.4f", msg, got, want)
	}
}

func findEntry(entries []*models.CalcuttaEntry, id string) *models.CalcuttaEntry {
	for _, e := range entries {
		if e.ID == id {
			return e
		}
	}
	return nil
}

func findPortfolioTeam(teams []*models.CalcuttaPortfolioTeam, teamID string) *models.CalcuttaPortfolioTeam {
	for _, pt := range teams {
		if pt.TeamID == teamID {
			return pt
		}
	}
	return nil
}

// =============================================================================
// Phase 1: Pre-Tournament (0 wins, all teams alive)
// =============================================================================

func TestThatLifecyclePreTournamentEntriesHaveZeroTotalPoints(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta before the tournament starts (all teams have 0 wins)
	f := mustSeedLifecycleTournament(t, ctx)

	// WHEN listing entries
	entries, err := f.calcuttaRepo.GetEntries(ctx, f.calcutta.ID)
	if err != nil {
		t.Fatalf("getting entries: %v", err)
	}

	// THEN all 3 entries have 0 total points
	for _, e := range entries {
		assertFloat(t, e.TotalPoints, 0.0, "entry "+e.Name+" total points")
	}
	_ = f
}

func TestThatLifecyclePreTournamentPortfolioTeamsHaveZeroActualPoints(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta before the tournament starts
	f := mustSeedLifecycleTournament(t, ctx)

	// WHEN getting Alice's portfolio teams
	teams, err := f.calcuttaRepo.GetPortfolioTeams(ctx, f.aliceEntry.ID)
	if err != nil {
		t.Fatalf("getting portfolio teams: %v", err)
	}

	// THEN all of Alice's portfolio teams have 0 actual points
	for _, pt := range teams {
		assertFloat(t, pt.ActualPoints, 0.0, "portfolio team actual points")
	}
}

func TestThatLifecyclePreTournamentOwnershipPercentagesAreCorrect(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta before the tournament starts
	f := mustSeedLifecycleTournament(t, ctx)

	// WHEN getting Alice's portfolio teams
	teams, err := f.calcuttaRepo.GetPortfolioTeams(ctx, f.aliceEntry.ID)
	if err != nil {
		t.Fatalf("getting portfolio teams: %v", err)
	}

	// THEN Alice's ownership percentages are correct:
	//   A=40/50=0.80, B=30/100=0.30, C=20/100=0.20, D=10/50=0.20
	expected := map[string]float64{
		f.teamA.ID: 0.80,
		f.teamB.ID: 0.30,
		f.teamC.ID: 0.20,
		f.teamD.ID: 0.20,
	}
	for teamID, wantPct := range expected {
		pt := findPortfolioTeam(teams, teamID)
		if pt == nil {
			t.Fatalf("Alice missing portfolio team for team %s", teamID)
		}
		assertFloat(t, pt.OwnershipPercentage, wantPct, "ownership for team "+teamID)
	}
}

func TestThatLifecyclePreTournamentExpectedPointsReflectMaxPossible(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta before the tournament starts (max points per team = 35)
	f := mustSeedLifecycleTournament(t, ctx)

	// WHEN getting Alice's portfolio teams
	teams, err := f.calcuttaRepo.GetPortfolioTeams(ctx, f.aliceEntry.ID)
	if err != nil {
		t.Fatalf("getting portfolio teams: %v", err)
	}

	// THEN each team's expected = ownership * 35
	expected := map[string]float64{
		f.teamA.ID: 0.80 * 35, // 28.0
		f.teamB.ID: 0.30 * 35, // 10.5
		f.teamC.ID: 0.20 * 35, // 7.0
		f.teamD.ID: 0.20 * 35, // 7.0
	}
	for teamID, wantExp := range expected {
		pt := findPortfolioTeam(teams, teamID)
		if pt == nil {
			t.Fatalf("Alice missing portfolio team for team %s", teamID)
		}
		assertFloat(t, pt.ExpectedPoints, wantExp, "expected points for team "+teamID)
	}
}

func TestThatLifecyclePreTournamentPortfolioMaxPointsEqualsSumOfExpected(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta before the tournament starts
	f := mustSeedLifecycleTournament(t, ctx)

	// WHEN getting Alice's portfolio
	portfolio, err := f.calcuttaRepo.GetPortfolio(ctx, f.aliceEntry.ID)
	if err != nil {
		t.Fatalf("getting portfolio: %v", err)
	}

	// THEN Alice's max points = 28.0 + 10.5 + 7.0 + 7.0 = 52.5
	assertFloat(t, portfolio.MaximumPoints, 52.5, "Alice portfolio max points")
}

// =============================================================================
// Phase 3: Mid-Tournament (Team A, B win round 1; Team C, D eliminated)
// =============================================================================

func TestThatLifecycleMidTournamentWinnersHaveCorrectActualPoints(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta mid-tournament (round 1 complete)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToRound1(t, ctx, &f)

	// WHEN getting Alice's portfolio teams
	teams, err := f.calcuttaRepo.GetPortfolioTeams(ctx, f.aliceEntry.ID)
	if err != nil {
		t.Fatalf("getting portfolio teams: %v", err)
	}

	// THEN Alice's actual points: A=0.80*10=8.0, B=0.30*10=3.0, C=0.0, D=0.0
	expected := map[string]float64{
		f.teamA.ID: 8.0,
		f.teamB.ID: 3.0,
		f.teamC.ID: 0.0,
		f.teamD.ID: 0.0,
	}
	for teamID, wantPts := range expected {
		pt := findPortfolioTeam(teams, teamID)
		if pt == nil {
			t.Fatalf("Alice missing portfolio team for team %s", teamID)
		}
		assertFloat(t, pt.ActualPoints, wantPts, "actual points for team "+teamID)
	}
}

func TestThatLifecycleMidTournamentEliminatedTeamsLockExpectedPoints(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta mid-tournament (round 1 complete)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToRound1(t, ctx, &f)

	// WHEN getting Alice's portfolio teams
	teams, err := f.calcuttaRepo.GetPortfolioTeams(ctx, f.aliceEntry.ID)
	if err != nil {
		t.Fatalf("getting portfolio teams: %v", err)
	}

	// THEN eliminated teams have expected = actual = 0
	ptC := findPortfolioTeam(teams, f.teamC.ID)
	if ptC == nil {
		t.Fatalf("Alice missing portfolio team for team C")
	}
	assertFloat(t, ptC.ExpectedPoints, 0.0, "eliminated Team C expected points")

	ptD := findPortfolioTeam(teams, f.teamD.ID)
	if ptD == nil {
		t.Fatalf("Alice missing portfolio team for team D")
	}
	assertFloat(t, ptD.ExpectedPoints, 0.0, "eliminated Team D expected points")
}

func TestThatLifecycleMidTournamentSurvivingTeamsKeepMaxExpectedPoints(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta mid-tournament (round 1 complete)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToRound1(t, ctx, &f)

	// WHEN getting Alice's portfolio teams
	teams, err := f.calcuttaRepo.GetPortfolioTeams(ctx, f.aliceEntry.ID)
	if err != nil {
		t.Fatalf("getting portfolio teams: %v", err)
	}

	// THEN surviving teams keep max expected: A=0.80*35=28.0, B=0.30*35=10.5
	ptA := findPortfolioTeam(teams, f.teamA.ID)
	if ptA == nil {
		t.Fatalf("Alice missing portfolio team for team A")
	}
	assertFloat(t, ptA.ExpectedPoints, 28.0, "surviving Team A expected points")

	ptB := findPortfolioTeam(teams, f.teamB.ID)
	if ptB == nil {
		t.Fatalf("Alice missing portfolio team for team B")
	}
	assertFloat(t, ptB.ExpectedPoints, 10.5, "surviving Team B expected points")
}

func TestThatLifecycleMidTournamentTotalPointsReflectProgress(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta mid-tournament (round 1 complete)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToRound1(t, ctx, &f)

	// WHEN listing entries
	entries, err := f.calcuttaRepo.GetEntries(ctx, f.calcutta.ID)
	if err != nil {
		t.Fatalf("getting entries: %v", err)
	}

	// THEN Alice=11.0, Bob=4.0, Charlie=5.0
	aliceE := findEntry(entries, f.aliceEntry.ID)
	if aliceE == nil {
		t.Fatalf("Alice entry not found")
	}
	assertFloat(t, aliceE.TotalPoints, 11.0, "Alice total points")
}

func TestThatLifecycleMidTournamentEntriesRankByTotalPointsDescending(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta mid-tournament (round 1 complete)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToRound1(t, ctx, &f)

	// WHEN listing entries (returned sorted by total_points DESC)
	entries, err := f.calcuttaRepo.GetEntries(ctx, f.calcutta.ID)
	if err != nil {
		t.Fatalf("getting entries: %v", err)
	}

	// THEN order is Alice (11.0), Charlie (5.0), Bob (4.0)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].ID != f.aliceEntry.ID {
		t.Errorf("expected Alice first, got %s", entries[0].Name)
	}
	if entries[1].ID != f.charlieEntry.ID {
		t.Errorf("expected Charlie second, got %s", entries[1].Name)
	}
	if entries[2].ID != f.bobEntry.ID {
		t.Errorf("expected Bob third, got %s", entries[2].Name)
	}
}

func TestThatLifecycleMidTournamentSharedTeamSplitsPointsByOwnership(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a calcutta mid-tournament (round 1 complete)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToRound1(t, ctx, &f)

	// WHEN getting portfolio teams for all 3 entries
	aliceTeams, err := f.calcuttaRepo.GetPortfolioTeams(ctx, f.aliceEntry.ID)
	if err != nil {
		t.Fatalf("getting Alice portfolio teams: %v", err)
	}
	bobTeams, err := f.calcuttaRepo.GetPortfolioTeams(ctx, f.bobEntry.ID)
	if err != nil {
		t.Fatalf("getting Bob portfolio teams: %v", err)
	}
	charlieTeams, err := f.calcuttaRepo.GetPortfolioTeams(ctx, f.charlieEntry.ID)
	if err != nil {
		t.Fatalf("getting Charlie portfolio teams: %v", err)
	}

	// THEN Team B actual points: Alice=0.30*10=3.0, Bob=0.20*10=2.0, Charlie=0.50*10=5.0
	aliceBPt := findPortfolioTeam(aliceTeams, f.teamB.ID)
	bobBPt := findPortfolioTeam(bobTeams, f.teamB.ID)
	charlieBPt := findPortfolioTeam(charlieTeams, f.teamB.ID)

	assertFloat(t, aliceBPt.ActualPoints, 3.0, "Alice Team B actual points")
	assertFloat(t, bobBPt.ActualPoints, 2.0, "Bob Team B actual points")
	assertFloat(t, charlieBPt.ActualPoints, 5.0, "Charlie Team B actual points")
}

// =============================================================================
// Phase 4: Tournament Complete (Team A wins final)
// =============================================================================

func TestThatLifecycleCompletedChampionHasFullActualPoints(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a completed tournament (Team A champion with 2 wins)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToFinal(t, ctx, &f)

	// WHEN getting Alice's portfolio teams
	teams, err := f.calcuttaRepo.GetPortfolioTeams(ctx, f.aliceEntry.ID)
	if err != nil {
		t.Fatalf("getting portfolio teams: %v", err)
	}

	// THEN Team A actual = 0.80 * 35 = 28.0
	ptA := findPortfolioTeam(teams, f.teamA.ID)
	if ptA == nil {
		t.Fatalf("Alice missing portfolio team for team A")
	}
	assertFloat(t, ptA.ActualPoints, 28.0, "champion Team A actual points")
}

func TestThatLifecycleCompletedRunnerUpActualPointsReflectOneWin(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a completed tournament
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToFinal(t, ctx, &f)

	// WHEN getting Alice's portfolio teams
	teams, err := f.calcuttaRepo.GetPortfolioTeams(ctx, f.aliceEntry.ID)
	if err != nil {
		t.Fatalf("getting portfolio teams: %v", err)
	}

	// THEN Team B actual = 0.30 * 10 = 3.0 (runner-up, 1 win)
	ptB := findPortfolioTeam(teams, f.teamB.ID)
	if ptB == nil {
		t.Fatalf("Alice missing portfolio team for team B")
	}
	assertFloat(t, ptB.ActualPoints, 3.0, "runner-up Team B actual points")
}

func TestThatLifecycleCompletedChampionExpectedEqualsActual(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a completed tournament (all teams eliminated)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToFinal(t, ctx, &f)

	// Also eliminate the champion since tournament is complete
	f.teamA.IsEliminated = true
	if err := f.tournamentRepo.UpdateTournamentTeam(ctx, f.teamA); err != nil {
		t.Fatalf("eliminating champion: %v", err)
	}

	// WHEN getting Alice's portfolio teams
	teams, err := f.calcuttaRepo.GetPortfolioTeams(ctx, f.aliceEntry.ID)
	if err != nil {
		t.Fatalf("getting portfolio teams: %v", err)
	}

	// THEN Team A expected == actual (eliminated with max wins)
	ptA := findPortfolioTeam(teams, f.teamA.ID)
	if ptA == nil {
		t.Fatalf("Alice missing portfolio team for team A")
	}
	assertFloat(t, ptA.ExpectedPoints, ptA.ActualPoints, "champion expected == actual")
}

func TestThatLifecycleCompletedFinalTotalPointsAreCorrect(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a completed tournament
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToFinal(t, ctx, &f)

	// WHEN listing entries
	entries, err := f.calcuttaRepo.GetEntries(ctx, f.calcutta.ID)
	if err != nil {
		t.Fatalf("getting entries: %v", err)
	}

	// THEN Alice=31.0, Bob=9.0, Charlie=5.0
	aliceE := findEntry(entries, f.aliceEntry.ID)
	bobE := findEntry(entries, f.bobEntry.ID)
	charlieE := findEntry(entries, f.charlieEntry.ID)
	assertFloat(t, aliceE.TotalPoints, 31.0, "Alice final total points")
	assertFloat(t, bobE.TotalPoints, 9.0, "Bob final total points")
	assertFloat(t, charlieE.TotalPoints, 5.0, "Charlie final total points")
}

func TestThatLifecycleCompletedFinalRankingIsAliceBobCharlie(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a completed tournament
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToFinal(t, ctx, &f)

	// WHEN listing entries (sorted by total_points DESC)
	entries, err := f.calcuttaRepo.GetEntries(ctx, f.calcutta.ID)
	if err != nil {
		t.Fatalf("getting entries: %v", err)
	}

	// THEN ranking is Alice (31.0), Bob (9.0), Charlie (5.0)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].ID != f.aliceEntry.ID {
		t.Errorf("expected Alice 1st, got %s", entries[0].Name)
	}
	if entries[1].ID != f.bobEntry.ID {
		t.Errorf("expected Bob 2nd, got %s", entries[1].Name)
	}
	if entries[2].ID != f.charlieEntry.ID {
		t.Errorf("expected Charlie 3rd, got %s", entries[2].Name)
	}
}

func TestThatLifecycleCompletedPayoutsDistributeCorrectly(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a completed tournament
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToFinal(t, ctx, &f)

	// WHEN computing payouts from SQL-derived total points
	entries, err := f.calcuttaRepo.GetEntries(ctx, f.calcutta.ID)
	if err != nil {
		t.Fatalf("getting entries: %v", err)
	}
	payouts, err := f.calcuttaRepo.GetPayouts(ctx, f.calcutta.ID)
	if err != nil {
		t.Fatalf("getting payouts: %v", err)
	}

	_, results := calcutta.ComputeEntryPlacementsAndPayouts(entries, payouts)

	// THEN Alice=500c (1st), Bob=300c (2nd), Charlie=0c (3rd)
	if results[f.aliceEntry.ID].PayoutCents != 500 {
		t.Errorf("expected Alice payout 500c, got %d", results[f.aliceEntry.ID].PayoutCents)
	}
	if results[f.bobEntry.ID].PayoutCents != 300 {
		t.Errorf("expected Bob payout 300c, got %d", results[f.bobEntry.ID].PayoutCents)
	}
	if results[f.charlieEntry.ID].PayoutCents != 0 {
		t.Errorf("expected Charlie payout 0c, got %d", results[f.charlieEntry.ID].PayoutCents)
	}
}
