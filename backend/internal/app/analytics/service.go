package analytics

import (
	"context"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Service struct {
	repo ports.AnalyticsRepo
}

func New(repo ports.AnalyticsRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetBestInvestments(ctx context.Context, limit int) ([]BestInvestmentResult, error) {
	limit = clampLimit(limit)

	data, err := s.repo.GetBestInvestments(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("getting best investments: %w", err)
	}

	results := make([]BestInvestmentResult, 0, len(data))
	for _, d := range data {
		results = append(results, BestInvestmentResult{
			TournamentName:   d.TournamentName,
			TournamentYear:   d.TournamentYear,
			CalcuttaID:       d.CalcuttaID,
			TeamID:           d.TeamID,
			SchoolName:       d.SchoolName,
			Seed:             d.Seed,
			Region:           d.Region,
			TeamPoints:       d.TeamPoints,
			TotalBid:         d.TotalBid,
			CalcuttaTotalBid: d.CalcuttaTotalBid,
			CalcuttaTotalPts: d.CalcuttaTotalPts,
			InvestmentShare:  d.InvestmentShare,
			PointsShare:      d.PointsShare,
			RawROI:           d.RawROI,
			NormalizedROI:    d.NormalizedROI,
		})
	}

	return results, nil
}

func (s *Service) GetBestCareers(ctx context.Context, limit int) ([]CareerLeaderboardResult, error) {
	limit = clampLimit(limit)

	data, err := s.repo.GetBestCareers(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("getting best careers: %w", err)
	}

	results := make([]CareerLeaderboardResult, 0, len(data))
	for _, d := range data {
		results = append(results, CareerLeaderboardResult{
			EntryName:              d.EntryName,
			Years:                  d.Years,
			BestFinish:             d.BestFinish,
			Wins:                   d.Wins,
			Podiums:                d.Podiums,
			InTheMoneys:            d.InTheMoneys,
			Top10s:                 d.Top10s,
			CareerEarningsCents:    d.CareerEarningsCents,
			ActiveInLatestCalcutta: d.ActiveInLatestCalcutta,
		})
	}

	return results, nil
}

func (s *Service) GetBestInvestmentBids(ctx context.Context, limit int) ([]InvestmentLeaderboardResult, error) {
	limit = clampLimit(limit)

	data, err := s.repo.GetBestInvestmentBids(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("getting best investment bids: %w", err)
	}

	results := make([]InvestmentLeaderboardResult, 0, len(data))
	for _, d := range data {
		results = append(results, InvestmentLeaderboardResult{
			TournamentName:      d.TournamentName,
			TournamentYear:      d.TournamentYear,
			CalcuttaID:          d.CalcuttaID,
			EntryID:             d.EntryID,
			EntryName:           d.EntryName,
			TeamID:              d.TeamID,
			SchoolName:          d.SchoolName,
			Seed:                d.Seed,
			Investment:          d.Investment,
			OwnershipPercentage: d.OwnershipPercentage,
			RawReturns:          d.RawReturns,
			NormalizedReturns:   d.NormalizedReturns,
		})
	}

	return results, nil
}

func (s *Service) GetBestEntries(ctx context.Context, limit int) ([]EntryLeaderboardResult, error) {
	limit = clampLimit(limit)

	data, err := s.repo.GetBestEntries(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("getting best entries: %w", err)
	}

	results := make([]EntryLeaderboardResult, 0, len(data))
	for _, d := range data {
		results = append(results, EntryLeaderboardResult{
			TournamentName:    d.TournamentName,
			TournamentYear:    d.TournamentYear,
			CalcuttaID:        d.CalcuttaID,
			EntryID:           d.EntryID,
			EntryName:         d.EntryName,
			TotalReturns:      d.TotalReturns,
			TotalParticipants: d.TotalParticipants,
			AverageReturns:    d.AverageReturns,
			NormalizedReturns: d.NormalizedReturns,
		})
	}

	return results, nil
}

func (s *Service) GetSeedInvestmentDistribution(ctx context.Context) (*SeedInvestmentDistributionResult, error) {
	data, err := s.repo.GetSeedInvestmentPoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting seed investment points: %w", err)
	}

	points := make([]SeedInvestmentPointResult, 0, len(data))
	bySeed := map[int][]float64{}

	for _, d := range data {
		points = append(points, SeedInvestmentPointResult{
			Seed:             d.Seed,
			TournamentName:   d.TournamentName,
			TournamentYear:   d.TournamentYear,
			CalcuttaID:       d.CalcuttaID,
			TeamID:           d.TeamID,
			SchoolName:       d.SchoolName,
			TotalBid:         d.TotalBid,
			CalcuttaTotalBid: d.CalcuttaTotalBid,
			NormalizedBid:    d.NormalizedBid,
		})

		bySeed[d.Seed] = append(bySeed[d.Seed], d.NormalizedBid)
	}

	summaries := CalculateSeedInvestmentSummaries(bySeed)

	return &SeedInvestmentDistributionResult{Points: points, Summaries: summaries}, nil
}

func (s *Service) GetSeedAnalytics(ctx context.Context) ([]SeedAnalyticsResult, float64, float64, error) {
	data, totalPoints, totalInvestment, err := s.repo.GetSeedAnalytics(ctx)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("getting seed analytics: %w", err)
	}

	// Convert ports data to input type for pure function
	input := make([]SeedAnalyticsInput, len(data))
	for i, d := range data {
		input[i] = SeedAnalyticsInput{
			Seed:            d.Seed,
			TotalPoints:     d.TotalPoints,
			TotalInvestment: d.TotalInvestment,
			TeamCount:       d.TeamCount,
		}
	}

	results := CalculateSeedAnalyticsResults(input, totalPoints, totalInvestment)
	return results, totalPoints, totalInvestment, nil
}

func (s *Service) GetRegionAnalytics(ctx context.Context) ([]RegionAnalyticsResult, float64, float64, error) {
	data, totalPoints, totalInvestment, err := s.repo.GetRegionAnalytics(ctx)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("getting region analytics: %w", err)
	}

	// Convert ports data to input type for pure function
	input := make([]RegionAnalyticsInput, len(data))
	for i, d := range data {
		input[i] = RegionAnalyticsInput{
			Region:          d.Region,
			TotalPoints:     d.TotalPoints,
			TotalInvestment: d.TotalInvestment,
			TeamCount:       d.TeamCount,
		}
	}

	results := CalculateRegionAnalyticsResults(input, totalPoints, totalInvestment)
	return results, totalPoints, totalInvestment, nil
}

func (s *Service) GetTeamAnalytics(ctx context.Context) ([]TeamAnalyticsResult, float64, error) {
	data, err := s.repo.GetTeamAnalytics(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("getting team analytics: %w", err)
	}

	// Convert ports data to input type for pure function
	input := make([]TeamAnalyticsInput, len(data))
	for i, d := range data {
		input[i] = TeamAnalyticsInput{
			SchoolID:        d.SchoolID,
			SchoolName:      d.SchoolName,
			TotalPoints:     d.TotalPoints,
			TotalInvestment: d.TotalInvestment,
			Appearances:     d.Appearances,
			TotalSeed:       d.TotalSeed,
		}
	}

	results, baselineROI := CalculateTeamAnalyticsResults(input)
	return results, baselineROI, nil
}

func (s *Service) GetSeedVarianceAnalytics(ctx context.Context) ([]SeedVarianceResult, error) {
	data, err := s.repo.GetSeedVarianceAnalytics(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting seed variance analytics: %w", err)
	}

	// Convert ports data to input type for pure function
	input := make([]SeedVarianceInput, len(data))
	for i, d := range data {
		input[i] = SeedVarianceInput{
			Seed:             d.Seed,
			InvestmentStdDev: d.InvestmentStdDev,
			PointsStdDev:     d.PointsStdDev,
			InvestmentMean:   d.InvestmentMean,
			PointsMean:       d.PointsMean,
			TeamCount:        d.TeamCount,
		}
	}

	return CalculateSeedVarianceResults(input), nil
}

func (s *Service) GetAllAnalytics(ctx context.Context) (*AnalyticsResult, error) {
	seedAnalytics, totalPoints, totalInvestment, err := s.GetSeedAnalytics(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting seed analytics: %w", err)
	}

	regionAnalytics, _, _, err := s.GetRegionAnalytics(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting region analytics: %w", err)
	}

	teamAnalytics, _, err := s.GetTeamAnalytics(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting team analytics: %w", err)
	}

	seedVarianceAnalytics, err := s.GetSeedVarianceAnalytics(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting seed variance analytics: %w", err)
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
