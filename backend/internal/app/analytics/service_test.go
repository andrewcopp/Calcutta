package analytics

import (
	"math"
	"testing"
)

// --- clampLimit ---

func TestThatClampLimitDefaultsToHundredForZero(t *testing.T) {
	// GIVEN a zero limit
	// WHEN clamping the limit
	got := clampLimit(0)

	// THEN the result is 100
	if got != 100 {
		t.Errorf("expected 100, got %d", got)
	}
}

func TestThatClampLimitDefaultsToHundredForNegativeValue(t *testing.T) {
	// GIVEN a negative limit
	// WHEN clamping the limit
	got := clampLimit(-5)

	// THEN the result is 100
	if got != 100 {
		t.Errorf("expected 100, got %d", got)
	}
}

func TestThatClampLimitCapsAtFiveHundred(t *testing.T) {
	// GIVEN a limit above 500
	// WHEN clamping the limit
	got := clampLimit(1000)

	// THEN the result is 500
	if got != 500 {
		t.Errorf("expected 500, got %d", got)
	}
}

func TestThatClampLimitPreservesValidValue(t *testing.T) {
	// GIVEN a limit within the valid range
	// WHEN clamping the limit
	got := clampLimit(50)

	// THEN the result is unchanged
	if got != 50 {
		t.Errorf("expected 50, got %d", got)
	}
}

func TestThatClampLimitPreservesLowerBoundary(t *testing.T) {
	// GIVEN a limit of exactly 1
	// WHEN clamping the limit
	got := clampLimit(1)

	// THEN the result is 1
	if got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}

func TestThatClampLimitPreservesUpperBoundary(t *testing.T) {
	// GIVEN a limit of exactly 500
	// WHEN clamping the limit
	got := clampLimit(500)

	// THEN the result is 500
	if got != 500 {
		t.Errorf("expected 500, got %d", got)
	}
}

// --- meanFloat64 ---

func TestThatMeanFloat64ReturnsZeroForEmptySlice(t *testing.T) {
	// GIVEN an empty slice
	// WHEN calculating the mean
	got := meanFloat64(nil)

	// THEN the result is 0
	if got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

func TestThatMeanFloat64ReturnsValueForSingleElement(t *testing.T) {
	// GIVEN a slice with a single value
	// WHEN calculating the mean
	got := meanFloat64([]float64{7.5})

	// THEN the result equals that value
	if got != 7.5 {
		t.Errorf("expected 7.5, got %v", got)
	}
}

func TestThatMeanFloat64ReturnsCorrectAverage(t *testing.T) {
	// GIVEN a slice of known values
	// WHEN calculating the mean
	got := meanFloat64([]float64{2, 4, 6})

	// THEN the result is the arithmetic mean
	if got != 4 {
		t.Errorf("expected 4, got %v", got)
	}
}

// --- stddevFloat64 ---

func TestThatStddevFloat64ReturnsZeroForEmptySlice(t *testing.T) {
	// GIVEN an empty slice
	// WHEN calculating the standard deviation
	got := stddevFloat64(nil, 0)

	// THEN the result is 0
	if got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

func TestThatStddevFloat64ReturnsZeroForSingleValue(t *testing.T) {
	// GIVEN a slice with a single value
	// WHEN calculating the standard deviation
	got := stddevFloat64([]float64{5}, 5)

	// THEN the result is 0
	if got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

func TestThatStddevFloat64CalculatesSampleStandardDeviation(t *testing.T) {
	// GIVEN values [2, 4, 6] with mean 4
	// sample variance = ((2-4)^2 + (4-4)^2 + (6-4)^2) / (3-1) = 8/2 = 4
	// sample stddev = sqrt(4) = 2
	// WHEN calculating the standard deviation
	got := stddevFloat64([]float64{2, 4, 6}, 4)

	// THEN the result is 2
	if got != 2 {
		t.Errorf("expected 2, got %v", got)
	}
}

// --- quantileSorted ---

func TestThatQuantileSortedReturnsZeroForEmptySlice(t *testing.T) {
	// GIVEN an empty slice
	// WHEN calculating the median
	got := quantileSorted(nil, 0.5)

	// THEN the result is 0
	if got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

func TestThatQuantileSortedReturnsMinForZeroQuantile(t *testing.T) {
	// GIVEN a sorted slice
	// WHEN calculating the 0th quantile
	got := quantileSorted([]float64{1, 2, 3, 4}, 0)

	// THEN the result is the minimum value
	if got != 1 {
		t.Errorf("expected 1, got %v", got)
	}
}

func TestThatQuantileSortedReturnsMaxForOneQuantile(t *testing.T) {
	// GIVEN a sorted slice
	// WHEN calculating the 1st quantile
	got := quantileSorted([]float64{1, 2, 3, 4}, 1)

	// THEN the result is the maximum value
	if got != 4 {
		t.Errorf("expected 4, got %v", got)
	}
}

func TestThatQuantileSortedReturnsMedianForEvenCount(t *testing.T) {
	// GIVEN a sorted slice with 4 elements
	// position = 0.5 * (4-1) = 1.5, interpolating between index 1 and 2
	// WHEN calculating the median
	got := quantileSorted([]float64{10, 20, 30, 40}, 0.5)

	// THEN the result is 25 (interpolated between 20 and 30)
	if got != 25 {
		t.Errorf("expected 25, got %v", got)
	}
}

func TestThatQuantileSortedReturnsExactValueWhenPositionIsInteger(t *testing.T) {
	// GIVEN a sorted slice with 5 elements
	// position = 0.5 * (5-1) = 2.0, exact index
	// WHEN calculating the median
	got := quantileSorted([]float64{10, 20, 30, 40, 50}, 0.5)

	// THEN the result is the middle element
	if got != 30 {
		t.Errorf("expected 30, got %v", got)
	}
}

func TestThatQuantileSortedReturnsMinForNegativeQuantile(t *testing.T) {
	// GIVEN a sorted slice
	// WHEN calculating a negative quantile
	got := quantileSorted([]float64{5, 10, 15}, -0.1)

	// THEN the result is the minimum value
	if got != 5 {
		t.Errorf("expected 5, got %v", got)
	}
}

func TestThatQuantileSortedReturnsMaxForQuantileAboveOne(t *testing.T) {
	// GIVEN a sorted slice
	// WHEN calculating a quantile above 1
	got := quantileSorted([]float64{5, 10, 15}, 1.5)

	// THEN the result is the maximum value
	if got != 15 {
		t.Errorf("expected 15, got %v", got)
	}
}

// --- CalculateSeedAnalyticsResults ---

func TestThatCalculateSeedAnalyticsResultsReturnsEmptySliceForEmptyInput(t *testing.T) {
	// GIVEN empty input data
	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(nil, 0, 0)

	// THEN the result has zero length
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestThatCalculateSeedAnalyticsResultsCalculatesPointsPercentage(t *testing.T) {
	// GIVEN a seed with 30 out of 100 total points
	data := []SeedAnalyticsInput{{Seed: 1, TotalPoints: 30, TotalInvestment: 10, TeamCount: 1}}

	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(data, 100, 50)

	// THEN points percentage is 30%
	if results[0].PointsPercentage != 30 {
		t.Errorf("expected 30, got %v", results[0].PointsPercentage)
	}
}

func TestThatCalculateSeedAnalyticsResultsCalculatesInvestmentPercentage(t *testing.T) {
	// GIVEN a seed with 10 out of 50 total investment
	data := []SeedAnalyticsInput{{Seed: 1, TotalPoints: 30, TotalInvestment: 10, TeamCount: 1}}

	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(data, 100, 50)

	// THEN investment percentage is 20%
	if results[0].InvestmentPercentage != 20 {
		t.Errorf("expected 20, got %v", results[0].InvestmentPercentage)
	}
}

func TestThatCalculateSeedAnalyticsResultsCalculatesAveragePoints(t *testing.T) {
	// GIVEN a seed with 30 total points across 3 teams
	data := []SeedAnalyticsInput{{Seed: 1, TotalPoints: 30, TotalInvestment: 10, TeamCount: 3}}

	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(data, 100, 50)

	// THEN average points is 10
	if results[0].AveragePoints != 10 {
		t.Errorf("expected 10, got %v", results[0].AveragePoints)
	}
}

func TestThatCalculateSeedAnalyticsResultsCalculatesAverageInvestment(t *testing.T) {
	// GIVEN a seed with 12 total investment across 4 teams
	data := []SeedAnalyticsInput{{Seed: 1, TotalPoints: 30, TotalInvestment: 12, TeamCount: 4}}

	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(data, 100, 50)

	// THEN average investment is 3
	if results[0].AverageInvestment != 3 {
		t.Errorf("expected 3, got %v", results[0].AverageInvestment)
	}
}

func TestThatCalculateSeedAnalyticsResultsCalculatesROIAsRatioToBaselineROI(t *testing.T) {
	// GIVEN a seed with 3x ROI and a baseline of 3x ROI
	// actual ROI = 30/10 = 3, baseline = 60/20 = 3, normalized = 3/3 = 1.0
	data := []SeedAnalyticsInput{{Seed: 1, TotalPoints: 30, TotalInvestment: 10, TeamCount: 1}}

	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(data, 60, 20)

	// THEN ROI equals 1.0
	if results[0].ROI != 1.0 {
		t.Errorf("expected 1.0, got %v", results[0].ROI)
	}
}

func TestThatCalculateSeedAnalyticsResultsReturnsZeroROIWhenSeedHasNoInvestment(t *testing.T) {
	// GIVEN a seed with zero investment
	data := []SeedAnalyticsInput{{Seed: 16, TotalPoints: 5, TotalInvestment: 0, TeamCount: 1}}

	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(data, 100, 50)

	// THEN ROI is zero (cannot divide by zero investment)
	if results[0].ROI != 0 {
		t.Errorf("expected 0, got %v", results[0].ROI)
	}
}

func TestThatCalculateSeedAnalyticsResultsReturnsZeroPercentagesWhenTotalsAreZero(t *testing.T) {
	// GIVEN seed data but zero overall totals
	data := []SeedAnalyticsInput{{Seed: 1, TotalPoints: 30, TotalInvestment: 10, TeamCount: 1}}

	// WHEN calculating seed analytics with zero totals
	results := CalculateSeedAnalyticsResults(data, 0, 0)

	// THEN points percentage is zero
	if results[0].PointsPercentage != 0 {
		t.Errorf("expected 0, got %v", results[0].PointsPercentage)
	}
}

func TestThatCalculateSeedAnalyticsResultsReturnsZeroAveragesForZeroTeamCount(t *testing.T) {
	// GIVEN a seed with zero team count
	data := []SeedAnalyticsInput{{Seed: 1, TotalPoints: 30, TotalInvestment: 10, TeamCount: 0}}

	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(data, 100, 50)

	// THEN average points is zero
	if results[0].AveragePoints != 0 {
		t.Errorf("expected 0, got %v", results[0].AveragePoints)
	}
}

func TestThatCalculateSeedAnalyticsResultsShowsOverperformanceWithROIAboveOne(t *testing.T) {
	// GIVEN a seed with ROI above the baseline
	// actual ROI = 40/10 = 4, baseline = 60/30 = 2, normalized = 4/2 = 2.0
	data := []SeedAnalyticsInput{{Seed: 1, TotalPoints: 40, TotalInvestment: 10, TeamCount: 1}}

	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(data, 60, 30)

	// THEN ROI is 2.0 (double the average performance)
	if results[0].ROI != 2.0 {
		t.Errorf("expected 2.0, got %v", results[0].ROI)
	}
}

// --- CalculateSeedVarianceResults ---

func TestThatCalculateSeedVarianceResultsReturnsEmptySliceForEmptyInput(t *testing.T) {
	// GIVEN empty input data
	// WHEN calculating seed variance
	results := CalculateSeedVarianceResults(nil)

	// THEN the result has zero length
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestThatCalculateSeedVarianceResultsCalculatesInvestmentCV(t *testing.T) {
	// GIVEN seed variance data with investment stddev 10 and mean 20
	data := []SeedVarianceInput{{Seed: 1, InvestmentStdDev: 10, InvestmentMean: 20, PointsStdDev: 3, PointsMean: 6, TeamCount: 4}}

	// WHEN calculating seed variance
	results := CalculateSeedVarianceResults(data)

	// THEN investment CV is 0.5 (10/20)
	if results[0].InvestmentCV != 0.5 {
		t.Errorf("expected 0.5, got %v", results[0].InvestmentCV)
	}
}

func TestThatCalculateSeedVarianceResultsCalculatesPointsCV(t *testing.T) {
	// GIVEN seed variance data with points stddev 3 and mean 6
	data := []SeedVarianceInput{{Seed: 1, InvestmentStdDev: 10, InvestmentMean: 20, PointsStdDev: 3, PointsMean: 6, TeamCount: 4}}

	// WHEN calculating seed variance
	results := CalculateSeedVarianceResults(data)

	// THEN points CV is 0.5 (3/6)
	if results[0].PointsCV != 0.5 {
		t.Errorf("expected 0.5, got %v", results[0].PointsCV)
	}
}

func TestThatCalculateSeedVarianceResultsCalculatesVarianceRatioAsInvestmentCVOverPointsCV(t *testing.T) {
	// GIVEN seed data where investment CV = 0.5 and points CV = 0.25
	data := []SeedVarianceInput{{Seed: 1, InvestmentStdDev: 10, InvestmentMean: 20, PointsStdDev: 3, PointsMean: 12, TeamCount: 4}}

	// WHEN calculating seed variance
	results := CalculateSeedVarianceResults(data)

	// THEN variance ratio is 2.0 (0.5 / 0.25)
	if results[0].VarianceRatio != 2.0 {
		t.Errorf("expected 2.0, got %v", results[0].VarianceRatio)
	}
}

func TestThatCalculateSeedVarianceResultsReturnsZeroInvestmentCVWhenMeanIsZero(t *testing.T) {
	// GIVEN seed data with zero investment mean
	data := []SeedVarianceInput{{Seed: 16, InvestmentStdDev: 5, InvestmentMean: 0, PointsStdDev: 3, PointsMean: 6, TeamCount: 4}}

	// WHEN calculating seed variance
	results := CalculateSeedVarianceResults(data)

	// THEN investment CV is zero
	if results[0].InvestmentCV != 0 {
		t.Errorf("expected 0, got %v", results[0].InvestmentCV)
	}
}

func TestThatCalculateSeedVarianceResultsReturnsZeroVarianceRatioWhenPointsCVIsZero(t *testing.T) {
	// GIVEN seed data where points stddev is zero (all teams score the same)
	data := []SeedVarianceInput{{Seed: 1, InvestmentStdDev: 10, InvestmentMean: 20, PointsStdDev: 0, PointsMean: 6, TeamCount: 4}}

	// WHEN calculating seed variance
	results := CalculateSeedVarianceResults(data)

	// THEN variance ratio is zero (cannot divide by zero points CV)
	if results[0].VarianceRatio != 0 {
		t.Errorf("expected 0, got %v", results[0].VarianceRatio)
	}
}

// --- CalculateRegionAnalyticsResults ---

func TestThatCalculateRegionAnalyticsResultsReturnsEmptySliceForEmptyInput(t *testing.T) {
	// GIVEN empty input data
	// WHEN calculating region analytics
	results := CalculateRegionAnalyticsResults(nil, 0, 0)

	// THEN the result has zero length
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestThatCalculateRegionAnalyticsResultsCalculatesPointsPercentage(t *testing.T) {
	// GIVEN a region with 40 out of 200 total points
	data := []RegionAnalyticsInput{{Region: "East", TotalPoints: 40, TotalInvestment: 20, TeamCount: 16}}

	// WHEN calculating region analytics
	results := CalculateRegionAnalyticsResults(data, 200, 100)

	// THEN points percentage is 20%
	if results[0].PointsPercentage != 20 {
		t.Errorf("expected 20, got %v", results[0].PointsPercentage)
	}
}

func TestThatCalculateRegionAnalyticsResultsCalculatesInvestmentPercentage(t *testing.T) {
	// GIVEN a region with 20 out of 100 total investment
	data := []RegionAnalyticsInput{{Region: "East", TotalPoints: 40, TotalInvestment: 20, TeamCount: 16}}

	// WHEN calculating region analytics
	results := CalculateRegionAnalyticsResults(data, 200, 100)

	// THEN investment percentage is 20%
	if results[0].InvestmentPercentage != 20 {
		t.Errorf("expected 20, got %v", results[0].InvestmentPercentage)
	}
}

func TestThatCalculateRegionAnalyticsResultsCalculatesAveragePoints(t *testing.T) {
	// GIVEN a region with 48 total points across 16 teams
	data := []RegionAnalyticsInput{{Region: "West", TotalPoints: 48, TotalInvestment: 32, TeamCount: 16}}

	// WHEN calculating region analytics
	results := CalculateRegionAnalyticsResults(data, 200, 100)

	// THEN average points is 3
	if results[0].AveragePoints != 3 {
		t.Errorf("expected 3, got %v", results[0].AveragePoints)
	}
}

func TestThatCalculateRegionAnalyticsResultsCalculatesAverageInvestment(t *testing.T) {
	// GIVEN a region with 32 total investment across 16 teams
	data := []RegionAnalyticsInput{{Region: "West", TotalPoints: 48, TotalInvestment: 32, TeamCount: 16}}

	// WHEN calculating region analytics
	results := CalculateRegionAnalyticsResults(data, 200, 100)

	// THEN average investment is 2
	if results[0].AverageInvestment != 2 {
		t.Errorf("expected 2, got %v", results[0].AverageInvestment)
	}
}

func TestThatCalculateRegionAnalyticsResultsCalculatesNormalizedROI(t *testing.T) {
	// GIVEN a region with ROI above baseline
	// actual ROI = 60/20 = 3, baseline = 200/100 = 2, normalized = 3/2 = 1.5
	data := []RegionAnalyticsInput{{Region: "East", TotalPoints: 60, TotalInvestment: 20, TeamCount: 16}}

	// WHEN calculating region analytics
	results := CalculateRegionAnalyticsResults(data, 200, 100)

	// THEN ROI is 1.5
	if results[0].ROI != 1.5 {
		t.Errorf("expected 1.5, got %v", results[0].ROI)
	}
}

func TestThatCalculateRegionAnalyticsResultsReturnsZeroROIWhenRegionHasNoInvestment(t *testing.T) {
	// GIVEN a region with zero investment
	data := []RegionAnalyticsInput{{Region: "South", TotalPoints: 10, TotalInvestment: 0, TeamCount: 16}}

	// WHEN calculating region analytics
	results := CalculateRegionAnalyticsResults(data, 200, 100)

	// THEN ROI is zero
	if results[0].ROI != 0 {
		t.Errorf("expected 0, got %v", results[0].ROI)
	}
}

func TestThatCalculateRegionAnalyticsResultsReturnsZeroAveragesWhenTeamCountIsZero(t *testing.T) {
	// GIVEN a region with zero team count
	data := []RegionAnalyticsInput{{Region: "North", TotalPoints: 40, TotalInvestment: 20, TeamCount: 0}}

	// WHEN calculating region analytics
	results := CalculateRegionAnalyticsResults(data, 200, 100)

	// THEN average points is zero
	if results[0].AveragePoints != 0 {
		t.Errorf("expected 0, got %v", results[0].AveragePoints)
	}
}

func TestThatCalculateRegionAnalyticsResultsPreservesRegionName(t *testing.T) {
	// GIVEN a region named "Midwest"
	data := []RegionAnalyticsInput{{Region: "Midwest", TotalPoints: 50, TotalInvestment: 25, TeamCount: 16}}

	// WHEN calculating region analytics
	results := CalculateRegionAnalyticsResults(data, 200, 100)

	// THEN the region name is preserved
	if results[0].Region != "Midwest" {
		t.Errorf("expected Midwest, got %v", results[0].Region)
	}
}

// --- CalculateTeamAnalyticsResults ---

func TestThatCalculateTeamAnalyticsResultsReturnsEmptySliceForEmptyInput(t *testing.T) {
	// GIVEN empty input data
	// WHEN calculating team analytics
	results, _ := CalculateTeamAnalyticsResults(nil)

	// THEN the result has zero length
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestThatCalculateTeamAnalyticsResultsReturnsZeroBaselineROIForEmptyInput(t *testing.T) {
	// GIVEN empty input data
	// WHEN calculating team analytics
	_, baselineROI := CalculateTeamAnalyticsResults(nil)

	// THEN baseline ROI is zero
	if baselineROI != 0 {
		t.Errorf("expected 0, got %v", baselineROI)
	}
}

func TestThatCalculateTeamAnalyticsResultsCalculatesBaselineROI(t *testing.T) {
	// GIVEN two teams with total points 100 and total investment 50
	data := []TeamAnalyticsInput{
		{SchoolID: "a", SchoolName: "Duke", TotalPoints: 60, TotalInvestment: 30, Appearances: 3, TotalSeed: 3},
		{SchoolID: "b", SchoolName: "UNC", TotalPoints: 40, TotalInvestment: 20, Appearances: 2, TotalSeed: 4},
	}

	// WHEN calculating team analytics
	_, baselineROI := CalculateTeamAnalyticsResults(data)

	// THEN baseline ROI is 2.0 (100/50)
	if baselineROI != 2.0 {
		t.Errorf("expected 2.0, got %v", baselineROI)
	}
}

func TestThatCalculateTeamAnalyticsResultsCalculatesAveragePoints(t *testing.T) {
	// GIVEN a team with 60 total points across 3 appearances
	data := []TeamAnalyticsInput{{SchoolID: "a", SchoolName: "Duke", TotalPoints: 60, TotalInvestment: 30, Appearances: 3, TotalSeed: 3}}

	// WHEN calculating team analytics
	results, _ := CalculateTeamAnalyticsResults(data)

	// THEN average points is 20
	if results[0].AveragePoints != 20 {
		t.Errorf("expected 20, got %v", results[0].AveragePoints)
	}
}

func TestThatCalculateTeamAnalyticsResultsCalculatesAverageInvestment(t *testing.T) {
	// GIVEN a team with 30 total investment across 3 appearances
	data := []TeamAnalyticsInput{{SchoolID: "a", SchoolName: "Duke", TotalPoints: 60, TotalInvestment: 30, Appearances: 3, TotalSeed: 3}}

	// WHEN calculating team analytics
	results, _ := CalculateTeamAnalyticsResults(data)

	// THEN average investment is 10
	if results[0].AverageInvestment != 10 {
		t.Errorf("expected 10, got %v", results[0].AverageInvestment)
	}
}

func TestThatCalculateTeamAnalyticsResultsCalculatesAverageSeed(t *testing.T) {
	// GIVEN a team with total seed 7 across 2 appearances
	data := []TeamAnalyticsInput{{SchoolID: "a", SchoolName: "Duke", TotalPoints: 60, TotalInvestment: 30, Appearances: 2, TotalSeed: 7}}

	// WHEN calculating team analytics
	results, _ := CalculateTeamAnalyticsResults(data)

	// THEN average seed is 3.5
	if results[0].AverageSeed != 3.5 {
		t.Errorf("expected 3.5, got %v", results[0].AverageSeed)
	}
}

func TestThatCalculateTeamAnalyticsResultsCalculatesNormalizedROI(t *testing.T) {
	// GIVEN a team with 4x ROI and a baseline of 2x
	// actual ROI = 60/15 = 4, baseline = (60+40)/(15+35) = 100/50 = 2, normalized = 4/2 = 2.0
	data := []TeamAnalyticsInput{
		{SchoolID: "a", SchoolName: "Duke", TotalPoints: 60, TotalInvestment: 15, Appearances: 3, TotalSeed: 3},
		{SchoolID: "b", SchoolName: "UNC", TotalPoints: 40, TotalInvestment: 35, Appearances: 2, TotalSeed: 8},
	}

	// WHEN calculating team analytics
	results, _ := CalculateTeamAnalyticsResults(data)

	// THEN Duke's ROI is 2.0
	if results[0].ROI != 2.0 {
		t.Errorf("expected 2.0, got %v", results[0].ROI)
	}
}

func TestThatCalculateTeamAnalyticsResultsReturnsZeroROIForTeamWithNoInvestment(t *testing.T) {
	// GIVEN a team with zero investment
	data := []TeamAnalyticsInput{
		{SchoolID: "a", SchoolName: "UMBC", TotalPoints: 5, TotalInvestment: 0, Appearances: 1, TotalSeed: 16},
		{SchoolID: "b", SchoolName: "Duke", TotalPoints: 50, TotalInvestment: 30, Appearances: 3, TotalSeed: 3},
	}

	// WHEN calculating team analytics
	results, _ := CalculateTeamAnalyticsResults(data)

	// THEN the zero-investment team has ROI of zero
	if results[0].ROI != 0 {
		t.Errorf("expected 0, got %v", results[0].ROI)
	}
}

func TestThatCalculateTeamAnalyticsResultsReturnsZeroAveragesForZeroAppearances(t *testing.T) {
	// GIVEN a team with zero appearances
	data := []TeamAnalyticsInput{{SchoolID: "a", SchoolName: "NoShow", TotalPoints: 0, TotalInvestment: 0, Appearances: 0, TotalSeed: 0}}

	// WHEN calculating team analytics
	results, _ := CalculateTeamAnalyticsResults(data)

	// THEN average points is zero
	if results[0].AveragePoints != 0 {
		t.Errorf("expected 0, got %v", results[0].AveragePoints)
	}
}

func TestThatCalculateTeamAnalyticsResultsPreservesSchoolName(t *testing.T) {
	// GIVEN a team named "Gonzaga"
	data := []TeamAnalyticsInput{{SchoolID: "g", SchoolName: "Gonzaga", TotalPoints: 40, TotalInvestment: 20, Appearances: 2, TotalSeed: 2}}

	// WHEN calculating team analytics
	results, _ := CalculateTeamAnalyticsResults(data)

	// THEN the school name is preserved
	if results[0].SchoolName != "Gonzaga" {
		t.Errorf("expected Gonzaga, got %v", results[0].SchoolName)
	}
}

func TestThatCalculateTeamAnalyticsResultsShowsUnderperformanceWithROIBelowOne(t *testing.T) {
	// GIVEN a team with ROI below baseline
	// actual ROI = 10/20 = 0.5, baseline = (10+90)/(20+30) = 100/50 = 2, normalized = 0.5/2 = 0.25
	data := []TeamAnalyticsInput{
		{SchoolID: "a", SchoolName: "Bust", TotalPoints: 10, TotalInvestment: 20, Appearances: 1, TotalSeed: 1},
		{SchoolID: "b", SchoolName: "Star", TotalPoints: 90, TotalInvestment: 30, Appearances: 3, TotalSeed: 6},
	}

	// WHEN calculating team analytics
	results, _ := CalculateTeamAnalyticsResults(data)

	// THEN the underperforming team has ROI of 0.25
	if results[0].ROI != 0.25 {
		t.Errorf("expected 0.25, got %v", results[0].ROI)
	}
}

// --- CalculateSeedInvestmentSummaries ---

func TestThatCalculateSeedInvestmentSummariesReturnsEmptySliceForEmptyMap(t *testing.T) {
	// GIVEN an empty map
	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(map[int][]float64{})

	// THEN the result has zero length
	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(summaries))
	}
}

func TestThatCalculateSeedInvestmentSummariesReturnsSummariesSortedBySeed(t *testing.T) {
	// GIVEN bids for seeds 3, 1, and 2 (out of order)
	bySeed := map[int][]float64{
		3: {0.10},
		1: {0.30},
		2: {0.20},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN the first summary is for seed 1
	if summaries[0].Seed != 1 {
		t.Errorf("expected seed 1 first, got seed %d", summaries[0].Seed)
	}
}

func TestThatCalculateSeedInvestmentSummariesReturnsSummariesInAscendingSeedOrder(t *testing.T) {
	// GIVEN bids for seeds 3, 1, and 2 (out of order)
	bySeed := map[int][]float64{
		3: {0.10},
		1: {0.30},
		2: {0.20},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN the last summary is for seed 3
	if summaries[2].Seed != 3 {
		t.Errorf("expected seed 3 last, got seed %d", summaries[2].Seed)
	}
}

func TestThatCalculateSeedInvestmentSummariesCalculatesCorrectCount(t *testing.T) {
	// GIVEN a seed with 4 bid values
	bySeed := map[int][]float64{
		1: {0.10, 0.20, 0.30, 0.40},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN the count is 4
	if summaries[0].Count != 4 {
		t.Errorf("expected 4, got %d", summaries[0].Count)
	}
}

func TestThatCalculateSeedInvestmentSummariesCalculatesMean(t *testing.T) {
	// GIVEN a seed with values [10, 20, 30]
	bySeed := map[int][]float64{
		1: {10, 20, 30},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN the mean is 20
	if summaries[0].Mean != 20 {
		t.Errorf("expected 20, got %v", summaries[0].Mean)
	}
}

func TestThatCalculateSeedInvestmentSummariesCalculatesMinValue(t *testing.T) {
	// GIVEN a seed with unsorted values
	bySeed := map[int][]float64{
		1: {30, 10, 20},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN the min is 10
	if summaries[0].Min != 10 {
		t.Errorf("expected 10, got %v", summaries[0].Min)
	}
}

func TestThatCalculateSeedInvestmentSummariesCalculatesMaxValue(t *testing.T) {
	// GIVEN a seed with unsorted values
	bySeed := map[int][]float64{
		1: {30, 10, 20},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN the max is 30
	if summaries[0].Max != 30 {
		t.Errorf("expected 30, got %v", summaries[0].Max)
	}
}

func TestThatCalculateSeedInvestmentSummariesCalculatesMedian(t *testing.T) {
	// GIVEN a seed with 5 values
	bySeed := map[int][]float64{
		1: {10, 20, 30, 40, 50},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN the median is 30
	if summaries[0].Median != 30 {
		t.Errorf("expected 30, got %v", summaries[0].Median)
	}
}

func TestThatCalculateSeedInvestmentSummariesCalculatesStdDev(t *testing.T) {
	// GIVEN a seed with values [2, 4, 6] (mean=4, sample stddev=2)
	bySeed := map[int][]float64{
		1: {2, 4, 6},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN stddev is 2
	if summaries[0].StdDev != 2 {
		t.Errorf("expected 2, got %v", summaries[0].StdDev)
	}
}

func TestThatCalculateSeedInvestmentSummariesHandlesSingleValuePerSeed(t *testing.T) {
	// GIVEN a seed with a single value
	bySeed := map[int][]float64{
		1: {42},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN mean equals the single value
	if summaries[0].Mean != 42 {
		t.Errorf("expected 42, got %v", summaries[0].Mean)
	}
}

func TestThatCalculateSeedInvestmentSummariesReturnsZeroStdDevForSingleValue(t *testing.T) {
	// GIVEN a seed with a single value
	bySeed := map[int][]float64{
		1: {42},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN stddev is zero (undefined for n=1)
	if summaries[0].StdDev != 0 {
		t.Errorf("expected 0, got %v", summaries[0].StdDev)
	}
}

func TestThatCalculateSeedInvestmentSummariesSkipsSeedsWithEmptyValues(t *testing.T) {
	// GIVEN a map with one seed having values and one empty
	bySeed := map[int][]float64{
		1: {10, 20},
		2: {},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN only one summary is returned (seed 2 is skipped)
	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
	}
}

func TestThatCalculateSeedInvestmentSummariesDoesNotMutateInputValues(t *testing.T) {
	// GIVEN unsorted input values
	values := []float64{30, 10, 20}
	bySeed := map[int][]float64{1: values}

	// WHEN calculating summaries
	_ = CalculateSeedInvestmentSummaries(bySeed)

	// THEN the original slice is not sorted
	if values[0] != 30 {
		t.Errorf("expected input values[0] to remain 30, got %v", values[0])
	}
}

func TestThatCalculateSeedInvestmentSummariesCalculatesQ1(t *testing.T) {
	// GIVEN a seed with 5 values [10, 20, 30, 40, 50]
	// Q1 position = 0.25 * (5-1) = 1.0, exact index -> 20
	bySeed := map[int][]float64{
		1: {10, 20, 30, 40, 50},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN Q1 is 20
	if summaries[0].Q1 != 20 {
		t.Errorf("expected 20, got %v", summaries[0].Q1)
	}
}

func TestThatCalculateSeedInvestmentSummariesCalculatesQ3(t *testing.T) {
	// GIVEN a seed with 5 values [10, 20, 30, 40, 50]
	// Q3 position = 0.75 * (5-1) = 3.0, exact index -> 40
	bySeed := map[int][]float64{
		1: {10, 20, 30, 40, 50},
	}

	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(bySeed)

	// THEN Q3 is 40
	if summaries[0].Q3 != 40 {
		t.Errorf("expected 40, got %v", summaries[0].Q3)
	}
}

// --- meanFloat64 edge cases ---

func TestThatMeanFloat64HandlesNegativeValues(t *testing.T) {
	// GIVEN values with negative numbers
	// WHEN calculating the mean
	got := meanFloat64([]float64{-10, 10})

	// THEN the result is 0
	if got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

// --- stddevFloat64 edge cases ---

func TestThatStddevFloat64HandlesIdenticalValues(t *testing.T) {
	// GIVEN a slice where all values are identical
	// WHEN calculating the standard deviation
	got := stddevFloat64([]float64{5, 5, 5}, 5)

	// THEN the result is 0
	if got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

// --- quantileSorted edge cases ---

func TestThatQuantileSortedReturnsSingleValueForSingleElementSlice(t *testing.T) {
	// GIVEN a single-element slice
	// WHEN calculating the median
	got := quantileSorted([]float64{42}, 0.5)

	// THEN the result is that single value
	if got != 42 {
		t.Errorf("expected 42, got %v", got)
	}
}

func TestThatQuantileSortedInterpolatesQ1ForFourElements(t *testing.T) {
	// GIVEN a 4-element sorted slice [10, 20, 30, 40]
	// Q1 position = 0.25 * 3 = 0.75, between index 0 and 1
	// interpolation: 10*(1-0.75) + 20*0.75 = 2.5 + 15 = 17.5
	// WHEN calculating Q1
	got := quantileSorted([]float64{10, 20, 30, 40}, 0.25)

	// THEN the result is 17.5
	if got != 17.5 {
		t.Errorf("expected 17.5, got %v", got)
	}
}

// --- CalculateSeedAnalyticsResults: multiple seeds ---

func TestThatCalculateSeedAnalyticsResultsHandlesMultipleSeeds(t *testing.T) {
	// GIVEN two seeds
	data := []SeedAnalyticsInput{
		{Seed: 1, TotalPoints: 60, TotalInvestment: 30, TeamCount: 4},
		{Seed: 16, TotalPoints: 10, TotalInvestment: 5, TeamCount: 4},
	}

	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(data, 100, 50)

	// THEN the result count matches the input count
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestThatCalculateSeedAnalyticsResultsPreservesSeedNumber(t *testing.T) {
	// GIVEN a 16-seed
	data := []SeedAnalyticsInput{{Seed: 16, TotalPoints: 5, TotalInvestment: 2, TeamCount: 4}}

	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(data, 100, 50)

	// THEN the seed number is preserved
	if results[0].Seed != 16 {
		t.Errorf("expected 16, got %d", results[0].Seed)
	}
}

// --- CalculateRegionAnalyticsResults: ROI edge case ---

func TestThatCalculateRegionAnalyticsResultsReturnsZeroROIWhenTotalInvestmentIsZero(t *testing.T) {
	// GIVEN a region with investment but zero total investment (baseline is zero)
	data := []RegionAnalyticsInput{{Region: "East", TotalPoints: 40, TotalInvestment: 20, TeamCount: 16}}

	// WHEN calculating region analytics with zero total investment
	results := CalculateRegionAnalyticsResults(data, 200, 0)

	// THEN ROI is zero (baseline ROI cannot be computed)
	if results[0].ROI != 0 {
		t.Errorf("expected 0, got %v", results[0].ROI)
	}
}

// --- CalculateTeamAnalyticsResults: average seed edge ---

func TestThatCalculateTeamAnalyticsResultsReturnsZeroAverageSeedForZeroAppearances(t *testing.T) {
	// GIVEN a team with zero appearances
	data := []TeamAnalyticsInput{{SchoolID: "x", SchoolName: "Ghost", TotalPoints: 0, TotalInvestment: 0, Appearances: 0, TotalSeed: 0}}

	// WHEN calculating team analytics
	results, _ := CalculateTeamAnalyticsResults(data)

	// THEN average seed is zero
	if results[0].AverageSeed != 0 {
		t.Errorf("expected 0, got %v", results[0].AverageSeed)
	}
}

// --- CalculateSeedInvestmentSummaries: nil map ---

func TestThatCalculateSeedInvestmentSummariesReturnsEmptySliceForNilMap(t *testing.T) {
	// GIVEN a nil map
	// WHEN calculating summaries
	summaries := CalculateSeedInvestmentSummaries(nil)

	// THEN the result has zero length
	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(summaries))
	}
}

// --- CalculateRegionAnalyticsResults: multiple regions ---

func TestThatCalculateRegionAnalyticsResultsHandlesMultipleRegions(t *testing.T) {
	// GIVEN four regions
	data := []RegionAnalyticsInput{
		{Region: "East", TotalPoints: 50, TotalInvestment: 25, TeamCount: 16},
		{Region: "West", TotalPoints: 50, TotalInvestment: 25, TeamCount: 16},
		{Region: "South", TotalPoints: 50, TotalInvestment: 25, TeamCount: 16},
		{Region: "Midwest", TotalPoints: 50, TotalInvestment: 25, TeamCount: 16},
	}

	// WHEN calculating region analytics
	results := CalculateRegionAnalyticsResults(data, 200, 100)

	// THEN all four results are returned
	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}
}

func TestThatCalculateRegionAnalyticsResultsCalculatesEqualPercentagesForEqualRegions(t *testing.T) {
	// GIVEN four equal regions
	data := []RegionAnalyticsInput{
		{Region: "East", TotalPoints: 50, TotalInvestment: 25, TeamCount: 16},
		{Region: "West", TotalPoints: 50, TotalInvestment: 25, TeamCount: 16},
		{Region: "South", TotalPoints: 50, TotalInvestment: 25, TeamCount: 16},
		{Region: "Midwest", TotalPoints: 50, TotalInvestment: 25, TeamCount: 16},
	}

	// WHEN calculating region analytics
	results := CalculateRegionAnalyticsResults(data, 200, 100)

	// THEN each region has 25% of points
	if results[0].PointsPercentage != 25 {
		t.Errorf("expected 25, got %v", results[0].PointsPercentage)
	}
}

// --- CalculateSeedVarianceResults: pass-through fields ---

func TestThatCalculateSeedVarianceResultsPreservesSeedNumber(t *testing.T) {
	// GIVEN variance data for seed 11
	data := []SeedVarianceInput{{Seed: 11, InvestmentStdDev: 5, InvestmentMean: 10, PointsStdDev: 2, PointsMean: 8, TeamCount: 4}}

	// WHEN calculating seed variance
	results := CalculateSeedVarianceResults(data)

	// THEN the seed number is preserved
	if results[0].Seed != 11 {
		t.Errorf("expected 11, got %d", results[0].Seed)
	}
}

func TestThatCalculateSeedVarianceResultsPreservesTeamCount(t *testing.T) {
	// GIVEN variance data with team count of 8
	data := []SeedVarianceInput{{Seed: 1, InvestmentStdDev: 5, InvestmentMean: 10, PointsStdDev: 2, PointsMean: 8, TeamCount: 8}}

	// WHEN calculating seed variance
	results := CalculateSeedVarianceResults(data)

	// THEN the team count is preserved
	if results[0].TeamCount != 8 {
		t.Errorf("expected 8, got %d", results[0].TeamCount)
	}
}

func TestThatCalculateSeedVarianceResultsReturnsZeroPointsCVWhenPointsMeanIsZero(t *testing.T) {
	// GIVEN seed data where points mean is zero (16 seeds that never score)
	data := []SeedVarianceInput{{Seed: 16, InvestmentStdDev: 5, InvestmentMean: 10, PointsStdDev: 0, PointsMean: 0, TeamCount: 4}}

	// WHEN calculating seed variance
	results := CalculateSeedVarianceResults(data)

	// THEN points CV is zero
	if results[0].PointsCV != 0 {
		t.Errorf("expected 0, got %v", results[0].PointsCV)
	}
}

// --- stddevFloat64: known calculation ---

func TestThatStddevFloat64CalculatesCorrectValueForKnownDataSet(t *testing.T) {
	// GIVEN values [1, 3, 5, 7, 9] with mean 5
	// sample variance = ((1-5)^2 + (3-5)^2 + (5-5)^2 + (7-5)^2 + (9-5)^2) / 4
	//                 = (16 + 4 + 0 + 4 + 16) / 4 = 40/4 = 10
	// sample stddev = sqrt(10) ~= 3.16227766
	// WHEN calculating the standard deviation
	got := stddevFloat64([]float64{1, 3, 5, 7, 9}, 5)

	// THEN the result matches the expected sample standard deviation
	expected := math.Sqrt(10)
	if math.Abs(got-expected) > 1e-10 {
		t.Errorf("expected %v, got %v", expected, got)
	}
}
