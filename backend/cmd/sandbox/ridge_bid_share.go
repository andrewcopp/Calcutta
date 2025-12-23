package main

import (
	"database/sql"
	"math"
	"sort"

	"gonum.org/v1/gonum/mat"
)

type ridgeBidShareModel struct {
	Beta            []float64
	SeedNetRtgMean  map[int]float64
	SeedNetRtgStd   map[int]float64
	Lambda          float64
	HasKenPomSignal bool
}

func fitRidgeBidShareModel(train []trainingBidRow) (*ridgeBidShareModel, bool) {
	if len(train) == 0 {
		return nil, false
	}

	seedMean := map[int]float64{}
	seedStd := map[int]float64{}

	// Compute within-seed KenPom mean/std for z-scoring.
	seedAgg := map[int]floatAgg{}
	seedSqAgg := map[int]floatAgg{}
	kenpomN := 0
	for _, r := range train {
		if !r.KenPomNetRtg.Valid {
			continue
		}
		kenpomN++
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
		mae, ok := ridgeLOOCVMAE(train, seedMean, seedStd, lam)
		if !ok {
			continue
		}
		if mae < bestMAE {
			bestMAE = mae
			bestLambda = lam
		}
	}

	beta, ok := ridgeSolve(train, seedMean, seedStd, bestLambda)
	if !ok {
		return nil, false
	}

	return &ridgeBidShareModel{
		Beta:            beta,
		SeedNetRtgMean:  seedMean,
		SeedNetRtgStd:   seedStd,
		Lambda:          bestLambda,
		HasKenPomSignal: kenpomN > 0,
	}, true
}

func ridgeLOOCVMAE(train []trainingBidRow, seedMean map[int]float64, seedStd map[int]float64, lambda float64) (float64, bool) {
	byCalcutta := map[string][]trainingBidRow{}
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
		trainFold := make([]trainingBidRow, 0, len(train))
		for id, rows := range byCalcutta {
			if id == holdoutID {
				continue
			}
			trainFold = append(trainFold, rows...)
		}
		beta, ok := ridgeSolve(trainFold, seedMean, seedStd, lambda)
		if !ok {
			continue
		}
		for _, r := range byCalcutta[holdoutID] {
			pred := ridgePredict(beta, seedMean, seedStd, r.Seed, nullableFloatToPtr(r.KenPomNetRtg))
			if pred < 0 {
				pred = 0
			}
			if pred > 1 {
				pred = 1
			}
			sumAbs += absFloat64(pred - r.BidShare)
			n++
		}
	}
	if n == 0 {
		return 0, false
	}
	return sumAbs / float64(n), true
}

func ridgeSolve(train []trainingBidRow, seedMean map[int]float64, seedStd map[int]float64, lambda float64) ([]float64, bool) {
	// Features:
	//  - 16 seed indicator columns
	//  - 1 KenPom z-score within seed
	p := 17

	// Build X^T X and X^T y without materializing X.
	xtx := mat.NewDense(p, p, nil)
	xty := mat.NewVecDense(p, nil)

	for _, r := range train {
		x := make([]float64, p)
		sIdx := r.Seed - 1
		if sIdx >= 0 && sIdx < 16 {
			x[sIdx] = 1
		}
		x[16] = ridgeKenPomZ(seedMean, seedStd, r.Seed, nullableFloatToPtr(r.KenPomNetRtg))

		for i := 0; i < p; i++ {
			if x[i] == 0 {
				continue
			}
			xty.SetVec(i, xty.AtVec(i)+x[i]*r.BidShare)
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

func ridgePredict(beta []float64, seedMean map[int]float64, seedStd map[int]float64, seed int, netRtg *float64) float64 {
	p := 17
	if len(beta) != p {
		return 0
	}
	sIdx := seed - 1
	out := 0.0
	if sIdx >= 0 && sIdx < 16 {
		out += beta[sIdx]
	}
	out += beta[16] * ridgeKenPomZ(seedMean, seedStd, seed, netRtg)
	return out
}

func ridgeKenPomZ(seedMean map[int]float64, seedStd map[int]float64, seed int, netRtg *float64) float64 {
	if netRtg == nil {
		return 0
	}
	mu, ok := seedMean[seed]
	if !ok {
		return 0
	}
	sd := seedStd[seed]
	if sd < 1e-9 {
		return 0
	}
	return (*netRtg - mu) / sd
}

func nullableFloatToPtr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	x := v.Float64
	return &x
}
