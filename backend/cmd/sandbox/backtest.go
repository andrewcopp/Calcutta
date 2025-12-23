package main

import (
	"context"
	"database/sql"
	"errors"
)

func runBacktest(ctx context.Context, db *sql.DB, startYear int, endYear int, trainYears int, excludeEntryName string, budget int, minTeams int, maxTeams int, minBid int, maxBid int, predModel string, investModel string, sigma float64) ([]BacktestRow, error) {
	rows := make([]BacktestRow, 0)
	for y := startYear; y <= endYear; y++ {
		calcuttaID, err := resolveSingleCalcuttaIDForYear(ctx, db, y)
		if err != nil {
			if errors.Is(err, ErrNoCalcuttaForYear) {
				continue
			}
			return nil, err
		}

		simRows, simSummary, err := runSimulateEntry(ctx, db, calcuttaID, trainYears, excludeEntryName, budget, minTeams, maxTeams, minBid, maxBid, predModel, investModel, sigma)
		if err != nil {
			if errors.Is(err, ErrNoTrainingData) {
				continue
			}
			return nil, err
		}

		datasetRows, err := queryTeamDataset(ctx, db, 0, calcuttaID, excludeEntryName)
		if err != nil {
			return nil, err
		}

		teamPointsByID := map[string]float64{}
		actualBaseBidByID := map[string]float64{}
		totalActualPoints := 0.0
		for _, r := range datasetRows {
			teamPointsByID[r.TeamID] = r.TeamPoints
			if excludeEntryName != "" {
				actualBaseBidByID[r.TeamID] = r.TotalCommunityBidExcl
			} else {
				actualBaseBidByID[r.TeamID] = r.TotalCommunityBid
			}
			totalActualPoints += r.TeamPoints
		}

		actualTotalMarketBid := 0.0
		if len(datasetRows) > 0 {
			if excludeEntryName != "" {
				actualTotalMarketBid = datasetRows[0].CalcuttaTotalExcl
			} else {
				actualTotalMarketBid = datasetRows[0].CalcuttaTotalCommunity
			}
		}

		realizedEntryPointsTotal := 0.0
		for _, r := range simRows {
			teamPoints := teamPointsByID[r.TeamID]
			actualBaseBid := actualBaseBidByID[r.TeamID]
			own := 0.0
			if r.RecommendedBid > 0 {
				own = float64(r.RecommendedBid) / (actualBaseBid + float64(r.RecommendedBid))
			}
			realizedEntryPointsTotal += teamPoints * own
		}

		realizedPointsShare := 0.0
		if totalActualPoints > 0 {
			realizedPointsShare = realizedEntryPointsTotal / totalActualPoints
		}
		realizedBidShare := 0.0
		if actualTotalMarketBid+float64(budget) > 0 {
			realizedBidShare = float64(budget) / (actualTotalMarketBid + float64(budget))
		}
		realizedNormROI := 0.0
		if realizedBidShare > 0 {
			realizedNormROI = realizedPointsShare / realizedBidShare
		}

		rows = append(rows, BacktestRow{
			TournamentYear:        y,
			CalcuttaID:            calcuttaID,
			TrainYears:            trainYears,
			ExcludeEntryName:      excludeEntryName,
			PredModel:             predModel,
			InvestModel:           investModel,
			Sigma:                 sigma,
			NumTeams:              simSummary.NumTeams,
			Budget:                simSummary.Budget,
			ExpectedPointsShare:   simSummary.ExpectedPointsShare,
			ExpectedBidShare:      simSummary.ExpectedBidShare,
			ExpectedNormalizedROI: simSummary.ExpectedNormalizedROI,
			RealizedPointsShare:   realizedPointsShare,
			RealizedBidShare:      realizedBidShare,
			RealizedNormalizedROI: realizedNormROI,
		})
	}

	return rows, nil
}
