package ports

import (
	"context"
	"time"
)

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

type CalcuttaPredictedInvestmentData struct {
	TeamID     string
	SchoolName string
	Seed       int
	Region     string
	Rational   float64
	Predicted  float64
	Delta      float64
}

type CalcuttaPredictedReturnsData struct {
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

type CalcuttaPredictedMarketShareData struct {
	TeamID         string
	SchoolName     string
	Seed           int
	Region         string
	RationalShare  float64
	PredictedShare float64
	DeltaPercent   float64
}

type TournamentPredictedAdvancementData struct {
	TeamID     string
	SchoolName string
	Seed       int
	Region     string
	ProbPI     float64
	ReachR64   float64
	ReachR32   float64
	ReachS16   float64
	ReachE8    float64
	ReachFF    float64
	ReachChamp float64
	WinChamp   float64
}

type CalcuttaSimulatedEntryData struct {
	TeamID         string
	SchoolName     string
	Seed           int
	Region         string
	ExpectedPoints float64
	ExpectedMarket float64
	OurBid         float64
}

type Algorithm struct {
	ID          string
	Kind        string
	Name        string
	Description *string
	ParamsJSON  []byte
	CreatedAt   time.Time
}

type GameOutcomeRun struct {
	ID           string
	AlgorithmID  string
	TournamentID string
	ParamsJSON   []byte
	GitSHA       *string
	CreatedAt    time.Time
}

type MarketShareRun struct {
	ID          string
	AlgorithmID string
	CalcuttaID  string
	ParamsJSON  []byte
	GitSHA      *string
	CreatedAt   time.Time
}

type LatestPredictionRuns struct {
	TournamentID     string
	GameOutcomeRunID *string
	MarketShareRunID *string
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
	GetCalcuttaPredictedInvestment(ctx context.Context, calcuttaID string, strategyGenerationRunID *string, marketShareRunID *string, gameOutcomeRunID *string) (*string, *string, []CalcuttaPredictedInvestmentData, error)
	GetCalcuttaPredictedReturns(ctx context.Context, calcuttaID string, strategyGenerationRunID *string, gameOutcomeRunID *string) (*string, *string, []CalcuttaPredictedReturnsData, error)
	GetTournamentPredictedAdvancement(ctx context.Context, tournamentID string, gameOutcomeRunID *string) (*string, []TournamentPredictedAdvancementData, error)
	GetCalcuttaPredictedMarketShare(ctx context.Context, calcuttaID string, marketShareRunID *string, gameOutcomeRunID *string) (*string, *string, []CalcuttaPredictedMarketShareData, error)
	GetCalcuttaSimulatedEntry(ctx context.Context, calcuttaID string, strategyGenerationRunID *string) (*string, []CalcuttaSimulatedEntryData, error)
	ListAlgorithms(ctx context.Context, kind *string) ([]Algorithm, error)
	ListGameOutcomeRunsByTournamentID(ctx context.Context, tournamentID string) ([]GameOutcomeRun, error)
	ListMarketShareRunsByCalcuttaID(ctx context.Context, calcuttaID string) ([]MarketShareRun, error)
	GetLatestPredictionRunsForCalcutta(ctx context.Context, calcuttaID string) (*LatestPredictionRuns, error)
}
