package services

import (
	"context"
	"log"
)

type AnalyticsService struct {
	repo *AnalyticsRepository
}

func NewAnalyticsService(repo *AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

type SeedAnalyticsResult struct {
	Seed                 int
	TotalPoints          float64
	TotalInvestment      float64
	PointsPercentage     float64
	InvestmentPercentage float64
	TeamCount            int
	AveragePoints        float64
	AverageInvestment    float64
	ROI                  float64
}

type RegionAnalyticsResult struct {
	Region               string
	TotalPoints          float64
	TotalInvestment      float64
	PointsPercentage     float64
	InvestmentPercentage float64
	TeamCount            int
	AveragePoints        float64
	AverageInvestment    float64
	ROI                  float64
}

type TeamAnalyticsResult struct {
	SchoolID          string
	SchoolName        string
	TotalPoints       float64
	TotalInvestment   float64
	Appearances       int
	AveragePoints     float64
	AverageInvestment float64
	AverageSeed       float64
	ROI               float64
}

type SeedVarianceResult struct {
	Seed             int
	InvestmentStdDev float64
	PointsStdDev     float64
	InvestmentMean   float64
	PointsMean       float64
	InvestmentCV     float64
	PointsCV         float64
	TeamCount        int
	VarianceRatio    float64
}

type AnalyticsResult struct {
	SeedAnalytics         []SeedAnalyticsResult
	RegionAnalytics       []RegionAnalyticsResult
	TeamAnalytics         []TeamAnalyticsResult
	SeedVarianceAnalytics []SeedVarianceResult
	TotalPoints           float64
	TotalInvestment       float64
	BaselineROI           float64
}

func (s *AnalyticsService) GetSeedAnalytics(ctx context.Context) ([]SeedAnalyticsResult, float64, float64, error) {
	data, totalPoints, totalInvestment, err := s.repo.GetSeedAnalytics(ctx)
	if err != nil {
		log.Printf("Error getting seed analytics: %v", err)
		return nil, 0, 0, err
	}

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

	return results, totalPoints, totalInvestment, nil
}

func (s *AnalyticsService) GetRegionAnalytics(ctx context.Context) ([]RegionAnalyticsResult, float64, float64, error) {
	data, totalPoints, totalInvestment, err := s.repo.GetRegionAnalytics(ctx)
	if err != nil {
		log.Printf("Error getting region analytics: %v", err)
		return nil, 0, 0, err
	}

	// Calculate baseline ROI (overall points per dollar)
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

		// Calculate normalized ROI
		if d.TotalInvestment > 0 && baselineROI > 0 {
			actualROI := d.TotalPoints / d.TotalInvestment
			results[i].ROI = actualROI / baselineROI
		}
	}

	return results, totalPoints, totalInvestment, nil
}

func (s *AnalyticsService) GetTeamAnalytics(ctx context.Context) ([]TeamAnalyticsResult, float64, error) {
	data, err := s.repo.GetTeamAnalytics(ctx)
	if err != nil {
		log.Printf("Error getting team analytics: %v", err)
		return nil, 0, err
	}

	// Calculate baseline ROI across all teams
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

		// Calculate normalized ROI
		if d.TotalInvestment > 0 && baselineROI > 0 {
			actualROI := d.TotalPoints / d.TotalInvestment
			results[i].ROI = actualROI / baselineROI
		}
	}

	return results, baselineROI, nil
}

func (s *AnalyticsService) GetSeedVarianceAnalytics(ctx context.Context) ([]SeedVarianceResult, error) {
	data, err := s.repo.GetSeedVarianceAnalytics(ctx)
	if err != nil {
		log.Printf("Error getting seed variance analytics: %v", err)
		return nil, err
	}

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

	return results, nil
}

func (s *AnalyticsService) GetAllAnalytics(ctx context.Context) (*AnalyticsResult, error) {
	seedAnalytics, totalPoints, totalInvestment, err := s.GetSeedAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	regionAnalytics, _, _, err := s.GetRegionAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	teamAnalytics, _, err := s.GetTeamAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	seedVarianceAnalytics, err := s.GetSeedVarianceAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	var baselineROI float64
	if totalInvestment > 0 {
		baselineROI = totalPoints / totalInvestment
	}

	return &AnalyticsResult{
		SeedAnalytics:         seedAnalytics,
		RegionAnalytics:       regionAnalytics,
		TeamAnalytics:         teamAnalytics,
		SeedVarianceAnalytics: seedVarianceAnalytics,
		TotalPoints:           totalPoints,
		TotalInvestment:       totalInvestment,
		BaselineROI:           baselineROI,
	}, nil
}
