package main

import (
	"context"
	"database/sql"
	"fmt"
)

func runSimulateEntry(ctx context.Context, db *sql.DB, targetCalcuttaID string, trainYears int, excludeEntryName string, budget int, minTeams int, maxTeams int, minBid int, maxBid int, predModel string, sigma float64) ([]SimulateRow, *SimulateSummary, error) {
	if budget <= 0 {
		return nil, nil, fmt.Errorf("budget must be > 0")
	}
	if minTeams <= 0 || maxTeams <= 0 || minTeams > maxTeams {
		return nil, nil, fmt.Errorf("invalid team constraints: min_teams=%d max_teams=%d", minTeams, maxTeams)
	}
	if minBid <= 0 || maxBid <= 0 || minBid > maxBid {
		return nil, nil, fmt.Errorf("invalid bid constraints: min_bid=%d max_bid=%d", minBid, maxBid)
	}
	if budget < minTeams*minBid {
		return nil, nil, fmt.Errorf("budget too small to satisfy minimums: budget=%d min_teams=%d min_bid=%d", budget, minTeams, minBid)
	}
	if budget > maxTeams*maxBid {
		return nil, nil, fmt.Errorf("budget too large to satisfy maximums: budget=%d max_teams=%d max_bid=%d", budget, maxTeams, maxBid)
	}

	targetRows, err := queryTeamDataset(ctx, db, 0, targetCalcuttaID, excludeEntryName)
	if err != nil {
		return nil, nil, err
	}

	predPointsByTeam, err := predictedPointsByTeam(ctx, db, targetCalcuttaID, targetRows, trainYears, predModel, sigma)
	if err != nil {
		return nil, nil, err
	}

	totalPredPoints := 0.0
	for _, r := range targetRows {
		totalPredPoints += predPointsByTeam[r.TeamID]
	}

	totalMarketBid := 0.0
	if len(targetRows) > 0 {
		if excludeEntryName != "" {
			totalMarketBid = targetRows[0].CalcuttaTotalExcl
		} else {
			totalMarketBid = targetRows[0].CalcuttaTotalCommunity
		}
	}

	type candidate struct {
		idx        int
		predPoints float64
		baseBid    float64
		bid        int
		seed       int
	}

	candidates := make([]candidate, 0, len(targetRows))
	for i, r := range targetRows {
		baseBid := r.TotalCommunityBid
		if excludeEntryName != "" {
			baseBid = r.TotalCommunityBidExcl
		}
		candidates = append(candidates, candidate{
			idx:        i,
			predPoints: predPointsByTeam[r.TeamID],
			baseBid:    baseBid,
			bid:        0,
			seed:       r.Seed,
		})
	}

	predPointsVec := make([]float64, len(candidates))
	baseBidsVec := make([]float64, len(candidates))
	for i := range candidates {
		predPointsVec[i] = candidates[i].predPoints
		baseBidsVec[i] = candidates[i].baseBid
	}

	bids, selected, err := dpAllocateBids(predPointsVec, baseBidsVec, budget, minTeams, maxTeams, minBid, maxBid)
	if err != nil {
		return nil, nil, err
	}
	for i := range candidates {
		candidates[i].bid = bids[i]
	}

	byIdx := make([]candidate, len(candidates))
	for _, c := range candidates {
		byIdx[c.idx] = c
	}

	simRows := make([]SimulateRow, 0, selected)
	expectedEntryPointsTotal := 0.0
	for i, r := range targetRows {
		c := byIdx[i]
		if c.bid <= 0 {
			continue
		}
		baseBid := c.baseBid
		predPoints := c.predPoints
		predPointsShare := 0.0
		if totalPredPoints > 0 {
			predPointsShare = predPoints / totalPredPoints
		}
		baseShare := 0.0
		if excludeEntryName != "" {
			baseShare = r.NormalizedBidExcl
		} else {
			baseShare = r.NormalizedBid
		}
		ownership := 0.0
		if c.bid > 0 {
			ownership = float64(c.bid) / (baseBid + float64(c.bid))
		}
		expectedEntryPoints := predPoints * ownership
		expectedEntryPointsTotal += expectedEntryPoints
		simRows = append(simRows, SimulateRow{
			TournamentName:      r.TournamentName,
			TournamentYear:      r.TournamentYear,
			CalcuttaID:          r.CalcuttaID,
			PredModel:           predModel,
			Sigma:               sigma,
			TeamID:              r.TeamID,
			SchoolName:          r.SchoolName,
			Seed:                r.Seed,
			Region:              r.Region,
			BaseMarketBid:       baseBid,
			BaseMarketBidShare:  baseShare,
			PredPoints:          predPoints,
			PredPointsShare:     predPointsShare,
			RecommendedBid:      c.bid,
			OwnershipAfter:      ownership,
			ExpectedEntryPoints: expectedEntryPoints,
		})
	}

	pointsShare := 0.0
	if totalPredPoints > 0 {
		pointsShare = expectedEntryPointsTotal / totalPredPoints
	}
	bidShare := 0.0
	if totalMarketBid+float64(budget) > 0 {
		bidShare = float64(budget) / (totalMarketBid + float64(budget))
	}
	normROI := 0.0
	if bidShare > 0 {
		normROI = pointsShare / bidShare
	}

	summary := &SimulateSummary{
		Budget:                budget,
		NumTeams:              selected,
		ExpectedPointsShare:   pointsShare,
		ExpectedBidShare:      bidShare,
		ExpectedNormalizedROI: normROI,
	}

	return simRows, summary, nil
}
