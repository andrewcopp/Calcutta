package lab

import (
	"fmt"
	"math"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// BuildEvaluationSummary produces a structured narrative for an evaluation.
// It is a pure function with no database access.
func BuildEvaluationSummary(
	entryResults []models.LabEvaluationEntryResult,
	ourBids []models.LabEvaluationEntryBid,
	baselineEval *models.LabEvaluationDetail,
) *models.LabEvaluationSummary {
	nEntries := len(entryResults)

	// Find "Our Strategy" in entry results
	var ourResult *models.LabEvaluationEntryResult
	for i := range entryResults {
		if entryResults[i].EntryName == models.LabStrategyEntryName {
			ourResult = &entryResults[i]
			break
		}
	}

	// Compute percentile rank: 1 - (rank-1)/(nEntries-1)
	// Rank 1 of 10 => 1.0, rank 10 of 10 => 0.0
	percentile := 0.0
	if ourResult != nil && nEntries > 1 {
		percentile = 1.0 - float64(ourResult.Rank-1)/float64(nEntries-1)
	} else if ourResult != nil && nEntries == 1 {
		percentile = 1.0
	}

	// Baseline comparison
	var baseline *models.LabEvaluationBaselineComparison
	if baselineEval != nil && ourResult != nil &&
		baselineEval.MeanNormalizedPayout != nil && ourResult.MeanNormalizedPayout != nil {
		delta := *ourResult.MeanNormalizedPayout - *baselineEval.MeanNormalizedPayout

		var pTop1Delta float64
		if baselineEval.PTop1 != nil && ourResult.PTop1 != nil {
			pTop1Delta = *ourResult.PTop1 - *baselineEval.PTop1
		}

		interp := interpretDelta(delta)
		baseline = &models.LabEvaluationBaselineComparison{
			MeanPayoutDelta: delta,
			PTop1Delta:      pTop1Delta,
			Interpretation:  interp,
		}
	}

	// Top 3 holdings sorted by bid points descending
	topHoldings := buildTopHoldings(ourBids, 3)

	// Generate key insight
	insight := buildKeyInsight(ourResult, nEntries, baseline)

	return &models.LabEvaluationSummary{
		PercentileRank: percentile,
		VsBaseline:     baseline,
		NEntries:       nEntries,
		TopHoldings:    topHoldings,
		KeyInsight:     insight,
	}
}

// interpretDelta returns "better", "similar", or "worse" based on the
// mean payout delta. Within 5% of baseline is considered "similar".
func interpretDelta(delta float64) string {
	if math.Abs(delta) <= 0.05 {
		return "similar"
	}
	if delta > 0 {
		return "better"
	}
	return "worse"
}

// buildTopHoldings returns the top N bids sorted by bid points descending.
func buildTopHoldings(bids []models.LabEvaluationEntryBid, n int) []models.LabEvaluationTopHolding {
	if len(bids) == 0 {
		return []models.LabEvaluationTopHolding{}
	}

	sorted := make([]models.LabEvaluationEntryBid, len(bids))
	copy(sorted, bids)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].BidPoints > sorted[j].BidPoints
	})

	limit := n
	if limit > len(sorted) {
		limit = len(sorted)
	}

	out := make([]models.LabEvaluationTopHolding, limit)
	for i := 0; i < limit; i++ {
		out[i] = models.LabEvaluationTopHolding{
			SchoolName: sorted[i].SchoolName,
			Seed:       sorted[i].Seed,
			BidPoints:  sorted[i].BidPoints,
		}
	}
	return out
}

// buildKeyInsight generates a one-sentence summary of the evaluation.
func buildKeyInsight(
	ourResult *models.LabEvaluationEntryResult,
	nEntries int,
	baseline *models.LabEvaluationBaselineComparison,
) string {
	if ourResult == nil {
		return "Our Strategy was not found in this evaluation."
	}

	base := fmt.Sprintf("Ranked #%d of %d entries", ourResult.Rank, nEntries)

	if baseline != nil {
		sign := "+"
		if baseline.MeanPayoutDelta < 0 {
			sign = ""
		}
		return fmt.Sprintf(
			"%s (%s%.2fx mean payout vs naive baseline).",
			base, sign, baseline.MeanPayoutDelta,
		)
	}

	if ourResult.MeanNormalizedPayout != nil {
		return fmt.Sprintf("%s with %.2fx mean payout.", base, *ourResult.MeanNormalizedPayout)
	}

	return base + "."
}
