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
	res, err := s.svc.GetAllAnalytics(ctx)
	return res, err
}

func (s *Service) GetBestInvestments(ctx context.Context, limit int) ([]services.BestInvestmentResult, error) {
	res, err := s.svc.GetBestInvestments(ctx, limit)
	return res, err
}

func (s *Service) GetBestInvestmentBids(ctx context.Context, limit int) ([]services.InvestmentLeaderboardResult, error) {
	res, err := s.svc.GetBestInvestmentBids(ctx, limit)
	return res, err
}

func (s *Service) GetBestEntries(ctx context.Context, limit int) ([]services.EntryLeaderboardResult, error) {
	res, err := s.svc.GetBestEntries(ctx, limit)
	return res, err
}

func (s *Service) GetBestCareers(ctx context.Context, limit int) ([]services.CareerLeaderboardResult, error) {
	res, err := s.svc.GetBestCareers(ctx, limit)
	return res, err
}

func (s *Service) GetSeedInvestmentDistribution(ctx context.Context) (*services.SeedInvestmentDistributionResult, error) {
	res, err := s.svc.GetSeedInvestmentDistribution(ctx)
	return res, err
}

func (s *Service) GetSeedAnalytics(ctx context.Context) ([]services.SeedAnalyticsResult, float64, float64, error) {
	res, totalPoints, totalInvestment, err := s.svc.GetSeedAnalytics(ctx)
	return res, totalPoints, totalInvestment, err
}

func (s *Service) GetRegionAnalytics(ctx context.Context) ([]services.RegionAnalyticsResult, float64, float64, error) {
	res, totalPoints, totalInvestment, err := s.svc.GetRegionAnalytics(ctx)
	return res, totalPoints, totalInvestment, err
}

func (s *Service) GetTeamAnalytics(ctx context.Context) ([]services.TeamAnalyticsResult, float64, error) {
	res, baselineROI, err := s.svc.GetTeamAnalytics(ctx)
	return res, baselineROI, err
}

func (s *Service) GetSeedVarianceAnalytics(ctx context.Context) ([]services.SeedVarianceResult, error) {
	res, err := s.svc.GetSeedVarianceAnalytics(ctx)
	return res, err
}
