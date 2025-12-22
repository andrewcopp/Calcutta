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
	SchoolName     string
	Seed           int
	Region         string
	BaseMarketBid  float64
	PredPoints     float64
	RecommendedBid int
	OwnershipAfter float64
	PredInvestment float64
	PredReturn     float64
	PredROI        float64
	MinBidReturn   float64
	MinBidROI      float64
	MaxBidReturn   float64
	MaxBidROI      float64
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

func runReport(ctx context.Context, db *sql.DB, w io.Writer, startYear int, endYear int, trainYears int, excludeEntryName string, budget int, minTeams int, maxTeams int, minBid int, maxBid int, predModel string, sigma float64) error {
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
			return err
		}

		predPointsVec := make([]float64, 0, len(datasetRows))
		baseBidsVec := make([]float64, 0, len(datasetRows))
		teamIDs := make([]string, 0, len(datasetRows))
		for _, r := range datasetRows {
			baseBid := r.TotalCommunityBid
			if excludeEntryName != "" {
				baseBid = r.TotalCommunityBidExcl
			}
			teamIDs = append(teamIDs, r.TeamID)
			predPointsVec = append(predPointsVec, predPointsByTeam[r.TeamID])
			baseBidsVec = append(baseBidsVec, baseBid)
		}

		bidsVec, _, err := dpAllocateBids(predPointsVec, baseBidsVec, budget, minTeams, maxTeams, minBid, maxBid)
		if err != nil {
			return err
		}

		teamPoints := map[string]float64{}
		teamBaseBid := map[string]float64{}
		for _, r := range datasetRows {
			teamPoints[r.TeamID] = r.TeamPoints
			if excludeEntryName != "" {
				teamBaseBid[r.TeamID] = r.TotalCommunityBidExcl
			} else {
				teamBaseBid[r.TeamID] = r.TotalCommunityBid
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
			base := teamBaseBid[r.TeamID]
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
			base := teamBaseBid[r.TeamID]
			predPts := predPointsByTeam[r.TeamID]
			bid := botBid[r.TeamID]
			den := base + bid
			own := 0.0
			if den > 0 {
				own = bid / den
			}
			predReturn := predPts * own
			predInvestment := bid
			predROI := 0.0
			if bid > 0 && den > 0 {
				predROI = predReturn / bid
			}

			minDen := base + float64(minBid)
			minOwn := 0.0
			if minDen > 0 {
				minOwn = float64(minBid) / minDen
			}
			minRet := predPts * minOwn
			minROI := 0.0
			if minBid > 0 && minDen > 0 {
				minROI = minRet / float64(minBid)
			}

			maxDen := base + float64(maxBid)
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
				SchoolName:     r.SchoolName,
				Seed:           r.Seed,
				Region:         r.Region,
				BaseMarketBid:  base,
				PredPoints:     predPts,
				RecommendedBid: int(math.Round(bid)),
				OwnershipAfter: own,
				PredInvestment: predInvestment,
				PredReturn:     predReturn,
				PredROI:        predROI,
				MinBidReturn:   minRet,
				MinBidROI:      minROI,
				MaxBidReturn:   maxRet,
				MaxBidROI:      maxROI,
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
			den := teamBaseBid[b.TeamID] + botBid[b.TeamID]
			if den <= 0 {
				continue
			}
			entryPoints[b.EntryName] += teamPoints[b.TeamID] * (b.Bid / den)
		}

		botPoints := 0.0
		for teamID, bid := range botBid {
			den := teamBaseBid[teamID] + bid
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
		fmt.Fprintln(w, "How the simulated entry is computed:")
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "We estimate each team’s expected points (PredPoints). For a hypothetical bid x against the market’s existing investment (BaseMarketBid), we model our ownership as x/(BaseMarketBid+x), so the expected points we get from that team is PredPoints * x/(BaseMarketBid+x).\n\n")
		fmt.Fprintf(w, "The optimizer chooses bids (integer dollars) to maximize the sum of expected points across teams, subject to: total budget=%d, team count in [%d,%d], per-team bid in [%d,%d].\n\n", budget, minTeams, maxTeams, minBid, maxBid)
		fmt.Fprintln(w, "Recommended bids (teams we actually invest in):")
		fmt.Fprintln(w, "")
		for i, r := range lines {
			fmt.Fprintf(w, "%d. %s (%d-seed) — bid=%d, own=%.2f%%, team_points=%.0f, realized=%.2f\n", i+1, r.SchoolName, r.Seed, r.Bid, r.Ownership*100, r.TeamPoints, r.RealizedPoints)
		}
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Per-team predicted economics (all teams in the field):")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "seed | team | region | base_market_bid | pred_points | bid | pred_investment | ownership_after | pred_return | pred_roi | return@min_bid | roi@min_bid | return@max_bid | roi@max_bid")
		fmt.Fprintln(w, "--- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---:")
		for _, pl := range predLines {
			fmt.Fprintf(
				w,
				"%d | %s | %s | %.0f | %.2f | %d | %.0f | %.2f%% | %.2f | %.4f | %.2f | %.4f | %.2f | %.4f\n",
				pl.Seed,
				pl.SchoolName,
				pl.Region,
				pl.BaseMarketBid,
				pl.PredPoints,
				pl.RecommendedBid,
				pl.PredInvestment,
				pl.OwnershipAfter*100,
				pl.PredReturn,
				pl.PredROI,
				pl.MinBidReturn,
				pl.MinBidROI,
				pl.MaxBidReturn,
				pl.MaxBidROI,
			)
		}
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
