package db

import (
	"context"
	"errors"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *AnalyticsRepository) GetSeedAnalytics(ctx context.Context) ([]ports.SeedAnalyticsData, float64, float64, error) {
	rows, err := r.q.GetSeedAnalytics(ctx)
	if err != nil {
		return nil, 0, 0, err
	}

	out := make([]ports.SeedAnalyticsData, 0, len(rows))
	var totalPoints float64
	var totalInvestment float64
	for _, row := range rows {
		out = append(out, ports.SeedAnalyticsData{
			Seed:            int(row.Seed),
			TotalPoints:     row.TotalPoints,
			TotalInvestment: row.TotalInvestment,
			TeamCount:       int(row.TeamCount),
		})
		totalPoints += row.TotalPoints
		totalInvestment += row.TotalInvestment
	}

	return out, totalPoints, totalInvestment, nil
}

func (r *AnalyticsRepository) GetRegionAnalytics(ctx context.Context) ([]ports.RegionAnalyticsData, float64, float64, error) {
	rows, err := r.q.GetRegionAnalytics(ctx)
	if err != nil {
		return nil, 0, 0, err
	}

	out := make([]ports.RegionAnalyticsData, 0, len(rows))
	var totalPoints float64
	var totalInvestment float64
	for _, row := range rows {
		out = append(out, ports.RegionAnalyticsData{
			Region:          row.Region,
			TotalPoints:     row.TotalPoints,
			TotalInvestment: row.TotalInvestment,
			TeamCount:       int(row.TeamCount),
		})
		totalPoints += row.TotalPoints
		totalInvestment += row.TotalInvestment
	}

	return out, totalPoints, totalInvestment, nil
}

func (r *AnalyticsRepository) GetTeamAnalytics(ctx context.Context) ([]ports.TeamAnalyticsData, error) {
	rows, err := r.q.GetTeamAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]ports.TeamAnalyticsData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.TeamAnalyticsData{
			SchoolID:        row.SchoolID,
			SchoolName:      row.SchoolName,
			TotalPoints:     row.TotalPoints,
			TotalInvestment: row.TotalInvestment,
			Appearances:     int(row.Appearances),
			TotalSeed:       int(row.TotalSeed),
		})
	}

	return out, nil
}

func (r *AnalyticsRepository) GetCalcuttaPredictedInvestment(ctx context.Context, calcuttaID string, strategyGenerationRunID *string) (*string, []ports.CalcuttaPredictedInvestmentData, error) {
	if strategyGenerationRunID != nil && *strategyGenerationRunID != "" {
		runID := *strategyGenerationRunID
		rows, err := r.q.GetCalcuttaPredictedInvestmentByStrategyGenerationRunID(ctx, sqlc.GetCalcuttaPredictedInvestmentByStrategyGenerationRunIDParams{
			CalcuttaID:              calcuttaID,
			StrategyGenerationRunID: runID,
		})
		if err != nil {
			return nil, nil, err
		}

		out := make([]ports.CalcuttaPredictedInvestmentData, 0, len(rows))
		for _, row := range rows {
			out = append(out, ports.CalcuttaPredictedInvestmentData{
				TeamID:     row.TeamID,
				SchoolName: row.SchoolName,
				Seed:       int(row.Seed),
				Region:     row.Region,
				Rational:   row.Rational,
				Predicted:  row.Predicted,
				Delta:      row.Delta,
			})
		}

		return &runID, out, nil
	}

	latestID, err := r.q.GetLatestStrategyGenerationRunIDByCoreCalcuttaID(ctx, calcuttaID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, err
		}
	} else {
		rows, err := r.q.GetCalcuttaPredictedInvestmentByStrategyGenerationRunID(ctx, sqlc.GetCalcuttaPredictedInvestmentByStrategyGenerationRunIDParams{
			CalcuttaID:              calcuttaID,
			StrategyGenerationRunID: latestID,
		})
		if err != nil {
			return nil, nil, err
		}

		out := make([]ports.CalcuttaPredictedInvestmentData, 0, len(rows))
		for _, row := range rows {
			out = append(out, ports.CalcuttaPredictedInvestmentData{
				TeamID:     row.TeamID,
				SchoolName: row.SchoolName,
				Seed:       int(row.Seed),
				Region:     row.Region,
				Rational:   row.Rational,
				Predicted:  row.Predicted,
				Delta:      row.Delta,
			})
		}

		return &latestID, out, nil
	}

	rows, err := r.q.GetCalcuttaPredictedInvestment(ctx, calcuttaID)
	if err != nil {
		return nil, nil, err
	}

	out := make([]ports.CalcuttaPredictedInvestmentData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.CalcuttaPredictedInvestmentData{
			TeamID:     row.TeamID,
			SchoolName: row.SchoolName,
			Seed:       int(row.Seed),
			Region:     row.Region,
			Rational:   row.Rational,
			Predicted:  row.Predicted,
			Delta:      row.Delta,
		})
	}

	return nil, out, nil
}

func (r *AnalyticsRepository) GetCalcuttaPredictedReturns(ctx context.Context, calcuttaID string, strategyGenerationRunID *string) (*string, []ports.CalcuttaPredictedReturnsData, error) {
	if strategyGenerationRunID != nil && *strategyGenerationRunID != "" {
		runID := *strategyGenerationRunID
		rows, err := r.q.GetCalcuttaPredictedReturnsByStrategyGenerationRunID(ctx, sqlc.GetCalcuttaPredictedReturnsByStrategyGenerationRunIDParams{
			CalcuttaID:              calcuttaID,
			StrategyGenerationRunID: runID,
		})
		if err != nil {
			return nil, nil, err
		}

		out := make([]ports.CalcuttaPredictedReturnsData, 0, len(rows))
		for _, row := range rows {
			out = append(out, ports.CalcuttaPredictedReturnsData{
				TeamID:        row.TeamID,
				SchoolName:    row.SchoolName,
				Seed:          int(row.Seed),
				Region:        row.Region,
				ProbPI:        row.ProbPi,
				ProbR64:       row.ProbR64,
				ProbR32:       row.ProbR32,
				ProbS16:       row.ProbS16,
				ProbE8:        row.ProbE8,
				ProbFF:        row.ProbFf,
				ProbChamp:     row.ProbChamp,
				ExpectedValue: row.ExpectedValue,
			})
		}

		return &runID, out, nil
	}

	latestID, err := r.q.GetLatestStrategyGenerationRunIDByCoreCalcuttaID(ctx, calcuttaID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, err
		}
	} else {
		rows, err := r.q.GetCalcuttaPredictedReturnsByStrategyGenerationRunID(ctx, sqlc.GetCalcuttaPredictedReturnsByStrategyGenerationRunIDParams{
			CalcuttaID:              calcuttaID,
			StrategyGenerationRunID: latestID,
		})
		if err != nil {
			return nil, nil, err
		}

		out := make([]ports.CalcuttaPredictedReturnsData, 0, len(rows))
		for _, row := range rows {
			out = append(out, ports.CalcuttaPredictedReturnsData{
				TeamID:        row.TeamID,
				SchoolName:    row.SchoolName,
				Seed:          int(row.Seed),
				Region:        row.Region,
				ProbPI:        row.ProbPi,
				ProbR64:       row.ProbR64,
				ProbR32:       row.ProbR32,
				ProbS16:       row.ProbS16,
				ProbE8:        row.ProbE8,
				ProbFF:        row.ProbFf,
				ProbChamp:     row.ProbChamp,
				ExpectedValue: row.ExpectedValue,
			})
		}

		return &latestID, out, nil
	}

	rows, err := r.q.GetCalcuttaPredictedReturns(ctx, calcuttaID)
	if err != nil {
		return nil, nil, err
	}

	out := make([]ports.CalcuttaPredictedReturnsData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.CalcuttaPredictedReturnsData{
			TeamID:        row.TeamID,
			SchoolName:    row.SchoolName,
			Seed:          int(row.Seed),
			Region:        row.Region,
			ProbPI:        row.ProbPi,
			ProbR64:       row.ProbR64,
			ProbR32:       row.ProbR32,
			ProbS16:       row.ProbS16,
			ProbE8:        row.ProbE8,
			ProbFF:        row.ProbFf,
			ProbChamp:     row.ProbChamp,
			ExpectedValue: row.ExpectedValue,
		})
	}

	return nil, out, nil
}

func (r *AnalyticsRepository) GetCalcuttaSimulatedEntry(ctx context.Context, calcuttaID string, strategyGenerationRunID *string) (*string, []ports.CalcuttaSimulatedEntryData, error) {
	if strategyGenerationRunID != nil && *strategyGenerationRunID != "" {
		runID := *strategyGenerationRunID
		rows, err := r.q.GetCalcuttaSimulatedEntryByStrategyGenerationRunID(ctx, sqlc.GetCalcuttaSimulatedEntryByStrategyGenerationRunIDParams{
			StrategyGenerationRunID: runID,
			CalcuttaID:              calcuttaID,
		})
		if err != nil {
			return nil, nil, err
		}

		out := make([]ports.CalcuttaSimulatedEntryData, 0, len(rows))
		for _, row := range rows {
			out = append(out, ports.CalcuttaSimulatedEntryData{
				TeamID:         row.TeamID,
				SchoolName:     row.SchoolName,
				Seed:           int(row.Seed),
				Region:         row.Region,
				ExpectedPoints: row.ExpectedPoints,
				ExpectedMarket: row.ExpectedMarket,
				OurBid:         row.OurBid,
			})
		}

		return &runID, out, nil
	}

	// Prefer latest strategy generation run if it exists.
	latestID, err := r.q.GetLatestStrategyGenerationRunIDByCoreCalcuttaID(ctx, calcuttaID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, err
		}
	} else {
		rows, err := r.q.GetCalcuttaSimulatedEntryByStrategyGenerationRunID(ctx, sqlc.GetCalcuttaSimulatedEntryByStrategyGenerationRunIDParams{
			StrategyGenerationRunID: latestID,
			CalcuttaID:              calcuttaID,
		})
		if err != nil {
			return nil, nil, err
		}

		out := make([]ports.CalcuttaSimulatedEntryData, 0, len(rows))
		for _, row := range rows {
			out = append(out, ports.CalcuttaSimulatedEntryData{
				TeamID:         row.TeamID,
				SchoolName:     row.SchoolName,
				Seed:           int(row.Seed),
				Region:         row.Region,
				ExpectedPoints: row.ExpectedPoints,
				ExpectedMarket: row.ExpectedMarket,
				OurBid:         row.OurBid,
			})
		}

		return &latestID, out, nil
	}

	// Legacy fallback.
	rows, err := r.q.GetCalcuttaSimulatedEntry(ctx, calcuttaID)
	if err != nil {
		return nil, nil, err
	}

	out := make([]ports.CalcuttaSimulatedEntryData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.CalcuttaSimulatedEntryData{
			TeamID:         row.TeamID,
			SchoolName:     row.SchoolName,
			Seed:           int(row.Seed),
			Region:         row.Region,
			ExpectedPoints: row.ExpectedPoints,
			ExpectedMarket: row.ExpectedMarket,
			OurBid:         row.OurBid,
		})
	}

	return nil, out, nil
}

func (r *AnalyticsRepository) GetSeedVarianceAnalytics(ctx context.Context) ([]ports.SeedVarianceData, error) {
	rows, err := r.q.GetSeedVarianceAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]ports.SeedVarianceData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.SeedVarianceData{
			Seed:             int(row.Seed),
			InvestmentStdDev: row.InvestmentStddev,
			PointsStdDev:     row.PointsStddev,
			InvestmentMean:   row.InvestmentMean,
			PointsMean:       row.PointsMean,
			TeamCount:        int(row.TeamCount),
		})
	}
	return out, nil
}

func (r *AnalyticsRepository) GetSeedInvestmentPoints(ctx context.Context) ([]ports.SeedInvestmentPointData, error) {
	rows, err := r.q.GetSeedInvestmentPoints(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]ports.SeedInvestmentPointData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.SeedInvestmentPointData{
			Seed:             int(row.Seed),
			TournamentName:   row.TournamentName,
			TournamentYear:   int(row.TournamentYear),
			CalcuttaID:       row.CalcuttaID,
			TeamID:           row.TeamID,
			SchoolName:       row.SchoolName,
			TotalBid:         row.TotalBid,
			CalcuttaTotalBid: row.CalcuttaTotalBid,
			NormalizedBid:    row.NormalizedBid,
		})
	}

	return out, nil
}

func (r *AnalyticsRepository) GetBestInvestments(ctx context.Context, limit int) ([]ports.BestInvestmentData, error) {
	rows, err := r.q.GetBestInvestments(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	out := make([]ports.BestInvestmentData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.BestInvestmentData{
			TournamentName:   row.TournamentName,
			TournamentYear:   int(row.TournamentYear),
			CalcuttaID:       row.CalcuttaID,
			TeamID:           row.TeamID,
			SchoolName:       row.SchoolName,
			Seed:             int(row.Seed),
			Region:           row.Region,
			TeamPoints:       row.TeamPoints,
			TotalBid:         row.TotalBid,
			CalcuttaTotalBid: row.CalcuttaTotalBid,
			CalcuttaTotalPts: row.CalcuttaTotalPoints,
			InvestmentShare:  row.InvestmentShare,
			PointsShare:      row.PointsShare,
			RawROI:           row.RawRoi,
			NormalizedROI:    row.NormalizedRoi,
		})
	}

	return out, nil
}

func (r *AnalyticsRepository) GetBestInvestmentBids(ctx context.Context, limit int) ([]ports.InvestmentLeaderboardData, error) {
	rows, err := r.q.GetBestInvestmentBids(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	out := make([]ports.InvestmentLeaderboardData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.InvestmentLeaderboardData{
			TournamentName:      row.TournamentName,
			TournamentYear:      int(row.TournamentYear),
			CalcuttaID:          row.CalcuttaID,
			EntryID:             row.EntryID,
			EntryName:           row.EntryName,
			TeamID:              row.TeamID,
			SchoolName:          row.SchoolName,
			Seed:                int(row.Seed),
			Investment:          row.Investment,
			OwnershipPercentage: row.OwnershipPercentage,
			RawReturns:          row.RawReturns,
			NormalizedReturns:   row.NormalizedReturns,
		})
	}

	return out, nil
}

func (r *AnalyticsRepository) GetBestEntries(ctx context.Context, limit int) ([]ports.EntryLeaderboardData, error) {
	rows, err := r.q.GetBestEntries(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	out := make([]ports.EntryLeaderboardData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.EntryLeaderboardData{
			TournamentName:    row.TournamentName,
			TournamentYear:    int(row.TournamentYear),
			CalcuttaID:        row.CalcuttaID,
			EntryID:           row.EntryID,
			EntryName:         row.EntryName,
			TotalReturns:      row.TotalReturns,
			TotalParticipants: int(row.TotalParticipants),
			AverageReturns:    row.AverageReturns,
			NormalizedReturns: row.NormalizedReturns,
		})
	}

	return out, nil
}

func (r *AnalyticsRepository) GetBestCareers(ctx context.Context, limit int) ([]ports.CareerLeaderboardData, error) {
	rows, err := r.q.GetBestCareers(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	out := make([]ports.CareerLeaderboardData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.CareerLeaderboardData{
			EntryName:              row.EntryName,
			Years:                  int(row.Years),
			BestFinish:             int(row.BestFinish),
			Wins:                   int(row.Wins),
			Podiums:                int(row.Podiums),
			InTheMoneys:            int(row.InTheMoneys),
			Top10s:                 int(row.Top10s),
			CareerEarningsCents:    int(row.CareerEarningsCents),
			ActiveInLatestCalcutta: row.ActiveInLatestCalcutta,
		})
	}

	return out, nil
}
