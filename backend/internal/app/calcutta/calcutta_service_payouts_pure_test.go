package calcutta

import (
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatComputeEntryPlacementsAndPayoutsReturnsNilWhenEntriesNil(t *testing.T) {
	sorted, results := ComputeEntryPlacementsAndPayouts(nil, nil)
	if sorted != nil || results != nil {
		t.Fatalf("expected nil outputs")
	}
}

func TestThatComputeEntryPlacementsAndPayoutsSortsByTotalPointsDescending(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 20, CreatedAt: time.Unix(2, 0)}

	sorted, _ := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e1, e2}, nil)

	if sorted[0].ID != "e2" {
		t.Fatalf("expected first entry to be e2, got %q", sorted[0].ID)
	}
}

func TestThatComputeEntryPlacementsAndPayoutsSortsTiesByCreatedDescending(t *testing.T) {
	eOld := &models.CalcuttaEntry{ID: "old", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}
	eNew := &models.CalcuttaEntry{ID: "new", TotalPoints: 10, CreatedAt: time.Unix(2, 0)}

	sorted, _ := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{eOld, eNew}, nil)

	if sorted[0].ID != "new" {
		t.Fatalf("expected first entry to be new, got %q", sorted[0].ID)
	}
}

func TestThatComputeEntryPlacementsAndPayoutsMarksTiesWithinEpsilon(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10.00000, CreatedAt: time.Unix(2, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 10.00001, CreatedAt: time.Unix(1, 0)}

	_, results := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e1, e2}, nil)

	if !results["e1"].IsTied {
		t.Fatalf("expected e1 to be tied")
	}
}

func TestThatComputeEntryPlacementsAndPayoutsSetsFinishPositionOneForTopEntry(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 20, CreatedAt: time.Unix(1, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}

	_, results := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e2, e1}, nil)

	if results["e1"].FinishPosition != 1 {
		t.Fatalf("expected finish position 1, got %d", results["e1"].FinishPosition)
	}
}

func TestThatComputeEntryPlacementsAndPayoutsSplitsPayoutAcrossTieGroup(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, CreatedAt: time.Unix(2, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}
	p1 := &models.CalcuttaPayout{Position: 1, AmountCents: 100}
	p2 := &models.CalcuttaPayout{Position: 2, AmountCents: 50}

	_, results := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e1, e2}, []*models.CalcuttaPayout{p1, p2})

	if results["e1"].PayoutCents != 75 {
		t.Fatalf("expected payout 75, got %d", results["e1"].PayoutCents)
	}
}

func TestThatComputeEntryPlacementsAndPayoutsDistributesRemainderToEarlierEntryInSortedOrder(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, CreatedAt: time.Unix(2, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}
	p1 := &models.CalcuttaPayout{Position: 1, AmountCents: 100}
	p2 := &models.CalcuttaPayout{Position: 2, AmountCents: 99}

	_, results := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e1, e2}, []*models.CalcuttaPayout{p1, p2})

	if results["e1"].PayoutCents != 100 {
		t.Fatalf("expected payout 100, got %d", results["e1"].PayoutCents)
	}
}

func TestThatComputeEntryPlacementsAndPayoutsSetsInTheMoneyWhenPayoutPositive(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}
	p1 := &models.CalcuttaPayout{Position: 1, AmountCents: 1}

	_, results := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e1}, []*models.CalcuttaPayout{p1})

	if !results["e1"].InTheMoney {
		t.Fatalf("expected in the money")
	}
}

func TestThatComputeEntryPlacementsAndPayoutsDoesNotMutateInputEntries(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, FinishPosition: 999, CreatedAt: time.Unix(1, 0)}

	_, _ = ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e1}, nil)

	if e1.FinishPosition != 999 {
		t.Fatalf("expected input entry to remain unmodified")
	}
}

func TestThatThreeWayTiePoolsAndSplitsPayoutsEvenly(t *testing.T) {
	// GIVEN three entries tied at the same score with payouts for positions 1-3
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, CreatedAt: time.Unix(3, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 10, CreatedAt: time.Unix(2, 0)}
	e3 := &models.CalcuttaEntry{ID: "e3", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}
	p1 := &models.CalcuttaPayout{Position: 1, AmountCents: 300}
	p2 := &models.CalcuttaPayout{Position: 2, AmountCents: 150}
	p3 := &models.CalcuttaPayout{Position: 3, AmountCents: 150}

	// WHEN computing placements and payouts
	_, results := ComputeEntryPlacementsAndPayouts(
		[]*models.CalcuttaEntry{e1, e2, e3},
		[]*models.CalcuttaPayout{p1, p2, p3},
	)

	// THEN each entry receives (300+150+150)/3 = 200 cents
	if results["e1"].PayoutCents != 200 {
		t.Fatalf("expected payout 200, got %d", results["e1"].PayoutCents)
	}
}

func TestThatThreeWayTieDistributesRemainderToEarliestEntriesInSortedOrder(t *testing.T) {
	// GIVEN three entries tied with a total pool that doesn't divide evenly by 3
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, CreatedAt: time.Unix(3, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 10, CreatedAt: time.Unix(2, 0)}
	e3 := &models.CalcuttaEntry{ID: "e3", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}
	p1 := &models.CalcuttaPayout{Position: 1, AmountCents: 100}
	p2 := &models.CalcuttaPayout{Position: 2, AmountCents: 100}
	p3 := &models.CalcuttaPayout{Position: 3, AmountCents: 100}

	// WHEN computing placements and payouts (300/3 = 100 each, 0 remainder)
	_, results := ComputeEntryPlacementsAndPayouts(
		[]*models.CalcuttaEntry{e1, e2, e3},
		[]*models.CalcuttaPayout{p1, p2, p3},
	)

	// THEN with remainder=1 scenario: use 301 total (100+101+100)
	// Actually test with odd total: 100+100+101 = 301
	// Re-setup with odd total
	p3odd := &models.CalcuttaPayout{Position: 3, AmountCents: 101}
	_, results = ComputeEntryPlacementsAndPayouts(
		[]*models.CalcuttaEntry{e1, e2, e3},
		[]*models.CalcuttaPayout{p1, p2, p3odd},
	)

	// THEN 301/3 = 100 base, remainder=1 → first entry in sorted order gets extra cent
	if results["e1"].PayoutCents != 101 {
		t.Fatalf("expected first entry to get 101, got %d", results["e1"].PayoutCents)
	}
	if results["e2"].PayoutCents != 100 {
		t.Fatalf("expected second entry to get 100, got %d", results["e2"].PayoutCents)
	}
}

func TestThatTieOutsidePayoutPositionsResultsInZeroPayout(t *testing.T) {
	// GIVEN two entries tied at positions 3-4, with payouts only for positions 1-2
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 20, CreatedAt: time.Unix(4, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 15, CreatedAt: time.Unix(3, 0)}
	e3 := &models.CalcuttaEntry{ID: "e3", TotalPoints: 10, CreatedAt: time.Unix(2, 0)}
	e4 := &models.CalcuttaEntry{ID: "e4", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}
	p1 := &models.CalcuttaPayout{Position: 1, AmountCents: 200}
	p2 := &models.CalcuttaPayout{Position: 2, AmountCents: 100}

	// WHEN computing placements and payouts
	_, results := ComputeEntryPlacementsAndPayouts(
		[]*models.CalcuttaEntry{e1, e2, e3, e4},
		[]*models.CalcuttaPayout{p1, p2},
	)

	// THEN tied entries at positions 3-4 are marked tied with zero payout
	if !results["e3"].IsTied {
		t.Fatalf("expected e3 to be tied")
	}
	if results["e3"].PayoutCents != 0 {
		t.Fatalf("expected payout 0, got %d", results["e3"].PayoutCents)
	}
}

func TestThatAllEntriesTiedResultsInEvenPayoutSplit(t *testing.T) {
	// GIVEN four entries all tied at the same score with payouts for all positions
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, CreatedAt: time.Unix(4, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 10, CreatedAt: time.Unix(3, 0)}
	e3 := &models.CalcuttaEntry{ID: "e3", TotalPoints: 10, CreatedAt: time.Unix(2, 0)}
	e4 := &models.CalcuttaEntry{ID: "e4", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}
	p1 := &models.CalcuttaPayout{Position: 1, AmountCents: 400}
	p2 := &models.CalcuttaPayout{Position: 2, AmountCents: 200}
	p3 := &models.CalcuttaPayout{Position: 3, AmountCents: 200}
	p4 := &models.CalcuttaPayout{Position: 4, AmountCents: 200}

	// WHEN computing placements and payouts
	_, results := ComputeEntryPlacementsAndPayouts(
		[]*models.CalcuttaEntry{e1, e2, e3, e4},
		[]*models.CalcuttaPayout{p1, p2, p3, p4},
	)

	// THEN total pool is 1000, each gets 1000/4 = 250
	if results["e1"].PayoutCents != 250 {
		t.Fatalf("expected payout 250, got %d", results["e1"].PayoutCents)
	}
}

func TestThatSingleEntryPoolGetsFullPayoutWithNoTie(t *testing.T) {
	// GIVEN a single entry with a payout
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}
	p1 := &models.CalcuttaPayout{Position: 1, AmountCents: 500}

	// WHEN computing placements and payouts
	_, results := ComputeEntryPlacementsAndPayouts(
		[]*models.CalcuttaEntry{e1},
		[]*models.CalcuttaPayout{p1},
	)

	// THEN the single entry is not tied and gets the full payout
	if results["e1"].IsTied {
		t.Fatalf("expected single entry to not be tied")
	}
	if results["e1"].PayoutCents != 500 {
		t.Fatalf("expected payout 500, got %d", results["e1"].PayoutCents)
	}
}

func TestThatTieGroupSpanningPayoutBoundaryPoolsOnlyDefinedPayouts(t *testing.T) {
	// GIVEN entries where a tie spans positions 2-4, but only positions 1-3 have payouts
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 20, CreatedAt: time.Unix(4, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 10, CreatedAt: time.Unix(3, 0)}
	e3 := &models.CalcuttaEntry{ID: "e3", TotalPoints: 10, CreatedAt: time.Unix(2, 0)}
	e4 := &models.CalcuttaEntry{ID: "e4", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}
	p1 := &models.CalcuttaPayout{Position: 1, AmountCents: 300}
	p2 := &models.CalcuttaPayout{Position: 2, AmountCents: 150}
	p3 := &models.CalcuttaPayout{Position: 3, AmountCents: 50}

	// WHEN computing placements and payouts
	_, results := ComputeEntryPlacementsAndPayouts(
		[]*models.CalcuttaEntry{e1, e2, e3, e4},
		[]*models.CalcuttaPayout{p1, p2, p3},
	)

	// THEN tie group at positions 2-4 pools payouts from positions 2+3+4 (150+50+0=200), split 3 ways
	// 200/3 = 66 base, remainder=2 → first two in sorted order get 67
	if results["e2"].PayoutCents != 67 {
		t.Fatalf("expected payout 67, got %d", results["e2"].PayoutCents)
	}
	if results["e4"].PayoutCents != 66 {
		t.Fatalf("expected payout 66, got %d", results["e4"].PayoutCents)
	}
}

func TestThatNonTiedEntryAfterTieGroupGetsCorrectFinishPosition(t *testing.T) {
	// GIVEN a 2-way tie at position 1, followed by a non-tied entry
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 20, CreatedAt: time.Unix(3, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 20, CreatedAt: time.Unix(2, 0)}
	e3 := &models.CalcuttaEntry{ID: "e3", TotalPoints: 10, CreatedAt: time.Unix(1, 0)}

	// WHEN computing placements and payouts
	_, results := ComputeEntryPlacementsAndPayouts(
		[]*models.CalcuttaEntry{e1, e2, e3},
		nil,
	)

	// THEN the entry after the tie group gets finish position 3 (skips position 2)
	if results["e3"].FinishPosition != 3 {
		t.Fatalf("expected finish position 3, got %d", results["e3"].FinishPosition)
	}
}

func TestThatEmptyEntriesSliceReturnsEmptyResults(t *testing.T) {
	// GIVEN an empty non-nil entries slice
	entries := []*models.CalcuttaEntry{}

	// WHEN computing placements and payouts
	sorted, results := ComputeEntryPlacementsAndPayouts(entries, nil)

	// THEN returns empty non-nil outputs
	if len(sorted) != 0 {
		t.Fatalf("expected empty sorted slice, got %d entries", len(sorted))
	}
	if len(results) != 0 {
		t.Fatalf("expected empty results map, got %d entries", len(results))
	}
}
