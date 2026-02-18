package analytics

import (
	"reflect"
	"testing"
)

func TestThatMeanFloat64ReturnsZeroForEmptySlice(t *testing.T) {
	// GIVEN an empty slice
	// WHEN calculating the mean
	got := meanFloat64(nil)

	// THEN the result is 0
	if got != 0 {
		t.Errorf("expected mean to be 0, got %v", got)
	}
}

func TestThatStddevFloat64ReturnsZeroForSingleValue(t *testing.T) {
	// GIVEN a slice with a single value
	// WHEN calculating the standard deviation
	got := stddevFloat64([]float64{5}, 5)

	// THEN the result is 0
	if got != 0 {
		t.Errorf("expected stddev to be 0, got %v", got)
	}
}

func TestThatQuantileSortedReturnsMinForZeroQuantile(t *testing.T) {
	// GIVEN a sorted slice
	// WHEN calculating the 0th quantile
	got := quantileSorted([]float64{1, 2, 3, 4}, 0)

	// THEN the result is the minimum value
	if got != 1 {
		t.Errorf("expected quantile to be 1, got %v", got)
	}
}

func TestThatQuantileSortedReturnsMaxForOneQuantile(t *testing.T) {
	// GIVEN a sorted slice
	// WHEN calculating the 1st quantile
	got := quantileSorted([]float64{1, 2, 3, 4}, 1)

	// THEN the result is the maximum value
	if got != 4 {
		t.Errorf("expected quantile to be 4, got %v", got)
	}
}

func TestThatCalculateSeedAnalyticsResultsCalculatesROIAsRatioToBaselineROI(t *testing.T) {
	// GIVEN seed data with 3x ROI and baseline ROI of 3x
	data := []SeedAnalyticsInput{{
		Seed:            1,
		TotalPoints:     30,
		TotalInvestment: 10,
		TeamCount:       1,
	}}
	totalPoints := 60.0
	totalInvestment := 20.0

	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(data, totalPoints, totalInvestment)

	// THEN ROI equals 1.0 (actual ROI / baseline ROI = 3/3 = 1)
	got := 0.0
	if len(results) > 0 {
		got = results[0].ROI
	}
	want := 1.0

	if got != want {
		t.Errorf("expected roi %v, got %v", want, got)
	}
}

func TestThatCalculateSeedVarianceResultsCalculatesVarianceRatioAsInvestmentCVOverPointsCV(t *testing.T) {
	// GIVEN seed variance data with investment CV of 0.5 and points CV of 0.5
	data := []SeedVarianceInput{{
		Seed:             1,
		InvestmentStdDev: 10,
		InvestmentMean:   20,
		PointsStdDev:     3,
		PointsMean:       6,
		TeamCount:        1,
	}}

	// WHEN calculating seed variance analytics
	results := CalculateSeedVarianceResults(data)

	// THEN variance ratio equals investment CV / points CV
	got := 0.0
	if len(results) > 0 {
		got = results[0].VarianceRatio
	}
	want := (10.0 / 20.0) / (3.0 / 6.0)

	if got != want {
		t.Errorf("expected variance ratio %v, got %v", want, got)
	}
}

func TestThatCalculateSeedAnalyticsResultsReturnsEmptySliceForEmptyInput(t *testing.T) {
	// GIVEN empty input data
	// WHEN calculating seed analytics
	results := CalculateSeedAnalyticsResults(nil, 0, 0)

	// THEN the result is an empty slice
	got := results
	want := []SeedAnalyticsResult{}
	if got == nil {
		got = []SeedAnalyticsResult{}
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected empty results")
	}
}
