package main

import (
	"context"
	"database/sql"
	"fmt"
)

type investmentTrainingContext struct {
	targetYear       int
	minYear          int
	maxYear          int
	seedBidShareMean map[int]float64
}

func buildInvestmentTrainingContext(ctx context.Context, db *sql.DB, targetCalcuttaID string, trainYears int, excludeEntryName string) (*investmentTrainingContext, error) {
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

	seedPointsMean, seedBidShareMean, err := computeSeedMeans(ctx, db, targetCalcuttaID, trainYears, minYear, maxYear, excludeEntryName)
	if err != nil {
		return nil, err
	}
	if len(seedPointsMean) == 0 {
		return nil, fmt.Errorf("%w: investment model target_year=%d train_years=%d", ErrNoTrainingData, targetYear, trainYears)
	}

	return &investmentTrainingContext{
		targetYear:       targetYear,
		minYear:          minYear,
		maxYear:          maxYear,
		seedBidShareMean: seedBidShareMean,
	}, nil
}

func (c *investmentTrainingContext) seedFallback(seed int) float64 {
	if c == nil {
		return 0
	}
	if v, ok := c.seedBidShareMean[seed]; ok {
		return v
	}
	return 0
}

type seedInvestmentModel struct{}

func (m seedInvestmentModel) Name() string { return "seed" }

func (m seedInvestmentModel) PredictBidShareByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, excludeEntryName string) (map[string]float64, error) {
	c, err := buildInvestmentTrainingContext(ctx, db, targetCalcuttaID, trainYears, excludeEntryName)
	if err != nil {
		return nil, err
	}
	out := make(map[string]float64, len(targetRows))
	for _, r := range targetRows {
		out[r.TeamID] = c.seedFallback(r.Seed)
	}
	return out, nil
}

type ridgeInvestmentModel struct{}

func (m ridgeInvestmentModel) Name() string { return "ridge" }

func (m ridgeInvestmentModel) PredictBidShareByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, excludeEntryName string) (map[string]float64, error) {
	c, err := buildInvestmentTrainingContext(ctx, db, targetCalcuttaID, trainYears, excludeEntryName)
	if err != nil {
		return nil, err
	}

	train, err := queryTrainingBidShares(ctx, db, targetCalcuttaID, trainYears, c.minYear, c.maxYear, excludeEntryName)
	if err != nil {
		return nil, err
	}
	model, ok := fitRidgeBidShareModel(train)
	if !ok {
		return nil, fmt.Errorf("insufficient training data for ridge investment model")
	}

	out := make(map[string]float64, len(targetRows))
	for _, r := range targetRows {
		pred := ridgePredict(model.Beta, model.SeedNetRtgMean, model.SeedNetRtgStd, r.Seed, r.KenPomNetRtg)
		out[r.TeamID] = clamp01(pred)
	}
	return out, nil
}

type ridgeReturnsInvestmentModel struct{}

func (m ridgeReturnsInvestmentModel) Name() string { return "ridge-returns" }

func (m ridgeReturnsInvestmentModel) PredictBidShareByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, excludeEntryName string) (map[string]float64, error) {
	c, err := buildInvestmentTrainingContext(ctx, db, targetCalcuttaID, trainYears, excludeEntryName)
	if err != nil {
		return nil, err
	}

	sigma := 11.0
	train, err := queryTrainingBidShares(ctx, db, targetCalcuttaID, trainYears, c.minYear, c.maxYear, excludeEntryName)
	if err != nil {
		return nil, err
	}

	byCalcutta := map[string]map[string]float64{}
	for _, tr := range train {
		if _, ok := byCalcutta[tr.CalcuttaID]; ok {
			continue
		}
		m, err := returnsShareByTeamForCalcutta(ctx, db, tr.CalcuttaID, sigma)
		if err != nil {
			return nil, err
		}
		byCalcutta[tr.CalcuttaID] = m
	}

	trainRows := make([]trainReturnsRow, 0, len(train))
	for _, tr := range train {
		expShare := 0.0
		if m, ok := byCalcutta[tr.CalcuttaID]; ok {
			if v, ok2 := m[tr.TeamID]; ok2 {
				expShare = v
			}
		}
		baseline := c.seedBidShareMean[tr.Seed]
		trainRows = append(trainRows, trainReturnsRow{
			CalcuttaID:     tr.CalcuttaID,
			TeamID:         tr.TeamID,
			Seed:           tr.Seed,
			KenPomNetRtg:   tr.KenPomNetRtg,
			BidShare:       tr.BidShare,
			SeedBaseline:   baseline,
			ExpPointsShare: expShare,
			ROIFeature:     ridgeReturnsROIFeature(expShare, baseline),
		})
	}

	model, ok := fitRidgeReturnsBidShareModel(trainRows)
	if !ok {
		return nil, fmt.Errorf("insufficient training data for ridge-returns investment model")
	}

	targetReturnsShare, err := returnsShareByTeamForCalcutta(ctx, db, targetCalcuttaID, sigma)
	if err != nil {
		return nil, err
	}

	out := make(map[string]float64, len(targetRows))
	for _, r := range targetRows {
		expShare := targetReturnsShare[r.TeamID]
		baseline := c.seedBidShareMean[r.Seed]
		roiFeat := ridgeReturnsROIFeature(expShare, baseline)
		predResid := ridgeReturnsPredict(model.Beta, model.SeedNetRtgMean, model.SeedNetRtgStd, r.Seed, r.KenPomNetRtg, expShare, roiFeat)
		pred := baseline + predResid
		out[r.TeamID] = clamp01(pred)
	}
	return out, nil
}

type seedPodInvestmentModel struct{}

func (m seedPodInvestmentModel) Name() string { return "seed-pod" }

func (m seedPodInvestmentModel) PredictBidShareByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, excludeEntryName string) (map[string]float64, error) {
	c, err := buildInvestmentTrainingContext(ctx, db, targetCalcuttaID, trainYears, excludeEntryName)
	if err != nil {
		return nil, err
	}

	train, err := queryTrainingBidShares(ctx, db, targetCalcuttaID, trainYears, c.minYear, c.maxYear, excludeEntryName)
	if err != nil {
		return nil, err
	}
	means := seedPodRankMeans(train)
	podRankByTeam := seedPodRankByTeam(targetRows)

	out := make(map[string]float64, len(targetRows))
	for _, r := range targetRows {
		rank := podRankByTeam[r.TeamID]
		if v, ok := means[seedIntKey{Seed: r.Seed, Val: rank}]; ok {
			out[r.TeamID] = v
		} else {
			out[r.TeamID] = c.seedFallback(r.Seed)
		}
	}
	return out, nil
}

type seedKenPomDeltaInvestmentModel struct{}

func (m seedKenPomDeltaInvestmentModel) Name() string { return "seed-kenpom-delta" }

func (m seedKenPomDeltaInvestmentModel) PredictBidShareByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, excludeEntryName string) (map[string]float64, error) {
	c, err := buildInvestmentTrainingContext(ctx, db, targetCalcuttaID, trainYears, excludeEntryName)
	if err != nil {
		return nil, err
	}

	train, err := queryTrainingBidShares(ctx, db, targetCalcuttaID, trainYears, c.minYear, c.maxYear, excludeEntryName)
	if err != nil {
		return nil, err
	}
	means := seedKenPomDeltaBinMeans(train)
	deltaBinByTeam := seedKenPomDeltaBinByTeam(targetRows)

	out := make(map[string]float64, len(targetRows))
	for _, r := range targetRows {
		bin := deltaBinByTeam[r.TeamID]
		if v, ok := means[seedIntKey{Seed: r.Seed, Val: bin}]; ok {
			out[r.TeamID] = v
		} else {
			out[r.TeamID] = c.seedFallback(r.Seed)
		}
	}
	return out, nil
}

type seedKenPomRankInvestmentModel struct{}

func (m seedKenPomRankInvestmentModel) Name() string { return "seed-kenpom-rank" }

func (m seedKenPomRankInvestmentModel) PredictBidShareByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, excludeEntryName string) (map[string]float64, error) {
	c, err := buildInvestmentTrainingContext(ctx, db, targetCalcuttaID, trainYears, excludeEntryName)
	if err != nil {
		return nil, err
	}

	train, err := queryTrainingBidShares(ctx, db, targetCalcuttaID, trainYears, c.minYear, c.maxYear, excludeEntryName)
	if err != nil {
		return nil, err
	}
	means := seedKenPomRankBucketMeans(train)
	bucketByTeam := kenPomRankBucketByTeam(targetRows)

	out := make(map[string]float64, len(targetRows))
	for _, r := range targetRows {
		b := bucketByTeam[r.TeamID]
		if v, ok := means[seedIntKey{Seed: r.Seed, Val: b}]; ok {
			out[r.TeamID] = v
		} else {
			out[r.TeamID] = c.seedFallback(r.Seed)
		}
	}
	return out, nil
}

type kenPomRankInvestmentModel struct{}

func (m kenPomRankInvestmentModel) Name() string { return "kenpom-rank" }

func (m kenPomRankInvestmentModel) PredictBidShareByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, excludeEntryName string) (map[string]float64, error) {
	c, err := buildInvestmentTrainingContext(ctx, db, targetCalcuttaID, trainYears, excludeEntryName)
	if err != nil {
		return nil, err
	}

	train, err := queryTrainingBidShares(ctx, db, targetCalcuttaID, trainYears, c.minYear, c.maxYear, excludeEntryName)
	if err != nil {
		return nil, err
	}
	bucketMeans := kenPomRankBucketMeans(train)
	bucketByTeam := kenPomRankBucketByTeam(targetRows)

	overallMean := 0.0
	if a, ok := bucketMeans[-1]; ok {
		overallMean = a
	}

	out := make(map[string]float64, len(targetRows))
	for _, r := range targetRows {
		b := bucketByTeam[r.TeamID]
		if v, ok := bucketMeans[b]; ok {
			out[r.TeamID] = v
		} else {
			out[r.TeamID] = overallMean
		}
	}
	return out, nil
}

type kenPomScoreInvestmentModel struct{}

func (m kenPomScoreInvestmentModel) Name() string { return "kenpom-score" }

func (m kenPomScoreInvestmentModel) PredictBidShareByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, excludeEntryName string) (map[string]float64, error) {
	c, err := buildInvestmentTrainingContext(ctx, db, targetCalcuttaID, trainYears, excludeEntryName)
	if err != nil {
		return nil, err
	}

	train, err := queryTrainingBidShares(ctx, db, targetCalcuttaID, trainYears, c.minYear, c.maxYear, excludeEntryName)
	if err != nil {
		return nil, err
	}
	intercept, slope, ok := fitKenPomScoreOLS(train)
	if !ok {
		return nil, fmt.Errorf("insufficient training data for kenpom-score investment model")
	}
	out := make(map[string]float64, len(targetRows))
	for _, r := range targetRows {
		pred := intercept
		if r.KenPomNetRtg != nil {
			pred = intercept + slope*(*r.KenPomNetRtg)
		}
		out[r.TeamID] = clamp01(pred)
	}
	return out, nil
}

func init() {
	RegisterInvestmentModel(seedInvestmentModel{})
	RegisterInvestmentModel(seedPodInvestmentModel{})
	RegisterInvestmentModel(seedKenPomDeltaInvestmentModel{})
	RegisterInvestmentModel(seedKenPomRankInvestmentModel{})
	RegisterInvestmentModel(kenPomRankInvestmentModel{})
	RegisterInvestmentModel(kenPomScoreInvestmentModel{})
	RegisterInvestmentModel(ridgeInvestmentModel{})
	RegisterInvestmentModel(ridgeReturnsInvestmentModel{})
}
