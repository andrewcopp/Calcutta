package calcutta_evaluations

import (
	"math"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation"
)

// --- calculateSimulationOutcomes tests ---

func TestThatPointsAreDistributedByProportionalOwnership(t *testing.T) {
	// GIVEN two entries with 60/40 bid split on one team that scored 100
	entries := map[string]*Entry{
		"alice": {Name: "Alice", Teams: map[string]int{"teamA": 60}},
		"bob":   {Name: "Bob", Teams: map[string]int{"teamA": 40}},
	}
	teamResults := []TeamSimResult{{TeamID: "teamA", Points: 100}}

	// WHEN calculating simulation outcomes
	results, err := CalculateSimulationOutcomes(1, entries, teamResults, map[int]int{1: 1000, 2: 500}, 1000)

	// THEN Alice gets 60 points and Bob gets 40 points
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	alicePoints := findResult(results, "Alice").TotalPoints
	if alicePoints != 60.0 {
		t.Errorf("expected Alice to have 60 points, got %v", alicePoints)
	}
}

func TestThatRank1IsAssignedToHighestScorer(t *testing.T) {
	// GIVEN two entries where Alice scores higher
	entries := map[string]*Entry{
		"alice": {Name: "Alice", Teams: map[string]int{"teamA": 100}},
		"bob":   {Name: "Bob", Teams: map[string]int{"teamB": 100}},
	}
	teamResults := []TeamSimResult{
		{TeamID: "teamA", Points: 200},
		{TeamID: "teamB", Points: 50},
	}

	// WHEN calculating simulation outcomes
	results, _ := CalculateSimulationOutcomes(1, entries, teamResults, map[int]int{1: 1000}, 1000)

	// THEN Alice is rank 1
	if findResult(results, "Alice").Rank != 1 {
		t.Errorf("expected Alice to be rank 1, got %d", findResult(results, "Alice").Rank)
	}
}

func TestThatTiesAreBrokenByNameAlphabetically(t *testing.T) {
	// GIVEN two entries with identical scores
	entries := map[string]*Entry{
		"bob":   {Name: "Bob", Teams: map[string]int{"teamA": 50}},
		"alice": {Name: "Alice", Teams: map[string]int{"teamA": 50}},
	}
	teamResults := []TeamSimResult{{TeamID: "teamA", Points: 100}}

	// WHEN calculating simulation outcomes
	results, _ := CalculateSimulationOutcomes(1, entries, teamResults, map[int]int{1: 1000, 2: 500}, 1000)

	// THEN Alice (alphabetically first) gets rank 1
	if findResult(results, "Alice").Rank != 1 {
		t.Errorf("expected Alice to be rank 1 (alphabetical tiebreak), got %d", findResult(results, "Alice").Rank)
	}
}

func TestThatNormalizedPayoutIsDividedByFirstPlacePayout(t *testing.T) {
	// GIVEN one entry that wins first place with payout 1000 and firstPlacePayout 2000
	entries := map[string]*Entry{
		"alice": {Name: "Alice", Teams: map[string]int{"teamA": 100}},
	}
	teamResults := []TeamSimResult{{TeamID: "teamA", Points: 100}}

	// WHEN calculating with firstPlacePayout = 2000
	results, _ := CalculateSimulationOutcomes(1, entries, teamResults, map[int]int{1: 1000}, 2000)

	// THEN normalized payout is 0.5 (1000/2000)
	if results[0].NormalizedPayout != 0.5 {
		t.Errorf("expected normalized payout 0.5, got %v", results[0].NormalizedPayout)
	}
}

func TestThatNormalizedPayoutIsZeroWhenFirstPlacePayoutIsZero(t *testing.T) {
	// GIVEN an entry with a payout
	entries := map[string]*Entry{
		"alice": {Name: "Alice", Teams: map[string]int{"teamA": 100}},
	}
	teamResults := []TeamSimResult{{TeamID: "teamA", Points: 100}}

	// WHEN firstPlacePayout is zero
	results, _ := CalculateSimulationOutcomes(1, entries, teamResults, map[int]int{1: 500}, 0)

	// THEN normalized payout is 0.0
	if results[0].NormalizedPayout != 0.0 {
		t.Errorf("expected normalized payout 0.0, got %v", results[0].NormalizedPayout)
	}
}

func TestThatTeamWithZeroTotalBidsDoesNotPanic(t *testing.T) {
	// GIVEN entries where no one bids on teamA but it scores points
	entries := map[string]*Entry{
		"alice": {Name: "Alice", Teams: map[string]int{"teamB": 100}},
	}
	teamResults := []TeamSimResult{
		{TeamID: "teamA", Points: 200},
		{TeamID: "teamB", Points: 50},
	}

	// WHEN calculating simulation outcomes
	results, err := CalculateSimulationOutcomes(1, entries, teamResults, map[int]int{1: 1000}, 1000)

	// THEN no panic and Alice gets points only from teamB
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].TotalPoints != 50.0 {
		t.Errorf("expected 50 points, got %v", results[0].TotalPoints)
	}
}

func TestThatSingleEntryGetsRankOne(t *testing.T) {
	// GIVEN a single entry
	entries := map[string]*Entry{
		"alice": {Name: "Alice", Teams: map[string]int{"teamA": 100}},
	}
	teamResults := []TeamSimResult{{TeamID: "teamA", Points: 50}}

	// WHEN calculating simulation outcomes
	results, _ := CalculateSimulationOutcomes(1, entries, teamResults, map[int]int{1: 1000}, 1000)

	// THEN Alice is rank 1
	if results[0].Rank != 1 {
		t.Errorf("expected rank 1, got %d", results[0].Rank)
	}
}

func TestThatEmptyEntriesReturnsEmptySlice(t *testing.T) {
	// GIVEN no entries
	entries := map[string]*Entry{}
	teamResults := []TeamSimResult{{TeamID: "teamA", Points: 100}}

	// WHEN calculating simulation outcomes
	results, err := CalculateSimulationOutcomes(1, entries, teamResults, map[int]int{1: 1000}, 1000)

	// THEN result slice is empty
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// --- calculatePerformanceMetrics tests ---

func TestThatMeanPayoutIsAverageOfNormalizedPayouts(t *testing.T) {
	// GIVEN results with normalized payouts 1.0, 0.5, 0.0
	results := []SimulationResult{
		{SimID: 1, EntryName: "Alice", NormalizedPayout: 1.0},
		{SimID: 2, EntryName: "Alice", NormalizedPayout: 0.5},
		{SimID: 3, EntryName: "Alice", NormalizedPayout: 0.0},
	}

	// WHEN calculating performance metrics
	perf := CalculatePerformanceMetrics(results)

	// THEN mean is 0.5
	if perf["Alice"].MeanPayout != 0.5 {
		t.Errorf("expected mean 0.5, got %v", perf["Alice"].MeanPayout)
	}
}

func TestThatMedianPayoutIsMiddleValue(t *testing.T) {
	// GIVEN results with normalized payouts 0.0, 0.5, 1.0 (sorted)
	results := []SimulationResult{
		{SimID: 1, EntryName: "Alice", NormalizedPayout: 0.0},
		{SimID: 2, EntryName: "Alice", NormalizedPayout: 0.5},
		{SimID: 3, EntryName: "Alice", NormalizedPayout: 1.0},
	}

	// WHEN calculating performance metrics
	perf := CalculatePerformanceMetrics(results)

	// THEN median is the middle value (index 1 of sorted [0.0, 0.5, 1.0] -> 0.5)
	if perf["Alice"].MedianPayout != 0.5 {
		t.Errorf("expected median 0.5, got %v", perf["Alice"].MedianPayout)
	}
}

func TestThatPTop1CountsSimsWithNormalizedPayoutAtLeastOne(t *testing.T) {
	// GIVEN 4 results where 1 has payout >= 1.0
	results := []SimulationResult{
		{SimID: 1, EntryName: "Alice", NormalizedPayout: 1.0},
		{SimID: 2, EntryName: "Alice", NormalizedPayout: 0.5},
		{SimID: 3, EntryName: "Alice", NormalizedPayout: 0.0},
		{SimID: 4, EntryName: "Alice", NormalizedPayout: 0.3},
	}

	// WHEN calculating performance metrics
	perf := CalculatePerformanceMetrics(results)

	// THEN PTop1 is 0.25 (1 out of 4)
	if perf["Alice"].PTop1 != 0.25 {
		t.Errorf("expected PTop1 0.25, got %v", perf["Alice"].PTop1)
	}
}

func TestThatPInMoneyCountsSimsWithPayoutGreaterThanZero(t *testing.T) {
	// GIVEN 4 results where 2 have payout > 0
	results := []SimulationResult{
		{SimID: 1, EntryName: "Alice", NormalizedPayout: 1.0},
		{SimID: 2, EntryName: "Alice", NormalizedPayout: 0.5},
		{SimID: 3, EntryName: "Alice", NormalizedPayout: 0.0},
		{SimID: 4, EntryName: "Alice", NormalizedPayout: 0.0},
	}

	// WHEN calculating performance metrics
	perf := CalculatePerformanceMetrics(results)

	// THEN PInMoney is 0.5 (2 out of 4)
	if perf["Alice"].PInMoney != 0.5 {
		t.Errorf("expected PInMoney 0.5, got %v", perf["Alice"].PInMoney)
	}
}

func TestThatTotalSimsEqualsInputCount(t *testing.T) {
	// GIVEN 3 simulation results for Alice
	results := []SimulationResult{
		{SimID: 1, EntryName: "Alice", NormalizedPayout: 1.0},
		{SimID: 2, EntryName: "Alice", NormalizedPayout: 0.5},
		{SimID: 3, EntryName: "Alice", NormalizedPayout: 0.0},
	}

	// WHEN calculating performance metrics
	perf := CalculatePerformanceMetrics(results)

	// THEN TotalSims is 3
	if perf["Alice"].TotalSims != 3 {
		t.Errorf("expected TotalSims 3, got %d", perf["Alice"].TotalSims)
	}
}

func TestThatPerformanceMetricsReturnsEmptyMapForNoResults(t *testing.T) {
	// GIVEN no results
	results := []SimulationResult{}

	// WHEN calculating performance metrics
	perf := CalculatePerformanceMetrics(results)

	// THEN map is empty
	if len(perf) != 0 {
		t.Errorf("expected empty map, got %d entries", len(perf))
	}
}

// --- ConvertSimulationResults tests ---

func TestThatConvertSimulationResultsConvertsWinsAndByesToPoints(t *testing.T) {
	// GIVEN a team with 2 wins and 1 bye, and scoring rules that award 10 per progress level
	simResults := []simulation.TeamSimulationResult{
		{SimID: 0, TeamID: "teamA", Wins: 2, Byes: 1},
	}
	rules := []scoring.Rule{
		{WinIndex: 1, PointsAwarded: 10},
		{WinIndex: 2, PointsAwarded: 20},
		{WinIndex: 3, PointsAwarded: 30},
	}

	// WHEN converting simulation results
	result := ConvertSimulationResults(simResults, 1, rules)

	// THEN the team gets 60 points (10+20+30 for progress=3)
	if result[0][0].Points != 60 {
		t.Errorf("expected 60 points, got %d", result[0][0].Points)
	}
}

func TestThatConvertSimulationResultsGroupsBySimID(t *testing.T) {
	// GIVEN results from two different simulations
	simResults := []simulation.TeamSimulationResult{
		{SimID: 0, TeamID: "teamA", Wins: 1, Byes: 1},
		{SimID: 0, TeamID: "teamB", Wins: 0, Byes: 1},
		{SimID: 1, TeamID: "teamA", Wins: 0, Byes: 1},
		{SimID: 1, TeamID: "teamB", Wins: 1, Byes: 1},
	}
	rules := []scoring.Rule{{WinIndex: 1, PointsAwarded: 10}}

	// WHEN converting simulation results
	result := ConvertSimulationResults(simResults, 2, rules)

	// THEN there are exactly 2 simulation groups
	if len(result) != 2 {
		t.Errorf("expected 2 simulation groups, got %d", len(result))
	}
}

func TestThatConvertSimulationResultsReturnsEmptyMapForEmptyInput(t *testing.T) {
	// GIVEN no simulation results
	simResults := []simulation.TeamSimulationResult{}
	rules := []scoring.Rule{{WinIndex: 1, PointsAwarded: 10}}

	// WHEN converting simulation results
	result := ConvertSimulationResults(simResults, 0, rules)

	// THEN the map is empty
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

// --- helpers ---

func findResult(results []SimulationResult, name string) SimulationResult {
	for _, r := range results {
		if r.EntryName == name {
			return r
		}
	}
	return SimulationResult{}
}

// floatEq checks approximate float equality.
func floatEq(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
