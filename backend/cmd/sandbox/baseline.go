package main

import (
	"context"
	"database/sql"
	"fmt"
)

func runSeedBaseline(ctx context.Context, db *sql.DB, targetCalcuttaID string, trainYears int, excludeEntryName string) ([]BaselineRow, *BaselineSummary, error) {
	targetRows, err := queryTeamDataset(ctx, db, 0, targetCalcuttaID, excludeEntryName)
	if err != nil {
		return nil, nil, err
	}

	targetYear, err := calcuttaYear(ctx, db, targetCalcuttaID)
	if err != nil {
		return nil, nil, err
	}

	maxYear := targetYear - 1
	minYear := 0
	if trainYears > 0 {
		minYear = targetYear - trainYears
	}
	if trainYears > 0 && maxYear < minYear {
		return nil, nil, fmt.Errorf("invalid training window: target_year=%d train_years=%d", targetYear, trainYears)
	}

	seedPointsMean, seedBidShareMean, err := computeSeedMeans(ctx, db, targetCalcuttaID, trainYears, minYear, maxYear, excludeEntryName)
	if err != nil {
		return nil, nil, err
	}
	if len(seedPointsMean) == 0 {
		return nil, nil, fmt.Errorf("no training data found for baseline: target_year=%d train_years=%d", targetYear, trainYears)
	}

	var totalActualPoints float64
	var totalPredPoints float64
	for _, r := range targetRows {
		totalActualPoints += r.TeamPoints
		totalPredPoints += seedPointsMean[r.Seed]
	}

	baselineRows := make([]BaselineRow, 0, len(targetRows))
	var absPointsErrSum float64
	var absBidShareErrSum float64

	for _, r := range targetRows {
		predPoints := seedPointsMean[r.Seed]
		predBidShare := seedBidShareMean[r.Seed]

		actualBidShareUsed := r.NormalizedBid
		if excludeEntryName != "" {
			actualBidShareUsed = r.NormalizedBidExcl
		}

		var actualPointsShare float64
		if totalActualPoints > 0 {
			actualPointsShare = r.TeamPoints / totalActualPoints
		}
		var predPointsShare float64
		if totalPredPoints > 0 {
			predPointsShare = predPoints / totalPredPoints
		}

		var actualROI float64
		if actualBidShareUsed > 0 {
			actualROI = actualPointsShare / actualBidShareUsed
		}
		var predROI float64
		if predBidShare > 0 {
			predROI = predPointsShare / predBidShare
		}

		absPointsErrSum += absFloat64(predPoints - r.TeamPoints)
		absBidShareErrSum += absFloat64(predBidShare - actualBidShareUsed)

		baselineRows = append(baselineRows, BaselineRow{
			TournamentName:     r.TournamentName,
			TournamentYear:     r.TournamentYear,
			CalcuttaID:         r.CalcuttaID,
			TeamID:             r.TeamID,
			SchoolName:         r.SchoolName,
			Seed:               r.Seed,
			Region:             r.Region,
			ActualPoints:       r.TeamPoints,
			PredPoints:         predPoints,
			ActualBidShare:     r.NormalizedBid,
			ActualBidShareExcl: r.NormalizedBidExcl,
			PredBidShare:       predBidShare,
			ActualROI:          actualROI,
			PredROI:            predROI,
		})
	}

	summary := &BaselineSummary{}
	if len(targetRows) > 0 {
		summary.PointsMAE = absPointsErrSum / float64(len(targetRows))
		summary.BidShareMAE = absBidShareErrSum / float64(len(targetRows))
	}

	return baselineRows, summary, nil
}
