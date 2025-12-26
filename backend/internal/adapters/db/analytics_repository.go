package db

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/pkg/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *AnalyticsRepository) GetSeedAnalytics(ctx context.Context) ([]services.SeedAnalyticsData, float64, float64, error) {
	rows, err := r.q.GetSeedAnalytics(ctx)
	if err != nil {
		return nil, 0, 0, err
	}

	out := make([]services.SeedAnalyticsData, 0, len(rows))
	var totalPoints float64
	var totalInvestment float64
	for _, row := range rows {
		out = append(out, services.SeedAnalyticsData{
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

func (r *AnalyticsRepository) GetRegionAnalytics(ctx context.Context) ([]services.RegionAnalyticsData, float64, float64, error) {
	rows, err := r.q.GetRegionAnalytics(ctx)
	if err != nil {
		return nil, 0, 0, err
	}

	out := make([]services.RegionAnalyticsData, 0, len(rows))
	var totalPoints float64
	var totalInvestment float64
	for _, row := range rows {
		out = append(out, services.RegionAnalyticsData{
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

func (r *AnalyticsRepository) GetTeamAnalytics(ctx context.Context) ([]services.TeamAnalyticsData, error) {
	rows, err := r.q.GetTeamAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]services.TeamAnalyticsData, 0, len(rows))
	for _, row := range rows {
		out = append(out, services.TeamAnalyticsData{
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

func (r *AnalyticsRepository) GetSeedVarianceAnalytics(ctx context.Context) ([]services.SeedVarianceData, error) {
	rows, err := r.q.GetSeedVarianceAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]services.SeedVarianceData, 0, len(rows))
	for _, row := range rows {
		out = append(out, services.SeedVarianceData{
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

func (r *AnalyticsRepository) GetSeedInvestmentPoints(ctx context.Context) ([]services.SeedInvestmentPointData, error) {
	rows, err := r.q.GetSeedInvestmentPoints(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]services.SeedInvestmentPointData, 0, len(rows))
	for _, row := range rows {
		out = append(out, services.SeedInvestmentPointData{
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

func (r *AnalyticsRepository) GetBestInvestments(ctx context.Context, limit int) ([]services.BestInvestmentData, error) {
	rows, err := r.q.GetBestInvestments(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	out := make([]services.BestInvestmentData, 0, len(rows))
	for _, row := range rows {
		out = append(out, services.BestInvestmentData{
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

func (r *AnalyticsRepository) GetBestInvestmentBids(ctx context.Context, limit int) ([]services.InvestmentLeaderboardData, error) {
	rows, err := r.q.GetBestInvestmentBids(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	out := make([]services.InvestmentLeaderboardData, 0, len(rows))
	for _, row := range rows {
		out = append(out, services.InvestmentLeaderboardData{
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

func (r *AnalyticsRepository) GetBestEntries(ctx context.Context, limit int) ([]services.EntryLeaderboardData, error) {
	rows, err := r.q.GetBestEntries(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	out := make([]services.EntryLeaderboardData, 0, len(rows))
	for _, row := range rows {
		out = append(out, services.EntryLeaderboardData{
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

func (r *AnalyticsRepository) GetBestCareers(ctx context.Context, limit int) ([]services.CareerLeaderboardData, error) {
	rows, err := r.q.GetBestCareers(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	out := make([]services.CareerLeaderboardData, 0, len(rows))
	for _, row := range rows {
		out = append(out, services.CareerLeaderboardData{
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
