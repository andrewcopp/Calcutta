package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
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

func runReport(ctx context.Context, db *sql.DB, w io.Writer, startYear int, endYear int, trainYears int, excludeEntryName string, budget int, minTeams int, maxTeams int, minBid int, maxBid int) error {
	for y := startYear; y <= endYear; y++ {
		calcuttaID, err := resolveSingleCalcuttaIDForYear(ctx, db, y)
		if err != nil {
			if errors.Is(err, ErrNoCalcuttaForYear) {
				continue
			}
			return err
		}

		simRows, _, err := runSimulateEntry(ctx, db, calcuttaID, trainYears, excludeEntryName, budget, minTeams, maxTeams, minBid, maxBid)
		if err != nil {
			return err
		}

		datasetRows, err := queryTeamDataset(ctx, db, 0, calcuttaID, excludeEntryName)
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
		for _, r := range simRows {
			botBid[r.TeamID] = float64(r.RecommendedBid)
		}

		lines := make([]simulatedLine, 0, len(simRows))
		for _, r := range simRows {
			base := teamBaseBid[r.TeamID]
			bid := float64(r.RecommendedBid)
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
				Bid:            r.RecommendedBid,
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
		fmt.Fprintf(w, "Simulated entry (budget=%d, teams=%d):\n\n", budget, len(lines))
		for i, r := range lines {
			fmt.Fprintf(w, "%d. %s (%d-seed) — bid=%d, own=%.2f%%, team_points=%.0f, realized=%.2f\n", i+1, r.SchoolName, r.Seed, r.Bid, r.Ownership*100, r.TeamPoints, r.RealizedPoints)
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
