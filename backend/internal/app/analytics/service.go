package analytics

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/pkg/services"
)

type Service struct {
	svc *services.AnalyticsService
}

func New(svc *services.AnalyticsService) *Service {
	return &Service{svc: svc}
}

func (s *Service) GetAllAnalytics(ctx context.Context) (*services.AnalyticsResult, error) {
	return s.svc.GetAllAnalytics(ctx)
}

func (s *Service) GetBestInvestments(ctx context.Context, limit int) ([]services.BestInvestmentResult, error) {
	return s.svc.GetBestInvestments(ctx, limit)
}

func (s *Service) GetBestInvestmentBids(ctx context.Context, limit int) ([]services.InvestmentLeaderboardResult, error) {
	return s.svc.GetBestInvestmentBids(ctx, limit)
}

func (s *Service) GetBestEntries(ctx context.Context, limit int) ([]services.EntryLeaderboardResult, error) {
	return s.svc.GetBestEntries(ctx, limit)
}

func (s *Service) GetBestCareers(ctx context.Context, limit int) ([]services.CareerLeaderboardResult, error) {
	return s.svc.GetBestCareers(ctx, limit)
}

func (s *Service) GetSeedInvestmentDistribution(ctx context.Context) (*services.SeedInvestmentDistributionResult, error) {
	return s.svc.GetSeedInvestmentDistribution(ctx)
}

func (s *Service) GetSeedAnalytics(ctx context.Context) ([]services.SeedAnalyticsResult, float64, float64, error) {
	return s.svc.GetSeedAnalytics(ctx)
}

func (s *Service) GetRegionAnalytics(ctx context.Context) ([]services.RegionAnalyticsResult, float64, float64, error) {
	return s.svc.GetRegionAnalytics(ctx)
}

func (s *Service) GetTeamAnalytics(ctx context.Context) ([]services.TeamAnalyticsResult, float64, error) {
	return s.svc.GetTeamAnalytics(ctx)
}

func (s *Service) GetSeedVarianceAnalytics(ctx context.Context) ([]services.SeedVarianceResult, error) {
	return s.svc.GetSeedVarianceAnalytics(ctx)
}
