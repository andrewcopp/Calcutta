package simulated_calcutta

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service handles simulated calcutta analysis
type Service struct {
	pool *pgxpool.Pool
}

// New creates a new simulated calcutta service
func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// SimulationResult represents the outcome of one simulation
type SimulationResult struct {
	SimID        int
	EntryName    string
	TotalPoints  int
	Rank         int
	PayoutPoints int
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

// CalculateSimulatedCalcutta calculates entry outcomes for all simulations
func (s *Service) CalculateSimulatedCalcutta(ctx context.Context, tournamentID string, runID string) error {
	excludedEntry := os.Getenv("EXCLUDED_ENTRY_NAME")
	if excludedEntry == "" {
		excludedEntry = "Andrew Copp" // Default
	}

	log.Printf("Calculating simulated calcutta for tournament %s, run %s", tournamentID, runID)
	log.Printf("Excluding entry: %s", excludedEntry)

	// Get all entries and their bids
	entries, err := s.getEntries(ctx, tournamentID, runID, excludedEntry)
	if err != nil {
		return fmt.Errorf("failed to get entries: %w", err)
	}

	log.Printf("Found %d entries (excluding %s)", len(entries), excludedEntry)

	// Get all simulations
	simulations, err := s.getSimulations(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("failed to get simulations: %w", err)
	}

	log.Printf("Found %d simulations", len(simulations))

	// Calculate outcomes for each simulation in parallel
	results := make(chan []SimulationResult, len(simulations))
	errors := make(chan error, len(simulations))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit to 10 concurrent goroutines

	for simID := range simulations {
		wg.Add(1)
		go func(sid int) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			simResults, err := s.calculateSimulationOutcomes(ctx, sid, entries, simulations[sid])
			if err != nil {
				errors <- fmt.Errorf("simulation %d: %w", sid, err)
				return
			}
			results <- simResults
		}(simID)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Collect results
	var allResults []SimulationResult
	for simResults := range results {
		allResults = append(allResults, simResults...)
	}

	// Check for errors
	for err := range errors {
		if err != nil {
			return err
		}
	}

	log.Printf("Calculated %d total outcomes", len(allResults))

	// Write results to database
	if err := s.writeSimulationOutcomes(ctx, runID, allResults); err != nil {
		return fmt.Errorf("failed to write simulation outcomes: %w", err)
	}

	// Calculate and write aggregated performance metrics
	performance := s.calculatePerformanceMetrics(allResults)
	if err := s.writePerformanceMetrics(ctx, runID, performance); err != nil {
		return fmt.Errorf("failed to write performance metrics: %w", err)
	}

	log.Printf("Successfully calculated simulated calcutta for %d entries", len(entries))
	return nil
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

func (s *Service) getEntries(ctx context.Context, tournamentID string, runID string, excludedEntry string) (map[string]*Entry, error) {
	// Get actual entries from bronze_entry_bids
	query := `
		SELECT DISTINCT eb.entry_name, eb.team_id, eb.bid_amount_points
		FROM bronze_entry_bids eb
		JOIN bronze_calcuttas c ON eb.calcutta_id = c.id
		WHERE c.tournament_id = $1
		  AND eb.entry_name != $2
	`

	rows, err := s.pool.Query(ctx, query, tournamentID, excludedEntry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make(map[string]*Entry)
	for rows.Next() {
		var entryName, teamID string
		var bidPoints int
		if err := rows.Scan(&entryName, &teamID, &bidPoints); err != nil {
			return nil, err
		}

		if entries[entryName] == nil {
			entries[entryName] = &Entry{
				Name:  entryName,
				Teams: make(map[string]int),
			}
		}
		entries[entryName].Teams[teamID] = bidPoints
	}

	// Add our simulated entry from gold_recommended_entry_bids
	ourQuery := `
		SELECT team_id, recommended_bid_points
		FROM gold_recommended_entry_bids
		WHERE run_id = $1
	`

	ourRows, err := s.pool.Query(ctx, ourQuery, runID)
	if err != nil {
		return nil, err
	}
	defer ourRows.Close()

	ourEntry := &Entry{
		Name:  "Our Strategy",
		Teams: make(map[string]int),
	}

	for ourRows.Next() {
		var teamID string
		var bidPoints int
		if err := ourRows.Scan(&teamID, &bidPoints); err != nil {
			return nil, err
		}
		ourEntry.Teams[teamID] = bidPoints
	}

	if len(ourEntry.Teams) > 0 {
		entries["Our Strategy"] = ourEntry
	}

	return entries, nil
}

func (s *Service) getSimulations(ctx context.Context, tournamentID string) (map[int][]TeamSimResult, error) {
	query := `
		SELECT sim_id, team_id, wins
		FROM silver_simulated_tournaments
		WHERE tournament_id = $1
		ORDER BY sim_id, team_id
	`

	rows, err := s.pool.Query(ctx, query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	simulations := make(map[int][]TeamSimResult)
	pointsPerWin := map[int]int{0: 0, 1: 50, 2: 150, 3: 300, 4: 500, 5: 750, 6: 1050}

	for rows.Next() {
		var simID, wins int
		var teamID string
		if err := rows.Scan(&simID, &teamID, &wins); err != nil {
			return nil, err
		}

		points := pointsPerWin[wins]
		simulations[simID] = append(simulations[simID], TeamSimResult{
			TeamID: teamID,
			Points: points,
		})
	}

	return simulations, nil
}

func (s *Service) calculateSimulationOutcomes(ctx context.Context, simID int, entries map[string]*Entry, teamResults []TeamSimResult) ([]SimulationResult, error) {
	// Build team points map for this simulation
	teamPoints := make(map[string]int)
	for _, tr := range teamResults {
		teamPoints[tr.TeamID] = tr.Points
	}

	// Calculate total points for each entry
	type entryScore struct {
		name   string
		points int
	}

	var scores []entryScore
	for _, entry := range entries {
		totalPoints := 0
		for teamID, bidPoints := range entry.Teams {
			if points, ok := teamPoints[teamID]; ok {
				totalPoints += points * bidPoints / 100 // Proportional ownership
			}
		}
		scores = append(scores, entryScore{name: entry.Name, points: totalPoints})
	}

	// Sort by points descending to determine ranks
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].points > scores[j].points
	})

	// Assign ranks and payouts
	results := make([]SimulationResult, len(scores))
	for i, score := range scores {
		rank := i + 1
		payout := s.calculatePayout(rank, len(scores))

		results[i] = SimulationResult{
			SimID:        simID,
			EntryName:    score.name,
			TotalPoints:  score.points,
			Rank:         rank,
			PayoutPoints: payout,
		}
	}

	return results, nil
}

func (s *Service) calculatePayout(rank int, totalEntries int) int {
	// Simple payout structure: top 3 get paid
	// 1st: 50%, 2nd: 30%, 3rd: 20%
	totalPool := totalEntries * 100 // Each entry pays 100 points

	switch rank {
	case 1:
		return totalPool * 50 / 100
	case 2:
		return totalPool * 30 / 100
	case 3:
		return totalPool * 20 / 100
	default:
		return 0
	}
}

func (s *Service) writeSimulationOutcomes(ctx context.Context, runID string, results []SimulationResult) error {
	// Delete existing results for this run
	_, err := s.pool.Exec(ctx, "DELETE FROM gold_entry_simulation_outcomes WHERE run_id = $1", runID)
	if err != nil {
		return err
	}

	// Batch insert new results
	batch := &pgx.Batch{}
	for _, r := range results {
		batch.Queue(`
			INSERT INTO gold_entry_simulation_outcomes (run_id, entry_name, sim_id, payout_points, rank)
			VALUES ($1, $2, $3, $4, $5)
		`, runID, r.EntryName, r.SimID, r.PayoutPoints, r.Rank)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(results); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) calculatePerformanceMetrics(results []SimulationResult) map[string]*EntryPerformance {
	entryPayouts := make(map[string][]int)

	for _, r := range results {
		entryPayouts[r.EntryName] = append(entryPayouts[r.EntryName], r.PayoutPoints)
	}

	performance := make(map[string]*EntryPerformance)
	for entryName, payouts := range entryPayouts {
		sort.Ints(payouts)

		// Calculate metrics
		var sum int
		var top1Count, inMoneyCount int
		for _, payout := range payouts {
			sum += payout
			if payout > 0 {
				inMoneyCount++
			}
			// Top 1 is when payout is 50% of pool (assuming standard payout structure)
			if payout >= len(payouts)*100*50/100 {
				top1Count++
			}
		}

		mean := float64(sum) / float64(len(payouts))
		median := float64(payouts[len(payouts)/2])

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

func (s *Service) writePerformanceMetrics(ctx context.Context, runID string, performance map[string]*EntryPerformance) error {
	// Delete existing performance for this run
	_, err := s.pool.Exec(ctx, "DELETE FROM gold_entry_performance WHERE run_id = $1", runID)
	if err != nil {
		return err
	}

	// Insert new performance metrics
	for _, p := range performance {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO gold_entry_performance (run_id, entry_name, mean_payout, median_payout, p_top1, p_in_money)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, runID, p.EntryName, p.MeanPayout, p.MedianPayout, p.PTop1, p.PInMoney)
		if err != nil {
			return err
		}
	}

	return nil
}
