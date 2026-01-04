package analytics

import (
	appanalytics "github.com/andrewcopp/Calcutta/backend/internal/app/analytics"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Service = appanalytics.Service

type AnalyticsResult = appanalytics.AnalyticsResult
type SeedAnalyticsResult = appanalytics.SeedAnalyticsResult
type RegionAnalyticsResult = appanalytics.RegionAnalyticsResult
type TeamAnalyticsResult = appanalytics.TeamAnalyticsResult
type SeedVarianceResult = appanalytics.SeedVarianceResult

type SeedInvestmentDistributionResult = appanalytics.SeedInvestmentDistributionResult
type SeedInvestmentPointResult = appanalytics.SeedInvestmentPointResult
type SeedInvestmentSummaryResult = appanalytics.SeedInvestmentSummaryResult

type BestInvestmentResult = appanalytics.BestInvestmentResult
type InvestmentLeaderboardResult = appanalytics.InvestmentLeaderboardResult
type EntryLeaderboardResult = appanalytics.EntryLeaderboardResult
type CareerLeaderboardResult = appanalytics.CareerLeaderboardResult

func New(repo ports.AnalyticsRepo) *Service {
	return appanalytics.New(repo)
}
