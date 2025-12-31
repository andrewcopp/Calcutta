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

func (r *MLAnalyticsRepository) GetTeamPerformance(ctx context.Context, year int, teamKey string) (*ports.TeamPerformance, error) {
	row, err := r.q.GetTeamPerformanceByKey(ctx, teamKey)
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
		TeamKey:           row.TeamKey,
		SchoolName:        row.SchoolName,
		Seed:              int(row.Seed),
		Region:            row.Region,
		KenpomNet:         floatPtrFromPgNumeric(row.KenpomNet),
		TotalSims:         int(row.TotalSims),
		AvgWins:           row.AvgWins,
		AvgPoints:         row.AvgPoints,
		PChampion:         floatPtrFromPgNumeric(row.PChampion),
		PFinals:           floatPtrFromPgNumeric(row.PFinals),
		PFinalFour:        floatPtrFromPgNumeric(row.PFinalFour),
		PEliteEight:       floatPtrFromPgNumeric(row.PEliteEight),
		PSweetSixteen:     floatPtrFromPgNumeric(row.PSweetSixteen),
		PRound32:          floatPtrFromPgNumeric(row.PRound32),
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
			TeamKey:               row.TeamKey,
			SchoolName:            row.SchoolName,
			Seed:                  int(row.Seed),
			Region:                row.Region,
			ExpectedPoints:        floatFromPgNumeric(row.ExpectedPoints),
			PredictedMarketShare:  floatFromPgNumeric(row.PredictedMarketShare),
			PredictedMarketPoints: row.PredictedMarketPoints,
			PChampion:             floatPtrFromPgNumeric(row.PChampion),
			KenpomNet:             floatPtrFromPgNumeric(row.KenpomNet),
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
		CalcuttaKey:  runRow.CalcuttaKey,
		Strategy:     runRow.Strategy,
		NSims:        int(runRow.NSims),
		Seed:         int(runRow.Seed),
		BudgetPoints: int(runRow.BudgetPoints),
		RunTimestamp: runRow.RunTimestamp.Time,
	}

	// Get portfolio bids
	bidRows, err := r.q.GetOurEntryBidsByRunID(ctx, runID)
	if err != nil {
		return nil, err
	}

	portfolio := make([]ports.OurEntryBid, 0, len(bidRows))
	for _, row := range bidRows {
		portfolio = append(portfolio, ports.OurEntryBid{
			TeamKey:               row.TeamKey,
			SchoolName:            row.SchoolName,
			Seed:                  int(row.Seed),
			Region:                row.Region,
			BidAmountPoints:       int(row.BidAmountPoints),
			ExpectedPoints:        floatFromPgNumeric(row.ExpectedPoints),
			PredictedMarketPoints: floatFromPgNumeric(row.PredictedMarketPoints),
			ActualMarketPoints:    floatFromPgNumeric(row.ActualMarketPoints),
			OurOwnership:          floatFromPgNumeric(row.OurOwnership),
			ExpectedROI:           floatFromPgNumeric(row.ExpectedRoi),
			OurROI:                floatFromPgNumeric(row.OurRoi),
			ROIDegradation:        floatFromPgNumeric(row.RoiDegradation),
		})
	}

	// Get performance summary
	perfRow, err := r.q.GetEntryPerformanceByRunID(ctx, runID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ports.OurEntryDetails{
				Run:       run,
				Portfolio: portfolio,
				Summary:   ports.EntryPerformanceSummary{},
			}, nil
		}
		return nil, err
	}

	summary := ports.EntryPerformanceSummary{
		MeanNormalizedPayout: floatFromPgNumeric(perfRow.MeanNormalizedPayout),
		PTop1:                floatFromPgNumeric(perfRow.PTop1),
		PInMoney:             floatFromPgNumeric(perfRow.PInMoney),
		PercentileRank:       floatPtrFromPgNumeric(perfRow.PercentileRank),
	}

	return &ports.OurEntryDetails{
		Run:       run,
		Portfolio: portfolio,
		Summary:   summary,
	}, nil
}

func (r *MLAnalyticsRepository) GetEntryRankings(ctx context.Context, year int, runID string, limit, offset int) ([]ports.EntryRanking, error) {
	rows, err := r.q.GetEntryRankingsByRunID(ctx, sqlc.GetEntryRankingsByRunIDParams{
		Column1: runID,
		Column2: int32(limit),
		Column3: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	out := make([]ports.EntryRanking, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.EntryRanking{
			Rank:                 int(row.Rank),
			EntryKey:             row.EntryKey,
			IsOurStrategy:        row.IsOurStrategy,
			NTeams:               int(row.NTeams),
			TotalBidPoints:       int(row.TotalBidPoints),
			MeanNormalizedPayout: floatFromPgNumeric(row.MeanNormalizedPayout),
			PercentileRank:       floatFromPgNumeric(row.PercentileRank),
			PTop1:                floatFromPgNumeric(row.PTop1),
			PInMoney:             floatFromPgNumeric(row.PInMoney),
			TotalEntries:         int(row.TotalEntries),
		})
	}

	return out, nil
}

func (r *MLAnalyticsRepository) GetEntrySimulations(ctx context.Context, year int, runID string, entryKey string, limit, offset int) (*ports.EntrySimulationDrillDown, error) {
	// Get simulations
	simRows, err := r.q.GetEntrySimulationsByKey(ctx, sqlc.GetEntrySimulationsByKeyParams{
		Column1: runID,
		Column2: entryKey,
		Column3: int32(limit),
		Column4: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	simulations := make([]ports.EntrySimulationOutcome, 0, len(simRows))
	for _, row := range simRows {
		simulations = append(simulations, ports.EntrySimulationOutcome{
			SimID:            int(row.SimID),
			PayoutCents:      int(row.PayoutCents),
			TotalPoints:      floatFromPgNumeric(row.TotalPoints),
			FinishPosition:   int(row.FinishPosition),
			IsTied:           row.IsTied,
			NormalizedPayout: floatFromPgNumeric(row.NormalizedPayout),
			NEntries:         int(row.NEntries),
		})
	}

	// Get summary
	summaryRow, err := r.q.GetEntrySimulationSummary(ctx, sqlc.GetEntrySimulationSummaryParams{
		Column1: runID,
		Column2: entryKey,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ports.EntrySimulationDrillDown{
				EntryKey:    entryKey,
				RunID:       runID,
				Simulations: simulations,
				Summary:     ports.EntrySimulationSummary{},
			}, nil
		}
		return nil, err
	}

	summary := ports.EntrySimulationSummary{
		TotalSimulations:     int(summaryRow.TotalSimulations),
		MeanPayoutCents:      summaryRow.MeanPayoutCents,
		MeanPoints:           summaryRow.MeanPoints,
		MeanNormalizedPayout: summaryRow.MeanNormalizedPayout,
		P50PayoutCents:       int(summaryRow.P50PayoutCents),
		P90PayoutCents:       int(summaryRow.P90PayoutCents),
	}

	return &ports.EntrySimulationDrillDown{
		EntryKey:    entryKey,
		RunID:       runID,
		Simulations: simulations,
		Summary:     summary,
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
				TeamKey:    row.TeamKey,
				SchoolName: row.SchoolName,
				Seed:       int(row.Seed),
				Region:     row.Region,
				BidAmount:  int(row.BidAmount),
			})
			totalBid += int(row.BidAmount)
		}
	} else {
		rows, err := r.q.GetActualEntryPortfolio(ctx, sqlc.GetActualEntryPortfolioParams{
			RunID:    runID,
			EntryKey: entryKey,
		})
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			teams = append(teams, ports.EntryPortfolioTeam{
				TeamKey:    row.TeamKey,
				SchoolName: row.SchoolName,
				Seed:       int(row.Seed),
				Region:     row.Region,
				BidAmount:  int(row.BidAmount),
			})
			totalBid += int(row.BidAmount)
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
			CalcuttaKey:  row.CalcuttaKey,
			Strategy:     row.Strategy,
			NSims:        int(row.NSims),
			Seed:         int(row.Seed),
			BudgetPoints: int(row.BudgetPoints),
			RunTimestamp: row.RunTimestamp.Time,
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
