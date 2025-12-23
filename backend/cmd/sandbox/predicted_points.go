package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func computeMeanCalcuttaTotalBid(ctx context.Context, db *sql.DB, excludeCalcuttaID string, trainYears int, minYear int, maxYear int, excludeEntryName string) (float64, error) {
	query := `
		WITH calcutta_totals AS (
			SELECT
				c.id as calcutta_id,
				COALESCE(substring(t.name from '([0-9]{4})')::int, 0) as tournament_year,
				COALESCE(SUM(
					CASE
						WHEN $5 <> '' AND ce.name = $5 THEN 0
						ELSE COALESCE(cet.bid, 0)
					END
				), 0)::float as total_bid
			FROM calcuttas c
			JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
			LEFT JOIN calcutta_entries ce ON ce.calcutta_id = c.id AND ce.deleted_at IS NULL
			LEFT JOIN calcutta_entry_teams cet ON cet.entry_id = ce.id AND cet.deleted_at IS NULL
			WHERE c.deleted_at IS NULL
				AND c.id <> $1::uuid
				AND (
					COALESCE(substring(t.name from '([0-9]{4})')::int, 0) <> 0
					AND COALESCE(substring(t.name from '([0-9]{4})')::int, 0) <= $4
					AND ($2 = 0 OR COALESCE(substring(t.name from '([0-9]{4})')::int, 0) >= $3)
				)
			GROUP BY c.id, tournament_year
		)
		SELECT AVG(total_bid)::float
		FROM calcutta_totals
		WHERE tournament_year <> 0
	`

	var mean sql.NullFloat64
	if err := db.QueryRowContext(ctx, query, excludeCalcuttaID, trainYears, minYear, maxYear, excludeEntryName).Scan(&mean); err != nil {
		return 0, err
	}
	if !mean.Valid {
		return 0, fmt.Errorf("%w: mean calcutta total bid", ErrNoTrainingData)
	}
	return mean.Float64, nil
}

func predictedMarketBidsByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, investModel string, excludeEntryName string) (map[string]float64, map[string]float64, float64, error) {
	targetYear, err := calcuttaYear(ctx, db, targetCalcuttaID)
	if err != nil {
		return nil, nil, 0, err
	}
	if investModel == "" {
		investModel = "seed"
	}

	maxYear := targetYear - 1
	minYear := 0
	if trainYears > 0 {
		minYear = targetYear - trainYears
	}
	if trainYears > 0 && maxYear < minYear {
		return nil, nil, 0, fmt.Errorf("invalid training window: target_year=%d train_years=%d", targetYear, trainYears)
	}

	model, err := GetInvestmentModel(investModel)
	if err != nil {
		allowed := strings.Join(ListInvestmentModelNames(), "|")
		return nil, nil, 0, fmt.Errorf("unknown invest-model %q (expected %s)", investModel, allowed)
	}
	if model == nil {
		allowed := strings.Join(ListInvestmentModelNames(), "|")
		return nil, nil, 0, fmt.Errorf("unknown invest-model %q (expected %s)", investModel, allowed)
	}

	predScoreByTeam, err := model.PredictBidShareByTeam(ctx, db, targetCalcuttaID, targetRows, trainYears, excludeEntryName)
	if err != nil {
		return nil, nil, 0, err
	}

	predTotalMarketBid, err := computeMeanCalcuttaTotalBid(ctx, db, targetCalcuttaID, trainYears, minYear, maxYear, excludeEntryName)
	if err != nil {
		return nil, nil, 0, err
	}

	teamIDs := make([]string, 0, len(targetRows))
	for _, r := range targetRows {
		teamIDs = append(teamIDs, r.TeamID)
	}
	shares, _, normErr := normalizeNonNegativeScoresForTeams(teamIDs, predScoreByTeam)
	if normErr != nil {
		return nil, nil, 0, fmt.Errorf("%w: market bid model target_year=%d train_years=%d invest_model=%s", ErrNoTrainingData, targetYear, trainYears, investModel)
	}

	predBidByTeam := make(map[string]float64, len(targetRows))
	predBidShareByTeam := make(map[string]float64, len(targetRows))
	for _, r := range targetRows {
		share := shares[r.TeamID]
		predBidShareByTeam[r.TeamID] = share
		predBidByTeam[r.TeamID] = predTotalMarketBid * share
	}

	return predBidByTeam, predBidShareByTeam, predTotalMarketBid, nil
}

func predictedPointsByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, predModel string, sigma float64) (map[string]float64, error) {
	model, err := GetPointsModel(predModel)
	if err != nil {
		allowed := strings.Join(ListPointsModelNames(), "|")
		return nil, fmt.Errorf("unknown pred-model %q (expected %s)", predModel, allowed)
	}
	if model == nil {
		allowed := strings.Join(ListPointsModelNames(), "|")
		return nil, fmt.Errorf("unknown pred-model %q (expected %s)", predModel, allowed)
	}
	return model.PredictPointsByTeam(ctx, db, targetCalcuttaID, targetRows, trainYears, sigma)
}
