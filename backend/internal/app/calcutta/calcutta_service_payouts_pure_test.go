package calcutta

import (
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

func TestThatComputeEntryPlacementsAndPayoutsReturnsNilWhenEntriesNil(t *testing.T) {
	sorted, results := ComputeEntryPlacementsAndPayouts(nil, nil)
	if sorted != nil || results != nil {
		t.Fatalf("expected nil outputs")
	}
}

func TestThatComputeEntryPlacementsAndPayoutsSortsByTotalPointsDescending(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, Created: time.Unix(1, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 20, Created: time.Unix(2, 0)}

	sorted, _ := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e1, e2}, nil)

	if sorted[0].ID != "e2" {
		t.Fatalf("expected first entry to be e2, got %q", sorted[0].ID)
	}
}

func TestThatComputeEntryPlacementsAndPayoutsSortsTiesByCreatedDescending(t *testing.T) {
	eOld := &models.CalcuttaEntry{ID: "old", TotalPoints: 10, Created: time.Unix(1, 0)}
	eNew := &models.CalcuttaEntry{ID: "new", TotalPoints: 10, Created: time.Unix(2, 0)}

	sorted, _ := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{eOld, eNew}, nil)

	if sorted[0].ID != "new" {
		t.Fatalf("expected first entry to be new, got %q", sorted[0].ID)
	}
}

func TestThatComputeEntryPlacementsAndPayoutsMarksTiesWithinEpsilon(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10.00000, Created: time.Unix(2, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 10.00001, Created: time.Unix(1, 0)}

	_, results := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e1, e2}, nil)

	if !results["e1"].IsTied {
		t.Fatalf("expected e1 to be tied")
	}
}

func TestThatComputeEntryPlacementsAndPayoutsSetsFinishPositionOneForTopEntry(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 20, Created: time.Unix(1, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 10, Created: time.Unix(1, 0)}

	_, results := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e2, e1}, nil)

	if results["e1"].FinishPosition != 1 {
		t.Fatalf("expected finish position 1, got %d", results["e1"].FinishPosition)
	}
}

func TestThatComputeEntryPlacementsAndPayoutsSplitsPayoutAcrossTieGroup(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, Created: time.Unix(2, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 10, Created: time.Unix(1, 0)}
	p1 := &models.CalcuttaPayout{Position: 1, AmountCents: 100}
	p2 := &models.CalcuttaPayout{Position: 2, AmountCents: 50}

	_, results := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e1, e2}, []*models.CalcuttaPayout{p1, p2})

	if results["e1"].PayoutCents != 75 {
		t.Fatalf("expected payout 75, got %d", results["e1"].PayoutCents)
	}
}

func TestThatComputeEntryPlacementsAndPayoutsDistributesRemainderToEarlierEntryInSortedOrder(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, Created: time.Unix(2, 0)}
	e2 := &models.CalcuttaEntry{ID: "e2", TotalPoints: 10, Created: time.Unix(1, 0)}
	p1 := &models.CalcuttaPayout{Position: 1, AmountCents: 100}
	p2 := &models.CalcuttaPayout{Position: 2, AmountCents: 99}

	_, results := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e1, e2}, []*models.CalcuttaPayout{p1, p2})

	if results["e1"].PayoutCents != 100 {
		t.Fatalf("expected payout 100, got %d", results["e1"].PayoutCents)
	}
}

func TestThatComputeEntryPlacementsAndPayoutsSetsInTheMoneyWhenPayoutPositive(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, Created: time.Unix(1, 0)}
	p1 := &models.CalcuttaPayout{Position: 1, AmountCents: 1}

	_, results := ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e1}, []*models.CalcuttaPayout{p1})

	if !results["e1"].InTheMoney {
		t.Fatalf("expected in the money")
	}
}

func TestThatComputeEntryPlacementsAndPayoutsDoesNotMutateInputEntries(t *testing.T) {
	e1 := &models.CalcuttaEntry{ID: "e1", TotalPoints: 10, FinishPosition: 999, Created: time.Unix(1, 0)}

	_, _ = ComputeEntryPlacementsAndPayouts([]*models.CalcuttaEntry{e1}, nil)

	if e1.FinishPosition != 999 {
		t.Fatalf("expected input entry to remain unmodified")
	}
}
