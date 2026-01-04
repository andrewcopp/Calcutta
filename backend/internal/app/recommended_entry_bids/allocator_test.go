package recommended_entry_bids

import "testing"

func TestThatAllocateBidsSpendsBudgetAndRespectsTeamCountConstraints(t *testing.T) {
	teams := []Team{
		{ID: "a", ExpectedPoints: 10, MarketPoints: 10},
		{ID: "b", ExpectedPoints: 9, MarketPoints: 10},
		{ID: "c", ExpectedPoints: 8, MarketPoints: 10},
		{ID: "d", ExpectedPoints: 7, MarketPoints: 10},
	}

	res, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: 10,
		MinTeams:     2,
		MaxTeams:     3,
		MinBidPoints: 1,
		MaxBidPoints: 7,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(res.Bids) < 2 || len(res.Bids) > 3 {
		t.Fatalf("expected team count in [2,3], got %d", len(res.Bids))
	}

	sum := 0
	for _, bid := range res.Bids {
		sum += bid
	}
	if sum != 10 {
		t.Fatalf("expected to spend entire budget=10, got %d", sum)
	}
}

func TestThatAllocateBidsRespectsPerTeamMax(t *testing.T) {
	teams := []Team{
		{ID: "a", ExpectedPoints: 100, MarketPoints: 1},
		{ID: "b", ExpectedPoints: 50, MarketPoints: 1},
		{ID: "c", ExpectedPoints: 25, MarketPoints: 1},
	}

	res, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: 10,
		MinTeams:     1,
		MaxTeams:     3,
		MinBidPoints: 1,
		MaxBidPoints: 3,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for teamID, bid := range res.Bids {
		if bid > 3 {
			t.Fatalf("expected bid <= 3 for team %s, got %d", teamID, bid)
		}
	}

	sum := 0
	for _, bid := range res.Bids {
		sum += bid
	}
	capacity := 3 * 3
	if sum != capacity {
		t.Fatalf("expected to spend to capacity=%d, got %d", capacity, sum)
	}
}

func TestThatAllocateBidsReturnsEmptyWhenMinTeamsImpossible(t *testing.T) {
	teams := []Team{{ID: "a", ExpectedPoints: 10, MarketPoints: 1}}
	res, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: 2,
		MinTeams:     3,
		MaxTeams:     10,
		MinBidPoints: 1,
		MaxBidPoints: 50,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res.Bids) != 0 {
		t.Fatalf("expected empty bids, got %v", res.Bids)
	}
}

func TestThatAllocateBidsIsDeterministic(t *testing.T) {
	teams := []Team{
		{ID: "a", ExpectedPoints: 10, MarketPoints: 9},
		{ID: "b", ExpectedPoints: 10, MarketPoints: 9},
		{ID: "c", ExpectedPoints: 9, MarketPoints: 9},
		{ID: "d", ExpectedPoints: 8, MarketPoints: 9},
	}

	p := AllocationParams{
		BudgetPoints: 25,
		MinTeams:     2,
		MaxTeams:     3,
		MinBidPoints: 1,
		MaxBidPoints: 10,
	}

	first, err := AllocateBids(teams, p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	second, err := AllocateBids(teams, p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(first.Bids) != len(second.Bids) {
		t.Fatalf("expected same number of bids, got %d and %d", len(first.Bids), len(second.Bids))
	}
	for k, v := range first.Bids {
		if second.Bids[k] != v {
			t.Fatalf("expected deterministic bids, mismatch on %s: %d vs %d", k, v, second.Bids[k])
		}
	}
}
