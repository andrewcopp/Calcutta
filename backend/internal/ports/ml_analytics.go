package ports

import (
	"context"
	"time"
)

// ML Analytics Ports - For tournament simulation and entry evaluation data

// TournamentSimStats represents aggregated simulation statistics
type TournamentSimStats struct {
	TournamentKey string
	Season        int
	NSims         int
	NTeams        int
	AvgProgress   float64
	MaxProgress   int
}

// TeamPerformance represents a team's performance across all simulations
type TeamPerformance struct {
	TeamKey           string
	SchoolName        string
	Seed              int
	Region            string
	KenpomNet         *float64
	TotalSims         int
	AvgWins           float64
	AvgPoints         float64
	PChampion         *float64
	PFinals           *float64
	PFinalFour        *float64
	PEliteEight       *float64
	PSweetSixteen     *float64
	PRound32          *float64
	RoundDistribution map[string]int // e.g., "R64": 0, "R32": 150, etc.
}

// TeamPrediction represents ML predictions and investment metrics for a team
type TeamPrediction struct {
	TeamKey               string
	SchoolName            string
	Seed                  int
	Region                string
	ExpectedPoints        float64
	PredictedMarketShare  float64
	PredictedMarketPoints float64
	PChampion             *float64
	KenpomNet             *float64
}

// OptimizationRun represents a strategy execution
type OptimizationRun struct {
	RunID        string
	CalcuttaKey  string
	Strategy     string
	NSims        int
	Seed         int
	BudgetPoints int
	RunTimestamp time.Time
}

// OurEntryDetails represents our optimized entry with portfolio and performance
type OurEntryDetails struct {
	Run       OptimizationRun
	Portfolio []OurEntryBid
	Summary   EntryPerformanceSummary
}

// OurEntryBid represents a single team in our portfolio
type OurEntryBid struct {
	TeamKey               string
	SchoolName            string
	Seed                  int
	Region                string
	BidAmountPoints       int
	ExpectedPoints        float64
	PredictedMarketPoints float64
	ActualMarketPoints    float64
	OurOwnership          float64
	ExpectedROI           float64
	OurROI                float64
	ROIDegradation        float64
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
	TeamKey    string
	SchoolName string
	Seed       int
	Region     string
	BidAmount  int
}

// MLAnalyticsRepo defines the interface for ML analytics data access
type MLAnalyticsRepo interface {
	// Tournament simulations
	GetTournamentSimStats(ctx context.Context, year int) (*TournamentSimStats, error)
	GetTeamPerformance(ctx context.Context, year int, teamKey string) (*TeamPerformance, error)

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
