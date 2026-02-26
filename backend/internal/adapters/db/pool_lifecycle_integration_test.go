//go:build integration

package db_test

import (
	"context"
	"math"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/app/pool"
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

func findOwnershipDetail(details []*models.OwnershipDetail, teamID string) *models.OwnershipDetail {
	for _, od := range details {
		if od.TeamID == teamID {
			return od
		}
	}
	return nil
}

// =============================================================================
// Phase 1: Pre-Tournament (0 wins, all teams alive)
// =============================================================================

func TestThatLifecyclePreTournamentPortfoliosHaveZeroTotalReturns(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool before the tournament starts (all teams have 0 wins)
	f := mustSeedLifecycleTournament(t, ctx)

	// WHEN listing portfolios
	portfolios, returnsByPortfolio, err := f.poolRepo.GetPortfolios(ctx, f.pool.ID)
	if err != nil {
		t.Fatalf("getting portfolios: %v", err)
	}

	// THEN all 3 portfolios have 0 total returns
	for _, p := range portfolios {
		assertFloat(t, returnsByPortfolio[p.ID], 0.0, "portfolio "+p.Name+" total returns")
	}
	_ = f
}

func TestThatLifecyclePreTournamentOwnershipDetailsHaveZeroActualReturns(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool before the tournament starts
	f := mustSeedLifecycleTournament(t, ctx)

	// WHEN getting Alice's ownership details
	details, err := f.poolRepo.GetOwnershipDetails(ctx, f.alicePortfolio.ID)
	if err != nil {
		t.Fatalf("getting ownership details: %v", err)
	}

	// THEN all of Alice's ownership details have 0 actual returns
	for _, od := range details {
		assertFloat(t, od.ActualReturns, 0.0, "ownership detail actual returns")
	}
}

func TestThatLifecyclePreTournamentOwnershipPercentagesAreCorrect(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool before the tournament starts
	f := mustSeedLifecycleTournament(t, ctx)

	// WHEN getting Alice's ownership details
	details, err := f.poolRepo.GetOwnershipDetails(ctx, f.alicePortfolio.ID)
	if err != nil {
		t.Fatalf("getting ownership details: %v", err)
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
		od := findOwnershipDetail(details, teamID)
		if od == nil {
			t.Fatalf("Alice missing ownership detail for team %s", teamID)
		}
		assertFloat(t, od.OwnershipPercentage, wantPct, "ownership for team "+teamID)
	}
}

func TestThatLifecyclePreTournamentExpectedReturnsReflectMaxPossible(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool before the tournament starts (max returns per team = 35)
	f := mustSeedLifecycleTournament(t, ctx)

	// WHEN getting Alice's ownership details
	details, err := f.poolRepo.GetOwnershipDetails(ctx, f.alicePortfolio.ID)
	if err != nil {
		t.Fatalf("getting ownership details: %v", err)
	}

	// THEN each team's expected = ownership * 35
	expected := map[string]float64{
		f.teamA.ID: 0.80 * 35, // 28.0
		f.teamB.ID: 0.30 * 35, // 10.5
		f.teamC.ID: 0.20 * 35, // 7.0
		f.teamD.ID: 0.20 * 35, // 7.0
	}
	for teamID, wantExp := range expected {
		od := findOwnershipDetail(details, teamID)
		if od == nil {
			t.Fatalf("Alice missing ownership detail for team %s", teamID)
		}
		assertFloat(t, od.ExpectedReturns, wantExp, "expected returns for team "+teamID)
	}
}

func TestThatLifecyclePreTournamentOwnershipSummaryMaxReturnsEqualsSumOfExpected(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool before the tournament starts
	f := mustSeedLifecycleTournament(t, ctx)

	// WHEN getting Alice's ownership summary
	summary, err := f.poolRepo.GetOwnershipSummary(ctx, f.alicePortfolio.ID)
	if err != nil {
		t.Fatalf("getting ownership summary: %v", err)
	}

	// THEN Alice's max returns = 28.0 + 10.5 + 7.0 + 7.0 = 52.5
	assertFloat(t, summary.MaximumReturns, 52.5, "Alice ownership summary max returns")
}

// =============================================================================
// Phase 3: Mid-Tournament (Team A, B win round 1; Team C, D eliminated)
// =============================================================================

func TestThatLifecycleMidTournamentWinnersHaveCorrectActualReturns(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool mid-tournament (round 1 complete)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToRound1(t, ctx, &f)

	// WHEN getting Alice's ownership details
	details, err := f.poolRepo.GetOwnershipDetails(ctx, f.alicePortfolio.ID)
	if err != nil {
		t.Fatalf("getting ownership details: %v", err)
	}

	// THEN Alice's actual returns: A=0.80*10=8.0, B=0.30*10=3.0, C=0.0, D=0.0
	expected := map[string]float64{
		f.teamA.ID: 8.0,
		f.teamB.ID: 3.0,
		f.teamC.ID: 0.0,
		f.teamD.ID: 0.0,
	}
	for teamID, wantReturns := range expected {
		od := findOwnershipDetail(details, teamID)
		if od == nil {
			t.Fatalf("Alice missing ownership detail for team %s", teamID)
		}
		assertFloat(t, od.ActualReturns, wantReturns, "actual returns for team "+teamID)
	}
}

func TestThatLifecycleMidTournamentEliminatedTeamsLockExpectedReturns(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool mid-tournament (round 1 complete)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToRound1(t, ctx, &f)

	// WHEN getting Alice's ownership details
	details, err := f.poolRepo.GetOwnershipDetails(ctx, f.alicePortfolio.ID)
	if err != nil {
		t.Fatalf("getting ownership details: %v", err)
	}

	// THEN eliminated teams have expected = actual = 0
	odC := findOwnershipDetail(details, f.teamC.ID)
	if odC == nil {
		t.Fatalf("Alice missing ownership detail for team C")
	}
	assertFloat(t, odC.ExpectedReturns, 0.0, "eliminated Team C expected returns")

	odD := findOwnershipDetail(details, f.teamD.ID)
	if odD == nil {
		t.Fatalf("Alice missing ownership detail for team D")
	}
	assertFloat(t, odD.ExpectedReturns, 0.0, "eliminated Team D expected returns")
}

func TestThatLifecycleMidTournamentSurvivingTeamsKeepMaxExpectedReturns(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool mid-tournament (round 1 complete)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToRound1(t, ctx, &f)

	// WHEN getting Alice's ownership details
	details, err := f.poolRepo.GetOwnershipDetails(ctx, f.alicePortfolio.ID)
	if err != nil {
		t.Fatalf("getting ownership details: %v", err)
	}

	// THEN surviving teams keep max expected: A=0.80*35=28.0, B=0.30*35=10.5
	odA := findOwnershipDetail(details, f.teamA.ID)
	if odA == nil {
		t.Fatalf("Alice missing ownership detail for team A")
	}
	assertFloat(t, odA.ExpectedReturns, 28.0, "surviving Team A expected returns")

	odB := findOwnershipDetail(details, f.teamB.ID)
	if odB == nil {
		t.Fatalf("Alice missing ownership detail for team B")
	}
	assertFloat(t, odB.ExpectedReturns, 10.5, "surviving Team B expected returns")
}

func TestThatLifecycleMidTournamentTotalReturnsReflectProgress(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool mid-tournament (round 1 complete)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToRound1(t, ctx, &f)

	// WHEN listing portfolios
	_, returnsByPortfolio, err := f.poolRepo.GetPortfolios(ctx, f.pool.ID)
	if err != nil {
		t.Fatalf("getting portfolios: %v", err)
	}

	// THEN Alice=11.0, Bob=4.0, Charlie=5.0
	assertFloat(t, returnsByPortfolio[f.alicePortfolio.ID], 11.0, "Alice total returns")
}

func TestThatLifecycleMidTournamentPortfoliosRankByTotalReturnsDescending(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool mid-tournament (round 1 complete)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToRound1(t, ctx, &f)

	// WHEN listing portfolios (returned sorted by total_returns DESC)
	portfolios, _, err := f.poolRepo.GetPortfolios(ctx, f.pool.ID)
	if err != nil {
		t.Fatalf("getting portfolios: %v", err)
	}

	// THEN order is Alice (11.0), Charlie (5.0), Bob (4.0)
	if len(portfolios) != 3 {
		t.Fatalf("expected 3 portfolios, got %d", len(portfolios))
	}
	if portfolios[0].ID != f.alicePortfolio.ID {
		t.Errorf("expected Alice first, got %s", portfolios[0].Name)
	}
	if portfolios[1].ID != f.charliePortfolio.ID {
		t.Errorf("expected Charlie second, got %s", portfolios[1].Name)
	}
	if portfolios[2].ID != f.bobPortfolio.ID {
		t.Errorf("expected Bob third, got %s", portfolios[2].Name)
	}
}

func TestThatLifecycleMidTournamentSharedTeamSplitsReturnsByOwnership(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a pool mid-tournament (round 1 complete)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToRound1(t, ctx, &f)

	// WHEN getting ownership details for all 3 portfolios
	aliceDetails, err := f.poolRepo.GetOwnershipDetails(ctx, f.alicePortfolio.ID)
	if err != nil {
		t.Fatalf("getting Alice ownership details: %v", err)
	}
	bobDetails, err := f.poolRepo.GetOwnershipDetails(ctx, f.bobPortfolio.ID)
	if err != nil {
		t.Fatalf("getting Bob ownership details: %v", err)
	}
	charlieDetails, err := f.poolRepo.GetOwnershipDetails(ctx, f.charliePortfolio.ID)
	if err != nil {
		t.Fatalf("getting Charlie ownership details: %v", err)
	}

	// THEN Team B actual returns: Alice=0.30*10=3.0, Bob=0.20*10=2.0, Charlie=0.50*10=5.0
	aliceBOd := findOwnershipDetail(aliceDetails, f.teamB.ID)
	bobBOd := findOwnershipDetail(bobDetails, f.teamB.ID)
	charlieBOd := findOwnershipDetail(charlieDetails, f.teamB.ID)

	assertFloat(t, aliceBOd.ActualReturns, 3.0, "Alice Team B actual returns")
	assertFloat(t, bobBOd.ActualReturns, 2.0, "Bob Team B actual returns")
	assertFloat(t, charlieBOd.ActualReturns, 5.0, "Charlie Team B actual returns")
}

// =============================================================================
// Phase 4: Tournament Complete (Team A wins final)
// =============================================================================

func TestThatLifecycleCompletedChampionHasFullActualReturns(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a completed tournament (Team A champion with 2 wins)
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToFinal(t, ctx, &f)

	// WHEN getting Alice's ownership details
	details, err := f.poolRepo.GetOwnershipDetails(ctx, f.alicePortfolio.ID)
	if err != nil {
		t.Fatalf("getting ownership details: %v", err)
	}

	// THEN Team A actual = 0.80 * 35 = 28.0
	odA := findOwnershipDetail(details, f.teamA.ID)
	if odA == nil {
		t.Fatalf("Alice missing ownership detail for team A")
	}
	assertFloat(t, odA.ActualReturns, 28.0, "champion Team A actual returns")
}

func TestThatLifecycleCompletedRunnerUpActualReturnsReflectOneWin(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a completed tournament
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToFinal(t, ctx, &f)

	// WHEN getting Alice's ownership details
	details, err := f.poolRepo.GetOwnershipDetails(ctx, f.alicePortfolio.ID)
	if err != nil {
		t.Fatalf("getting ownership details: %v", err)
	}

	// THEN Team B actual = 0.30 * 10 = 3.0 (runner-up, 1 win)
	odB := findOwnershipDetail(details, f.teamB.ID)
	if odB == nil {
		t.Fatalf("Alice missing ownership detail for team B")
	}
	assertFloat(t, odB.ActualReturns, 3.0, "runner-up Team B actual returns")
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

	// WHEN getting Alice's ownership details
	details, err := f.poolRepo.GetOwnershipDetails(ctx, f.alicePortfolio.ID)
	if err != nil {
		t.Fatalf("getting ownership details: %v", err)
	}

	// THEN Team A expected == actual (eliminated with max wins)
	odA := findOwnershipDetail(details, f.teamA.ID)
	if odA == nil {
		t.Fatalf("Alice missing ownership detail for team A")
	}
	assertFloat(t, odA.ExpectedReturns, odA.ActualReturns, "champion expected == actual")
}

func TestThatLifecycleCompletedFinalTotalReturnsAreCorrect(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		if err := testutil.TruncateAll(ctx, pool); err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	})

	// GIVEN a completed tournament
	f := mustSeedLifecycleTournament(t, ctx)
	mustAdvanceToFinal(t, ctx, &f)

	// WHEN listing portfolios
	_, returnsByPortfolio, err := f.poolRepo.GetPortfolios(ctx, f.pool.ID)
	if err != nil {
		t.Fatalf("getting portfolios: %v", err)
	}

	// THEN Alice=31.0, Bob=9.0, Charlie=5.0
	assertFloat(t, returnsByPortfolio[f.alicePortfolio.ID], 31.0, "Alice final total returns")
	assertFloat(t, returnsByPortfolio[f.bobPortfolio.ID], 9.0, "Bob final total returns")
	assertFloat(t, returnsByPortfolio[f.charliePortfolio.ID], 5.0, "Charlie final total returns")
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

	// WHEN listing portfolios (sorted by total_returns DESC)
	portfolios, _, err := f.poolRepo.GetPortfolios(ctx, f.pool.ID)
	if err != nil {
		t.Fatalf("getting portfolios: %v", err)
	}

	// THEN ranking is Alice (31.0), Bob (9.0), Charlie (5.0)
	if len(portfolios) != 3 {
		t.Fatalf("expected 3 portfolios, got %d", len(portfolios))
	}
	if portfolios[0].ID != f.alicePortfolio.ID {
		t.Errorf("expected Alice 1st, got %s", portfolios[0].Name)
	}
	if portfolios[1].ID != f.bobPortfolio.ID {
		t.Errorf("expected Bob 2nd, got %s", portfolios[1].Name)
	}
	if portfolios[2].ID != f.charliePortfolio.ID {
		t.Errorf("expected Charlie 3rd, got %s", portfolios[2].Name)
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

	// WHEN computing payouts from SQL-derived total returns
	portfolios, returnsByPortfolio, err := f.poolRepo.GetPortfolios(ctx, f.pool.ID)
	if err != nil {
		t.Fatalf("getting portfolios: %v", err)
	}
	payouts, err := f.poolRepo.GetPayouts(ctx, f.pool.ID)
	if err != nil {
		t.Fatalf("getting payouts: %v", err)
	}

	standings := pool.ComputeStandings(portfolios, returnsByPortfolio, payouts)
	standingsByID := make(map[string]*models.PortfolioStanding, len(standings))
	for _, s := range standings {
		standingsByID[s.PortfolioID] = s
	}

	// THEN Alice=500c (1st), Bob=300c (2nd), Charlie=0c (3rd)
	if standingsByID[f.alicePortfolio.ID].PayoutCents != 500 {
		t.Errorf("expected Alice payout 500c, got %d", standingsByID[f.alicePortfolio.ID].PayoutCents)
	}
	if standingsByID[f.bobPortfolio.ID].PayoutCents != 300 {
		t.Errorf("expected Bob payout 300c, got %d", standingsByID[f.bobPortfolio.ID].PayoutCents)
	}
	if standingsByID[f.charliePortfolio.ID].PayoutCents != 0 {
		t.Errorf("expected Charlie payout 0c, got %d", standingsByID[f.charliePortfolio.ID].PayoutCents)
	}
}
