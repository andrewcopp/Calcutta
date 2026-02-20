package analytics

import "math"

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
