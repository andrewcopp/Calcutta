package main

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
)

type seedIntKey struct {
	Seed int
	Val  int
}

func seedKenPomRankBucketMeans(rows []trainingBidRow) map[seedIntKey]float64 {
	byCalcutta := map[string][]trainingBidRow{}
	for _, r := range rows {
		byCalcutta[r.CalcuttaID] = append(byCalcutta[r.CalcuttaID], r)
	}

	agg := map[seedIntKey]floatAgg{}
	for _, calcuttaRows := range byCalcutta {
		buckets := kenPomRankBucketsTraining(calcuttaRows)
		for _, r := range calcuttaRows {
			key := seedIntKey{Seed: r.Seed, Val: buckets[r.TeamID]}
			a := agg[key]
			a.Sum += r.BidShare
			a.Count++
			agg[key] = a
		}
	}

	out := make(map[seedIntKey]float64, len(agg))
	for k, a := range agg {
		if a.Count > 0 {
			out[k] = a.Sum / float64(a.Count)
		}
	}
	return out
}

type floatAgg struct {
	Sum   float64
	Count int
}

func predictedBidShareByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, investModel string, excludeEntryName string) (map[string]float64, error) {
	if investModel == "" {
		investModel = "seed"
	}
	model, err := GetInvestmentModel(investModel)
	if err != nil {
		allowed := strings.Join(ListInvestmentModelNames(), "|")
		return nil, fmt.Errorf("unknown invest-model %q (expected %s)", investModel, allowed)
	}
	return model.PredictBidShareByTeam(ctx, db, targetCalcuttaID, targetRows, trainYears, excludeEntryName)
}

func seedPodRankMeans(rows []trainingBidRow) map[seedIntKey]float64 {
	byCalcutta := map[string][]trainingBidRow{}
	for _, r := range rows {
		byCalcutta[r.CalcuttaID] = append(byCalcutta[r.CalcuttaID], r)
	}

	agg := map[seedIntKey]floatAgg{}
	for _, calcuttaRows := range byCalcutta {
		bySeed := map[int][]trainingBidRow{}
		for _, r := range calcuttaRows {
			bySeed[r.Seed] = append(bySeed[r.Seed], r)
		}
		for seed, seedRows := range bySeed {
			ranks := podRanksTraining(seedRows)
			for _, r := range seedRows {
				key := seedIntKey{Seed: seed, Val: ranks[r.TeamID]}
				a := agg[key]
				a.Sum += r.BidShare
				a.Count++
				agg[key] = a
			}
		}
	}

	out := make(map[seedIntKey]float64, len(agg))
	for k, a := range agg {
		if a.Count > 0 {
			out[k] = a.Sum / float64(a.Count)
		}
	}
	return out
}

func podRanksTraining(rows []trainingBidRow) map[string]int {
	type item struct {
		TeamID string
		X      sql.NullFloat64
	}

	items := make([]item, 0, len(rows))
	for _, r := range rows {
		items = append(items, item{TeamID: r.TeamID, X: r.KenPomNetRtg})
	}

	sort.Slice(items, func(i, j int) bool {
		a := items[i].X
		b := items[j].X
		if a.Valid != b.Valid {
			return a.Valid
		}
		if !a.Valid && !b.Valid {
			return items[i].TeamID < items[j].TeamID
		}
		if a.Float64 == b.Float64 {
			return items[i].TeamID < items[j].TeamID
		}
		return a.Float64 > b.Float64
	})

	out := make(map[string]int, len(items))
	for i := range items {
		out[items[i].TeamID] = i + 1
	}
	return out
}

func seedPodRankByTeam(targetRows []TeamDatasetRow) map[string]int {
	bySeed := map[int][]TeamDatasetRow{}
	for _, r := range targetRows {
		bySeed[r.Seed] = append(bySeed[r.Seed], r)
	}

	out := make(map[string]int, len(targetRows))
	for _, seedRows := range bySeed {
		type item struct {
			TeamID string
			X      *float64
		}
		items := make([]item, 0, len(seedRows))
		for _, r := range seedRows {
			items = append(items, item{TeamID: r.TeamID, X: r.KenPomNetRtg})
		}
		sort.Slice(items, func(i, j int) bool {
			a := items[i].X
			b := items[j].X
			if (a != nil) != (b != nil) {
				return a != nil
			}
			if a == nil && b == nil {
				return items[i].TeamID < items[j].TeamID
			}
			if *a == *b {
				return items[i].TeamID < items[j].TeamID
			}
			return *a > *b
		})
		for i := range items {
			out[items[i].TeamID] = i + 1
		}
	}
	return out
}

func seedKenPomDeltaBinMeans(rows []trainingBidRow) map[seedIntKey]float64 {
	byCalcutta := map[string][]trainingBidRow{}
	for _, r := range rows {
		byCalcutta[r.CalcuttaID] = append(byCalcutta[r.CalcuttaID], r)
	}

	agg := map[seedIntKey]floatAgg{}
	for _, calcuttaRows := range byCalcutta {
		bySeed := map[int][]trainingBidRow{}
		for _, r := range calcuttaRows {
			bySeed[r.Seed] = append(bySeed[r.Seed], r)
		}
		for seed, seedRows := range bySeed {
			bins := seedKenPomDeltaBinsTraining(seedRows)
			for _, r := range seedRows {
				key := seedIntKey{Seed: seed, Val: bins[r.TeamID]}
				a := agg[key]
				a.Sum += r.BidShare
				a.Count++
				agg[key] = a
			}
		}
	}

	out := make(map[seedIntKey]float64, len(agg))
	for k, a := range agg {
		if a.Count > 0 {
			out[k] = a.Sum / float64(a.Count)
		}
	}
	return out
}

func seedKenPomDeltaBinsTraining(rows []trainingBidRow) map[string]int {
	vals := make([]float64, 0, len(rows))
	for _, r := range rows {
		if r.KenPomNetRtg.Valid {
			vals = append(vals, r.KenPomNetRtg.Float64)
		}
	}
	mean, std := meanStd(vals)

	out := make(map[string]int, len(rows))
	for _, r := range rows {
		bin := 0
		if r.KenPomNetRtg.Valid && std > 0 {
			z := (r.KenPomNetRtg.Float64 - mean) / std
			if z <= -1 {
				bin = -1
			} else if z >= 1 {
				bin = 1
			} else {
				bin = 0
			}
		}
		out[r.TeamID] = bin
	}
	return out
}

func seedKenPomDeltaBinByTeam(targetRows []TeamDatasetRow) map[string]int {
	bySeed := map[int][]TeamDatasetRow{}
	for _, r := range targetRows {
		bySeed[r.Seed] = append(bySeed[r.Seed], r)
	}

	out := make(map[string]int, len(targetRows))
	for _, seedRows := range bySeed {
		vals := make([]float64, 0, len(seedRows))
		for _, r := range seedRows {
			if r.KenPomNetRtg != nil {
				vals = append(vals, *r.KenPomNetRtg)
			}
		}
		mean, std := meanStd(vals)
		for _, r := range seedRows {
			bin := 0
			if r.KenPomNetRtg != nil && std > 0 {
				z := (*r.KenPomNetRtg - mean) / std
				if z <= -1 {
					bin = -1
				} else if z >= 1 {
					bin = 1
				}
			}
			out[r.TeamID] = bin
		}
	}
	return out
}

func kenPomRankBucketMeans(rows []trainingBidRow) map[int]float64 {
	byCalcutta := map[string][]trainingBidRow{}
	for _, r := range rows {
		byCalcutta[r.CalcuttaID] = append(byCalcutta[r.CalcuttaID], r)
	}

	agg := map[int]floatAgg{}
	overall := floatAgg{}
	for _, calcuttaRows := range byCalcutta {
		buckets := kenPomRankBucketsTraining(calcuttaRows)
		for _, r := range calcuttaRows {
			b := buckets[r.TeamID]
			a := agg[b]
			a.Sum += r.BidShare
			a.Count++
			agg[b] = a
			overall.Sum += r.BidShare
			overall.Count++
		}
	}

	out := make(map[int]float64, len(agg)+1)
	for k, a := range agg {
		if a.Count > 0 {
			out[k] = a.Sum / float64(a.Count)
		}
	}
	if overall.Count > 0 {
		out[-1] = overall.Sum / float64(overall.Count)
	}
	return out
}

func kenPomRankBucketsTraining(rows []trainingBidRow) map[string]int {
	type item struct {
		TeamID string
		X      sql.NullFloat64
	}
	items := make([]item, 0, len(rows))
	for _, r := range rows {
		items = append(items, item{TeamID: r.TeamID, X: r.KenPomNetRtg})
	}

	sort.Slice(items, func(i, j int) bool {
		a := items[i].X
		b := items[j].X
		if a.Valid != b.Valid {
			return a.Valid
		}
		if !a.Valid && !b.Valid {
			return items[i].TeamID < items[j].TeamID
		}
		if a.Float64 == b.Float64 {
			return items[i].TeamID < items[j].TeamID
		}
		return a.Float64 > b.Float64
	})

	out := make(map[string]int, len(items))
	for i := range items {
		rank := i + 1
		bucket := kenPomRankBucket(rank, items[i].X.Valid)
		out[items[i].TeamID] = bucket
	}
	return out
}

func kenPomRankBucketByTeam(targetRows []TeamDatasetRow) map[string]int {
	type item struct {
		TeamID string
		X      *float64
	}
	items := make([]item, 0, len(targetRows))
	for _, r := range targetRows {
		items = append(items, item{TeamID: r.TeamID, X: r.KenPomNetRtg})
	}

	sort.Slice(items, func(i, j int) bool {
		a := items[i].X
		b := items[j].X
		if (a != nil) != (b != nil) {
			return a != nil
		}
		if a == nil && b == nil {
			return items[i].TeamID < items[j].TeamID
		}
		if *a == *b {
			return items[i].TeamID < items[j].TeamID
		}
		return *a > *b
	})

	out := make(map[string]int, len(items))
	for i := range items {
		rank := i + 1
		valid := items[i].X != nil
		out[items[i].TeamID] = kenPomRankBucket(rank, valid)
	}
	return out
}

func kenPomRankBucket(rank int, valid bool) int {
	if !valid {
		return 99
	}
	switch {
	case rank <= 4:
		return 1
	case rank <= 8:
		return 2
	case rank <= 16:
		return 3
	case rank <= 32:
		return 4
	case rank <= 64:
		return 5
	default:
		return 6
	}
}

func fitKenPomScoreOLS(rows []trainingBidRow) (float64, float64, bool) {
	xs := make([]float64, 0, len(rows))
	ys := make([]float64, 0, len(rows))
	for _, r := range rows {
		if !r.KenPomNetRtg.Valid {
			continue
		}
		xs = append(xs, r.KenPomNetRtg.Float64)
		ys = append(ys, r.BidShare)
	}
	if len(xs) < 2 {
		return 0, 0, false
	}

	xMean, _ := meanStd(xs)
	yMean, _ := meanStd(ys)

	var num float64
	var den float64
	for i := range xs {
		dx := xs[i] - xMean
		num += dx * (ys[i] - yMean)
		den += dx * dx
	}
	if den == 0 {
		return yMean, 0, true
	}
	slope := num / den
	intercept := yMean - slope*xMean
	return intercept, slope, true
}

func meanStd(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	if len(values) < 2 {
		return mean, 0
	}
	var ss float64
	for _, v := range values {
		d := v - mean
		ss += d * d
	}
	std := math.Sqrt(ss / float64(len(values)-1))
	return mean, std
}
