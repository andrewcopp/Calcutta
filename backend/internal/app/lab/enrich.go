package lab

import (
	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// SeedExpectedPoints provides fallback expected-point estimates keyed by seed.
// Used only when an entry has no predictions from Python.
var SeedExpectedPoints = map[int]float64{
	1: 320, 2: 240, 3: 160, 4: 120, 5: 80, 6: 64, 7: 48, 8: 40,
	9: 32, 10: 24, 11: 20, 12: 16, 13: 8, 14: 4, 15: 2, 16: 1,
}

// EnrichEntry transforms a raw entry (from the database) into an enriched entry
// with naive allocation, edge percent, and ROI calculations. This function is
// pure (no DB access) -- all required data is passed in via the raw entry.
func EnrichEntry(raw *models.LabEntryRaw) *models.LabEntryDetailEnriched {
	result := &models.LabEntryDetailEnriched{
		ID:                    raw.ID,
		InvestmentModelID:     raw.InvestmentModelID,
		CalcuttaID:            raw.CalcuttaID,
		GameOutcomeKind:       raw.GameOutcomeKind,
		GameOutcomeParamsJSON: raw.GameOutcomeParamsJSON,
		OptimizerKind:         raw.OptimizerKind,
		OptimizerParamsJSON:   raw.OptimizerParamsJSON,
		StartingStateKey:      raw.StartingStateKey,
		CreatedAt:             raw.CreatedAt,
		UpdatedAt:             raw.UpdatedAt,
		ModelName:             raw.ModelName,
		ModelKind:             raw.ModelKind,
		CalcuttaName:          raw.CalcuttaName,
		NEvaluations:          raw.NEvaluations,
		HasPredictions:        raw.HasPredictions,
	}

	teamExpectedPoints := buildTeamExpectedPoints(raw)
	totalExpectedPoints := sumValues(teamExpectedPoints)

	// Calculate total budget from our bids (for naive allocation comparison).
	totalBudget := 0
	for _, b := range raw.Bids {
		totalBudget += b.BidPoints
	}

	result.Bids = enrichBids(raw, teamExpectedPoints, totalExpectedPoints, totalBudget)

	if raw.HasPredictions {
		result.Predictions = enrichPredictions(raw, teamExpectedPoints, totalExpectedPoints)
	}

	return result
}

// buildTeamExpectedPoints returns expected points per team, using predictions
// when available and falling back to seed-based estimates otherwise.
func buildTeamExpectedPoints(raw *models.LabEntryRaw) map[string]float64 {
	m := make(map[string]float64, len(raw.Teams))
	if raw.HasPredictions {
		for _, p := range raw.Predictions {
			m[p.TeamID] = p.ExpectedPoints
		}
		return m
	}

	// Fallback: seed-based estimates (only used when no predictions exist).
	for tid, ti := range raw.Teams {
		m[tid] = SeedExpectedPoints[ti.Seed]
	}
	return m
}

// sumValues returns the sum of all values in the map.
func sumValues(m map[string]float64) float64 {
	total := 0.0
	for _, v := range m {
		total += v
	}
	return total
}

// enrichBids builds enriched bids for ALL teams (not just those with bids > 0),
// computing naive allocation and edge percent for each.
func enrichBids(raw *models.LabEntryRaw, teamExpectedPoints map[string]float64, totalExpectedPoints float64, totalBudget int) []models.LabEnrichedBid {
	// Index raw bids by team ID.
	bidPointsByTeam := make(map[string]int, len(raw.Bids))
	expectedROIByTeam := make(map[string]*float64, len(raw.Bids))
	for _, b := range raw.Bids {
		bidPointsByTeam[b.TeamID] = b.BidPoints
		expectedROIByTeam[b.TeamID] = b.ExpectedROI
	}

	enriched := make([]models.LabEnrichedBid, 0, len(raw.Teams))
	for tid, ti := range raw.Teams {
		bidPoints := bidPointsByTeam[tid] // 0 if not in map

		// Naive allocation: team's expected points / total expected points * budget
		naiveShare := teamExpectedPoints[tid] / totalExpectedPoints
		naivePoints := int(naiveShare * float64(totalBudget))

		// Edge: (naive - bid) / naive * 100 (positive = undervalued opportunity)
		edgePercent := 0.0
		if naivePoints > 0 {
			edgePercent = float64(naivePoints-bidPoints) / float64(naivePoints) * 100
		}

		enriched = append(enriched, models.LabEnrichedBid{
			TeamID:      tid,
			SchoolName:  ti.Name,
			Seed:        ti.Seed,
			Region:      ti.Region,
			BidPoints:   bidPoints,
			NaivePoints: naivePoints,
			EdgePercent: edgePercent,
			ExpectedROI: expectedROIByTeam[tid],
		})
	}
	return enriched
}

// enrichPredictions builds enriched predictions with predicted bid points,
// expected ROI, naive allocation, and edge percent.
func enrichPredictions(raw *models.LabEntryRaw, teamExpectedPoints map[string]float64, totalExpectedPoints float64) []models.LabEnrichedPrediction {
	enriched := make([]models.LabEnrichedPrediction, 0, len(raw.Predictions))
	for _, p := range raw.Predictions {
		ti, ok := raw.Teams[p.TeamID]
		if !ok {
			continue
		}

		// Convert market share to bid points (using total POOL budget, not our single entry budget).
		predictedBidPoints := int(p.PredictedMarketShare * float64(raw.TotalPoolBudget))

		// Expected ROI = expected points / predicted market bid
		expectedROI := 0.0
		if predictedBidPoints > 0 {
			expectedROI = p.ExpectedPoints / float64(predictedBidPoints)
		}

		// Naive allocation for comparison (also uses total pool budget for predictions).
		naiveShare := teamExpectedPoints[p.TeamID] / totalExpectedPoints
		naivePoints := int(naiveShare * float64(raw.TotalPoolBudget))

		// Edge: (naive - predicted) / naive * 100
		edgePercent := 0.0
		if naivePoints > 0 {
			edgePercent = float64(naivePoints-predictedBidPoints) / float64(naivePoints) * 100
		}

		enriched = append(enriched, models.LabEnrichedPrediction{
			TeamID:               p.TeamID,
			SchoolName:           ti.Name,
			Seed:                 ti.Seed,
			Region:               ti.Region,
			PredictedMarketShare: p.PredictedMarketShare,
			PredictedBidPoints:   predictedBidPoints,
			ExpectedPoints:       p.ExpectedPoints,
			ExpectedROI:          expectedROI,
			NaivePoints:          naivePoints,
			EdgePercent:          edgePercent,
		})
	}
	return enriched
}
