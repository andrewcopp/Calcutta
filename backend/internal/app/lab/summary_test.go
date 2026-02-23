package lab

import (
	"math"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func ptrFloat(v float64) *float64 { return &v }

func makeEntryResults(names []string, payouts []float64) []models.LabEvaluationEntryResult {
	out := make([]models.LabEvaluationEntryResult, len(names))
	for i, n := range names {
		p := payouts[i]
		out[i] = models.LabEvaluationEntryResult{
			ID:                   "id-" + n,
			EntryName:            n,
			MeanNormalizedPayout: &p,
			Rank:                 i + 1,
		}
	}
	return out
}

func TestThatPercentileRankIsCorrectForFirstPlace(t *testing.T) {
	// GIVEN Our Strategy is ranked #1 of 10
	results := makeEntryResults(
		[]string{models.LabStrategyEntryName, "B", "C", "D", "E", "F", "G", "H", "I", "J"},
		[]float64{2.0, 1.9, 1.8, 1.7, 1.6, 1.5, 1.4, 1.3, 1.2, 1.1},
	)

	// WHEN building summary
	summary := BuildEvaluationSummary(results, nil, nil)

	// THEN percentile is 1.0
	if math.Abs(summary.PercentileRank-1.0) > 0.001 {
		t.Errorf("expected percentile 1.0, got %f", summary.PercentileRank)
	}
}

func TestThatPercentileRankIsCorrectForLastPlace(t *testing.T) {
	// GIVEN Our Strategy is ranked #10 of 10
	names := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", models.LabStrategyEntryName}
	payouts := []float64{2.0, 1.9, 1.8, 1.7, 1.6, 1.5, 1.4, 1.3, 1.2, 1.1}
	results := makeEntryResults(names, payouts)

	// WHEN building summary
	summary := BuildEvaluationSummary(results, nil, nil)

	// THEN percentile is 0.0
	if math.Abs(summary.PercentileRank-0.0) > 0.001 {
		t.Errorf("expected percentile 0.0, got %f", summary.PercentileRank)
	}
}

func TestThatBaselineComparisonShowsBetterWhenPayoutHigher(t *testing.T) {
	// GIVEN Our Strategy has 1.5x mean payout and baseline has 1.0x
	results := makeEntryResults(
		[]string{models.LabStrategyEntryName, "B"},
		[]float64{1.5, 1.0},
	)
	baseline := &models.LabEvaluationDetail{}
	baseline.MeanNormalizedPayout = ptrFloat(1.0)
	baseline.PTop1 = ptrFloat(0.10)

	results[0].PTop1 = ptrFloat(0.15)

	// WHEN building summary
	summary := BuildEvaluationSummary(results, nil, baseline)

	// THEN interpretation is "better"
	if summary.VsBaseline == nil {
		t.Fatal("expected VsBaseline to be non-nil")
	}
	if summary.VsBaseline.Interpretation != "better" {
		t.Errorf("expected interpretation 'better', got %q", summary.VsBaseline.Interpretation)
	}
}

func TestThatBaselineComparisonIsNilWhenNoBaseline(t *testing.T) {
	// GIVEN no baseline evaluation
	results := makeEntryResults(
		[]string{models.LabStrategyEntryName},
		[]float64{1.5},
	)

	// WHEN building summary with nil baseline
	summary := BuildEvaluationSummary(results, nil, nil)

	// THEN VsBaseline is nil
	if summary.VsBaseline != nil {
		t.Error("expected VsBaseline to be nil")
	}
}

func TestThatTopHoldingsLimitedToThree(t *testing.T) {
	// GIVEN 5 bids
	bids := []models.LabEvaluationEntryBid{
		{SchoolName: "Duke", Seed: 1, BidPoints: 20},
		{SchoolName: "UNC", Seed: 2, BidPoints: 15},
		{SchoolName: "Kansas", Seed: 1, BidPoints: 18},
		{SchoolName: "Kentucky", Seed: 3, BidPoints: 10},
		{SchoolName: "Gonzaga", Seed: 4, BidPoints: 8},
	}
	results := makeEntryResults(
		[]string{models.LabStrategyEntryName},
		[]float64{1.5},
	)

	// WHEN building summary
	summary := BuildEvaluationSummary(results, bids, nil)

	// THEN only top 3 holdings returned, sorted by bid points
	if len(summary.TopHoldings) != 3 {
		t.Fatalf("expected 3 top holdings, got %d", len(summary.TopHoldings))
	}
	if summary.TopHoldings[0].SchoolName != "Duke" {
		t.Errorf("expected top holding Duke, got %s", summary.TopHoldings[0].SchoolName)
	}
}

func TestThatKeyInsightContainsPayoutDelta(t *testing.T) {
	// GIVEN Our Strategy with baseline comparison
	results := makeEntryResults(
		[]string{models.LabStrategyEntryName, "B", "C"},
		[]float64{1.8, 1.5, 1.2},
	)
	results[0].PTop1 = ptrFloat(0.20)
	baseline := &models.LabEvaluationDetail{}
	baseline.MeanNormalizedPayout = ptrFloat(1.3)
	baseline.PTop1 = ptrFloat(0.10)

	// WHEN building summary
	summary := BuildEvaluationSummary(results, nil, baseline)

	// THEN key insight contains the delta
	if summary.KeyInsight == "" {
		t.Error("expected non-empty key insight")
	}
	if !containsSubstring(summary.KeyInsight, "0.50x") {
		t.Errorf("expected key insight to contain delta, got %q", summary.KeyInsight)
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsHelper(s, sub))
}

func containsHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
