package analytics

import (
	"context"
	"reflect"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type stubAnalyticsRepo struct {
	seedData           []ports.SeedAnalyticsData
	totalPoints        float64
	totalInvestment    float64
	seedAnalyticsErr   error
	seedVarianceData   []ports.SeedVarianceData
	seedVarianceErr    error
	regionData         []ports.RegionAnalyticsData
	regionTotalPoints  float64
	regionTotalInv     float64
	regionAnalyticsErr error
	teamData           []ports.TeamAnalyticsData
	teamAnalyticsErr   error
}

func (s *stubAnalyticsRepo) GetSeedAnalytics(ctx context.Context) ([]ports.SeedAnalyticsData, float64, float64, error) {
	return s.seedData, s.totalPoints, s.totalInvestment, s.seedAnalyticsErr
}

func (s *stubAnalyticsRepo) GetRegionAnalytics(ctx context.Context) ([]ports.RegionAnalyticsData, float64, float64, error) {
	return s.regionData, s.regionTotalPoints, s.regionTotalInv, s.regionAnalyticsErr
}

func (s *stubAnalyticsRepo) GetTeamAnalytics(ctx context.Context) ([]ports.TeamAnalyticsData, error) {
	return s.teamData, s.teamAnalyticsErr
}

func (s *stubAnalyticsRepo) GetSeedVarianceAnalytics(ctx context.Context) ([]ports.SeedVarianceData, error) {
	return s.seedVarianceData, s.seedVarianceErr
}

func (s *stubAnalyticsRepo) GetSeedInvestmentPoints(ctx context.Context) ([]ports.SeedInvestmentPointData, error) {
	return nil, nil
}

func (s *stubAnalyticsRepo) GetBestInvestments(ctx context.Context, limit int) ([]ports.BestInvestmentData, error) {
	return nil, nil
}

func (s *stubAnalyticsRepo) GetBestInvestmentBids(ctx context.Context, limit int) ([]ports.InvestmentLeaderboardData, error) {
	return nil, nil
}

func (s *stubAnalyticsRepo) GetBestEntries(ctx context.Context, limit int) ([]ports.EntryLeaderboardData, error) {
	return nil, nil
}

func (s *stubAnalyticsRepo) GetBestCareers(ctx context.Context, limit int) ([]ports.CareerLeaderboardData, error) {
	return nil, nil
}

func (s *stubAnalyticsRepo) GetCalcuttaPredictedInvestment(ctx context.Context, calcuttaID string, strategyGenerationRunID *string, marketShareRunID *string) (*string, *string, []ports.CalcuttaPredictedInvestmentData, error) {
	return nil, nil, nil, nil
}

func (s *stubAnalyticsRepo) GetCalcuttaPredictedReturns(ctx context.Context, calcuttaID string, strategyGenerationRunID *string, gameOutcomeRunID *string) (*string, *string, []ports.CalcuttaPredictedReturnsData, error) {
	return nil, nil, nil, nil
}

func (s *stubAnalyticsRepo) GetCalcuttaSimulatedEntry(ctx context.Context, calcuttaID string, strategyGenerationRunID *string) (*string, []ports.CalcuttaSimulatedEntryData, error) {
	return nil, nil, nil
}

func (s *stubAnalyticsRepo) ListAlgorithms(ctx context.Context, kind *string) ([]ports.Algorithm, error) {
	return nil, nil
}

func (s *stubAnalyticsRepo) ListGameOutcomeRunsByTournamentID(ctx context.Context, tournamentID string) ([]ports.GameOutcomeRun, error) {
	return nil, nil
}

func (s *stubAnalyticsRepo) ListMarketShareRunsByCalcuttaID(ctx context.Context, calcuttaID string) ([]ports.MarketShareRun, error) {
	return nil, nil
}

func (s *stubAnalyticsRepo) GetLatestPredictionRunsForCalcutta(ctx context.Context, calcuttaID string) (*ports.LatestPredictionRuns, error) {
	return nil, nil
}

func TestThatMeanFloat64ReturnsZeroForEmptySlice(t *testing.T) {
	got := meanFloat64(nil)

	if got != 0 {
		t.Errorf("expected mean to be 0, got %v", got)
	}
}

func TestThatStddevFloat64ReturnsZeroForSingleValue(t *testing.T) {
	got := stddevFloat64([]float64{5}, 5)

	if got != 0 {
		t.Errorf("expected stddev to be 0, got %v", got)
	}
}

func TestThatQuantileSortedReturnsMinForZeroQuantile(t *testing.T) {
	got := quantileSorted([]float64{1, 2, 3, 4}, 0)

	if got != 1 {
		t.Errorf("expected quantile to be 1, got %v", got)
	}
}

func TestThatQuantileSortedReturnsMaxForOneQuantile(t *testing.T) {
	got := quantileSorted([]float64{1, 2, 3, 4}, 1)

	if got != 4 {
		t.Errorf("expected quantile to be 4, got %v", got)
	}
}

func TestThatGetSeedAnalyticsCalculatesROIAsRatioToBaselineROI(t *testing.T) {
	repo := &stubAnalyticsRepo{
		seedData: []ports.SeedAnalyticsData{{
			Seed:            1,
			TotalPoints:     30,
			TotalInvestment: 10,
			TeamCount:       1,
		}},
		totalPoints:     60,
		totalInvestment: 20,
	}
	service := New(repo)

	results, _, _, err := service.GetSeedAnalytics(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := 0.0
	if len(results) > 0 {
		got = results[0].ROI
	}
	want := 1.0

	if got != want {
		t.Errorf("expected roi %v, got %v", want, got)
	}
}

func TestThatGetSeedVarianceAnalyticsCalculatesVarianceRatioAsInvestmentCVOverPointsCV(t *testing.T) {
	repo := &stubAnalyticsRepo{
		seedVarianceData: []ports.SeedVarianceData{{
			Seed:             1,
			InvestmentStdDev: 10,
			InvestmentMean:   20,
			PointsStdDev:     3,
			PointsMean:       6,
			TeamCount:        1,
		}},
	}
	service := New(repo)

	results, err := service.GetSeedVarianceAnalytics(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := 0.0
	if len(results) > 0 {
		got = results[0].VarianceRatio
	}
	want := (10.0 / 20.0) / (3.0 / 6.0)

	if got != want {
		t.Errorf("expected variance ratio %v, got %v", want, got)
	}
}

func TestThatGetSeedAnalyticsReturnsEmptySliceWhenRepoReturnsEmptyData(t *testing.T) {
	repo := &stubAnalyticsRepo{seedData: nil, totalPoints: 0, totalInvestment: 0}
	service := New(repo)

	results, _, _, err := service.GetSeedAnalytics(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := results
	want := []SeedAnalyticsResult{}
	if got == nil {
		got = []SeedAnalyticsResult{}
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected empty results")
	}
}
