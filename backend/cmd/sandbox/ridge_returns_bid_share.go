package main

import (
	"context"
	"database/sql"
	"math"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/pkg/services"
	"gonum.org/v1/gonum/mat"
)

type ridgeReturnsBidShareModel struct {
	Beta           []float64
	SeedNetRtgMean map[int]float64
	SeedNetRtgStd  map[int]float64
	Lambda         float64
}

type trainReturnsRow struct {
	CalcuttaID     string
	TeamID         string
	Seed           int
	KenPomNetRtg   sql.NullFloat64
	BidShare       float64
	SeedBaseline   float64
	ExpPointsShare float64
	ROIFeature     float64
}

func ridgeReturnsROIFeature(expPointsShare float64, seedBaseline float64) float64 {
	denom := seedBaseline
	if denom < 1e-9 {
		denom = 1e-9
	}
	roi := expPointsShare / denom
	if roi < 0 {
		roi = 0
	}
	if roi > 10 {
		roi = 10
	}
	return math.Log1p(roi)
}

func ridgeReturnsSeg(seed int) (isTop4 float64, isTop8 float64) {
	if seed >= 1 && seed <= 4 {
		isTop4 = 1
	}
	if seed >= 1 && seed <= 8 {
		isTop8 = 1
	}
	return isTop4, isTop8
}

func returnsShareByTeamForCalcutta(ctx context.Context, db *sql.DB, calcuttaID string, sigma float64) (map[string]float64, error) {
	meta, err := tournamentMetaForCalcuttaID(ctx, db, calcuttaID)
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

	sum := 0.0
	for _, p := range pred {
		sum += p.ExpectedPoints
	}
	out := make(map[string]float64, len(pred))
	if sum <= 0 {
		return out, nil
	}
	for _, p := range pred {
		out[p.TeamID] = p.ExpectedPoints / sum
	}
	return out, nil
}

func fitRidgeReturnsBidShareModel(train []trainReturnsRow) (*ridgeReturnsBidShareModel, bool) {
	if len(train) == 0 {
		return nil, false
	}

	seedMean := map[int]float64{}
	seedStd := map[int]float64{}
	seedAgg := map[int]floatAgg{}
	seedSqAgg := map[int]floatAgg{}
	for _, r := range train {
		if !r.KenPomNetRtg.Valid {
			continue
		}
		a := seedAgg[r.Seed]
		a.Sum += r.KenPomNetRtg.Float64
		a.Count++
		seedAgg[r.Seed] = a
		b := seedSqAgg[r.Seed]
		b.Sum += r.KenPomNetRtg.Float64 * r.KenPomNetRtg.Float64
		b.Count++
		seedSqAgg[r.Seed] = b
	}
	for seed, a := range seedAgg {
		if a.Count == 0 {
			continue
		}
		mu := a.Sum / float64(a.Count)
		seedMean[seed] = mu
		sq := seedSqAgg[seed]
		if sq.Count > 0 {
			v := sq.Sum/float64(sq.Count) - mu*mu
			if v < 0 {
				v = 0
			}
			seedStd[seed] = math.Sqrt(v)
		}
	}

	lambdas := []float64{0.1, 1.0, 10.0}
	bestLambda := lambdas[0]
	bestMAE := math.Inf(1)
	for _, lam := range lambdas {
		mae, ok := ridgeReturnsLOOCVMAE(train, seedMean, seedStd, lam)
		if !ok {
			continue
		}
		if mae < bestMAE {
			bestMAE = mae
			bestLambda = lam
		}
	}

	beta, ok := ridgeReturnsSolve(train, seedMean, seedStd, bestLambda)
	if !ok {
		return nil, false
	}

	return &ridgeReturnsBidShareModel{Beta: beta, SeedNetRtgMean: seedMean, SeedNetRtgStd: seedStd, Lambda: bestLambda}, true
}

func ridgeReturnsLOOCVMAE(train []trainReturnsRow, seedMean map[int]float64, seedStd map[int]float64, lambda float64) (float64, bool) {
	byCalcutta := map[string][]trainReturnsRow{}
	for _, r := range train {
		byCalcutta[r.CalcuttaID] = append(byCalcutta[r.CalcuttaID], r)
	}
	if len(byCalcutta) < 2 {
		return 0, false
	}

	folds := make([]string, 0, len(byCalcutta))
	for id := range byCalcutta {
		folds = append(folds, id)
	}
	sort.Strings(folds)

	sumAbs := 0.0
	n := 0
	for _, holdoutID := range folds {
		trainFold := make([]trainReturnsRow, 0, len(train))
		for id, rows := range byCalcutta {
			if id == holdoutID {
				continue
			}
			trainFold = append(trainFold, rows...)
		}
		beta, ok := ridgeReturnsSolve(trainFold, seedMean, seedStd, lambda)
		if !ok {
			continue
		}
		for _, r := range byCalcutta[holdoutID] {
			predResid := ridgeReturnsPredict(beta, seedMean, seedStd, r.Seed, nullableFloatToPtr(r.KenPomNetRtg), r.ExpPointsShare, r.ROIFeature)
			actualResid := r.BidShare - r.SeedBaseline
			sumAbs += absFloat64(predResid - actualResid)
			n++
		}
	}
	if n == 0 {
		return 0, false
	}
	return sumAbs / float64(n), true
}

func ridgeReturnsSolve(train []trainReturnsRow, seedMean map[int]float64, seedStd map[int]float64, lambda float64) ([]float64, bool) {
	// Features:
	//  - 16 seed indicator columns
	//  - 1 KenPom z-score within seed
	//  - 1 expected points share
	//  - 1 ROI-like feature (log1p(capped ROI))
	//  - segment interactions to capture different market behavior for favorites
	p := 23

	xtx := mat.NewDense(p, p, nil)
	xty := mat.NewVecDense(p, nil)

	for _, r := range train {
		x := make([]float64, p)
		sIdx := r.Seed - 1
		if sIdx >= 0 && sIdx < 16 {
			x[sIdx] = 1
		}
		x[16] = ridgeKenPomZ(seedMean, seedStd, r.Seed, nullableFloatToPtr(r.KenPomNetRtg))
		x[17] = r.ExpPointsShare
		x[18] = r.ROIFeature
		top4, top8 := ridgeReturnsSeg(r.Seed)
		x[19] = r.ExpPointsShare * top4
		x[20] = r.ExpPointsShare * top8
		x[21] = r.ROIFeature * top4
		x[22] = r.ROIFeature * top8

		y := r.BidShare - r.SeedBaseline

		for i := 0; i < p; i++ {
			if x[i] == 0 {
				continue
			}
			xty.SetVec(i, xty.AtVec(i)+x[i]*y)
			for j := 0; j < p; j++ {
				if x[j] == 0 {
					continue
				}
				xtx.Set(i, j, xtx.At(i, j)+x[i]*x[j])
			}
		}
	}

	for i := 0; i < p; i++ {
		xtx.Set(i, i, xtx.At(i, i)+lambda)
	}

	betaVec := mat.NewVecDense(p, nil)
	if err := betaVec.SolveVec(xtx, xty); err != nil {
		return nil, false
	}

	out := make([]float64, p)
	for i := 0; i < p; i++ {
		out[i] = betaVec.AtVec(i)
	}
	return out, true
}

func ridgeReturnsPredict(beta []float64, seedMean map[int]float64, seedStd map[int]float64, seed int, netRtg *float64, expPointsShare float64, roiFeat float64) float64 {
	p := 23
	if len(beta) != p {
		return 0
	}
	sIdx := seed - 1
	out := 0.0
	if sIdx >= 0 && sIdx < 16 {
		out += beta[sIdx]
	}
	out += beta[16] * ridgeKenPomZ(seedMean, seedStd, seed, netRtg)
	out += beta[17] * expPointsShare
	out += beta[18] * roiFeat
	top4, top8 := ridgeReturnsSeg(seed)
	out += beta[19] * (expPointsShare * top4)
	out += beta[20] * (expPointsShare * top8)
	out += beta[21] * (roiFeat * top4)
	out += beta[22] * (roiFeat * top8)
	return out
}
