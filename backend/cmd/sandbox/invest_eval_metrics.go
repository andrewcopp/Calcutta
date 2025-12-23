package main

import (
	"math"
	"sort"
)

type seedErrAgg struct {
	AbsSum float64
	Count  int
}

func meanBySeed(values map[string]float64, seedByTeam map[string]int) map[int]float64 {
	agg := map[int]floatAgg{}
	for teamID, v := range values {
		seed, ok := seedByTeam[teamID]
		if !ok {
			continue
		}
		a := agg[seed]
		a.Sum += v
		a.Count++
		agg[seed] = a
	}
	out := make(map[int]float64, len(agg))
	for seed, a := range agg {
		if a.Count > 0 {
			out[seed] = a.Sum / float64(a.Count)
		}
	}
	return out
}

func stddevBySeed(values map[string]float64, seedByTeam map[string]int, meanBySeed map[int]float64) map[int]float64 {
	sse := map[int]floatAgg{}
	for teamID, v := range values {
		seed, ok := seedByTeam[teamID]
		if !ok {
			continue
		}
		mu := meanBySeed[seed]
		d := v - mu
		a := sse[seed]
		a.Sum += d * d
		a.Count++
		sse[seed] = a
	}
	out := make(map[int]float64, len(sse))
	for seed, a := range sse {
		if a.Count > 0 {
			out[seed] = math.Sqrt(a.Sum / float64(a.Count))
		}
	}
	return out
}

func equalWeightBySeedMAE(actualByTeam map[string]float64, predByTeam map[string]float64, seedByTeam map[string]int) float64 {
	bySeed := map[int]floatAgg{}
	for teamID, actual := range actualByTeam {
		seed, ok := seedByTeam[teamID]
		if !ok {
			continue
		}
		pred := predByTeam[teamID]
		a := bySeed[seed]
		a.Sum += absFloat64(pred - actual)
		a.Count++
		bySeed[seed] = a
	}
	return overallEqualWeightBySeedMAE(bySeed)
}

func overallEqualWeightBySeedMAE(bySeed map[int]floatAgg) float64 {
	sum := 0.0
	seeds := 0
	for _, a := range bySeed {
		if a.Count == 0 {
			continue
		}
		sum += a.Sum / float64(a.Count)
		seeds++
	}
	if seeds == 0 {
		return 0
	}
	return sum / float64(seeds)
}

func topKResidualHitCounts(actualByTeam map[string]float64, predByTeam map[string]float64, seedByTeam map[string]int, meanActualBySeed map[int]float64, meanPredBySeed map[int]float64, k int) (int, int, int) {
	return topKResidualHitCountsAllowed(actualByTeam, predByTeam, seedByTeam, meanActualBySeed, meanPredBySeed, k, nil)
}

func topKResidualHitCountsAllowed(actualByTeam map[string]float64, predByTeam map[string]float64, seedByTeam map[string]int, meanActualBySeed map[int]float64, meanPredBySeed map[int]float64, k int, allowedTeamIDs map[string]struct{}) (int, int, int) {
	type item struct {
		TeamID      string
		ResidActual float64
		ResidPred   float64
	}
	items := make([]item, 0, len(actualByTeam))
	for teamID, actual := range actualByTeam {
		if allowedTeamIDs != nil {
			if _, ok := allowedTeamIDs[teamID]; !ok {
				continue
			}
		}
		seed, ok := seedByTeam[teamID]
		if !ok {
			continue
		}
		pred := predByTeam[teamID]
		items = append(items, item{
			TeamID:      teamID,
			ResidActual: actual - meanActualBySeed[seed],
			ResidPred:   pred - meanPredBySeed[seed],
		})
	}
	if k <= 0 || len(items) == 0 {
		return 0, 0, 0
	}
	if k > len(items) {
		k = len(items)
	}

	underActual := make([]item, len(items))
	copy(underActual, items)
	sort.Slice(underActual, func(i, j int) bool {
		if underActual[i].ResidActual == underActual[j].ResidActual {
			return underActual[i].TeamID < underActual[j].TeamID
		}
		return underActual[i].ResidActual < underActual[j].ResidActual
	})
	underPred := make([]item, len(items))
	copy(underPred, items)
	sort.Slice(underPred, func(i, j int) bool {
		if underPred[i].ResidPred == underPred[j].ResidPred {
			return underPred[i].TeamID < underPred[j].TeamID
		}
		return underPred[i].ResidPred < underPred[j].ResidPred
	})

	overActual := make([]item, len(items))
	copy(overActual, items)
	sort.Slice(overActual, func(i, j int) bool {
		if overActual[i].ResidActual == overActual[j].ResidActual {
			return overActual[i].TeamID < overActual[j].TeamID
		}
		return overActual[i].ResidActual > overActual[j].ResidActual
	})
	overPred := make([]item, len(items))
	copy(overPred, items)
	sort.Slice(overPred, func(i, j int) bool {
		if overPred[i].ResidPred == overPred[j].ResidPred {
			return overPred[i].TeamID < overPred[j].TeamID
		}
		return overPred[i].ResidPred > overPred[j].ResidPred
	})

	trueUnder := map[string]struct{}{}
	trueOver := map[string]struct{}{}
	for i := 0; i < k; i++ {
		trueUnder[underActual[i].TeamID] = struct{}{}
		trueOver[overActual[i].TeamID] = struct{}{}
	}

	underHits := 0
	overHits := 0
	for i := 0; i < k; i++ {
		if _, ok := trueUnder[underPred[i].TeamID]; ok {
			underHits++
		}
		if _, ok := trueOver[overPred[i].TeamID]; ok {
			overHits++
		}
	}
	return underHits, overHits, k
}
