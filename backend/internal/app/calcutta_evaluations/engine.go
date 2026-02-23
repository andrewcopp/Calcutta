package calcutta_evaluations

import (
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation"
)

// SimulationResult represents the outcome of one simulation
type SimulationResult struct {
	SimID            int
	EntryName        string
	TotalPoints      float64
	Rank             int
	PayoutCents      int
	NormalizedPayout float64
}

// EntryPerformance represents aggregated performance metrics for an entry
type EntryPerformance struct {
	EntryName    string
	MeanPayout   float64
	MedianPayout float64
	PTop1        float64
	PInMoney     float64
	TotalSims    int
}

// Entry represents an entry with their team bids
type Entry struct {
	Name  string
	Teams map[string]int // team_id -> bid_points
}

// TeamSimResult represents a team's result in one simulation
type TeamSimResult struct {
	TeamID string
	Points int
}

// CalculateSimulationOutcomes computes each entry's total points, rank, and
// payout for a single simulation. It is a pure function with no side effects.
func CalculateSimulationOutcomes(simID int, entries map[string]*Entry, teamResults []TeamSimResult, payouts map[int]int, firstPlacePayout int) ([]SimulationResult, error) {

	// Build team points map for this simulation
	teamPoints := make(map[string]int)
	for _, tr := range teamResults {
		teamPoints[tr.TeamID] = tr.Points
	}

	// Calculate total bids per team across all entries
	totalBidsPerTeam := make(map[string]int)
	for _, entry := range entries {
		for teamID, bidPoints := range entry.Teams {
			totalBidsPerTeam[teamID] += bidPoints
		}
	}

	// Calculate total points for each entry
	type entryScore struct {
		name   string
		points float64
	}

	var scores []entryScore
	for _, entry := range entries {
		totalPoints := 0.0
		for teamID, bidPoints := range entry.Teams {
			if points, ok := teamPoints[teamID]; ok {
				totalBids := totalBidsPerTeam[teamID]
				if totalBids > 0 {
					// Proportional ownership: (my_bid / total_bids) * team_points
					totalPoints += float64(points) * float64(bidPoints) / float64(totalBids)
				}
			}
		}
		scores = append(scores, entryScore{name: entry.Name, points: totalPoints})
	}

	// Sort by points descending to determine ranks
	// Use stable tie-breaking by name to ensure deterministic results
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].points != scores[j].points {
			return scores[i].points > scores[j].points
		}
		return scores[i].name < scores[j].name
	})

	// Assign ranks and payouts
	results := make([]SimulationResult, len(scores))
	for i, score := range scores {
		rank := i + 1

		// Get payout from actual payout structure
		payoutCents := 0
		if amount, ok := payouts[rank]; ok {
			payoutCents = amount
		}

		// Normalize by first place payout
		normalizedPayout := 0.0
		if firstPlacePayout > 0 {
			normalizedPayout = float64(payoutCents) / float64(firstPlacePayout)
		}

		results[i] = SimulationResult{
			SimID:            simID,
			EntryName:        score.name,
			TotalPoints:      score.points,
			Rank:             rank,
			PayoutCents:      payoutCents,
			NormalizedPayout: normalizedPayout,
		}
	}

	return results, nil
}

// ConvertSimulationResults bridges simulation output to evaluation input by
// grouping TeamSimulationResult records by SimID and converting each team's
// wins/byes into points using the provided scoring rules.
func ConvertSimulationResults(
	simResults []simulation.TeamSimulationResult,
	nTeams int,
	rules []scoring.Rule,
) map[int][]TeamSimResult {
	out := make(map[int][]TeamSimResult)
	for _, sr := range simResults {
		points := scoring.PointsForProgress(rules, sr.Wins, sr.Byes)
		out[sr.SimID] = append(out[sr.SimID], TeamSimResult{
			TeamID: sr.TeamID,
			Points: points,
		})
	}
	return out
}

// CalculatePerformanceMetrics aggregates simulation results into per-entry
// performance statistics (mean/median payout, P(top1), P(in money)).
func CalculatePerformanceMetrics(results []SimulationResult) map[string]*EntryPerformance {
	entryPayouts := make(map[string][]float64)

	for _, r := range results {
		entryPayouts[r.EntryName] = append(entryPayouts[r.EntryName], r.NormalizedPayout)
	}

	performance := make(map[string]*EntryPerformance)
	for entryName, payouts := range entryPayouts {
		sort.Float64s(payouts)

		// Calculate metrics
		var sum float64
		var top1Count, inMoneyCount int
		for _, payout := range payouts {
			sum += payout
			if payout > 0 {
				inMoneyCount++
			}
			// Top 1 is when normalized payout is 1.0 (first place)
			if payout >= 1.0 {
				top1Count++
			}
		}

		mean := sum / float64(len(payouts))
		median := payouts[len(payouts)/2]

		performance[entryName] = &EntryPerformance{
			EntryName:    entryName,
			MeanPayout:   mean,
			MedianPayout: median,
			PTop1:        float64(top1Count) / float64(len(payouts)),
			PInMoney:     float64(inMoneyCount) / float64(len(payouts)),
			TotalSims:    len(payouts),
		}
	}

	return performance
}
