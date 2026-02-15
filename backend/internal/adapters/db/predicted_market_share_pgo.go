package db

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

func computeCalcuttaPredictedMarketShareFromPGO(
	ctx context.Context,
	pool *pgxpool.Pool,
	bracketBuilder ports.BracketBuilder,
	calcuttaID string,
	marketShareRunID *string,
	gameOutcomeRunID *string,
) (*string, *string, []ports.CalcuttaPredictedMarketShareData, error) {
	if calcuttaID == "" {
		return nil, nil, nil, errors.New("calcuttaID is required")
	}

	coreTournamentID, _, err := loadTournamentForCalcutta(ctx, pool, calcuttaID)
	if err != nil {
		return nil, nil, nil, err
	}

	selectedMarketShareRunID, pmsByTeamID, err := loadPredictedMarketShareForCalcutta(ctx, pool, calcuttaID, coreTournamentID, marketShareRunID)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(pmsByTeamID) == 0 {
		return selectedMarketShareRunID, nil, nil, fmt.Errorf("no predicted_market_share found for calcutta_id=%s", calcuttaID)
	}

	selectedGameOutcomeRunID, returns, err := computeCalcuttaPredictedReturnsFromPGO(ctx, pool, bracketBuilder, calcuttaID, gameOutcomeRunID)
	if err != nil {
		return selectedMarketShareRunID, nil, nil, err
	}
	if len(returns) == 0 {
		return selectedMarketShareRunID, selectedGameOutcomeRunID, nil, errors.New("no predicted returns computed")
	}

	totalEV := 0.0
	for _, r := range returns {
		if r.ExpectedValue > 0 {
			totalEV += r.ExpectedValue
		}
	}

	out := make([]ports.CalcuttaPredictedMarketShareData, 0, len(returns))
	for _, r := range returns {
		predictedShare := pmsByTeamID[r.TeamID]

		rationalShare := 0.0
		if totalEV > 0 && r.ExpectedValue > 0 {
			rationalShare = r.ExpectedValue / totalEV
		}

		deltaPercent := 0.0
		if rationalShare > 0 {
			deltaPercent = ((predictedShare / rationalShare) - 1.0) * 100.0
		}

		out = append(out, ports.CalcuttaPredictedMarketShareData{
			TeamID:         r.TeamID,
			SchoolName:     r.SchoolName,
			Seed:           r.Seed,
			Region:         r.Region,
			RationalShare:  rationalShare,
			PredictedShare: predictedShare,
			DeltaPercent:   deltaPercent,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].PredictedShare != out[j].PredictedShare {
			return out[i].PredictedShare > out[j].PredictedShare
		}
		if out[i].Seed != out[j].Seed {
			return out[i].Seed < out[j].Seed
		}
		return out[i].SchoolName < out[j].SchoolName
	})

	return selectedMarketShareRunID, selectedGameOutcomeRunID, out, nil
}
