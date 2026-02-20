package analytics

// SeedAnalyticsResult holds aggregated analytics for a single seed value.
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

// RegionAnalyticsResult holds aggregated analytics for a single region.
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

// TeamAnalyticsResult holds aggregated analytics for a single school/team.
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

// SeedVarianceResult holds variance analytics for a single seed value.
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

// SeedInvestmentPointResult holds a single data point for seed investment distribution.
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

// SeedInvestmentSummaryResult holds summary statistics for a single seed's investment distribution.
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

// BestInvestmentResult holds data for a single top-performing investment.
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

// InvestmentLeaderboardResult holds data for a single investment bid leaderboard entry.
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

// EntryLeaderboardResult holds data for a single entry leaderboard entry.
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

// CareerLeaderboardResult holds career-level statistics for a participant.
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

// AnalyticsResult holds the combined analytics results for all dimensions.
type AnalyticsResult struct {
	SeedAnalytics         []SeedAnalyticsResult
	RegionAnalytics       []RegionAnalyticsResult
	TeamAnalytics         []TeamAnalyticsResult
	SeedVarianceAnalytics []SeedVarianceResult
	TotalPoints           float64
	TotalInvestment       float64
	BaselineROI           float64
}

// SeedInvestmentDistributionResult holds the full seed investment distribution data.
type SeedInvestmentDistributionResult struct {
	Points    []SeedInvestmentPointResult
	Summaries []SeedInvestmentSummaryResult
}

// SeedAnalyticsInput represents the raw data needed to calculate seed analytics.
// This mirrors ports.SeedAnalyticsData but is defined here to avoid circular dependencies in tests.
type SeedAnalyticsInput struct {
	Seed            int
	TotalPoints     float64
	TotalInvestment float64
	TeamCount       int
}

// SeedVarianceInput represents the raw data needed to calculate seed variance analytics.
type SeedVarianceInput struct {
	Seed             int
	InvestmentStdDev float64
	PointsStdDev     float64
	InvestmentMean   float64
	PointsMean       float64
	TeamCount        int
}
