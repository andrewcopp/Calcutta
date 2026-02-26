package pool

import (
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func newTestPortfolio(id string, createdAt time.Time) *models.Portfolio {
	return &models.Portfolio{
		ID:        id,
		Name:      id,
		PoolID:    "pool-1",
		Status:    "submitted",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
}

func newTestPayout(position, amountCents int) *models.PoolPayout {
	return &models.PoolPayout{
		ID:          "payout-1",
		PoolID:      "pool-1",
		Position:    position,
		AmountCents: amountCents,
	}
}

func standingsByID(standings []*models.PortfolioStanding) map[string]*models.PortfolioStanding {
	m := make(map[string]*models.PortfolioStanding, len(standings))
	for _, s := range standings {
		m[s.PortfolioID] = s
	}
	return m
}

func TestThatComputeStandingsReturnsNilWhenPortfoliosNil(t *testing.T) {
	standings := ComputeStandings(nil, nil, nil)
	if standings != nil {
		t.Fatalf("expected nil output")
	}
}

func TestThatComputeStandingsSortsByTotalReturnsDescending(t *testing.T) {
	p1 := newTestPortfolio("p1", time.Unix(1, 0))
	p2 := newTestPortfolio("p2", time.Unix(2, 0))

	returns := map[string]float64{"p1": 10, "p2": 20}
	standings := ComputeStandings([]*models.Portfolio{p1, p2}, returns, nil)

	if standings[0].PortfolioID != "p2" {
		t.Fatalf("expected first standing to be p2, got %q", standings[0].PortfolioID)
	}
}

func TestThatComputeStandingsSortsTiesByCreatedDescending(t *testing.T) {
	pOld := newTestPortfolio("old", time.Unix(1, 0))
	pNew := newTestPortfolio("new", time.Unix(2, 0))

	returns := map[string]float64{"old": 10, "new": 10}
	standings := ComputeStandings([]*models.Portfolio{pOld, pNew}, returns, nil)

	if standings[0].PortfolioID != "new" {
		t.Fatalf("expected first standing to be new, got %q", standings[0].PortfolioID)
	}
}

func TestThatComputeStandingsMarksTiesWithinEpsilon(t *testing.T) {
	p1 := newTestPortfolio("p1", time.Unix(2, 0))
	p2 := newTestPortfolio("p2", time.Unix(1, 0))

	returns := map[string]float64{"p1": 10.00000, "p2": 10.00001}
	byID := standingsByID(ComputeStandings([]*models.Portfolio{p1, p2}, returns, nil))

	if !byID["p1"].IsTied {
		t.Fatalf("expected p1 to be tied")
	}
}

func TestThatComputeStandingsSetsFinishPositionOneForTopPortfolio(t *testing.T) {
	p1 := newTestPortfolio("p1", time.Unix(1, 0))
	p2 := newTestPortfolio("p2", time.Unix(1, 0))

	returns := map[string]float64{"p1": 20, "p2": 10}
	byID := standingsByID(ComputeStandings([]*models.Portfolio{p2, p1}, returns, nil))

	if byID["p1"].FinishPosition != 1 {
		t.Fatalf("expected finish position 1, got %d", byID["p1"].FinishPosition)
	}
}

func TestThatComputeStandingsSplitsPayoutAcrossTieGroup(t *testing.T) {
	p1 := newTestPortfolio("p1", time.Unix(2, 0))
	p2 := newTestPortfolio("p2", time.Unix(1, 0))

	returns := map[string]float64{"p1": 10, "p2": 10}

	pay1 := newTestPayout(1, 100)
	pay2 := newTestPayout(2, 50)

	byID := standingsByID(ComputeStandings([]*models.Portfolio{p1, p2}, returns, []*models.PoolPayout{pay1, pay2}))

	if byID["p1"].PayoutCents != 75 {
		t.Fatalf("expected payout 75, got %d", byID["p1"].PayoutCents)
	}
}

func TestThatComputeStandingsDistributesRemainderToEarlierPortfolioInSortedOrder(t *testing.T) {
	p1 := newTestPortfolio("p1", time.Unix(2, 0))
	p2 := newTestPortfolio("p2", time.Unix(1, 0))

	returns := map[string]float64{"p1": 10, "p2": 10}

	pay1 := newTestPayout(1, 100)
	pay2 := newTestPayout(2, 99)

	byID := standingsByID(ComputeStandings([]*models.Portfolio{p1, p2}, returns, []*models.PoolPayout{pay1, pay2}))

	if byID["p1"].PayoutCents != 100 {
		t.Fatalf("expected payout 100, got %d", byID["p1"].PayoutCents)
	}
}

func TestThatComputeStandingsSetsInTheMoneyWhenPayoutPositive(t *testing.T) {
	p1 := newTestPortfolio("p1", time.Unix(1, 0))

	returns := map[string]float64{"p1": 10}

	pay1 := newTestPayout(1, 1)

	byID := standingsByID(ComputeStandings([]*models.Portfolio{p1}, returns, []*models.PoolPayout{pay1}))

	if !byID["p1"].InTheMoney {
		t.Fatalf("expected in the money")
	}
}

func TestThatComputeStandingsDoesNotMutateInputPortfolios(t *testing.T) {
	p1 := newTestPortfolio("p1", time.Unix(1, 0))
	origName := p1.Name

	returns := map[string]float64{"p1": 10}
	_ = ComputeStandings([]*models.Portfolio{p1}, returns, nil)

	if p1.Name != origName {
		t.Fatalf("expected input portfolio to remain unmodified")
	}
}

func TestThatThreeWayTiePoolsAndSplitsPayoutsEvenly(t *testing.T) {
	// GIVEN three portfolios tied at the same score with payouts for positions 1-3
	p1 := newTestPortfolio("p1", time.Unix(3, 0))
	p2 := newTestPortfolio("p2", time.Unix(2, 0))
	p3 := newTestPortfolio("p3", time.Unix(1, 0))

	returns := map[string]float64{"p1": 10, "p2": 10, "p3": 10}

	pay1 := newTestPayout(1, 300)
	pay2 := newTestPayout(2, 150)
	pay3 := newTestPayout(3, 150)

	// WHEN computing standings
	byID := standingsByID(ComputeStandings(
		[]*models.Portfolio{p1, p2, p3},
		returns,
		[]*models.PoolPayout{pay1, pay2, pay3},
	))

	// THEN each portfolio receives (300+150+150)/3 = 200 cents
	if byID["p1"].PayoutCents != 200 {
		t.Fatalf("expected payout 200, got %d", byID["p1"].PayoutCents)
	}
}

// threeWayTieRemainderScenario returns standings for a 3-way tie where the pool
// doesn't divide evenly (301/3 = 100 base, remainder=1).
func threeWayTieRemainderScenario() map[string]*models.PortfolioStanding {
	p1 := newTestPortfolio("p1", time.Unix(3, 0))
	p2 := newTestPortfolio("p2", time.Unix(2, 0))
	p3 := newTestPortfolio("p3", time.Unix(1, 0))

	returns := map[string]float64{"p1": 10, "p2": 10, "p3": 10}

	pay1 := newTestPayout(1, 100)
	pay2 := newTestPayout(2, 100)
	pay3odd := newTestPayout(3, 101)

	return standingsByID(ComputeStandings(
		[]*models.Portfolio{p1, p2, p3},
		returns,
		[]*models.PoolPayout{pay1, pay2, pay3odd},
	))
}

func TestThatThreeWayTieRemainderFirstPortfolioGetsRemainderCent(t *testing.T) {
	// GIVEN three portfolios tied with a total pool that doesn't divide evenly by 3
	byID := threeWayTieRemainderScenario()

	// THEN first portfolio in sorted order gets extra cent (301/3 = 100 + 1 remainder)
	if byID["p1"].PayoutCents != 101 {
		t.Fatalf("expected first portfolio to get 101, got %d", byID["p1"].PayoutCents)
	}
}

func TestThatThreeWayTieRemainderSecondPortfolioGetsBasePayout(t *testing.T) {
	// GIVEN three portfolios tied with a total pool that doesn't divide evenly by 3
	byID := threeWayTieRemainderScenario()

	// THEN second portfolio in sorted order gets base payout
	if byID["p2"].PayoutCents != 100 {
		t.Fatalf("expected second portfolio to get 100, got %d", byID["p2"].PayoutCents)
	}
}

// tieOutsidePayoutScenario returns standings for portfolios where a tie at positions 3-4
// falls outside the payout range (only positions 1-2 have payouts).
func tieOutsidePayoutScenario() map[string]*models.PortfolioStanding {
	p1 := newTestPortfolio("p1", time.Unix(4, 0))
	p2 := newTestPortfolio("p2", time.Unix(3, 0))
	p3 := newTestPortfolio("p3", time.Unix(2, 0))
	p4 := newTestPortfolio("p4", time.Unix(1, 0))

	returns := map[string]float64{"p1": 20, "p2": 15, "p3": 10, "p4": 10}

	pay1 := newTestPayout(1, 200)
	pay2 := newTestPayout(2, 100)

	return standingsByID(ComputeStandings(
		[]*models.Portfolio{p1, p2, p3, p4},
		returns,
		[]*models.PoolPayout{pay1, pay2},
	))
}

func TestThatTieOutsidePayoutPositionsMarksTied(t *testing.T) {
	// GIVEN two portfolios tied at positions 3-4, with payouts only for positions 1-2
	byID := tieOutsidePayoutScenario()

	// THEN tied portfolios at positions 3-4 are marked as tied
	if !byID["p3"].IsTied {
		t.Fatalf("expected p3 to be tied")
	}
}

func TestThatTieOutsidePayoutPositionsResultsInZeroPayout(t *testing.T) {
	// GIVEN two portfolios tied at positions 3-4, with payouts only for positions 1-2
	byID := tieOutsidePayoutScenario()

	// THEN tied portfolios at positions 3-4 receive zero payout
	if byID["p3"].PayoutCents != 0 {
		t.Fatalf("expected payout 0, got %d", byID["p3"].PayoutCents)
	}
}

func TestThatAllPortfoliosTiedResultsInEvenPayoutSplit(t *testing.T) {
	// GIVEN four portfolios all tied at the same score with payouts for all positions
	p1 := newTestPortfolio("p1", time.Unix(4, 0))
	p2 := newTestPortfolio("p2", time.Unix(3, 0))
	p3 := newTestPortfolio("p3", time.Unix(2, 0))
	p4 := newTestPortfolio("p4", time.Unix(1, 0))

	returns := map[string]float64{"p1": 10, "p2": 10, "p3": 10, "p4": 10}

	pay1 := newTestPayout(1, 400)
	pay2 := newTestPayout(2, 200)
	pay3 := newTestPayout(3, 200)
	pay4 := newTestPayout(4, 200)

	// WHEN computing standings
	byID := standingsByID(ComputeStandings(
		[]*models.Portfolio{p1, p2, p3, p4},
		returns,
		[]*models.PoolPayout{pay1, pay2, pay3, pay4},
	))

	// THEN total pool is 1000, each gets 1000/4 = 250
	if byID["p1"].PayoutCents != 250 {
		t.Fatalf("expected payout 250, got %d", byID["p1"].PayoutCents)
	}
}

// singlePortfolioPoolScenario returns standings for a single portfolio with a 500-cent payout.
func singlePortfolioPoolScenario() map[string]*models.PortfolioStanding {
	p1 := newTestPortfolio("p1", time.Unix(1, 0))

	returns := map[string]float64{"p1": 10}

	pay1 := newTestPayout(1, 500)

	return standingsByID(ComputeStandings(
		[]*models.Portfolio{p1},
		returns,
		[]*models.PoolPayout{pay1},
	))
}

func TestThatSinglePortfolioPoolIsNotTied(t *testing.T) {
	// GIVEN a single portfolio with a payout
	byID := singlePortfolioPoolScenario()

	// THEN the single portfolio is not tied
	if byID["p1"].IsTied {
		t.Fatalf("expected single portfolio to not be tied")
	}
}

func TestThatSinglePortfolioPoolGetsFullPayout(t *testing.T) {
	// GIVEN a single portfolio with a payout
	byID := singlePortfolioPoolScenario()

	// THEN the single portfolio gets the full payout
	if byID["p1"].PayoutCents != 500 {
		t.Fatalf("expected payout 500, got %d", byID["p1"].PayoutCents)
	}
}

// tieSpanningPayoutBoundaryScenario returns standings for portfolios where a 3-way tie
// at positions 2-4 spans the payout boundary (only positions 1-3 have payouts).
func tieSpanningPayoutBoundaryScenario() map[string]*models.PortfolioStanding {
	p1 := newTestPortfolio("p1", time.Unix(4, 0))
	p2 := newTestPortfolio("p2", time.Unix(3, 0))
	p3 := newTestPortfolio("p3", time.Unix(2, 0))
	p4 := newTestPortfolio("p4", time.Unix(1, 0))

	returns := map[string]float64{"p1": 20, "p2": 10, "p3": 10, "p4": 10}

	pay1 := newTestPayout(1, 300)
	pay2 := newTestPayout(2, 150)
	pay3 := newTestPayout(3, 50)

	return standingsByID(ComputeStandings(
		[]*models.Portfolio{p1, p2, p3, p4},
		returns,
		[]*models.PoolPayout{pay1, pay2, pay3},
	))
}

func TestThatTieGroupSpanningPayoutBoundaryRemainderToEarlierPortfolio(t *testing.T) {
	// GIVEN portfolios where a tie spans positions 2-4, but only positions 1-3 have payouts
	byID := tieSpanningPayoutBoundaryScenario()

	// THEN tie group pools payouts from positions 2+3+4 (150+50+0=200), split 3 ways
	// 200/3 = 66 base, remainder=2 -> first portfolio in sorted order gets 67
	if byID["p2"].PayoutCents != 67 {
		t.Fatalf("expected payout 67, got %d", byID["p2"].PayoutCents)
	}
}

func TestThatTieGroupSpanningPayoutBoundaryBaseToLaterPortfolio(t *testing.T) {
	// GIVEN portfolios where a tie spans positions 2-4, but only positions 1-3 have payouts
	byID := tieSpanningPayoutBoundaryScenario()

	// THEN the last portfolio in sorted order gets the base payout (no remainder)
	if byID["p4"].PayoutCents != 66 {
		t.Fatalf("expected payout 66, got %d", byID["p4"].PayoutCents)
	}
}

func TestThatNonTiedPortfolioAfterTieGroupGetsCorrectFinishPosition(t *testing.T) {
	// GIVEN a 2-way tie at position 1, followed by a non-tied portfolio
	p1 := newTestPortfolio("p1", time.Unix(3, 0))
	p2 := newTestPortfolio("p2", time.Unix(2, 0))
	p3 := newTestPortfolio("p3", time.Unix(1, 0))

	returns := map[string]float64{"p1": 20, "p2": 20, "p3": 10}

	// WHEN computing standings
	byID := standingsByID(ComputeStandings(
		[]*models.Portfolio{p1, p2, p3},
		returns,
		nil,
	))

	// THEN the portfolio after the tie group gets finish position 3 (skips position 2)
	if byID["p3"].FinishPosition != 3 {
		t.Fatalf("expected finish position 3, got %d", byID["p3"].FinishPosition)
	}
}

func TestThatEmptyPortfoliosSliceReturnsEmptyResults(t *testing.T) {
	// GIVEN an empty non-nil portfolios slice
	portfolios := []*models.Portfolio{}

	// WHEN computing standings
	standings := ComputeStandings(portfolios, nil, nil)

	// THEN returns empty non-nil output
	if len(standings) != 0 {
		t.Fatalf("expected empty standings slice, got %d portfolios", len(standings))
	}
}
