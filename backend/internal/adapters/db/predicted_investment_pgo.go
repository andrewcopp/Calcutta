package db

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

func computeCalcuttaPredictedInvestmentFromPGO(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) ([]ports.CalcuttaPredictedInvestmentData, error) {
	if calcuttaID == "" {
		return nil, errors.New("calcuttaID is required")
	}

	coreTournamentID, budgetPoints, err := loadTournamentAndBudgetForCalcutta(ctx, pool, calcuttaID)
	if err != nil {
		return nil, err
	}

	poolSize, err := loadCalcuttaPoolSizePoints(ctx, pool, calcuttaID, budgetPoints)
	if err != nil {
		return nil, err
	}

	pmsByTeamID, err := loadTournamentPredictedMarketShare(ctx, pool, coreTournamentID)
	if err != nil {
		return nil, err
	}
	if len(pmsByTeamID) == 0 {
		return nil, fmt.Errorf("no predicted_market_share found for tournament_id=%s", coreTournamentID)
	}

	returns, err := computeCalcuttaPredictedReturnsFromPGO(ctx, pool, calcuttaID)
	if err != nil {
		return nil, err
	}
	if len(returns) == 0 {
		return nil, errors.New("no predicted returns computed")
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

	return out, nil
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

func loadTournamentPredictedMarketShare(ctx context.Context, pool *pgxpool.Pool, tournamentID string) (map[string]float64, error) {
	rows, err := pool.Query(ctx, `
		SELECT team_id, predicted_share
		FROM derived.predicted_market_share
		WHERE tournament_id = $1::uuid
			AND calcutta_id IS NULL
			AND run_id IS NULL
			AND deleted_at IS NULL
	`, tournamentID)
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
