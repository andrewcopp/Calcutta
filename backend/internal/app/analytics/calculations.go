package analytics

import (
	"math"
	"sort"
)

// clampLimit constrains a pagination limit to the range [1, 500], defaulting to 100 when non-positive.
func clampLimit(limit int) int {
	if limit <= 0 {
		return 100
	}
	if limit > 500 {
		return 500
	}
	return limit
}

// CalculateSeedAnalyticsResults is a pure function that calculates seed analytics from raw data.
// This can be tested without mocking repositories.
func CalculateSeedAnalyticsResults(data []SeedAnalyticsInput, totalPoints, totalInvestment float64) []SeedAnalyticsResult {
	// Calculate baseline ROI (overall points per dollar)
	var baselineROI float64
	if totalInvestment > 0 {
		baselineROI = totalPoints / totalInvestment
	}

	results := make([]SeedAnalyticsResult, len(data))
	for i, d := range data {
		results[i] = SeedAnalyticsResult{
			Seed:            d.Seed,
			TotalPoints:     d.TotalPoints,
			TotalInvestment: d.TotalInvestment,
			TeamCount:       d.TeamCount,
		}

		if totalPoints > 0 {
			results[i].PointsPercentage = (d.TotalPoints / totalPoints) * 100
		}
		if totalInvestment > 0 {
			results[i].InvestmentPercentage = (d.TotalInvestment / totalInvestment) * 100
		}
		if d.TeamCount > 0 {
			results[i].AveragePoints = d.TotalPoints / float64(d.TeamCount)
			results[i].AverageInvestment = d.TotalInvestment / float64(d.TeamCount)
		}

		// Calculate normalized ROI
		// ROI = (actual points per dollar) / (baseline points per dollar)
		// 1.0 = average, >1.0 = over-performance, <1.0 = under-performance
		if d.TotalInvestment > 0 && baselineROI > 0 {
			actualROI := d.TotalPoints / d.TotalInvestment
			results[i].ROI = actualROI / baselineROI
		}
	}

	return results
}

// CalculateSeedVarianceResults is a pure function that calculates seed variance analytics from raw data.
// This can be tested without mocking repositories.
func CalculateSeedVarianceResults(data []SeedVarianceInput) []SeedVarianceResult {
	results := make([]SeedVarianceResult, len(data))
	for i, d := range data {
		results[i] = SeedVarianceResult{
			Seed:             d.Seed,
			InvestmentStdDev: d.InvestmentStdDev,
			PointsStdDev:     d.PointsStdDev,
			InvestmentMean:   d.InvestmentMean,
			PointsMean:       d.PointsMean,
			TeamCount:        d.TeamCount,
		}

		// Calculate coefficient of variation (CV) = stddev / mean
		// CV allows comparison of variability across different scales
		if d.InvestmentMean > 0 {
			results[i].InvestmentCV = d.InvestmentStdDev / d.InvestmentMean
		}
		if d.PointsMean > 0 {
			results[i].PointsCV = d.PointsStdDev / d.PointsMean
		}

		// Variance Ratio: Investment CV / Points CV
		// Ratio > 1 means investment varies more than performance (ugly duckling indicator)
		if results[i].PointsCV > 0 {
			results[i].VarianceRatio = results[i].InvestmentCV / results[i].PointsCV
		}
	}

	return results
}

// CalculateRegionAnalyticsResults is a pure function that calculates region analytics from raw data.
// This can be tested without mocking repositories.
func CalculateRegionAnalyticsResults(data []RegionAnalyticsInput, totalPoints, totalInvestment float64) []RegionAnalyticsResult {
	var baselineROI float64
	if totalInvestment > 0 {
		baselineROI = totalPoints / totalInvestment
	}

	results := make([]RegionAnalyticsResult, len(data))
	for i, d := range data {
		results[i] = RegionAnalyticsResult{
			Region:          d.Region,
			TotalPoints:     d.TotalPoints,
			TotalInvestment: d.TotalInvestment,
			TeamCount:       d.TeamCount,
		}

		if totalPoints > 0 {
			results[i].PointsPercentage = (d.TotalPoints / totalPoints) * 100
		}
		if totalInvestment > 0 {
			results[i].InvestmentPercentage = (d.TotalInvestment / totalInvestment) * 100
		}
		if d.TeamCount > 0 {
			results[i].AveragePoints = d.TotalPoints / float64(d.TeamCount)
			results[i].AverageInvestment = d.TotalInvestment / float64(d.TeamCount)
		}

		if d.TotalInvestment > 0 && baselineROI > 0 {
			actualROI := d.TotalPoints / d.TotalInvestment
			results[i].ROI = actualROI / baselineROI
		}
	}

	return results
}

// CalculateTeamAnalyticsResults is a pure function that calculates team analytics from raw data.
// It returns the computed results and the baseline ROI used for normalization.
func CalculateTeamAnalyticsResults(data []TeamAnalyticsInput) ([]TeamAnalyticsResult, float64) {
	var totalPoints, totalInvestment float64
	for _, d := range data {
		totalPoints += d.TotalPoints
		totalInvestment += d.TotalInvestment
	}

	var baselineROI float64
	if totalInvestment > 0 {
		baselineROI = totalPoints / totalInvestment
	}

	results := make([]TeamAnalyticsResult, len(data))
	for i, d := range data {
		results[i] = TeamAnalyticsResult{
			SchoolID:        d.SchoolID,
			SchoolName:      d.SchoolName,
			TotalPoints:     d.TotalPoints,
			TotalInvestment: d.TotalInvestment,
			Appearances:     d.Appearances,
		}

		if d.Appearances > 0 {
			results[i].AveragePoints = d.TotalPoints / float64(d.Appearances)
			results[i].AverageInvestment = d.TotalInvestment / float64(d.Appearances)
			results[i].AverageSeed = float64(d.TotalSeed) / float64(d.Appearances)
		}

		if d.TotalInvestment > 0 && baselineROI > 0 {
			actualROI := d.TotalPoints / d.TotalInvestment
			results[i].ROI = actualROI / baselineROI
		}
	}

	return results, baselineROI
}

// CalculateSeedInvestmentSummaries computes descriptive statistics (mean, stddev, quartiles)
// for each seed's normalized bid values. The input map keys are seed numbers and values are
// the normalized bid amounts for that seed.
func CalculateSeedInvestmentSummaries(bySeed map[int][]float64) []SeedInvestmentSummaryResult {
	seeds := make([]int, 0, len(bySeed))
	for seed := range bySeed {
		seeds = append(seeds, seed)
	}
	sort.Ints(seeds)

	summaries := make([]SeedInvestmentSummaryResult, 0, len(seeds))
	for _, seed := range seeds {
		values := make([]float64, len(bySeed[seed]))
		copy(values, bySeed[seed])
		sort.Float64s(values)

		count := len(values)
		if count == 0 {
			continue
		}

		mean := meanFloat64(values)
		stddev := stddevFloat64(values, mean)

		summaries = append(summaries, SeedInvestmentSummaryResult{
			Seed:   seed,
			Count:  count,
			Mean:   mean,
			StdDev: stddev,
			Min:    values[0],
			Q1:     quantileSorted(values, 0.25),
			Median: quantileSorted(values, 0.50),
			Q3:     quantileSorted(values, 0.75),
			Max:    values[count-1],
		})
	}

	return summaries
}

func meanFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stddevFloat64(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}

	var sumSquares float64
	for _, v := range values {
		d := v - mean
		sumSquares += d * d
	}

	return math.Sqrt(sumSquares / float64(len(values)-1))
}

func quantileSorted(sortedValues []float64, q float64) float64 {
	n := len(sortedValues)
	if n == 0 {
		return 0
	}
	if q <= 0 {
		return sortedValues[0]
	}
	if q >= 1 {
		return sortedValues[n-1]
	}

	pos := q * float64(n-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi {
		return sortedValues[lo]
	}

	w := pos - float64(lo)
	return sortedValues[lo]*(1-w) + sortedValues[hi]*w
}
