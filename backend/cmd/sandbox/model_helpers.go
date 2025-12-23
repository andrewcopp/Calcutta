package main

import (
	"fmt"
	"strings"
)

func canonicalModelName(name string) string {
	return strings.TrimSpace(strings.ToLower(name))
}

func clampFloat64(v float64, lo float64, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clamp01(v float64) float64 {
	return clampFloat64(v, 0, 1)
}

func normalizeNonNegativeScoresForTeams(teamIDs []string, scoresByTeam map[string]float64) (map[string]float64, float64, error) {
	sum := 0.0
	for _, teamID := range teamIDs {
		v := scoresByTeam[teamID]
		if v < 0 {
			v = 0
		}
		sum += v
	}
	if sum <= 0 {
		return nil, 0, fmt.Errorf("sum of non-negative scores must be > 0")
	}
	shares := make(map[string]float64, len(teamIDs))
	for _, teamID := range teamIDs {
		v := scoresByTeam[teamID]
		if v < 0 {
			v = 0
		}
		shares[teamID] = v / sum
	}
	return shares, sum, nil
}
