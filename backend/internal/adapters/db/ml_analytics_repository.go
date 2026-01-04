package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/google/uuid"
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
func derefInt32ML(v any) int32 {
	switch x := v.(type) {
	case int32:
		return x
	case *int32:
		if x == nil {
			return 0
		}
		return *x
	case nil:
		return 0
	default:
		return 0
	}
}

func derefStringML(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case *string:
		if x == nil {
			return ""
		}
		return *x
	case nil:
		return ""
	default:
		return ""
	}
}

func stringFromInterfaceML(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	case nil:
		return ""
	default:
		return fmt.Sprint(x)
	}
}

func uuidToStringPtr(v pgtype.UUID) *string {
	if !v.Valid {
		return nil
	}
	u := uuid.UUID(v.Bytes)
	str := u.String()
	return &str
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

func (r *MLAnalyticsRepository) GetTournamentSimStatsByCoreTournamentID(ctx context.Context, coreTournamentID string) (*ports.TournamentSimStatsByID, error) {
	row, err := r.q.GetTournamentSimStatsByCoreTournamentID(ctx, coreTournamentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &ports.TournamentSimStatsByID{
		TournamentID:     row.TournamentID,
		Season:           int(row.Season),
		TotalSimulations: int(row.TotalSimulations),
		TotalPredictions: int(row.TotalPredictions),
		MeanWins:         row.MeanWins,
		MedianWins:       row.MedianWins,
		MaxWins:          int(row.MaxWins),
		LastUpdated:      row.LastUpdated.Time,
	}, nil
}

func (r *MLAnalyticsRepository) GetTeamPerformanceByCalcutta(ctx context.Context, calcuttaID string, teamID string) (*ports.TeamPerformance, error) {
	row, err := r.q.GetTeamPerformanceByCalcutta(ctx, sqlc.GetTeamPerformanceByCalcuttaParams{
		CalcuttaID: calcuttaID,
		TeamID:     teamID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var roundDist map[string]int
	if err := json.Unmarshal(row.RoundDistribution, &roundDist); err != nil {
		return nil, err
	}

	return &ports.TeamPerformance{
		TeamID:            row.TeamID,
		SchoolName:        row.SchoolName,
		Seed:              int(derefInt32ML(row.Seed)),
		Region:            derefStringML(row.Region),
		KenpomNet:         row.KenpomNet,
		TotalSims:         int(row.TotalSims),
		AvgWins:           row.AvgWins,
		AvgPoints:         row.AvgPoints,
		RoundDistribution: roundDist,
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
		Seed:              int(derefInt32ML(row.Seed)),
		Region:            derefStringML(row.Region),
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
			Seed:       int(derefInt32ML(row.Seed)),
			Region:     derefStringML(row.Region),
			KenpomNet:  row.KenpomNet,
		})
	}

	return out, nil
}

func (r *MLAnalyticsRepository) ListTournamentSimulationBatchesByCoreTournamentID(ctx context.Context, coreTournamentID string) ([]ports.TournamentSimulationBatch, error) {
	rows, err := r.q.ListTournamentSimulationBatchesByCoreTournamentID(ctx, coreTournamentID)
	if err != nil {
		return nil, err
	}

	out := make([]ports.TournamentSimulationBatch, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.TournamentSimulationBatch{
			ID:                   row.ID,
			TournamentID:         row.TournamentID,
			SimulationStateID:    row.SimulationStateID,
			NSims:                int(row.NSims),
			Seed:                 int(row.Seed),
			ProbabilitySourceKey: row.ProbabilitySourceKey,
			CreatedAt:            row.CreatedAt.Time,
		})
	}

	return out, nil
}

func (r *MLAnalyticsRepository) ListCalcuttaEvaluationRunsByCoreCalcuttaID(ctx context.Context, calcuttaID string) ([]ports.CalcuttaEvaluationRun, error) {
	rows, err := r.q.ListCalcuttaEvaluationRunsByCoreCalcuttaID(ctx, calcuttaID)
	if err != nil {
		return nil, err
	}

	out := make([]ports.CalcuttaEvaluationRun, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.CalcuttaEvaluationRun{
			ID:                    row.ID,
			SimulatedTournamentID: row.SimulatedTournamentID,
			CalcuttaSnapshotID:    uuidToStringPtr(row.CalcuttaSnapshotID),
			Purpose:               row.Purpose,
			CreatedAt:             row.CreatedAt.Time,
		})
	}

	return out, nil
}

func (r *MLAnalyticsRepository) ListStrategyGenerationRunsByCoreCalcuttaID(ctx context.Context, calcuttaID string) ([]ports.StrategyGenerationRun, error) {
	rows, err := r.q.ListStrategyGenerationRunsByCoreCalcuttaID(ctx, calcuttaID)
	if err != nil {
		return nil, err
	}

	out := make([]ports.StrategyGenerationRun, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.StrategyGenerationRun{
			ID:                    row.ID,
			RunKey:                row.RunKey,
			SimulatedTournamentID: uuidToStringPtr(row.SimulatedTournamentID),
			CalcuttaID:            uuidToStringPtr(row.CalcuttaID),
			Purpose:               row.Purpose,
			ReturnsModelKey:       row.ReturnsModelKey,
			InvestmentModelKey:    row.InvestmentModelKey,
			OptimizerKey:          row.OptimizerKey,
			ParamsJSON:            row.ParamsJson,
			GitSHA:                row.GitSha,
			CreatedAt:             row.CreatedAt.Time,
		})
	}

	return out, nil
}

func (r *MLAnalyticsRepository) GetSimulatedCalcuttaEntryRankings(ctx context.Context, calcuttaID string, calcuttaEvaluationRunID *string) (string, *string, []ports.SimulatedCalcuttaEntryRanking, error) {
	if calcuttaEvaluationRunID != nil && *calcuttaEvaluationRunID != "" {
		evalID := *calcuttaEvaluationRunID
		rows, err := r.q.GetEntryPerformanceByCalcuttaEvaluationRunID(ctx, evalID)
		if err != nil {
			return "", nil, nil, err
		}

		out := make([]ports.SimulatedCalcuttaEntryRanking, 0, len(rows))
		for _, row := range rows {
			out = append(out, ports.SimulatedCalcuttaEntryRanking{
				Rank:                   int(row.Rank),
				EntryName:              row.EntryName,
				MeanNormalizedPayout:   row.MeanNormalizedPayout,
				MedianNormalizedPayout: row.MedianNormalizedPayout,
				PTop1:                  row.PTop1,
				PInMoney:               row.PInMoney,
				TotalSimulations:       int(row.TotalSimulations),
			})
		}

		return "", &evalID, out, nil
	}

	// Prefer lineage-native evaluation runs when available.
	evalID, err := r.q.GetLatestCalcuttaEvaluationRunIDByCoreCalcuttaID(ctx, calcuttaID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return "", nil, nil, err
		}
	} else {
		rows, err := r.q.GetEntryPerformanceByCalcuttaEvaluationRunID(ctx, evalID)
		if err != nil {
			return "", nil, nil, err
		}

		out := make([]ports.SimulatedCalcuttaEntryRanking, 0, len(rows))
		for _, row := range rows {
			out = append(out, ports.SimulatedCalcuttaEntryRanking{
				Rank:                   int(row.Rank),
				EntryName:              row.EntryName,
				MeanNormalizedPayout:   row.MeanNormalizedPayout,
				MedianNormalizedPayout: row.MedianNormalizedPayout,
				PTop1:                  row.PTop1,
				PInMoney:               row.PInMoney,
				TotalSimulations:       int(row.TotalSimulations),
			})
		}

		return "", &evalID, out, nil
	}

	// Fallback: derive run_key and query by run_id.
	runID, err := r.q.GetLatestStrategyGenerationRunKeyByCoreCalcuttaID(ctx, calcuttaID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil, nil, nil
		}
		return "", nil, nil, err
	}

	rows, err := r.q.GetEntryPerformanceByRunID(ctx, runID)
	if err != nil {
		return "", nil, nil, err
	}

	out := make([]ports.SimulatedCalcuttaEntryRanking, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.SimulatedCalcuttaEntryRanking{
			Rank:                   int(row.Rank),
			EntryName:              row.EntryName,
			MeanNormalizedPayout:   row.MeanNormalizedPayout,
			MedianNormalizedPayout: row.MedianNormalizedPayout,
			PTop1:                  row.PTop1,
			PInMoney:               row.PInMoney,
			TotalSimulations:       int(row.TotalSimulations),
		})
	}

	return runID, nil, out, nil
}

func (r *MLAnalyticsRepository) GetOurEntryDetails(ctx context.Context, year int, runID string) (*ports.OurEntryDetails, error) {
	// Prefer strategy_generation_runs resolved by run_key.
	strategyRun, err := r.q.GetStrategyGenerationRunByRunKey(ctx, runID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, nil
	} else {
		strategy := ""
		switch v := strategyRun.Strategy.(type) {
		case string:
			strategy = v
		case []byte:
			strategy = string(v)
		case nil:
			strategy = ""
		default:
			strategy = fmt.Sprint(v)
		}
		if strategy == "" {
			strategy = "legacy"
		}

		run := ports.OptimizationRun{
			RunID:        derefStringML(strategyRun.RunID),
			Name:         stringFromInterfaceML(strategyRun.Name),
			CalcuttaID:   uuidToStringPtr(strategyRun.CalcuttaID),
			Strategy:     strategy,
			NSims:        int(strategyRun.NSims),
			Seed:         int(strategyRun.Seed),
			BudgetPoints: int(strategyRun.BudgetPoints),
			CreatedAt:    strategyRun.CreatedAt.Time,
		}

		bidRows, err := r.q.GetOurEntryBidsByStrategyGenerationRunID(ctx, strategyRun.ID)
		if err != nil {
			return nil, err
		}

		portfolio := make([]ports.OurEntryBid, 0, len(bidRows))
		for _, row := range bidRows {
			portfolio = append(portfolio, ports.OurEntryBid{
				TeamID:      row.TeamID,
				SchoolName:  row.SchoolName,
				Seed:        int(derefInt32ML(row.Seed)),
				Region:      derefStringML(row.Region),
				BidPoints:   int(row.BidPoints),
				ExpectedROI: row.ExpectedRoi,
			})
		}

		summary := ports.EntryPerformanceSummary{}
		perfRow, err := r.q.GetOurEntryPerformanceSummaryByRunKey(ctx, runID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
		} else {
			summary.MeanNormalizedPayout = perfRow.MeanNormalizedPayout
			summary.PTop1 = perfRow.PTop1
			summary.PInMoney = perfRow.PInMoney
			percentile := perfRow.PercentileRank
			summary.PercentileRank = &percentile
		}
		return &ports.OurEntryDetails{
			Run:       run,
			Portfolio: portfolio,
			Summary:   summary,
		}, nil
	}
}

func (r *MLAnalyticsRepository) GetEntryRankings(ctx context.Context, year int, runID string, limit, offset int) ([]ports.EntryRanking, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.q.GetEntryRankingsByRunKey(ctx, sqlc.GetEntryRankingsByRunKeyParams{
		PageOffset: int32(offset),
		PageLimit:  int32(limit),
		RunID:      runID,
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
			MeanNormalizedPayout: row.MeanNormalizedPayout,
			PercentileRank:       row.PercentileRank,
			PTop1:                row.PTop1,
			PInMoney:             row.PInMoney,
			TotalEntries:         int(row.TotalEntries),
		})
	}

	return out, nil
}

func (r *MLAnalyticsRepository) GetEntrySimulations(ctx context.Context, year int, runID string, entryKey string, limit, offset int) (*ports.EntrySimulationDrillDown, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 5000 {
		limit = 5000
	}
	if offset < 0 {
		offset = 0
	}

	summaryRow, err := r.q.GetEntrySimulationSummaryByRunKeyAndEntryName(ctx, sqlc.GetEntrySimulationSummaryByRunKeyAndEntryNameParams{
		RunID:     runID,
		EntryName: entryKey,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if summaryRow.TotalSimulations == 0 {
		return nil, nil
	}

	rows, err := r.q.GetEntrySimulationsByRunKeyAndEntryName(ctx, sqlc.GetEntrySimulationsByRunKeyAndEntryNameParams{
		RunID:      runID,
		EntryName:  entryKey,
		PageOffset: int32(offset),
		PageLimit:  int32(limit),
	})
	if err != nil {
		return nil, err
	}

	sims := make([]ports.EntrySimulationOutcome, 0, len(rows))
	for _, row := range rows {
		payoutCents := int(row.PayoutCents)
		sims = append(sims, ports.EntrySimulationOutcome{
			SimID:            int(row.SimID),
			PayoutCents:      payoutCents,
			TotalPoints:      row.PointsScored,
			FinishPosition:   int(row.Rank),
			IsTied:           row.IsTied,
			NormalizedPayout: row.NormalizedPayout,
			NEntries:         int(row.NEntries),
		})
	}

	return &ports.EntrySimulationDrillDown{
		EntryKey:    entryKey,
		RunID:       runID,
		Simulations: sims,
		Summary: ports.EntrySimulationSummary{
			TotalSimulations:     int(summaryRow.TotalSimulations),
			MeanPayoutCents:      summaryRow.MeanPayoutCents,
			MeanPoints:           summaryRow.MeanPoints,
			MeanNormalizedPayout: summaryRow.MeanNormalizedPayout,
			P50PayoutCents:       int(summaryRow.P50PayoutCents),
			P90PayoutCents:       int(summaryRow.P90PayoutCents),
		},
	}, nil
}

func (r *MLAnalyticsRepository) GetEntryPortfolio(ctx context.Context, year int, runID string, entryKey string) (*ports.EntryPortfolio, error) {
	teams := make([]ports.EntryPortfolioTeam, 0)
	totalBid := 0

	// Check if this is our strategy or an actual entry
	if entryKey == "our_strategy" {
		strategyRun, err := r.q.GetStrategyGenerationRunByRunKey(ctx, runID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
		} else {
			rows, err := r.q.GetEntryPortfolioByStrategyGenerationRunID(ctx, strategyRun.ID)
			if err != nil {
				return nil, err
			}
			for _, row := range rows {
				teams = append(teams, ports.EntryPortfolioTeam{
					TeamID:     row.TeamID,
					SchoolName: row.SchoolName,
					Seed:       int(derefInt32ML(row.Seed)),
					Region:     derefStringML(row.Region),
					BidPoints:  int(row.BidPoints),
				})
				totalBid += int(row.BidPoints)
			}
			return &ports.EntryPortfolio{
				EntryKey: entryKey,
				Teams:    teams,
				TotalBid: totalBid,
				NTeams:   len(teams),
			}, nil
		}

		// Legacy fallback: use run_id join.
		rows, err := r.q.GetEntryPortfolio(ctx, runID)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			teams = append(teams, ports.EntryPortfolioTeam{
				TeamID:     row.TeamID,
				SchoolName: row.SchoolName,
				Seed:       int(derefInt32ML(row.Seed)),
				Region:     derefStringML(row.Region),
				BidPoints:  int(row.BidPoints),
			})
			totalBid += int(row.BidPoints)
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
				TeamID:     row.TeamID,
				SchoolName: row.SchoolName,
				Seed:       int(derefInt32ML(row.Seed)),
				Region:     derefStringML(row.Region),
				BidPoints:  int(row.BidPoints),
			})
			totalBid += int(row.BidPoints)
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
		strategy := ""
		switch v := row.Strategy.(type) {
		case string:
			strategy = v
		case []byte:
			strategy = string(v)
		case nil:
			strategy = ""
		default:
			strategy = fmt.Sprint(v)
		}
		if strategy == "" {
			strategy = "legacy"
		}

		out = append(out, ports.OptimizationRun{
			RunID:        derefStringML(row.RunID),
			Name:         stringFromInterfaceML(row.Name),
			CalcuttaID:   uuidToStringPtr(row.CalcuttaID),
			Strategy:     strategy,
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
