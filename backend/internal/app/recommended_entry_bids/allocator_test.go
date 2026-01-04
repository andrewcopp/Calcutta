package recommended_entry_bids

import (
	"math"
	"testing"
)

func sumBids(bids map[string]int) int {
	sum := 0
	for _, bid := range bids {
		sum += bid
	}
	return sum
}

func totalObjective(teams []Team, bids map[string]int) float64 {
	total := 0.0
	for _, team := range teams {
		total += objective(team, bids[team.ID])
	}
	return total
}

func bruteForceBest(teams []Team, params AllocationParams) (float64, int, int) {
	negInf := math.Inf(-1)
	bestObj := negInf
	bestSpend := -1
	bestTeams := -1
	eps := 1e-9

	var rec func(i int, spend int, count int, bids map[string]int)
	rec = func(i int, spend int, count int, bids map[string]int) {
		if spend > params.BudgetPoints {
			return
		}
		if i == len(teams) {
			if count < params.MinTeams || count > params.MaxTeams {
				return
			}
			obj := totalObjective(teams, bids)
			if obj > bestObj+eps {
				bestObj = obj
				bestSpend = spend
				bestTeams = count
				return
			}
			if math.Abs(obj-bestObj) <= eps {
				if spend > bestSpend {
					bestSpend = spend
					bestTeams = count
					return
				}
				if spend == bestSpend && (bestTeams < 0 || count < bestTeams) {
					bestTeams = count
				}
			}
			return
		}

		team := teams[i]

		bids[team.ID] = 0
		rec(i+1, spend, count, bids)

		for bid := params.MinBidPoints; bid <= params.MaxBidPoints; bid++ {
			bids[team.ID] = bid
			rec(i+1, spend+bid, count+1, bids)
		}
	}

	rec(0, 0, 0, make(map[string]int))
	return bestObj, bestSpend, bestTeams
}

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

	if sumBids(res.Bids) != 10 {
		t.Fatalf("expected to spend entire budget=10, got %d", sumBids(res.Bids))
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

func TestThatAllocateBidsReturnsEmptyWhenNoTeams(t *testing.T) {
	res, err := AllocateBids(nil, AllocationParams{
		BudgetPoints: 10,
		MinTeams:     1,
		MaxTeams:     3,
		MinBidPoints: 1,
		MaxBidPoints: 10,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res.Bids) != 0 {
		t.Fatalf("expected empty bids, got %v", res.Bids)
	}
}

func TestThatAllocateBidsNormalizesParams(t *testing.T) {
	teams := []Team{
		{ID: "a", ExpectedPoints: 10, MarketPoints: 10},
		{ID: "b", ExpectedPoints: 9, MarketPoints: 10},
	}

	res, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: 5,
		MinTeams:     0,
		MaxTeams:     0,
		MinBidPoints: 0,
		MaxBidPoints: 0,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sumBids(res.Bids) != 1 {
		t.Fatalf("expected to spend 1 due to default max_bid=1, got %d", sumBids(res.Bids))
	}
}

func TestThatAllocateBidsMatchesBruteForceOptimumOnSmallInstance(t *testing.T) {
	teams := []Team{
		{ID: "a", ExpectedPoints: 10, MarketPoints: 1},
		{ID: "b", ExpectedPoints: 9, MarketPoints: 2},
		{ID: "c", ExpectedPoints: 8, MarketPoints: 3},
		{ID: "d", ExpectedPoints: 7, MarketPoints: 4},
	}
	p := AllocationParams{
		BudgetPoints: 6,
		MinTeams:     1,
		MaxTeams:     3,
		MinBidPoints: 1,
		MaxBidPoints: 4,
	}

	res, err := AllocateBids(teams, p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bestObj, bestSpend, bestTeams := bruteForceBest(teams, p)
	gotObj := totalObjective(teams, res.Bids)
	gotSpend := sumBids(res.Bids)
	gotTeams := len(res.Bids)

	eps := 1e-9
	if math.Abs(gotObj-bestObj) > eps || gotSpend != bestSpend || gotTeams != bestTeams {
		t.Fatalf("expected brute optimum (obj=%v spend=%d teams=%d), got (obj=%v spend=%d teams=%d) bids=%v", bestObj, bestSpend, bestTeams, gotObj, gotSpend, gotTeams, res.Bids)
	}
}

func TestThatAllocateBidsPrefersLargeCapacityTeamWhenSmallMarketTeamHasHigherNaiveROI(t *testing.T) {
	teams := []Team{
		{ID: "a", ExpectedPoints: 100, MarketPoints: 1},
		{ID: "b", ExpectedPoints: 300, MarketPoints: 100},
	}

	res, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: 50,
		MinTeams:     1,
		MaxTeams:     1,
		MinBidPoints: 1,
		MaxBidPoints: 50,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res.Bids) != 1 || res.Bids["b"] != 50 {
		t.Fatalf("expected to allocate entire budget to team b, got %v", res.Bids)
	}
}

func TestThatAllocateBidsAllocatesToHighCapacityTeamEvenWhenInitialROIIsHigherOnSmallMarketTeams(t *testing.T) {
	teams := []Team{
		{ID: "a", ExpectedPoints: 10, MarketPoints: 1},
		{ID: "b", ExpectedPoints: 10, MarketPoints: 1},
		{ID: "c", ExpectedPoints: 200, MarketPoints: 50},
	}

	res, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: 10,
		MinTeams:     1,
		MaxTeams:     2,
		MinBidPoints: 1,
		MaxBidPoints: 10,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if sumBids(res.Bids) != 10 {
		t.Fatalf("expected to spend entire budget=10, got %d", sumBids(res.Bids))
	}
	if len(res.Bids) < 1 || len(res.Bids) > 2 {
		t.Fatalf("expected team count in [1,2], got %d", len(res.Bids))
	}

	if res.Bids["c"] != 9 || res.Bids["a"] != 1 || res.Bids["b"] != 0 {
		t.Fatalf("expected c=9 a=1 b=0, got %v", res.Bids)
	}
}

func TestThatAllocateBidsAvoidsOverbiddingSmallMarketTeamWhenDiversificationIsAllowed(t *testing.T) {
	teams := []Team{
		{ID: "favorite", ExpectedPoints: 200, MarketPoints: 100},
		{ID: "longshot", ExpectedPoints: 20, MarketPoints: 0.1},
	}

	res, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: 10,
		MinTeams:     2,
		MaxTeams:     2,
		MinBidPoints: 1,
		MaxBidPoints: 10,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Bids["longshot"] != 1 || res.Bids["favorite"] != 9 {
		t.Fatalf("expected longshot=1 favorite=9, got %v", res.Bids)
	}
}

func TestThatAllocateBidsBreaksTiesByChoosingEarlierTeamID(t *testing.T) {
	teams := []Team{
		{ID: "a", ExpectedPoints: 10, MarketPoints: 10},
		{ID: "b", ExpectedPoints: 10, MarketPoints: 10},
	}

	res, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: 1,
		MinTeams:     1,
		MaxTeams:     1,
		MinBidPoints: 1,
		MaxBidPoints: 1,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res.Bids) != 1 || res.Bids["a"] != 1 {
		t.Fatalf("expected to pick team a on tie, got %v", res.Bids)
	}
}

func TestThatAllocateBidsBreaksGlobalTiesBySpendingMoreAndChoosingFewerTeams(t *testing.T) {
	teams := []Team{
		{ID: "a", ExpectedPoints: 0, MarketPoints: 10},
		{ID: "b", ExpectedPoints: 0, MarketPoints: 10},
	}

	res, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: 5,
		MinTeams:     1,
		MaxTeams:     2,
		MinBidPoints: 1,
		MaxBidPoints: 5,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sumBids(res.Bids) != 5 || len(res.Bids) != 1 || res.Bids["a"] != 5 {
		t.Fatalf("expected to spend 5 on a single team a, got %v", res.Bids)
	}
}

func TestThatAllocateBidsAvoidsPayingMinBidForSecondTeamWhenNotWorthIt(t *testing.T) {
	teams := []Team{
		{ID: "a", ExpectedPoints: 100, MarketPoints: 10},
		{ID: "b", ExpectedPoints: 5, MarketPoints: 1},
	}

	res, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: 6,
		MinTeams:     1,
		MaxTeams:     2,
		MinBidPoints: 3,
		MaxBidPoints: 6,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if sumBids(res.Bids) != 6 {
		t.Fatalf("expected to spend entire budget=6, got %d", sumBids(res.Bids))
	}
	if len(res.Bids) != 1 || res.Bids["a"] != 6 {
		t.Fatalf("expected a=6 only, got %v", res.Bids)
	}
}

func TestThatAllocateBidsSpendsAllBudgetOnBestTeamWhenMarketIsZero(t *testing.T) {
	teams := []Team{
		{ID: "a", ExpectedPoints: 10, MarketPoints: 0},
		{ID: "b", ExpectedPoints: 9, MarketPoints: 0},
	}

	res, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: 10,
		MinTeams:     1,
		MaxTeams:     2,
		MinBidPoints: 1,
		MaxBidPoints: 10,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if sumBids(res.Bids) != 10 {
		t.Fatalf("expected to spend entire budget=10, got %d", sumBids(res.Bids))
	}
	if len(res.Bids) != 2 || res.Bids["a"] != 9 || res.Bids["b"] != 1 {
		t.Fatalf("expected a=9 b=1 (since MarketPoints=0 makes objective flat in bid), got %v", res.Bids)
	}
}

func TestThatAllocateBidsMatchesBruteForceOptimumOnManyLongshotsVsFavorite(t *testing.T) {
	teams := []Team{
		{ID: "favorite", ExpectedPoints: 200, MarketPoints: 100},
		{ID: "l1", ExpectedPoints: 10, MarketPoints: 0.1},
		{ID: "l2", ExpectedPoints: 10, MarketPoints: 0.1},
		{ID: "l3", ExpectedPoints: 10, MarketPoints: 0.1},
		{ID: "l4", ExpectedPoints: 10, MarketPoints: 0.1},
	}

	p := AllocationParams{
		BudgetPoints: 10,
		MinTeams:     1,
		MaxTeams:     3,
		MinBidPoints: 1,
		MaxBidPoints: 10,
	}

	res, err := AllocateBids(teams, p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bestObj, bestSpend, bestTeams := bruteForceBest(teams, p)
	gotObj := totalObjective(teams, res.Bids)
	gotSpend := sumBids(res.Bids)
	gotTeams := len(res.Bids)

	eps := 1e-9
	if math.Abs(gotObj-bestObj) > eps || gotSpend != bestSpend || gotTeams != bestTeams {
		t.Fatalf("expected brute optimum (obj=%v spend=%d teams=%d), got (obj=%v spend=%d teams=%d) bids=%v", bestObj, bestSpend, bestTeams, gotObj, gotSpend, gotTeams, res.Bids)
	}
	if res.Bids["favorite"] <= 0 {
		t.Fatalf("expected favorite to receive some allocation, got %v", res.Bids)
	}
}
