package calcutta_evaluations

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// --- buildLabEvaluationResult tests ---

func TestThatBuildLabEvaluationResultReturnsMeanNormalizedPayout(t *testing.T) {
	// GIVEN simulation results where lab entry has normalized payouts 1.0, 0.5, 0.0 (mean = 0.5)
	allResults := []SimulationResult{
		{SimID: 1, EntryName: models.LabStrategyEntryName, NormalizedPayout: 1.0},
		{SimID: 2, EntryName: models.LabStrategyEntryName, NormalizedPayout: 0.5},
		{SimID: 3, EntryName: models.LabStrategyEntryName, NormalizedPayout: 0.0},
	}

	// WHEN building the lab evaluation result
	result, err := buildLabEvaluationResult(allResults, 3)

	// THEN the mean normalized payout is 0.5
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MeanNormalizedPayout != 0.5 {
		t.Errorf("expected MeanNormalizedPayout 0.5, got %v", result.MeanNormalizedPayout)
	}
}

func TestThatBuildLabEvaluationResultReturnsPTop1(t *testing.T) {
	// GIVEN simulation results where lab entry wins first place in 1 out of 4 sims
	allResults := []SimulationResult{
		{SimID: 1, EntryName: models.LabStrategyEntryName, NormalizedPayout: 1.0},
		{SimID: 2, EntryName: models.LabStrategyEntryName, NormalizedPayout: 0.5},
		{SimID: 3, EntryName: models.LabStrategyEntryName, NormalizedPayout: 0.0},
		{SimID: 4, EntryName: models.LabStrategyEntryName, NormalizedPayout: 0.0},
	}

	// WHEN building the lab evaluation result
	result, err := buildLabEvaluationResult(allResults, 4)

	// THEN PTop1 is 0.25 (1 out of 4 sims had payout >= 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PTop1 != 0.25 {
		t.Errorf("expected PTop1 0.25, got %v", result.PTop1)
	}
}

func TestThatBuildLabEvaluationResultReturnsErrorWhenLabEntryMissing(t *testing.T) {
	// GIVEN simulation results that do not contain the lab strategy entry
	allResults := []SimulationResult{
		{SimID: 1, EntryName: "SomeOtherEntry", NormalizedPayout: 1.0},
		{SimID: 2, EntryName: "SomeOtherEntry", NormalizedPayout: 0.5},
	}

	// WHEN building the lab evaluation result
	_, err := buildLabEvaluationResult(allResults, 2)

	// THEN an error is returned
	if err == nil {
		t.Error("expected error when lab entry is missing from results")
	}
}

// --- rankEntryPerformance tests ---

func TestThatRankEntryPerformanceRanksEntriesByMeanPayoutDescending(t *testing.T) {
	// GIVEN three entries with different mean payouts
	performance := map[string]*EntryPerformance{
		"Low":    {EntryName: "Low", MeanPayout: 0.1, PTop1: 0.0, PInMoney: 0.1},
		"High":   {EntryName: "High", MeanPayout: 0.9, PTop1: 0.5, PInMoney: 0.9},
		"Medium": {EntryName: "Medium", MeanPayout: 0.5, PTop1: 0.2, PInMoney: 0.5},
	}

	// WHEN ranking entry performance
	results := RankEntryPerformance(performance)

	// THEN the first entry has the highest mean payout
	if results[0].EntryName != "High" {
		t.Errorf("expected first ranked entry to be 'High', got '%s'", results[0].EntryName)
	}
}

func TestThatRankEntryPerformanceAssignsSequentialRanks(t *testing.T) {
	// GIVEN three entries with different mean payouts
	performance := map[string]*EntryPerformance{
		"A": {EntryName: "A", MeanPayout: 0.9},
		"B": {EntryName: "B", MeanPayout: 0.5},
		"C": {EntryName: "C", MeanPayout: 0.1},
	}

	// WHEN ranking entry performance
	results := RankEntryPerformance(performance)

	// THEN ranks are sequential: 1, 2, 3
	ranks := make([]int, len(results))
	for i, r := range results {
		ranks[i] = r.Rank
	}
	expected := []int{1, 2, 3}
	for i, rank := range ranks {
		if rank != expected[i] {
			t.Errorf("expected rank %d at position %d, got %d", expected[i], i, rank)
			return
		}
	}
}
