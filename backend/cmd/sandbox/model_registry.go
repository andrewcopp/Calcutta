package main

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
)

type InvestmentModel interface {
	Name() string
	PredictBidShareByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, excludeEntryName string) (map[string]float64, error)
}

type PointsModel interface {
	Name() string
	PredictPointsByTeam(ctx context.Context, db *sql.DB, targetCalcuttaID string, targetRows []TeamDatasetRow, trainYears int, sigma float64) (map[string]float64, error)
}

var (
	investmentModelsMu sync.RWMutex
	investmentModels   = map[string]InvestmentModel{}

	pointsModelsMu sync.RWMutex
	pointsModels   = map[string]PointsModel{}
)

func RegisterInvestmentModel(m InvestmentModel) {
	if m == nil {
		panic("RegisterInvestmentModel: nil model")
	}
	name := canonicalModelName(m.Name())
	if name == "" {
		panic("RegisterInvestmentModel: empty model name")
	}
	investmentModelsMu.Lock()
	defer investmentModelsMu.Unlock()
	if _, exists := investmentModels[name]; exists {
		panic(fmt.Sprintf("RegisterInvestmentModel: duplicate model name %q", name))
	}
	investmentModels[name] = m
}

func GetInvestmentModel(name string) (InvestmentModel, error) {
	key := canonicalModelName(name)
	investmentModelsMu.RLock()
	m := investmentModels[key]
	investmentModelsMu.RUnlock()
	if m == nil {
		return nil, fmt.Errorf("unknown invest-model %q", name)
	}
	return m, nil
}

func ListInvestmentModelNames() []string {
	investmentModelsMu.RLock()
	names := make([]string, 0, len(investmentModels))
	for name := range investmentModels {
		names = append(names, name)
	}
	investmentModelsMu.RUnlock()
	sort.Strings(names)
	return names
}

func RegisterPointsModel(m PointsModel) {
	if m == nil {
		panic("RegisterPointsModel: nil model")
	}
	name := canonicalModelName(m.Name())
	if name == "" {
		panic("RegisterPointsModel: empty model name")
	}
	pointsModelsMu.Lock()
	defer pointsModelsMu.Unlock()
	if _, exists := pointsModels[name]; exists {
		panic(fmt.Sprintf("RegisterPointsModel: duplicate model name %q", name))
	}
	pointsModels[name] = m
}

func GetPointsModel(name string) (PointsModel, error) {
	key := canonicalModelName(name)
	pointsModelsMu.RLock()
	m := pointsModels[key]
	pointsModelsMu.RUnlock()
	if m == nil {
		return nil, fmt.Errorf("unknown pred-model %q", name)
	}
	return m, nil
}

func ListPointsModelNames() []string {
	pointsModelsMu.RLock()
	names := make([]string, 0, len(pointsModels))
	for name := range pointsModels {
		names = append(names, name)
	}
	pointsModelsMu.RUnlock()
	sort.Strings(names)
	return names
}
