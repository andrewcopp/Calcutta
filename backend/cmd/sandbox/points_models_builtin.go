package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/pkg/services"
)

type seedPointsModel struct{}

func (m seedPointsModel) Name() string { return "seed" }

func (m seedPointsModel) PredictPointsByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, sigma float64) (map[string]float64, error) {
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
}

type kenPomPointsModel struct{}

func (m kenPomPointsModel) Name() string { return "kenpom" }

func (m kenPomPointsModel) PredictPointsByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, sigma float64) (map[string]float64, error) {
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
}

func init() {
	RegisterPointsModel(seedPointsModel{})
	RegisterPointsModel(kenPomPointsModel{})
}
