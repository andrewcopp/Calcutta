package main

import (
	"database/sql"
	"math"
	"sort"
)

type alphaCombo3 [3]float64

type seedStats struct {
	mean map[int]float64
	std  map[int]float64
}

func seedMeanStdFromTrainingRows(rows []trainingBidRow) seedStats {
	meanAgg := map[int]floatAgg{}
	sqAgg := map[int]floatAgg{}
	for _, r := range rows {
		a := meanAgg[r.Seed]
		a.Sum += r.BidShare
		a.Count++
		meanAgg[r.Seed] = a
		b := sqAgg[r.Seed]
		b.Sum += r.BidShare * r.BidShare
		b.Count++
		sqAgg[r.Seed] = b
	}
	mean := make(map[int]float64, len(meanAgg))
	std := make(map[int]float64, len(meanAgg))
	for seed, a := range meanAgg {
		if a.Count == 0 {
			continue
		}
		mu := a.Sum / float64(a.Count)
		mean[seed] = mu
		sq := sqAgg[seed]
		if sq.Count > 0 {
			v := sq.Sum/float64(sq.Count) - mu*mu
			if v < 0 {
				v = 0
			}
			s := math.Sqrt(v)
			if s < 1e-9 {
				s = 1e-9
			}
			std[seed] = s
		}
	}
	return seedStats{mean: mean, std: std}
}

func kenPomScoreRaw(intercept float64, slope float64, net sql.NullFloat64) float64 {
	if net.Valid {
		return intercept + slope*net.Float64
	}
	return intercept
}

func chooseShrinkAlphaLOOCVKenPomScore(train []trainingBidRow, candidateAlphas []float64) float64 {
	if len(candidateAlphas) == 0 {
		candidateAlphas = []float64{0, 0.25, 0.5, 0.75, 1.0}
	}

	byCalcutta := map[string][]trainingBidRow{}
	for _, r := range train {
		byCalcutta[r.CalcuttaID] = append(byCalcutta[r.CalcuttaID], r)
	}
	if len(byCalcutta) < 2 {
		return 0.5
	}

	foldIDs := make([]string, 0, len(byCalcutta))
	for id := range byCalcutta {
		foldIDs = append(foldIDs, id)
	}
	sort.Strings(foldIDs)

	scores := make(map[float64]float64, len(candidateAlphas))
	counts := make(map[float64]int, len(candidateAlphas))

	for _, holdoutID := range foldIDs {
		trainFold := make([]trainingBidRow, 0, len(train))
		for id, rows := range byCalcutta {
			if id == holdoutID {
				continue
			}
			trainFold = append(trainFold, rows...)
		}
		intercept, slope, ok := fitKenPomScoreOLS(trainFold)
		if !ok {
			continue
		}
		stats := seedMeanStdFromTrainingRows(trainFold)

		holdoutRows := byCalcutta[holdoutID]
		if len(holdoutRows) == 0 {
			continue
		}

		for _, alpha := range candidateAlphas {
			predSumBySeed := map[int]floatAgg{}
			predByRow := make([]float64, len(holdoutRows))
			for i, r := range holdoutRows {
				baseline := stats.mean[r.Seed]
				raw := kenPomScoreRaw(intercept, slope, r.KenPomNetRtg)
				pred := baseline + alpha*(raw-baseline)
				pred = clamp01(pred)
				predByRow[i] = pred
				a := predSumBySeed[r.Seed]
				a.Sum += pred
				a.Count++
				predSumBySeed[r.Seed] = a
			}

			meanPredBySeed := map[int]float64{}
			for seed, a := range predSumBySeed {
				if a.Count > 0 {
					meanPredBySeed[seed] = a.Sum / float64(a.Count)
				}
			}

			for i, r := range holdoutRows {
				muA := stats.mean[r.Seed]
				sdA := stats.std[r.Seed]
				muP := meanPredBySeed[r.Seed]
				if sdA < 1e-9 {
					sdA = 1e-9
				}
				residActual := r.BidShare - muA
				residPred := predByRow[i] - muP
				dr := residPred - residActual
				scores[alpha] += absFloat64(dr) / sdA
				counts[alpha]++
			}
		}
	}

	bestAlpha := candidateAlphas[0]
	bestScore := math.Inf(1)
	for _, alpha := range candidateAlphas {
		n := counts[alpha]
		if n == 0 {
			continue
		}
		s := scores[alpha] / float64(n)
		if s < bestScore-1e-12 {
			bestScore = s
			bestAlpha = alpha
			continue
		}
		// Tie-break: prefer smaller alpha (more shrink) when scores are effectively equal.
		if math.Abs(s-bestScore) <= 1e-12 && alpha < bestAlpha {
			bestAlpha = alpha
		}
	}
	return bestAlpha
}

func seedSegmentIdx(seed int) int {
	if seed >= 1 && seed <= 4 {
		return 0
	}
	if seed >= 1 && seed <= 8 {
		return 1
	}
	return 2
}

func chooseShrinkAlphaSegmentsLOOCVKenPomScore(train []trainingBidRow, candidateAlphas []float64) (float64, float64, float64) {
	if len(candidateAlphas) == 0 {
		candidateAlphas = []float64{0, 0.25, 0.5, 0.75, 1.0}
	}

	byCalcutta := map[string][]trainingBidRow{}
	for _, r := range train {
		byCalcutta[r.CalcuttaID] = append(byCalcutta[r.CalcuttaID], r)
	}
	if len(byCalcutta) < 2 {
		return 0.5, 0.5, 0.5
	}

	foldIDs := make([]string, 0, len(byCalcutta))
	for id := range byCalcutta {
		foldIDs = append(foldIDs, id)
	}
	sort.Strings(foldIDs)

	scores := map[alphaCombo3]float64{}
	counts := map[alphaCombo3]int{}

	for _, holdoutID := range foldIDs {
		trainFold := make([]trainingBidRow, 0, len(train))
		for id, rows := range byCalcutta {
			if id == holdoutID {
				continue
			}
			trainFold = append(trainFold, rows...)
		}
		intercept, slope, ok := fitKenPomScoreOLS(trainFold)
		if !ok {
			continue
		}
		stats := seedMeanStdFromTrainingRows(trainFold)

		holdoutRows := byCalcutta[holdoutID]
		if len(holdoutRows) == 0 {
			continue
		}

		for _, aTop4 := range candidateAlphas {
			for _, aTop8 := range candidateAlphas {
				for _, aOther := range candidateAlphas {
					combo := alphaCombo3{aTop4, aTop8, aOther}

					predByRow := make([]float64, len(holdoutRows))
					predAggBySeed := map[int]floatAgg{}
					for i, r := range holdoutRows {
						seg := seedSegmentIdx(r.Seed)
						alpha := combo[seg]
						baseline := stats.mean[r.Seed]
						raw := kenPomScoreRaw(intercept, slope, r.KenPomNetRtg)
						pred := baseline + alpha*(raw-baseline)
						pred = clamp01(pred)
						predByRow[i] = pred
						a := predAggBySeed[r.Seed]
						a.Sum += pred
						a.Count++
						predAggBySeed[r.Seed] = a
					}

					meanPredBySeed := map[int]float64{}
					for seed, a := range predAggBySeed {
						if a.Count > 0 {
							meanPredBySeed[seed] = a.Sum / float64(a.Count)
						}
					}

					for i, r := range holdoutRows {
						muA := stats.mean[r.Seed]
						sdA := stats.std[r.Seed]
						if sdA < 1e-9 {
							sdA = 1e-9
						}
						muP := meanPredBySeed[r.Seed]
						residActual := r.BidShare - muA
						residPred := predByRow[i] - muP
						dr := residPred - residActual
						scores[combo] += absFloat64(dr) / sdA
						counts[combo]++
					}
				}
			}
		}
	}

	bestCombo := alphaCombo3{candidateAlphas[0], candidateAlphas[0], candidateAlphas[0]}
	bestScore := math.Inf(1)
	bestInit := false
	for _, aTop4 := range candidateAlphas {
		for _, aTop8 := range candidateAlphas {
			for _, aOther := range candidateAlphas {
				combo := alphaCombo3{aTop4, aTop8, aOther}
				n := counts[combo]
				if n == 0 {
					continue
				}
				s := scores[combo] / float64(n)
				if !bestInit || s < bestScore-1e-12 {
					bestInit = true
					bestScore = s
					bestCombo = combo
					continue
				}
				if math.Abs(s-bestScore) <= 1e-12 {
					// Conservative tie-break: prefer more shrink overall.
					sumA := combo[0] + combo[1] + combo[2]
					sumB := bestCombo[0] + bestCombo[1] + bestCombo[2]
					if sumA < sumB-1e-12 {
						bestCombo = combo
						continue
					}
					if math.Abs(sumA-sumB) <= 1e-12 {
						// If still tied, prefer lexicographically smaller alphas.
						if combo[0] < bestCombo[0] || (combo[0] == bestCombo[0] && (combo[1] < bestCombo[1] || (combo[1] == bestCombo[1] && combo[2] < bestCombo[2]))) {
							bestCombo = combo
						}
					}
				}
			}
		}
	}

	return bestCombo[0], bestCombo[1], bestCombo[2]
}
