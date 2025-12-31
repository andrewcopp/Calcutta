package ports

import (
	"context"
	"time"
)

// ML Analytics Ports - For tournament simulation and entry evaluation data

// TournamentSimStats represents simulation statistics for a tournament
type TournamentSimStats struct {
	TournamentID string
	Season       int
	NSims        int
	NTeams       int
	AvgProgress  float64
	MaxProgress  int
}

// TeamPerformance represents a team's performance across simulations
type TeamPerformance struct {
	TeamID            string
	SchoolName        string
	Seed              int
	Region            string
	KenpomNet         *float64
	TotalSims         int
	AvgWins           float64
	AvgPoints         float64
	RoundDistribution map[string]int // e.g., "R64": 0, "R32": 150, etc.
}

// TeamPrediction represents ML predictions and investment metrics for a team
type TeamPrediction struct {
	TeamID     string
	SchoolName string
	Seed       int
	Region     string
	KenpomNet  *float64
}

// OptimizationRun represents a strategy execution
type OptimizationRun struct {
	RunID        string
	CalcuttaID   *string
	Strategy     string
	NSims        int
	Seed         int
	BudgetPoints int
	CreatedAt    time.Time
}

// OurEntryDetails represents our optimized entry with portfolio and performance
type OurEntryDetails struct {
	Run       OptimizationRun
	Portfolio []OurEntryBid
	Summary   EntryPerformanceSummary
}

// OurEntryBid represents a single team in our portfolio
type OurEntryBid struct {
	TeamID               string
	SchoolName           string
	Seed                 int
	Region               string
	RecommendedBidPoints int
	ExpectedROI          float64
}

// EntryRanking represents an entry's ranking in the competition
type EntryRanking struct {
	Rank                 int
	EntryKey             string
	IsOurStrategy        bool
	NTeams               int
	TotalBidPoints       int
	MeanNormalizedPayout float64
	PercentileRank       float64
	PTop1                float64
	PInMoney             float64
	TotalEntries         int
}

// EntrySimulationOutcome represents a single simulation result for an entry
type EntrySimulationOutcome struct {
	SimID            int
	PayoutCents      int
	TotalPoints      float64
	FinishPosition   int
	IsTied           bool
	NormalizedPayout float64
	NEntries         int
}

// EntrySimulationDrillDown represents all simulations for an entry with summary
type EntrySimulationDrillDown struct {
	EntryKey    string
	RunID       string
	Simulations []EntrySimulationOutcome
	Summary     EntrySimulationSummary
}

// EntrySimulationSummary represents aggregated stats for an entry's simulations
type EntrySimulationSummary struct {
	TotalSimulations     int
	MeanPayoutCents      float64
	MeanPoints           float64
	MeanNormalizedPayout float64
	P50PayoutCents       int
	P90PayoutCents       int
}

// EntryPerformanceSummary represents aggregated performance metrics
type EntryPerformanceSummary struct {
	MeanNormalizedPayout float64
	PTop1                float64
	PInMoney             float64
	PercentileRank       *float64
}

// EntryPortfolio represents the team composition for any entry
type EntryPortfolio struct {
	EntryKey string
	Teams    []EntryPortfolioTeam
	TotalBid int
	NTeams   int
}

// EntryPortfolioTeam represents a single team in an entry's portfolio
type EntryPortfolioTeam struct {
	TeamID          string
	SchoolName      string
	Seed            int
	Region          string
	BidAmountPoints int
}

// MLAnalyticsRepo defines the interface for ML analytics data access
type MLAnalyticsRepository interface {
	// Tournament simulations
	GetTournamentSimStats(ctx context.Context, year int) (*TournamentSimStats, error)
	GetTeamPerformance(ctx context.Context, year int, teamID string) (*TeamPerformance, error)

	// Team predictions
	GetTeamPredictions(ctx context.Context, year int, runID *string) ([]TeamPrediction, error)

	// Our entry
	GetOurEntryDetails(ctx context.Context, year int, runID string) (*OurEntryDetails, error)

	// Entry rankings
	GetEntryRankings(ctx context.Context, year int, runID string, limit, offset int) ([]EntryRanking, error)

	// Entry drill-down
	GetEntrySimulations(ctx context.Context, year int, runID string, entryKey string, limit, offset int) (*EntrySimulationDrillDown, error)

	// Entry portfolio
	GetEntryPortfolio(ctx context.Context, year int, runID string, entryKey string) (*EntryPortfolio, error)

	// Available runs
	GetOptimizationRuns(ctx context.Context, year int) ([]OptimizationRun, error)
}
