package dtos

import (
	"reflect"
	"testing"

	analytics "github.com/andrewcopp/Calcutta/backend/internal/app/analytics"
)

func TestThatToAnalyticsResponseReturnsZeroValueForNilInput(t *testing.T) {
	// GIVEN
	var in *analytics.AnalyticsResult

	// WHEN
	out := ToAnalyticsResponse(in)

	// THEN
	if !reflect.DeepEqual(out, AnalyticsResponse{}) {
		t.Fatalf("expected zero value response")
	}
}

func TestThatToAnalyticsResponseMapsAllFields(t *testing.T) {
	// GIVEN
	in := &analytics.AnalyticsResult{
		TotalPoints:     123.4,
		TotalInvestment: 56.7,
		BaselineROI:     2.176,
		SeedAnalytics: []analytics.SeedAnalyticsResult{{
			Seed:                 1,
			TotalPoints:          10,
			TotalInvestment:      5,
			PointsPercentage:     0.1,
			InvestmentPercentage: 0.2,
			TeamCount:            3,
			AveragePoints:        3.33,
			AverageInvestment:    1.67,
			ROI:                  2.0,
		}},
		RegionAnalytics: []analytics.RegionAnalyticsResult{{
			Region:               "East",
			TotalPoints:          20,
			TotalInvestment:      8,
			PointsPercentage:     0.25,
			InvestmentPercentage: 0.3,
			TeamCount:            4,
			AveragePoints:        5,
			AverageInvestment:    2,
			ROI:                  2.5,
		}},
		TeamAnalytics: []analytics.TeamAnalyticsResult{{
			SchoolID:          "s1",
			SchoolName:        "School 1",
			TotalPoints:       7,
			TotalInvestment:   3,
			Appearances:       2,
			AveragePoints:     3.5,
			AverageInvestment: 1.5,
			AverageSeed:       4.0,
			ROI:               2.333,
		}},
		SeedVarianceAnalytics: []analytics.SeedVarianceResult{{
			Seed:             2,
			InvestmentStdDev: 10,
			PointsStdDev:     5,
			InvestmentMean:   20,
			PointsMean:       10,
			InvestmentCV:     0.5,
			PointsCV:         0.5,
			TeamCount:        6,
			VarianceRatio:    1.2,
		}},
	}

	// WHEN
	out := ToAnalyticsResponse(in)

	// THEN
	want := AnalyticsResponse{
		TotalPoints:     123.4,
		TotalInvestment: 56.7,
		BaselineROI:     2.176,
		SeedAnalytics: []SeedAnalytics{{
			Seed:                 1,
			TotalPoints:          10,
			TotalInvestment:      5,
			PointsPercentage:     0.1,
			InvestmentPercentage: 0.2,
			TeamCount:            3,
			AveragePoints:        3.33,
			AverageInvestment:    1.67,
			ROI:                  2.0,
		}},
		RegionAnalytics: []RegionAnalytics{{
			Region:               "East",
			TotalPoints:          20,
			TotalInvestment:      8,
			PointsPercentage:     0.25,
			InvestmentPercentage: 0.3,
			TeamCount:            4,
			AveragePoints:        5,
			AverageInvestment:    2,
			ROI:                  2.5,
		}},
		TeamAnalytics: []TeamAnalytics{{
			SchoolID:          "s1",
			SchoolName:        "School 1",
			TotalPoints:       7,
			TotalInvestment:   3,
			Appearances:       2,
			AveragePoints:     3.5,
			AverageInvestment: 1.5,
			AverageSeed:       4.0,
			ROI:               2.333,
		}},
		SeedVarianceAnalytics: []SeedVarianceAnalytics{{
			Seed:             2,
			InvestmentStdDev: 10,
			PointsStdDev:     5,
			InvestmentMean:   20,
			PointsMean:       10,
			InvestmentCV:     0.5,
			PointsCV:         0.5,
			TeamCount:        6,
			VarianceRatio:    1.2,
		}},
	}

	if !reflect.DeepEqual(out, want) {
		t.Fatalf("expected response to match input mapping")
	}
}
