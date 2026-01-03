package analytics

import (
	"context"
	"log"
	"math"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Service struct {
	repo ports.AnalyticsRepo
}

func New(repo ports.AnalyticsRepo) *Service {
	return &Service{repo: repo}
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

type SeedInvestmentPointResult struct {
	Seed             int
	TournamentName   string
	TournamentYear   int
	CalcuttaID       string
	TeamID           string
	SchoolName       string
	TotalBid         float64
	CalcuttaTotalBid float64
	NormalizedBid    float64
}

type SeedInvestmentSummaryResult struct {
	Seed   int
	Count  int
	Mean   float64
	StdDev float64
	Min    float64
	Q1     float64
	Median float64
	Q3     float64
	Max    float64
}

type BestInvestmentResult struct {
	TournamentName   string
	TournamentYear   int
	CalcuttaID       string
	TeamID           string
	SchoolName       string
	Seed             int
	Region           string
	TeamPoints       float64
	TotalBid         float64
	CalcuttaTotalBid float64
	CalcuttaTotalPts float64
	InvestmentShare  float64
	PointsShare      float64
	RawROI           float64
	NormalizedROI    float64
}

type InvestmentLeaderboardResult struct {
	TournamentName      string
	TournamentYear      int
	CalcuttaID          string
	EntryID             string
	EntryName           string
	TeamID              string
	SchoolName          string
	Seed                int
	Investment          float64
	OwnershipPercentage float64
	RawReturns          float64
	NormalizedReturns   float64
}

type EntryLeaderboardResult struct {
	TournamentName    string
	TournamentYear    int
	CalcuttaID        string
	EntryID           string
	EntryName         string
	TotalReturns      float64
	TotalParticipants int
	AverageReturns    float64
	NormalizedReturns float64
}

type CareerLeaderboardResult struct {
	EntryName              string
	Years                  int
	BestFinish             int
	Wins                   int
	Podiums                int
	InTheMoneys            int
	Top10s                 int
	CareerEarningsCents    int
	ActiveInLatestCalcutta bool
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

type CalcuttaPredictedInvestmentResult struct {
	TeamID     string
	SchoolName string
	Seed       int
	Region     string
	Rational   float64
	Predicted  float64
	Delta      float64
}

type CalcuttaPredictedReturnsResult struct {
	TeamID        string
	SchoolName    string
	Seed          int
	Region        string
	ProbPI        float64
	ProbR64       float64
	ProbR32       float64
	ProbS16       float64
	ProbE8        float64
	ProbFF        float64
	ProbChamp     float64
	ExpectedValue float64
}

type CalcuttaSimulatedEntryResult struct {
	TeamID         string
	SchoolName     string
	Seed           int
	Region         string
	ExpectedPoints float64
	ExpectedMarket float64
	OurBid         float64
}

type SeedInvestmentDistributionResult struct {
	Points    []SeedInvestmentPointResult
	Summaries []SeedInvestmentSummaryResult
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
		log.Printf("Error getting best investments: %v", err)
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

func (s *Service) GetCalcuttaPredictedInvestment(ctx context.Context, calcuttaID string, strategyGenerationRunID *string) (*string, []CalcuttaPredictedInvestmentResult, error) {
	selectedID, data, err := s.repo.GetCalcuttaPredictedInvestment(ctx, calcuttaID, strategyGenerationRunID)
	if err != nil {
		log.Printf("Error getting predicted investment: %v", err)
		return nil, nil, err
	}

	results := make([]CalcuttaPredictedInvestmentResult, 0, len(data))
	for _, d := range data {
		results = append(results, CalcuttaPredictedInvestmentResult{
			TeamID:     d.TeamID,
			SchoolName: d.SchoolName,
			Seed:       d.Seed,
			Region:     d.Region,
			Rational:   d.Rational,
			Predicted:  d.Predicted,
			Delta:      d.Delta,
		})
	}

	return selectedID, results, nil
}

func (s *Service) GetCalcuttaPredictedReturns(ctx context.Context, calcuttaID string, strategyGenerationRunID *string) (*string, []CalcuttaPredictedReturnsResult, error) {
	selectedID, data, err := s.repo.GetCalcuttaPredictedReturns(ctx, calcuttaID, strategyGenerationRunID)
	if err != nil {
		log.Printf("Error getting predicted returns: %v", err)
		return nil, nil, err
	}

	results := make([]CalcuttaPredictedReturnsResult, 0, len(data))
	for _, d := range data {
		results = append(results, CalcuttaPredictedReturnsResult{
			TeamID:        d.TeamID,
			SchoolName:    d.SchoolName,
			Seed:          d.Seed,
			Region:        d.Region,
			ProbPI:        d.ProbPI,
			ProbR64:       d.ProbR64,
			ProbR32:       d.ProbR32,
			ProbS16:       d.ProbS16,
			ProbE8:        d.ProbE8,
			ProbFF:        d.ProbFF,
			ProbChamp:     d.ProbChamp,
			ExpectedValue: d.ExpectedValue,
		})
	}

	return selectedID, results, nil
}

func (s *Service) GetCalcuttaSimulatedEntry(ctx context.Context, calcuttaID string, strategyGenerationRunID *string) (*string, []CalcuttaSimulatedEntryResult, error) {
	selectedID, data, err := s.repo.GetCalcuttaSimulatedEntry(ctx, calcuttaID, strategyGenerationRunID)
	if err != nil {
		log.Printf("Error getting simulated entry: %v", err)
		return nil, nil, err
	}

	results := make([]CalcuttaSimulatedEntryResult, 0, len(data))
	for _, d := range data {
		results = append(results, CalcuttaSimulatedEntryResult{
			TeamID:         d.TeamID,
			SchoolName:     d.SchoolName,
			Seed:           d.Seed,
			Region:         d.Region,
			ExpectedPoints: d.ExpectedPoints,
			ExpectedMarket: d.ExpectedMarket,
			OurBid:         d.OurBid,
		})
	}

	return selectedID, results, nil
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
		log.Printf("Error getting best careers: %v", err)
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
		log.Printf("Error getting best investment bids: %v", err)
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
		log.Printf("Error getting best entries: %v", err)
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
		log.Printf("Error getting seed investment points: %v", err)
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

func (s *Service) GetSeedAnalytics(ctx context.Context) ([]SeedAnalyticsResult, float64, float64, error) {
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

func (s *Service) GetRegionAnalytics(ctx context.Context) ([]RegionAnalyticsResult, float64, float64, error) {
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

func (s *Service) GetTeamAnalytics(ctx context.Context) ([]TeamAnalyticsResult, float64, error) {
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

func (s *Service) GetSeedVarianceAnalytics(ctx context.Context) ([]SeedVarianceResult, error) {
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
