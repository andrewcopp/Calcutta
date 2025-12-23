package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
)

type entryResult struct {
	Name   string
	Points float64
}

type simulatedLine struct {
	SchoolName     string
	Seed           int
	Bid            int
	Ownership      float64
	TeamPoints     float64
	RealizedPoints float64
}

type predictedLine struct {
	SchoolName         string
	Seed               int
	Region             string
	PredMarketBid      float64
	ActualMarketBid    float64
	PredPoints         float64
	ActualPoints       float64
	RecommendedBid     int
	PredOwnershipAfter float64
	RealOwnershipAfter float64
	PredInvestment     float64
	PredReturn         float64
	PredROI            float64
	RealizedReturn     float64
	RealizedROI        float64
	MinBidReturn       float64
	MinBidROI          float64
	MaxBidReturn       float64
	MaxBidROI          float64
}

func ordinal(n int) string {
	if n%100 >= 11 && n%100 <= 13 {
		return fmt.Sprintf("%dth", n)
	}
	switch n % 10 {
	case 1:
		return fmt.Sprintf("%dst", n)
	case 2:
		return fmt.Sprintf("%dnd", n)
	case 3:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}

func runReport(ctx context.Context, db *sql.DB, w io.Writer, startYear int, endYear int, trainYears int, excludeEntryName string, budget int, minTeams int, maxTeams int, minBid int, maxBid int, predModel string, investModel string, sigma float64) error {
	for y := startYear; y <= endYear; y++ {
		calcuttaID, err := resolveSingleCalcuttaIDForYear(ctx, db, y)
		if err != nil {
			if errors.Is(err, ErrNoCalcuttaForYear) {
				continue
			}
			return err
		}

		datasetRows, err := queryTeamDataset(ctx, db, 0, calcuttaID, excludeEntryName)
		if err != nil {
			return err
		}

		predPointsByTeam, err := predictedPointsByTeam(ctx, db, calcuttaID, datasetRows, trainYears, predModel, sigma)
		if err != nil {
			if errors.Is(err, ErrNoTrainingData) {
				continue
			}
			return err
		}

		predMarketBidByTeam, _, predTotalMarketBid, err := predictedMarketBidsByTeam(ctx, db, calcuttaID, datasetRows, trainYears, investModel, excludeEntryName)
		if err != nil {
			if errors.Is(err, ErrNoTrainingData) {
				continue
			}
			return err
		}

		totalPredPoints := 0.0
		for _, r := range datasetRows {
			totalPredPoints += predPointsByTeam[r.TeamID]
		}

		predPointsVec := make([]float64, 0, len(datasetRows))
		baseBidsVec := make([]float64, 0, len(datasetRows))
		teamIDs := make([]string, 0, len(datasetRows))
		for _, r := range datasetRows {
			baseBid := predMarketBidByTeam[r.TeamID]
			teamIDs = append(teamIDs, r.TeamID)
			predPointsVec = append(predPointsVec, predPointsByTeam[r.TeamID])
			baseBidsVec = append(baseBidsVec, baseBid)
		}

		bidsVec, _, err := dpAllocateBids(predPointsVec, baseBidsVec, budget, minTeams, maxTeams, minBid, maxBid)
		if err != nil {
			return err
		}

		teamPoints := map[string]float64{}
		actualBaseBid := map[string]float64{}
		actualTotalMarketBid := 0.0
		for _, r := range datasetRows {
			teamPoints[r.TeamID] = r.TeamPoints
			if excludeEntryName != "" {
				actualBaseBid[r.TeamID] = r.TotalCommunityBidExcl
			} else {
				actualBaseBid[r.TeamID] = r.TotalCommunityBid
			}
		}
		if len(datasetRows) > 0 {
			if excludeEntryName != "" {
				actualTotalMarketBid = datasetRows[0].CalcuttaTotalExcl
			} else {
				actualTotalMarketBid = datasetRows[0].CalcuttaTotalCommunity
			}
		}

		botBid := map[string]float64{}
		for i, teamID := range teamIDs {
			botBid[teamID] = float64(bidsVec[i])
		}

		lines := make([]simulatedLine, 0)
		for _, r := range datasetRows {
			bid := botBid[r.TeamID]
			if bid <= 0 {
				continue
			}
			base := actualBaseBid[r.TeamID]
			den := base + bid
			own := 0.0
			if den > 0 {
				own = bid / den
			}
			tp := teamPoints[r.TeamID]
			realized := tp * own
			lines = append(lines, simulatedLine{
				SchoolName:     r.SchoolName,
				Seed:           r.Seed,
				Bid:            int(math.Round(bid)),
				Ownership:      own,
				TeamPoints:     tp,
				RealizedPoints: realized,
			})
		}

		sort.Slice(lines, func(i, j int) bool {
			if lines[i].Bid == lines[j].Bid {
				return lines[i].SchoolName < lines[j].SchoolName
			}
			return lines[i].Bid > lines[j].Bid
		})

		predLines := make([]predictedLine, 0, len(datasetRows))
		for _, r := range datasetRows {
			predBase := predMarketBidByTeam[r.TeamID]
			actBase := actualBaseBid[r.TeamID]
			predPts := predPointsByTeam[r.TeamID]
			actPts := teamPoints[r.TeamID]
			bid := botBid[r.TeamID]
			predDen := predBase + bid
			predOwn := 0.0
			if predDen > 0 {
				predOwn = bid / predDen
			}
			predReturn := predPts * predOwn
			predInvestment := bid
			predROI := 0.0
			if bid > 0 && predDen > 0 {
				predROI = predReturn / bid
			}

			realDen := actBase + bid
			realOwn := 0.0
			if realDen > 0 {
				realOwn = bid / realDen
			}
			realRet := actPts * realOwn
			realROI := 0.0
			if bid > 0 && realDen > 0 {
				realROI = realRet / bid
			}

			minDen := predBase + float64(minBid)
			minOwn := 0.0
			if minDen > 0 {
				minOwn = float64(minBid) / minDen
			}
			minRet := predPts * minOwn
			minROI := 0.0
			if minBid > 0 && minDen > 0 {
				minROI = minRet / float64(minBid)
			}

			maxDen := predBase + float64(maxBid)
			maxOwn := 0.0
			if maxDen > 0 {
				maxOwn = float64(maxBid) / maxDen
			}
			maxRet := predPts * maxOwn
			maxROI := 0.0
			if maxBid > 0 && maxDen > 0 {
				maxROI = maxRet / float64(maxBid)
			}

			predLines = append(predLines, predictedLine{
				SchoolName:         r.SchoolName,
				Seed:               r.Seed,
				Region:             r.Region,
				PredMarketBid:      predBase,
				ActualMarketBid:    actBase,
				PredPoints:         predPts,
				ActualPoints:       actPts,
				RecommendedBid:     int(math.Round(bid)),
				PredOwnershipAfter: predOwn,
				RealOwnershipAfter: realOwn,
				PredInvestment:     predInvestment,
				PredReturn:         predReturn,
				PredROI:            predROI,
				RealizedReturn:     realRet,
				RealizedROI:        realROI,
				MinBidReturn:       minRet,
				MinBidROI:          minROI,
				MaxBidReturn:       maxRet,
				MaxBidROI:          maxROI,
			})
		}

		sort.Slice(predLines, func(i, j int) bool {
			if predLines[i].Seed != predLines[j].Seed {
				return predLines[i].Seed < predLines[j].Seed
			}
			if predLines[i].RecommendedBid != predLines[j].RecommendedBid {
				return predLines[i].RecommendedBid > predLines[j].RecommendedBid
			}
			return predLines[i].SchoolName < predLines[j].SchoolName
		})

		bids, err := queryEntryBids(ctx, db, calcuttaID, excludeEntryName)
		if err != nil {
			return err
		}

		entryPoints := map[string]float64{}
		for _, b := range bids {
			den := actualBaseBid[b.TeamID] + botBid[b.TeamID]
			if den <= 0 {
				continue
			}
			entryPoints[b.EntryName] += teamPoints[b.TeamID] * (b.Bid / den)
		}

		botPoints := 0.0
		for teamID, bid := range botBid {
			den := actualBaseBid[teamID] + bid
			if den <= 0 {
				continue
			}
			botPoints += teamPoints[teamID] * (bid / den)
		}

		results := make([]entryResult, 0, len(entryPoints)+1)
		for name, pts := range entryPoints {
			results = append(results, entryResult{Name: name, Points: pts})
		}
		botName := "Simulated Entry"
		results = append(results, entryResult{Name: botName, Points: botPoints})

		sort.Slice(results, func(i, j int) bool {
			if results[i].Points == results[j].Points {
				return results[i].Name < results[j].Name
			}
			return results[i].Points > results[j].Points
		})

		rank := 0
		for i := range results {
			if results[i].Name == botName {
				rank = i + 1
				break
			}
		}

		fmt.Fprintf(w, "## %d\n\n", y)
		fmt.Fprintf(w, "Simulated entry (pred_model=%s, sigma=%.2f, budget=%d, teams=%d):\n\n", predModel, sigma, budget, len(lines))
		fmt.Fprintln(w, "Step 1) Predicted returns (points) for all teams")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "seed | team | region | pred_points | pred_points_share")
		fmt.Fprintln(w, "--- | --- | --- | ---: | ---:")
		for _, pl := range predLines {
			share := 0.0
			if totalPredPoints > 0 {
				share = pl.PredPoints / totalPredPoints
			}
			fmt.Fprintf(w, "%d | %s | %s | %.2f | %.4f\n", pl.Seed, pl.SchoolName, pl.Region, pl.PredPoints, share)
		}
		fmt.Fprintln(w, "")

		fmt.Fprintln(w, "Step 2) Predicted investments (market bids) for all teams")
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "pred_total_market_bid (train-only estimate): %.0f\n\n", predTotalMarketBid)
		fmt.Fprintln(w, "seed | team | region | pred_market_bid | pred_market_bid_share | implied_points_per_dollar")
		fmt.Fprintln(w, "--- | --- | --- | ---: | ---: | ---:")
		for _, pl := range predLines {
			share := 0.0
			if predTotalMarketBid > 0 {
				share = pl.PredMarketBid / predTotalMarketBid
			}
			implied := 0.0
			if pl.PredMarketBid > 0 {
				implied = pl.PredPoints / pl.PredMarketBid
			}
			fmt.Fprintf(w, "%d | %s | %s | %.0f | %.4f | %.4f\n", pl.Seed, pl.SchoolName, pl.Region, pl.PredMarketBid, share, implied)
		}
		fmt.Fprintln(w, "")

		fmt.Fprintln(w, "Step 3) Simulated entry selection (using predicted ROI / marginal ROI)")
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "Ownership model: own(x) = x/(PredMarketBid+x)\n")
		fmt.Fprintf(w, "Expected team contribution: PredPoints * own(x)\n")
		fmt.Fprintf(w, "Objective: maximize sum of expected contributions subject to budget=%d, teams in [%d,%d], bid in [%d,%d]\n\n", budget, minTeams, maxTeams, minBid, maxBid)
		fmt.Fprintln(w, "Top opportunities at minimum bid (across all teams):")
		fmt.Fprintln(w, "")
		type minBidOpportunity struct {
			Seed          int
			SchoolName    string
			Region        string
			PredMarketBid float64
			PredPoints    float64
			ROIMinBid     float64
			SelectedBid   int
		}
		minOpps := make([]minBidOpportunity, 0, len(predLines))
		for _, pl := range predLines {
			den := pl.PredMarketBid + float64(minBid)
			ret := 0.0
			if den > 0 {
				ret = pl.PredPoints * (float64(minBid) / den)
			}
			roi := 0.0
			if minBid > 0 {
				roi = ret / float64(minBid)
			}
			minOpps = append(minOpps, minBidOpportunity{
				Seed:          pl.Seed,
				SchoolName:    pl.SchoolName,
				Region:        pl.Region,
				PredMarketBid: pl.PredMarketBid,
				PredPoints:    pl.PredPoints,
				ROIMinBid:     roi,
				SelectedBid:   pl.RecommendedBid,
			})
		}
		sort.Slice(minOpps, func(i, j int) bool {
			if minOpps[i].ROIMinBid == minOpps[j].ROIMinBid {
				return minOpps[i].SchoolName < minOpps[j].SchoolName
			}
			return minOpps[i].ROIMinBid > minOpps[j].ROIMinBid
		})
		topK := 20
		if len(minOpps) < topK {
			topK = len(minOpps)
		}
		fmt.Fprintln(w, "seed | team | region | pred_market_bid | pred_points | roi@min_bid | selected_bid")
		fmt.Fprintln(w, "--- | --- | --- | ---: | ---: | ---: | ---:")
		for i := 0; i < topK; i++ {
			op := minOpps[i]
			fmt.Fprintf(w, "%d | %s | %s | %.0f | %.2f | %.4f | %d\n", op.Seed, op.SchoolName, op.Region, op.PredMarketBid, op.PredPoints, op.ROIMinBid, op.SelectedBid)
		}
		fmt.Fprintln(w, "")

		fmt.Fprintln(w, "Selected teams, with predicted ROI and marginal ROI at the last allocated dollar:")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "seed | team | region | pred_market_bid | bid | pred_own | pred_return | pred_roi | roi@min_bid | marginal_return_last_$ | marginal_roi_last_$")
		fmt.Fprintln(w, "--- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---:")
		for _, pl := range predLines {
			bid := pl.RecommendedBid
			if bid <= 0 {
				continue
			}
			base := pl.PredMarketBid
			predPts := pl.PredPoints

			own := 0.0
			if base+float64(bid) > 0 {
				own = float64(bid) / (base + float64(bid))
			}
			ret := predPts * own
			roi := 0.0
			if bid > 0 {
				roi = ret / float64(bid)
			}

			prevBid := bid - 1
			prevRet := 0.0
			if prevBid > 0 && base+float64(prevBid) > 0 {
				prevRet = predPts * (float64(prevBid) / (base + float64(prevBid)))
			}
			margRet := ret - prevRet
			margROI := margRet
			roiMinBid := 0.0
			if pl.PredMarketBid+float64(minBid) > 0 {
				roiMinBid = (pl.PredPoints * (float64(minBid) / (pl.PredMarketBid + float64(minBid)))) / float64(minBid)
			}

			fmt.Fprintf(w, "%d | %s | %s | %.0f | %d | %.2f%% | %.2f | %.4f | %.4f | %.4f | %.4f\n", pl.Seed, pl.SchoolName, pl.Region, base, bid, own*100, ret, roi, roiMinBid, margRet, margROI)
		}
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Selected teams (sorted by bid):")
		fmt.Fprintln(w, "")
		for i, r := range lines {
			fmt.Fprintf(w, "%d. %s (%d-seed) — bid=%d\n", i+1, r.SchoolName, r.Seed, r.Bid)
		}
		fmt.Fprintln(w, "")

		fmt.Fprintln(w, "Step 4) Simulated performance (realized outcomes)")
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "actual_total_market_bid: %.0f\n\n", actualTotalMarketBid)
		fmt.Fprintln(w, "seed | team | region | actual_market_bid | actual_points | bid | real_own | realized_return | realized_roi")
		fmt.Fprintln(w, "--- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---:")
		for _, pl := range predLines {
			if pl.RecommendedBid <= 0 {
				continue
			}
			fmt.Fprintf(w, "%d | %s | %s | %.0f | %.0f | %d | %.2f%% | %.2f | %.4f\n", pl.Seed, pl.SchoolName, pl.Region, pl.ActualMarketBid, pl.ActualPoints, pl.RecommendedBid, pl.RealOwnershipAfter*100, pl.RealizedReturn, pl.RealizedROI)
		}
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "\nWe would have finished %s place (%d/%d), scoring %.4f points.\n\n", ordinal(rank), rank, len(results), botPoints)

		topN := 3
		if len(results) < topN {
			topN = len(results)
		}
		fmt.Fprintf(w, "Top %d:\n\n", topN)
		for i := 0; i < topN; i++ {
			fmt.Fprintf(w, "%d. %s - %.4f\n", i+1, results[i].Name, results[i].Points)
		}
		fmt.Fprintln(w)
	}
	return nil
}
