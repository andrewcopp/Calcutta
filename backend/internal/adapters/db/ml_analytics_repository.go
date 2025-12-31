package db

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MLAnalyticsRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewMLAnalyticsRepository(pool *pgxpool.Pool) *MLAnalyticsRepository {
	return &MLAnalyticsRepository{pool: pool, q: sqlc.New(pool)}
}

// Helper functions for nullable types
func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func uuidToStringPtr(v pgtype.UUID) *string {
	if !v.Valid {
		return nil
	}
	s := v.Bytes.String()
	return &s
}

func (r *MLAnalyticsRepository) GetTournamentSimStats(ctx context.Context, year int) (*ports.TournamentSimStats, error) {
	row, err := r.q.GetTournamentSimStatsByYear(ctx, int32(year))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &ports.TournamentSimStats{
		TournamentID: row.TournamentID,
		Season:       int(row.Season),
		NSims:        int(row.NSims),
		NTeams:       int(row.NTeams),
		AvgProgress:  row.AvgProgress,
		MaxProgress:  int(row.MaxProgress),
	}, nil
}

func (r *MLAnalyticsRepository) GetTeamPerformance(ctx context.Context, year int, teamID string) (*ports.TeamPerformance, error) {
	row, err := r.q.GetTeamPerformanceByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Parse round distribution from JSONB
	var roundDist map[string]int
	if err := json.Unmarshal(row.RoundDistribution, &roundDist); err != nil {
		return nil, err
	}

	return &ports.TeamPerformance{
		TeamID:            row.TeamID,
		SchoolName:        row.SchoolName,
		Seed:              int(derefInt32(row.Seed)),
		Region:            derefString(row.Region),
		KenpomNet:         row.KenpomNet,
		TotalSims:         int(row.TotalSims),
		AvgWins:           row.AvgWins,
		AvgPoints:         row.AvgPoints,
		RoundDistribution: roundDist,
	}, nil
}

func (r *MLAnalyticsRepository) GetTeamPredictions(ctx context.Context, year int, runID *string) ([]ports.TeamPrediction, error) {
	rows, err := r.q.GetTeamPredictionsByYear(ctx, int32(year))
	if err != nil {
		return nil, err
	}

	out := make([]ports.TeamPrediction, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.TeamPrediction{
			TeamID:     row.TeamID,
			SchoolName: row.SchoolName,
			Seed:       int(derefInt32(row.Seed)),
			Region:     derefString(row.Region),
			KenpomNet:  row.KenpomNet,
		})
	}

	return out, nil
}

func (r *MLAnalyticsRepository) GetOurEntryDetails(ctx context.Context, year int, runID string) (*ports.OurEntryDetails, error) {
	// Get optimization run metadata
	runRow, err := r.q.GetOptimizationRunByID(ctx, runID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	run := ports.OptimizationRun{
		RunID:        runRow.RunID,
		CalcuttaID:   runRow.CalcuttaID,
		Strategy:     runRow.Strategy,
		NSims:        int(runRow.NSims),
		Seed:         int(runRow.Seed),
		BudgetPoints: int(runRow.BudgetPoints),
		CreatedAt:    runRow.CreatedAt.Time,
	}

	// Get portfolio bids
	bidRows, err := r.q.GetOurEntryBidsByRunID(ctx, runID)
	if err != nil {
		return nil, err
	}

	portfolio := make([]ports.OurEntryBid, 0, len(bidRows))
	for _, row := range bidRows {
		portfolio = append(portfolio, ports.OurEntryBid{
			TeamID:               row.TeamID,
			SchoolName:           row.SchoolName,
			Seed:                 int(row.Seed),
			Region:               row.Region,
			RecommendedBidPoints: int(row.RecommendedBidPoints),
			ExpectedROI:          floatFromPgNumeric(row.ExpectedRoi),
		})
	}

	// Entry performance queries removed in new schema
	// Return empty summary for now
	summary := ports.EntryPerformanceSummary{}

	return &ports.OurEntryDetails{
		Run:       run,
		Portfolio: portfolio,
		Summary:   summary,
	}, nil
}

func (r *MLAnalyticsRepository) GetEntryRankings(ctx context.Context, year int, runID string, limit, offset int) ([]ports.EntryRanking, error) {
	// Query removed in new schema - return empty for now
	return []ports.EntryRanking{}, nil
}

func (r *MLAnalyticsRepository) GetEntrySimulations(ctx context.Context, year int, runID string, entryKey string, limit, offset int) (*ports.EntrySimulationDrillDown, error) {
	// Query removed in new schema - return empty for now
	return &ports.EntrySimulationDrillDown{
		EntryKey:    entryKey,
		RunID:       runID,
		Simulations: []ports.EntrySimulationOutcome{},
		Summary:     ports.EntrySimulationSummary{},
	}, nil
}

func (r *MLAnalyticsRepository) GetEntryPortfolio(ctx context.Context, year int, runID string, entryKey string) (*ports.EntryPortfolio, error) {
	teams := make([]ports.EntryPortfolioTeam, 0)
	totalBid := 0

	// Check if this is our strategy or an actual entry
	if entryKey == "our_strategy" {
		rows, err := r.q.GetEntryPortfolio(ctx, runID)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			teams = append(teams, ports.EntryPortfolioTeam{
				TeamID:          row.TeamID,
				SchoolName:      row.SchoolName,
				Seed:            int(row.Seed),
				Region:          row.Region,
				BidAmountPoints: int(row.BidAmount),
			})
			totalBid += int(row.BidAmount)
		}
	} else {
		rows, err := r.q.GetActualEntryPortfolio(ctx, sqlc.GetActualEntryPortfolioParams{
			RunID:     runID,
			EntryName: entryKey,
		})
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			teams = append(teams, ports.EntryPortfolioTeam{
				TeamID:          row.TeamID,
				SchoolName:      row.SchoolName,
				Seed:            int(row.Seed),
				Region:          row.Region,
				BidAmountPoints: int(row.BidAmountPoints),
			})
			totalBid += int(row.BidAmountPoints)
		}
	}

	return &ports.EntryPortfolio{
		EntryKey: entryKey,
		Teams:    teams,
		TotalBid: totalBid,
		NTeams:   len(teams),
	}, nil
}

func (r *MLAnalyticsRepository) GetOptimizationRuns(ctx context.Context, year int) ([]ports.OptimizationRun, error) {
	rows, err := r.q.GetOptimizationRunsByYear(ctx, int32(year))
	if err != nil {
		return nil, err
	}

	out := make([]ports.OptimizationRun, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.OptimizationRun{
			RunID:        row.RunID,
			CalcuttaID:   row.CalcuttaID,
			Strategy:     row.Strategy,
			NSims:        int(row.NSims),
			Seed:         int(row.Seed),
			BudgetPoints: int(row.BudgetPoints),
			CreatedAt:    row.CreatedAt.Time,
		})
	}

	return out, nil
}

// Helper function to convert pgtype.Numeric to *float64
func floatPtrFromPgNumeric(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f64, err := n.Float64Value()
	if err != nil {
		return nil
	}
	val := f64.Float64
	return &val
}

// Helper function to convert pgtype.Numeric to float64
func floatFromPgNumeric(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f64, err := n.Float64Value()
	if err != nil {
		return 0
	}
	return f64.Float64
}
