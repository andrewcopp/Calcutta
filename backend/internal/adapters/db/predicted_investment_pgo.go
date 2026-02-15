package db

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

func computeCalcuttaPredictedInvestmentFromPGO(
	ctx context.Context,
	pool *pgxpool.Pool,
	bracketBuilder ports.BracketBuilder,
	calcuttaID string,
	marketShareRunID *string,
	gameOutcomeRunID *string,
) (*string, []ports.CalcuttaPredictedInvestmentData, error) {
	if calcuttaID == "" {
		return nil, nil, errors.New("calcuttaID is required")
	}
	if marketShareRunID == nil || *marketShareRunID == "" {
		return nil, nil, errors.New("market_share_run_id is required")
	}
	if gameOutcomeRunID == nil || *gameOutcomeRunID == "" {
		return nil, nil, errors.New("game_outcome_run_id is required")
	}

	coreTournamentID, budgetPoints, err := loadTournamentAndBudgetForCalcutta(ctx, pool, calcuttaID)
	if err != nil {
		return nil, nil, err
	}

	poolSize, err := loadCalcuttaPoolSizePoints(ctx, pool, calcuttaID, budgetPoints)
	if err != nil {
		return nil, nil, err
	}

	selectedMarketShareRunID, pmsByTeamID, err := loadPredictedMarketShareForCalcutta(ctx, pool, calcuttaID, coreTournamentID, marketShareRunID)
	if err != nil {
		return nil, nil, err
	}
	if len(pmsByTeamID) == 0 {
		return selectedMarketShareRunID, nil, fmt.Errorf("no predicted_market_share found for calcutta_id=%s", calcuttaID)
	}

	_, returns, err := computeCalcuttaPredictedReturnsFromPGO(ctx, pool, bracketBuilder, calcuttaID, gameOutcomeRunID)
	if err != nil {
		return selectedMarketShareRunID, nil, err
	}
	if len(returns) == 0 {
		return selectedMarketShareRunID, nil, errors.New("no predicted returns computed")
	}

	// Rational: normalize expected_value (points) to pool size.
	totalEV := 0.0
	for _, r := range returns {
		if r.ExpectedValue > 0 {
			totalEV += r.ExpectedValue
		}
	}
	if totalEV <= 0 {
		// If EV is degenerate, still return predicted market and zero rational.
		totalEV = 0
	}

	out := make([]ports.CalcuttaPredictedInvestmentData, 0, len(returns))
	for _, r := range returns {
		predictedShare := pmsByTeamID[r.TeamID]
		predicted := predictedShare * poolSize

		rational := computeRationalInvestment(r.ExpectedValue, totalEV, poolSize)
		delta := computeDeltaPercent(predicted, rational)

		out = append(out, ports.CalcuttaPredictedInvestmentData{
			TeamID:     r.TeamID,
			SchoolName: r.SchoolName,
			Seed:       r.Seed,
			Region:     r.Region,
			Rational:   rational,
			Predicted:  predicted,
			Delta:      delta,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Predicted != out[j].Predicted {
			return out[i].Predicted > out[j].Predicted
		}
		if out[i].Seed != out[j].Seed {
			return out[i].Seed < out[j].Seed
		}
		return out[i].SchoolName < out[j].SchoolName
	})

	return selectedMarketShareRunID, out, nil
}

func computeRationalInvestment(expectedValue float64, totalExpectedValue float64, poolSize float64) float64 {
	if poolSize <= 0 {
		return 0.0
	}
	if totalExpectedValue <= 0 {
		return 0.0
	}
	if expectedValue <= 0 {
		return 0.0
	}
	return (expectedValue / totalExpectedValue) * poolSize
}

func computeDeltaPercent(predicted float64, rational float64) float64 {
	if rational <= 0 {
		return 0.0
	}
	return ((predicted - rational) / rational) * 100.0
}

func loadTournamentAndBudgetForCalcutta(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) (string, float64, error) {
	var tournamentID string
	var budgetPoints *float64
	if err := pool.QueryRow(ctx, `
		SELECT c.tournament_id, c.budget_points
		FROM core.calcuttas c
		WHERE c.id = $1::uuid
			AND c.deleted_at IS NULL
		LIMIT 1
	`, calcuttaID).Scan(&tournamentID, &budgetPoints); err != nil {
		return "", 0, err
	}

	budget := 100.0
	if budgetPoints != nil && *budgetPoints > 0 {
		budget = *budgetPoints
	}
	return tournamentID, budget, nil
}

func loadCalcuttaPoolSizePoints(ctx context.Context, pool *pgxpool.Pool, calcuttaID string, budgetPoints float64) (float64, error) {
	var nEntries int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM core.entries e
		WHERE e.calcutta_id = $1::uuid
			AND e.deleted_at IS NULL
	`, calcuttaID).Scan(&nEntries); err != nil {
		return 0, err
	}

	entries := float64(nEntries)
	if entries <= 0 {
		entries = 47
	}
	if budgetPoints <= 0 {
		budgetPoints = 100
	}
	return entries * budgetPoints, nil
}

func loadPredictedMarketShareForCalcutta(
	ctx context.Context,
	pool *pgxpool.Pool,
	calcuttaID string,
	tournamentID string,
	marketShareRunID *string,
) (*string, map[string]float64, error) {
	if marketShareRunID == nil || *marketShareRunID == "" {
		return nil, nil, errors.New("market_share_run_id is required")
	}

	out, err := loadPredictedMarketShareByRunID(ctx, pool, *marketShareRunID)
	if err != nil {
		return marketShareRunID, nil, err
	}
	if len(out) == 0 {
		return marketShareRunID, nil, fmt.Errorf("no predicted_market_share rows for run_id=%s", *marketShareRunID)
	}
	return marketShareRunID, out, nil
}

func loadPredictedMarketShareByRunID(ctx context.Context, pool *pgxpool.Pool, runID string) (map[string]float64, error) {
	rows, err := pool.Query(ctx, `
		SELECT team_id, predicted_share
		FROM derived.predicted_market_share
		WHERE run_id = $1::uuid
			AND deleted_at IS NULL
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]float64)
	for rows.Next() {
		var teamID string
		var share float64
		if err := rows.Scan(&teamID, &share); err != nil {
			return nil, err
		}
		out[teamID] = share
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}
