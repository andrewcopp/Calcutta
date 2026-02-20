package analytics

import (
	"context"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Service struct {
	repo ports.AnalyticsRepo
}

func New(repo ports.AnalyticsRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetBestInvestments(ctx context.Context, limit int) ([]BestInvestmentResult, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	data, err := s.repo.GetBestInvestments(ctx, limit)
	if err != nil {
		return nil, err
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
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	data, err := s.repo.GetBestCareers(ctx, limit)
	if err != nil {
		return nil, err
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
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	data, err := s.repo.GetBestInvestmentBids(ctx, limit)
	if err != nil {
		return nil, err
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
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	data, err := s.repo.GetBestEntries(ctx, limit)
	if err != nil {
		return nil, err
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
		return nil, err
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

	seeds := make([]int, 0, len(bySeed))
	for seed := range bySeed {
		seeds = append(seeds, seed)
	}
	sort.Ints(seeds)

	summaries := make([]SeedInvestmentSummaryResult, 0, len(seeds))
	for _, seed := range seeds {
		values := bySeed[seed]
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

	return &SeedInvestmentDistributionResult{Points: points, Summaries: summaries}, nil
}

func (s *Service) GetSeedAnalytics(ctx context.Context) ([]SeedAnalyticsResult, float64, float64, error) {
	data, totalPoints, totalInvestment, err := s.repo.GetSeedAnalytics(ctx)
	if err != nil {
		return nil, 0, 0, err
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

func (s *Service) GetTeamAnalytics(ctx context.Context) ([]TeamAnalyticsResult, float64, error) {
	data, err := s.repo.GetTeamAnalytics(ctx)
	if err != nil {
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

func (s *Service) GetSeedVarianceAnalytics(ctx context.Context) ([]SeedVarianceResult, error) {
	data, err := s.repo.GetSeedVarianceAnalytics(ctx)
	if err != nil {
		return nil, err
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
