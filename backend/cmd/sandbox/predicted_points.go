package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/pkg/services"
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

func predictedMarketBidsByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, excludeEntryName string) (map[string]float64, map[string]float64, float64, error) {
	targetYear, err := calcuttaYear(ctx, db, targetCalcuttaID)
	if err != nil {
		return nil, nil, 0, err
	}

	maxYear := targetYear - 1
	minYear := 0
	if trainYears > 0 {
		minYear = targetYear - trainYears
	}
	if trainYears > 0 && maxYear < minYear {
		return nil, nil, 0, fmt.Errorf("invalid training window: target_year=%d train_years=%d", targetYear, trainYears)
	}

	_, seedBidShareMean, err := computeSeedMeans(ctx, db, targetCalcuttaID, trainYears, minYear, maxYear, excludeEntryName)
	if err != nil {
		return nil, nil, 0, err
	}
	if len(seedBidShareMean) == 0 {
		return nil, nil, 0, fmt.Errorf("%w: market bid model target_year=%d train_years=%d", ErrNoTrainingData, targetYear, trainYears)
	}

	predTotalMarketBid, err := computeMeanCalcuttaTotalBid(ctx, db, targetCalcuttaID, trainYears, minYear, maxYear, excludeEntryName)
	if err != nil {
		return nil, nil, 0, err
	}

	predBidByTeam := make(map[string]float64, len(targetRows))
	predBidShareByTeam := make(map[string]float64, len(targetRows))
	for _, r := range targetRows {
		share := seedBidShareMean[r.Seed]
		predBidShareByTeam[r.TeamID] = share
		predBidByTeam[r.TeamID] = predTotalMarketBid * share
	}

	return predBidByTeam, predBidShareByTeam, predTotalMarketBid, nil
}

func predictedPointsByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, predModel string, sigma float64) (map[string]float64, error) {
	switch predModel {
	case "seed":
		targetYear, err := calcuttaYear(ctx, db, targetCalcuttaID)
		if err != nil {
			return nil, err
		}

		maxYear := targetYear - 1
		minYear := 0
		if trainYears > 0 {
			minYear = targetYear - trainYears
		}
		if trainYears > 0 && maxYear < minYear {
			return nil, fmt.Errorf("invalid training window: target_year=%d train_years=%d", targetYear, trainYears)
		}

		seedPointsMean, _, err := computeSeedMeans(ctx, db, targetCalcuttaID, trainYears, minYear, maxYear, "")
		if err != nil {
			return nil, err
		}
		if len(seedPointsMean) == 0 {
			return nil, fmt.Errorf("%w: seed model target_year=%d train_years=%d", ErrNoTrainingData, targetYear, trainYears)
		}

		out := make(map[string]float64, len(targetRows))
		for _, r := range targetRows {
			out[r.TeamID] = seedPointsMean[r.Seed]
		}
		return out, nil

	case "kenpom":
		meta, err := tournamentMetaForCalcuttaID(ctx, db, targetCalcuttaID)
		if err != nil {
			return nil, err
		}

		repo := services.NewTournamentRepository(db)
		bracketService := services.NewBracketService(repo)
		returnsService := services.NewKenPomPredictedReturnsService(bracketService).WithSigma(sigma)

		pred, err := returnsService.GetPredictedReturnsPreTournament(ctx, meta.TournamentID)
		if err != nil {
			return nil, err
		}

		out := make(map[string]float64, len(pred))
		for _, p := range pred {
			out[p.TeamID] = p.ExpectedPoints
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unknown pred-model %q (expected seed|kenpom)", predModel)
	}
}
