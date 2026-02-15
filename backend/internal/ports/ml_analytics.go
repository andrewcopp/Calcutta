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

type TournamentSimStatsByID struct {
	TournamentID     string
	Season           int
	TotalSimulations int
	TotalPredictions int
	MeanWins         float64
	MedianWins       float64
	MaxWins          int
	LastUpdated      time.Time
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
	Name         string
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
	TeamID      string
	SchoolName  string
	Seed        int
	Region      string
	BidPoints   int
	ExpectedROI float64
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

// SimulatedCalcuttaEntryRanking represents an entry's performance in simulated calcuttas (analytics.entry_performance)
type SimulatedCalcuttaEntryRanking struct {
	Rank                   int
	EntryName              string
	MeanNormalizedPayout   float64
	MedianNormalizedPayout float64
	PTop1                  float64
	PInMoney               float64
	TotalSimulations       int
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
	TeamID     string
	SchoolName string
	Seed       int
	Region     string
	BidPoints  int
}

type TournamentSimulationBatch struct {
	ID                   string
	TournamentID         string
	SimulationStateID    string
	NSims                int
	Seed                 int
	ProbabilitySourceKey string
	CreatedAt            time.Time
}

type CalcuttaEvaluationRun struct {
	ID                    string
	SimulatedTournamentID string
	CalcuttaSnapshotID    *string
	Purpose               string
	CreatedAt             time.Time
}

type OptimizedEntry struct {
	ID                    string
	RunKey                *string
	SimulatedTournamentID *string
	CalcuttaID            *string
	Purpose               string
	ReturnsModelKey       string
	InvestmentModelKey    string
	OptimizerKind         string
	ParamsJSON            []byte
	GitSHA                *string
	CreatedAt             time.Time
}

// MLAnalyticsRepo defines the interface for ML analytics data access
type MLAnalyticsRepository interface {
	// Tournament simulations
	GetTournamentSimStats(ctx context.Context, year int) (*TournamentSimStats, error)
	GetTournamentSimStatsByCoreTournamentID(ctx context.Context, coreTournamentID string) (*TournamentSimStatsByID, error)
	GetTeamPerformance(ctx context.Context, year int, teamID string) (*TeamPerformance, error)
	GetTeamPerformanceByCalcutta(ctx context.Context, calcuttaID string, teamID string) (*TeamPerformance, error)

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

	// Calcutta-scoped simulated calcuttas
	GetSimulatedCalcuttaEntryRankings(ctx context.Context, calcuttaID string, calcuttaEvaluationRunID *string) (string, *string, []SimulatedCalcuttaEntryRanking, error)

	ListTournamentSimulationBatchesByCoreTournamentID(ctx context.Context, coreTournamentID string) ([]TournamentSimulationBatch, error)
	ListCalcuttaEvaluationRunsByCoreCalcuttaID(ctx context.Context, calcuttaID string) ([]CalcuttaEvaluationRun, error)
	ListOptimizedEntriesByCoreCalcuttaID(ctx context.Context, calcuttaID string) ([]OptimizedEntry, error)
}
