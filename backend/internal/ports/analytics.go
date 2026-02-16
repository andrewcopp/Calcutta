package ports

import "context"

type SeedAnalyticsData struct {
	Seed            int
	TotalPoints     float64
	TotalInvestment float64
	TeamCount       int
}

type RegionAnalyticsData struct {
	Region          string
	TotalPoints     float64
	TotalInvestment float64
	TeamCount       int
}

type TeamAnalyticsData struct {
	SchoolID        string
	SchoolName      string
	TotalPoints     float64
	TotalInvestment float64
	Appearances     int
	TotalSeed       int
}

type SeedVarianceData struct {
	Seed             int
	InvestmentStdDev float64
	PointsStdDev     float64
	InvestmentMean   float64
	PointsMean       float64
	TeamCount        int
}

type SeedInvestmentPointData struct {
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

type BestInvestmentData struct {
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

type InvestmentLeaderboardData struct {
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

type EntryLeaderboardData struct {
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

type CareerLeaderboardData struct {
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

type AnalyticsRepo interface {
	GetSeedAnalytics(ctx context.Context) ([]SeedAnalyticsData, float64, float64, error)
	GetRegionAnalytics(ctx context.Context) ([]RegionAnalyticsData, float64, float64, error)
	GetTeamAnalytics(ctx context.Context) ([]TeamAnalyticsData, error)
	GetSeedVarianceAnalytics(ctx context.Context) ([]SeedVarianceData, error)
	GetSeedInvestmentPoints(ctx context.Context) ([]SeedInvestmentPointData, error)
	GetBestInvestments(ctx context.Context, limit int) ([]BestInvestmentData, error)
	GetBestInvestmentBids(ctx context.Context, limit int) ([]InvestmentLeaderboardData, error)
	GetBestEntries(ctx context.Context, limit int) ([]EntryLeaderboardData, error)
	GetBestCareers(ctx context.Context, limit int) ([]CareerLeaderboardData, error)
}
