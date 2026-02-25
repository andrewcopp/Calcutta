package calcutta

import (
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
)

func standingsByID(standings []*models.EntryStanding) map[string]*models.EntryStanding {
	m := make(map[string]*models.EntryStanding, len(standings))
	for _, s := range standings {
		m[s.EntryID] = s
	}
	return m
}

func TestThatComputeStandingsReturnsNilWhenEntriesNil(t *testing.T) {
	standings := ComputeStandings(nil, nil, nil)
	if standings != nil {
		t.Fatalf("expected nil output")
	}
}

func TestThatComputeStandingsSortsByTotalPointsDescending(t *testing.T) {
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(1, 0)

	e2 := testutil.NewEntry()
	e2.ID = "e2"
	e2.CreatedAt = time.Unix(2, 0)

	points := map[string]float64{"e1": 10, "e2": 20}
	standings := ComputeStandings([]*models.CalcuttaEntry{e1, e2}, points, nil)

	if standings[0].EntryID != "e2" {
		t.Fatalf("expected first standing to be e2, got %q", standings[0].EntryID)
	}
}

func TestThatComputeStandingsSortsTiesByCreatedDescending(t *testing.T) {
	eOld := testutil.NewEntry()
	eOld.ID = "old"
	eOld.CreatedAt = time.Unix(1, 0)

	eNew := testutil.NewEntry()
	eNew.ID = "new"
	eNew.CreatedAt = time.Unix(2, 0)

	points := map[string]float64{"old": 10, "new": 10}
	standings := ComputeStandings([]*models.CalcuttaEntry{eOld, eNew}, points, nil)

	if standings[0].EntryID != "new" {
		t.Fatalf("expected first standing to be new, got %q", standings[0].EntryID)
	}
}

func TestThatComputeStandingsMarksTiesWithinEpsilon(t *testing.T) {
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(2, 0)

	e2 := testutil.NewEntry()
	e2.ID = "e2"
	e2.CreatedAt = time.Unix(1, 0)

	points := map[string]float64{"e1": 10.00000, "e2": 10.00001}
	byID := standingsByID(ComputeStandings([]*models.CalcuttaEntry{e1, e2}, points, nil))

	if !byID["e1"].IsTied {
		t.Fatalf("expected e1 to be tied")
	}
}

func TestThatComputeStandingsSetsFinishPositionOneForTopEntry(t *testing.T) {
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(1, 0)

	e2 := testutil.NewEntry()
	e2.ID = "e2"
	e2.CreatedAt = time.Unix(1, 0)

	points := map[string]float64{"e1": 20, "e2": 10}
	byID := standingsByID(ComputeStandings([]*models.CalcuttaEntry{e2, e1}, points, nil))

	if byID["e1"].FinishPosition != 1 {
		t.Fatalf("expected finish position 1, got %d", byID["e1"].FinishPosition)
	}
}

func TestThatComputeStandingsSplitsPayoutAcrossTieGroup(t *testing.T) {
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(2, 0)

	e2 := testutil.NewEntry()
	e2.ID = "e2"
	e2.CreatedAt = time.Unix(1, 0)

	points := map[string]float64{"e1": 10, "e2": 10}

	p1 := testutil.NewPayout()
	p1.Position = 1
	p1.AmountCents = 100

	p2 := testutil.NewPayout()
	p2.Position = 2
	p2.AmountCents = 50

	byID := standingsByID(ComputeStandings([]*models.CalcuttaEntry{e1, e2}, points, []*models.CalcuttaPayout{p1, p2}))

	if byID["e1"].PayoutCents != 75 {
		t.Fatalf("expected payout 75, got %d", byID["e1"].PayoutCents)
	}
}

func TestThatComputeStandingsDistributesRemainderToEarlierEntryInSortedOrder(t *testing.T) {
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(2, 0)

	e2 := testutil.NewEntry()
	e2.ID = "e2"
	e2.CreatedAt = time.Unix(1, 0)

	points := map[string]float64{"e1": 10, "e2": 10}

	p1 := testutil.NewPayout()
	p1.Position = 1
	p1.AmountCents = 100

	p2 := testutil.NewPayout()
	p2.Position = 2
	p2.AmountCents = 99

	byID := standingsByID(ComputeStandings([]*models.CalcuttaEntry{e1, e2}, points, []*models.CalcuttaPayout{p1, p2}))

	if byID["e1"].PayoutCents != 100 {
		t.Fatalf("expected payout 100, got %d", byID["e1"].PayoutCents)
	}
}

func TestThatComputeStandingsSetsInTheMoneyWhenPayoutPositive(t *testing.T) {
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(1, 0)

	points := map[string]float64{"e1": 10}

	p1 := testutil.NewPayout()
	p1.Position = 1
	p1.AmountCents = 1

	byID := standingsByID(ComputeStandings([]*models.CalcuttaEntry{e1}, points, []*models.CalcuttaPayout{p1}))

	if !byID["e1"].InTheMoney {
		t.Fatalf("expected in the money")
	}
}

func TestThatComputeStandingsDoesNotMutateInputEntries(t *testing.T) {
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(1, 0)
	origName := e1.Name

	points := map[string]float64{"e1": 10}
	_ = ComputeStandings([]*models.CalcuttaEntry{e1}, points, nil)

	if e1.Name != origName {
		t.Fatalf("expected input entry to remain unmodified")
	}
}

func TestThatThreeWayTiePoolsAndSplitsPayoutsEvenly(t *testing.T) {
	// GIVEN three entries tied at the same score with payouts for positions 1-3
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(3, 0)

	e2 := testutil.NewEntry()
	e2.ID = "e2"
	e2.CreatedAt = time.Unix(2, 0)

	e3 := testutil.NewEntry()
	e3.ID = "e3"
	e3.CreatedAt = time.Unix(1, 0)

	points := map[string]float64{"e1": 10, "e2": 10, "e3": 10}

	p1 := testutil.NewPayout()
	p1.Position = 1
	p1.AmountCents = 300

	p2 := testutil.NewPayout()
	p2.Position = 2
	p2.AmountCents = 150

	p3 := testutil.NewPayout()
	p3.Position = 3
	p3.AmountCents = 150

	// WHEN computing standings
	byID := standingsByID(ComputeStandings(
		[]*models.CalcuttaEntry{e1, e2, e3},
		points,
		[]*models.CalcuttaPayout{p1, p2, p3},
	))

	// THEN each entry receives (300+150+150)/3 = 200 cents
	if byID["e1"].PayoutCents != 200 {
		t.Fatalf("expected payout 200, got %d", byID["e1"].PayoutCents)
	}
}

// threeWayTieRemainderScenario returns standings for a 3-way tie where the pool
// doesn't divide evenly (301/3 = 100 base, remainder=1).
func threeWayTieRemainderScenario() map[string]*models.EntryStanding {
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(3, 0)

	e2 := testutil.NewEntry()
	e2.ID = "e2"
	e2.CreatedAt = time.Unix(2, 0)

	e3 := testutil.NewEntry()
	e3.ID = "e3"
	e3.CreatedAt = time.Unix(1, 0)

	points := map[string]float64{"e1": 10, "e2": 10, "e3": 10}

	p1 := testutil.NewPayout()
	p1.Position = 1
	p1.AmountCents = 100

	p2 := testutil.NewPayout()
	p2.Position = 2
	p2.AmountCents = 100

	p3odd := testutil.NewPayout()
	p3odd.Position = 3
	p3odd.AmountCents = 101

	return standingsByID(ComputeStandings(
		[]*models.CalcuttaEntry{e1, e2, e3},
		points,
		[]*models.CalcuttaPayout{p1, p2, p3odd},
	))
}

func TestThatThreeWayTieRemainderFirstEntryGetsRemainderCent(t *testing.T) {
	// GIVEN three entries tied with a total pool that doesn't divide evenly by 3
	byID := threeWayTieRemainderScenario()

	// THEN first entry in sorted order gets extra cent (301/3 = 100 + 1 remainder)
	if byID["e1"].PayoutCents != 101 {
		t.Fatalf("expected first entry to get 101, got %d", byID["e1"].PayoutCents)
	}
}

func TestThatThreeWayTieRemainderSecondEntryGetsBasePayout(t *testing.T) {
	// GIVEN three entries tied with a total pool that doesn't divide evenly by 3
	byID := threeWayTieRemainderScenario()

	// THEN second entry in sorted order gets base payout
	if byID["e2"].PayoutCents != 100 {
		t.Fatalf("expected second entry to get 100, got %d", byID["e2"].PayoutCents)
	}
}

// tieOutsidePayoutScenario returns standings for entries where a tie at positions 3-4
// falls outside the payout range (only positions 1-2 have payouts).
func tieOutsidePayoutScenario() map[string]*models.EntryStanding {
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(4, 0)

	e2 := testutil.NewEntry()
	e2.ID = "e2"
	e2.CreatedAt = time.Unix(3, 0)

	e3 := testutil.NewEntry()
	e3.ID = "e3"
	e3.CreatedAt = time.Unix(2, 0)

	e4 := testutil.NewEntry()
	e4.ID = "e4"
	e4.CreatedAt = time.Unix(1, 0)

	points := map[string]float64{"e1": 20, "e2": 15, "e3": 10, "e4": 10}

	p1 := testutil.NewPayout()
	p1.Position = 1
	p1.AmountCents = 200

	p2 := testutil.NewPayout()
	p2.Position = 2
	p2.AmountCents = 100

	return standingsByID(ComputeStandings(
		[]*models.CalcuttaEntry{e1, e2, e3, e4},
		points,
		[]*models.CalcuttaPayout{p1, p2},
	))
}

func TestThatTieOutsidePayoutPositionsMarksTied(t *testing.T) {
	// GIVEN two entries tied at positions 3-4, with payouts only for positions 1-2
	byID := tieOutsidePayoutScenario()

	// THEN tied entries at positions 3-4 are marked as tied
	if !byID["e3"].IsTied {
		t.Fatalf("expected e3 to be tied")
	}
}

func TestThatTieOutsidePayoutPositionsResultsInZeroPayout(t *testing.T) {
	// GIVEN two entries tied at positions 3-4, with payouts only for positions 1-2
	byID := tieOutsidePayoutScenario()

	// THEN tied entries at positions 3-4 receive zero payout
	if byID["e3"].PayoutCents != 0 {
		t.Fatalf("expected payout 0, got %d", byID["e3"].PayoutCents)
	}
}

func TestThatAllEntriesTiedResultsInEvenPayoutSplit(t *testing.T) {
	// GIVEN four entries all tied at the same score with payouts for all positions
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(4, 0)

	e2 := testutil.NewEntry()
	e2.ID = "e2"
	e2.CreatedAt = time.Unix(3, 0)

	e3 := testutil.NewEntry()
	e3.ID = "e3"
	e3.CreatedAt = time.Unix(2, 0)

	e4 := testutil.NewEntry()
	e4.ID = "e4"
	e4.CreatedAt = time.Unix(1, 0)

	points := map[string]float64{"e1": 10, "e2": 10, "e3": 10, "e4": 10}

	p1 := testutil.NewPayout()
	p1.Position = 1
	p1.AmountCents = 400

	p2 := testutil.NewPayout()
	p2.Position = 2
	p2.AmountCents = 200

	p3 := testutil.NewPayout()
	p3.Position = 3
	p3.AmountCents = 200

	p4 := testutil.NewPayout()
	p4.Position = 4
	p4.AmountCents = 200

	// WHEN computing standings
	byID := standingsByID(ComputeStandings(
		[]*models.CalcuttaEntry{e1, e2, e3, e4},
		points,
		[]*models.CalcuttaPayout{p1, p2, p3, p4},
	))

	// THEN total pool is 1000, each gets 1000/4 = 250
	if byID["e1"].PayoutCents != 250 {
		t.Fatalf("expected payout 250, got %d", byID["e1"].PayoutCents)
	}
}

// singleEntryPoolScenario returns standings for a single entry with a 500-cent payout.
func singleEntryPoolScenario() map[string]*models.EntryStanding {
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(1, 0)

	points := map[string]float64{"e1": 10}

	p1 := testutil.NewPayout()
	p1.Position = 1
	p1.AmountCents = 500

	return standingsByID(ComputeStandings(
		[]*models.CalcuttaEntry{e1},
		points,
		[]*models.CalcuttaPayout{p1},
	))
}

func TestThatSingleEntryPoolIsNotTied(t *testing.T) {
	// GIVEN a single entry with a payout
	byID := singleEntryPoolScenario()

	// THEN the single entry is not tied
	if byID["e1"].IsTied {
		t.Fatalf("expected single entry to not be tied")
	}
}

func TestThatSingleEntryPoolGetsFullPayout(t *testing.T) {
	// GIVEN a single entry with a payout
	byID := singleEntryPoolScenario()

	// THEN the single entry gets the full payout
	if byID["e1"].PayoutCents != 500 {
		t.Fatalf("expected payout 500, got %d", byID["e1"].PayoutCents)
	}
}

// tieSpanningPayoutBoundaryScenario returns standings for entries where a 3-way tie
// at positions 2-4 spans the payout boundary (only positions 1-3 have payouts).
func tieSpanningPayoutBoundaryScenario() map[string]*models.EntryStanding {
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(4, 0)

	e2 := testutil.NewEntry()
	e2.ID = "e2"
	e2.CreatedAt = time.Unix(3, 0)

	e3 := testutil.NewEntry()
	e3.ID = "e3"
	e3.CreatedAt = time.Unix(2, 0)

	e4 := testutil.NewEntry()
	e4.ID = "e4"
	e4.CreatedAt = time.Unix(1, 0)

	points := map[string]float64{"e1": 20, "e2": 10, "e3": 10, "e4": 10}

	p1 := testutil.NewPayout()
	p1.Position = 1
	p1.AmountCents = 300

	p2 := testutil.NewPayout()
	p2.Position = 2
	p2.AmountCents = 150

	p3 := testutil.NewPayout()
	p3.Position = 3
	p3.AmountCents = 50

	return standingsByID(ComputeStandings(
		[]*models.CalcuttaEntry{e1, e2, e3, e4},
		points,
		[]*models.CalcuttaPayout{p1, p2, p3},
	))
}

func TestThatTieGroupSpanningPayoutBoundaryRemainderToEarlierEntry(t *testing.T) {
	// GIVEN entries where a tie spans positions 2-4, but only positions 1-3 have payouts
	byID := tieSpanningPayoutBoundaryScenario()

	// THEN tie group pools payouts from positions 2+3+4 (150+50+0=200), split 3 ways
	// 200/3 = 66 base, remainder=2 -> first entry in sorted order gets 67
	if byID["e2"].PayoutCents != 67 {
		t.Fatalf("expected payout 67, got %d", byID["e2"].PayoutCents)
	}
}

func TestThatTieGroupSpanningPayoutBoundaryBaseToLaterEntry(t *testing.T) {
	// GIVEN entries where a tie spans positions 2-4, but only positions 1-3 have payouts
	byID := tieSpanningPayoutBoundaryScenario()

	// THEN the last entry in sorted order gets the base payout (no remainder)
	if byID["e4"].PayoutCents != 66 {
		t.Fatalf("expected payout 66, got %d", byID["e4"].PayoutCents)
	}
}

func TestThatNonTiedEntryAfterTieGroupGetsCorrectFinishPosition(t *testing.T) {
	// GIVEN a 2-way tie at position 1, followed by a non-tied entry
	e1 := testutil.NewEntry()
	e1.ID = "e1"
	e1.CreatedAt = time.Unix(3, 0)

	e2 := testutil.NewEntry()
	e2.ID = "e2"
	e2.CreatedAt = time.Unix(2, 0)

	e3 := testutil.NewEntry()
	e3.ID = "e3"
	e3.CreatedAt = time.Unix(1, 0)

	points := map[string]float64{"e1": 20, "e2": 20, "e3": 10}

	// WHEN computing standings
	byID := standingsByID(ComputeStandings(
		[]*models.CalcuttaEntry{e1, e2, e3},
		points,
		nil,
	))

	// THEN the entry after the tie group gets finish position 3 (skips position 2)
	if byID["e3"].FinishPosition != 3 {
		t.Fatalf("expected finish position 3, got %d", byID["e3"].FinishPosition)
	}
}

func TestThatEmptyEntriesSliceReturnsEmptyResults(t *testing.T) {
	// GIVEN an empty non-nil entries slice
	entries := []*models.CalcuttaEntry{}

	// WHEN computing standings
	standings := ComputeStandings(entries, nil, nil)

	// THEN returns empty non-nil output
	if len(standings) != 0 {
		t.Fatalf("expected empty standings slice, got %d entries", len(standings))
	}
}
