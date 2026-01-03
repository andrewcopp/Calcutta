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

func (s *Service) getLatestTournamentSimulationBatchID(ctx context.Context, bronzeTournamentID string, coreTournamentID string) (string, bool, error) {
	var batchID string
	err := s.pool.QueryRow(ctx, `
		SELECT b.id
		FROM analytics.tournament_simulation_batches b
		WHERE b.tournament_id = $2
			AND b.deleted_at IS NULL
			AND EXISTS (
				SELECT 1
				FROM analytics.simulated_tournaments st
				WHERE st.tournament_id = $1
					AND st.tournament_simulation_batch_id = b.id
					AND st.deleted_at IS NULL
			)
		ORDER BY b.created_at DESC
		LIMIT 1
	`, bronzeTournamentID, coreTournamentID).Scan(&batchID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", false, nil
		}
		return "", false, err
	}
	return batchID, true, nil
}

// New creates a new simulated calcutta service
func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

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

// CalculateSimulatedCalcutta calculates entry outcomes for all simulations
func (s *Service) CalculateSimulatedCalcutta(ctx context.Context, tournamentID string, runID string) error {
	// Get excluded entry name from environment (e.g., "Andrew Copp")
	excludedEntryName := os.Getenv("EXCLUDED_ENTRY_NAME")
	return s.calculateSimulatedCalcuttaInternal(ctx, tournamentID, runID, excludedEntryName, nil)
}

func (s *Service) CalculateSimulatedCalcuttaForEvaluationRun(
	ctx context.Context,
	bronzeTournamentID string,
	runID string,
	excludedEntryName string,
	tournamentSimulationBatchID *string,
) error {
	return s.calculateSimulatedCalcuttaInternal(ctx, bronzeTournamentID, runID, excludedEntryName, tournamentSimulationBatchID)
}

func (s *Service) calculateSimulatedCalcuttaInternal(
	ctx context.Context,
	bronzeTournamentID string,
	runID string,
	excludedEntryName string,
	tournamentSimulationBatchIDOverride *string,
) error {
	log.Printf("Calculating simulated calcutta for tournament %s, run %s", bronzeTournamentID, runID)
	if excludedEntryName != "" {
		log.Printf("Excluding entry name: %s", excludedEntryName)
	}

	cc, err := s.getCalcuttaContext(ctx, bronzeTournamentID)
	if err != nil {
		return fmt.Errorf("failed to resolve calcutta context: %w", err)
	}

	tournamentSimulationBatchID := ""
	if tournamentSimulationBatchIDOverride != nil {
		tournamentSimulationBatchID = *tournamentSimulationBatchIDOverride
	}

	if tournamentSimulationBatchID == "" {
		var ok bool
		tournamentSimulationBatchID, ok, err = s.getLatestTournamentSimulationBatchID(ctx, bronzeTournamentID, cc.TournamentID)
		if err != nil {
			return fmt.Errorf("failed to resolve latest tournament simulation batch: %w", err)
		}
		if !ok {
			tournamentStateSnapshotID, err := s.createTournamentStateSnapshot(ctx, cc.TournamentID)
			if err != nil {
				return fmt.Errorf("failed to create tournament state snapshot: %w", err)
			}

			tournamentSimulationBatchID, err = s.createTournamentSimulationBatch(ctx, bronzeTournamentID, cc.TournamentID, tournamentStateSnapshotID)
			if err != nil {
				return fmt.Errorf("failed to create tournament simulation batch: %w", err)
			}

			if err := s.attachSimulationBatchToSimulatedTournaments(ctx, bronzeTournamentID, tournamentSimulationBatchID); err != nil {
				return fmt.Errorf("failed to attach tournament_simulation_batch_id to simulated_tournaments: %w", err)
			}
		}
	}

	calcuttaSnapshotID, err := s.createCalcuttaSnapshot(ctx, cc.CalcuttaID, cc.TournamentID, bronzeTournamentID, runID, excludedEntryName)
	if err != nil {
		return fmt.Errorf("failed to create calcutta snapshot: %w", err)
	}

	calcuttaEvaluationRunID, err := s.createCalcuttaEvaluationRun(ctx, tournamentSimulationBatchID, calcuttaSnapshotID)
	if err != nil {
		return fmt.Errorf("failed to create calcutta evaluation run: %w", err)
	}

	// Get payout structure from database
	payouts, firstPlacePayout, err := s.getPayoutStructure(ctx, cc.CalcuttaID)
	if err != nil {
		return fmt.Errorf("failed to get payout structure: %w", err)
	}

	log.Printf("Found payout structure with %d positions, 1st place: %d cents", len(payouts), firstPlacePayout)

	// Get all entries and their bids
	entries, err := s.getEntries(ctx, bronzeTournamentID, cc, runID, excludedEntryName)
	if err != nil {
		return fmt.Errorf("failed to get entries: %w", err)
	}

	log.Printf("Found %d entries", len(entries))

	simulations, err := s.getSimulations(ctx, bronzeTournamentID, cc, tournamentSimulationBatchID)
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

			simResults, err := s.calculateSimulationOutcomes(ctx, sid, entries, simulations[sid], payouts, firstPlacePayout)
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
	if err := s.writeSimulationOutcomes(ctx, runID, calcuttaEvaluationRunID, allResults); err != nil {
		return fmt.Errorf("failed to write simulation outcomes: %w", err)
	}

	// Calculate and write aggregated performance metrics
	performance := s.calculatePerformanceMetrics(allResults)
	if err := s.writePerformanceMetrics(ctx, runID, calcuttaEvaluationRunID, performance); err != nil {
		return fmt.Errorf("failed to write performance metrics: %w", err)
	}

	log.Printf("Successfully calculated simulated calcutta for %d entries", len(entries))
	return nil
}

func (s *Service) createTournamentStateSnapshot(ctx context.Context, coreTournamentID string) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var snapshotID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO analytics.tournament_state_snapshots (tournament_id, source, description)
		VALUES ($1, 'simulated_calcutta', 'Autogenerated snapshot for simulated calcutta evaluation')
		RETURNING id
	`, coreTournamentID).Scan(&snapshotID); err != nil {
		return "", err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO analytics.tournament_state_snapshot_teams (
			tournament_state_snapshot_id,
			team_id,
			wins,
			byes,
			eliminated
		)
		SELECT
			$1,
			ct.id,
			ct.wins,
			ct.byes,
			ct.eliminated
		FROM core.teams ct
		WHERE ct.tournament_id = $2
			AND ct.deleted_at IS NULL
		ON CONFLICT (tournament_state_snapshot_id, team_id) DO NOTHING
	`, snapshotID, coreTournamentID)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return snapshotID, nil
}

func (s *Service) createTournamentSimulationBatch(ctx context.Context, bronzeTournamentID string, coreTournamentID string, tournamentStateSnapshotID string) (string, error) {
	var nSims int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT sim_id)::int
		FROM analytics.simulated_tournaments
		WHERE tournament_id = $1
	`, bronzeTournamentID).Scan(&nSims); err != nil {
		return "", err
	}

	var batchID string
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO analytics.tournament_simulation_batches (
			tournament_id,
			tournament_state_snapshot_id,
			n_sims,
			seed,
			probability_source_key
		)
		VALUES ($1, $2, $3, 0, 'legacy')
		RETURNING id
	`, coreTournamentID, tournamentStateSnapshotID, nSims).Scan(&batchID); err != nil {
		return "", err
	}

	return batchID, nil
}

func (s *Service) attachSimulationBatchToSimulatedTournaments(ctx context.Context, bronzeTournamentID string, tournamentSimulationBatchID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE analytics.simulated_tournaments
		SET tournament_simulation_batch_id = $2
		WHERE tournament_id = $1
			AND tournament_simulation_batch_id IS NULL
	`, bronzeTournamentID, tournamentSimulationBatchID)
	return err
}

func (s *Service) createCalcuttaSnapshot(ctx context.Context, calcuttaID string, coreTournamentID string, bronzeTournamentID string, runID string, excludedEntryName string) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var snapshotID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO core.calcutta_snapshots (base_calcutta_id, snapshot_type, description)
		VALUES ($1, 'simulated_calcutta', 'Autogenerated snapshot for simulated calcutta evaluation')
		RETURNING id
	`, calcuttaID).Scan(&snapshotID); err != nil {
		return "", err
	}

	// Snapshot payouts
	_, err = tx.Exec(ctx, `
		INSERT INTO core.calcutta_snapshot_payouts (calcutta_snapshot_id, position, amount_cents)
		SELECT $2, position, amount_cents
		FROM core.payouts
		WHERE calcutta_id = $1
			AND deleted_at IS NULL
	`, calcuttaID, snapshotID)
	if err != nil {
		return "", err
	}

	// Snapshot scoring rules
	_, err = tx.Exec(ctx, `
		INSERT INTO core.calcutta_snapshot_scoring_rules (calcutta_snapshot_id, win_index, points_awarded)
		SELECT $2, win_index, points_awarded
		FROM core.calcutta_scoring_rules
		WHERE calcutta_id = $1
			AND deleted_at IS NULL
	`, calcuttaID, snapshotID)
	if err != nil {
		return "", err
	}

	// Snapshot real entries and their bids
	entryRows, err := tx.Query(ctx, `
		SELECT id, name
		FROM core.entries
		WHERE calcutta_id = $1
			AND deleted_at IS NULL
			AND (name != $2 OR $2 = '')
		ORDER BY created_at ASC
	`, calcuttaID, excludedEntryName)
	if err != nil {
		return "", err
	}

	type entryRow struct {
		id   string
		name string
	}
	entries := make([]entryRow, 0)
	for entryRows.Next() {
		var entryID, entryName string
		if err := entryRows.Scan(&entryID, &entryName); err != nil {
			entryRows.Close()
			return "", err
		}
		entries = append(entries, entryRow{id: entryID, name: entryName})
	}
	if err := entryRows.Err(); err != nil {
		entryRows.Close()
		return "", err
	}
	entryRows.Close()

	for _, e := range entries {
		var snapshotEntryID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
			VALUES ($1, $2, $3, false)
			RETURNING id
		`, snapshotID, e.id, e.name).Scan(&snapshotEntryID); err != nil {
			return "", err
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			SELECT $1, team_id, bid_points
			FROM core.entry_teams
			WHERE entry_id = $2
				AND deleted_at IS NULL
		`, snapshotEntryID, e.id)
		if err != nil {
			return "", err
		}
	}

	// Snapshot our_strategy synthetic entry when a strategy_generation_run exists for this run_key.
	var strategyGenerationRunID string
	if err := tx.QueryRow(ctx, `
		SELECT id
		FROM lab_gold.strategy_generation_runs
		WHERE run_key = $1::text
			AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, runID).Scan(&strategyGenerationRunID); err == nil {
		var snapshotEntryID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
			VALUES ($1, NULL, 'Our Strategy', true)
			RETURNING id
		`, snapshotID).Scan(&snapshotEntryID); err != nil {
			return "", err
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			SELECT
				$1,
				tt.id,
				reb.recommended_bid_points
			FROM lab_gold.recommended_entry_bids reb
			JOIN lab_bronze.teams bt ON reb.team_id = bt.id
			JOIN core.schools s ON bt.school_name = s.name
			JOIN core.teams tt ON tt.school_id = s.id AND tt.tournament_id = $4
			WHERE reb.strategy_generation_run_id = $2::uuid
				AND bt.tournament_id = $3
				AND reb.deleted_at IS NULL
		`, snapshotEntryID, strategyGenerationRunID, bronzeTournamentID, coreTournamentID)
		if err != nil {
			return "", err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return snapshotID, nil
}

func (s *Service) createCalcuttaEvaluationRun(ctx context.Context, tournamentSimulationBatchID string, calcuttaSnapshotID string) (string, error) {
	var evalID string
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO analytics.calcutta_evaluation_runs (tournament_simulation_batch_id, calcutta_snapshot_id, purpose)
		VALUES ($1, $2, 'simulated_calcutta')
		RETURNING id
	`, tournamentSimulationBatchID, calcuttaSnapshotID).Scan(&evalID); err != nil {
		return "", err
	}
	return evalID, nil
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

type calcuttaContext struct {
	CalcuttaID   string
	TournamentID string
}

func (s *Service) getCalcuttaContext(ctx context.Context, bronzeTournamentID string) (*calcuttaContext, error) {
	// Resolve a single calcutta_id (and its tournament_id) for a given bronze tournament.
	// This intentionally isolates the season/name join in one place.
	query := `
		SELECT c.id, c.tournament_id
		FROM lab_bronze.tournaments bt
		JOIN core.tournaments t ON t.id = bt.core_tournament_id AND t.deleted_at IS NULL
		JOIN core.calcuttas c ON c.tournament_id = t.id AND c.deleted_at IS NULL
		WHERE bt.id = $1
		ORDER BY c.created_at DESC
		LIMIT 1
	`

	var calcuttaID, tournamentID string
	if err := s.pool.QueryRow(ctx, query, bronzeTournamentID).Scan(&calcuttaID, &tournamentID); err != nil {
		return nil, err
	}
	return &calcuttaContext{CalcuttaID: calcuttaID, TournamentID: tournamentID}, nil
}

func (s *Service) getEntries(ctx context.Context, bronzeTournamentID string, cc *calcuttaContext, runID string, excludedEntry string) (map[string]*Entry, error) {
	// Use canonical core tables for entries/bids.
	query := `
		SELECT
			ce.name as entry_name,
			cet.team_id,
			cet.bid_points as bid_points
		FROM core.entry_teams cet
		JOIN core.entries ce ON cet.entry_id = ce.id
		WHERE ce.calcutta_id = $1
		  AND cet.deleted_at IS NULL
		  AND ce.deleted_at IS NULL
		  AND (ce.name != $2 OR $2 = '')
	`

	rows, err := s.pool.Query(ctx, query, cc.CalcuttaID, excludedEntry)
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
	// Map bronze_teams IDs to core.teams IDs in this tournament.
	ourQuery := `
		SELECT 
			tt.id as tournament_team_id,
			greb.recommended_bid_points
		FROM lab_gold.recommended_entry_bids greb
		JOIN lab_bronze.teams bt ON greb.team_id = bt.id
		JOIN core.schools s ON bt.school_name = s.name
		JOIN core.teams tt ON tt.school_id = s.id AND tt.tournament_id = $3
		WHERE greb.run_id = $1
		  AND bt.tournament_id = $2
	`

	ourRows, err := s.pool.Query(ctx, ourQuery, runID, bronzeTournamentID, cc.TournamentID)
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

func (s *Service) getSimulations(ctx context.Context, bronzeTournamentID string, cc *calcuttaContext, tournamentSimulationBatchID string) (map[int][]TeamSimResult, error) {
	// Simulations are keyed by bronze tournaments; map bronze teams to core teams in the resolved tournament.
	// Compute points using canonical scoring rules.
	query := `
		SELECT
			sst.sim_id,
			tt.id as tournament_team_id,
			core.calcutta_points_for_progress($3, sst.wins, sst.byes) as points
		FROM analytics.simulated_tournaments sst
		JOIN lab_bronze.teams bt ON sst.team_id = bt.id
		JOIN core.schools s ON bt.school_name = s.name
		JOIN core.teams tt ON tt.school_id = s.id AND tt.tournament_id = $2
		WHERE sst.tournament_id = $1
			AND sst.tournament_simulation_batch_id = $4
		ORDER BY sst.sim_id, tt.id
	`

	rows, err := s.pool.Query(ctx, query, bronzeTournamentID, cc.TournamentID, cc.CalcuttaID, tournamentSimulationBatchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	simulations := make(map[int][]TeamSimResult)

	for rows.Next() {
		var simID, points int
		var teamID string
		if err := rows.Scan(&simID, &teamID, &points); err != nil {
			return nil, err
		}
		simulations[simID] = append(simulations[simID], TeamSimResult{
			TeamID: teamID,
			Points: points,
		})
	}

	return simulations, nil
}

func (s *Service) getPayoutStructure(ctx context.Context, calcuttaID string) (map[int]int, int, error) {
	query := `
		SELECT position, amount_cents
		FROM core.payouts
		WHERE calcutta_id = $1
		  AND deleted_at IS NULL
		ORDER BY position
	`

	rows, err := s.pool.Query(ctx, query, calcuttaID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	payouts := make(map[int]int)
	var firstPlacePayout int

	for rows.Next() {
		var position, amountCents int
		if err := rows.Scan(&position, &amountCents); err != nil {
			return nil, 0, err
		}
		payouts[position] = amountCents
		if position == 1 {
			firstPlacePayout = amountCents
		}
	}

	if firstPlacePayout == 0 {
		return nil, 0, fmt.Errorf("no first place payout found")
	}

	return payouts, firstPlacePayout, nil
}

func (s *Service) calculateSimulationOutcomes(ctx context.Context, simID int, entries map[string]*Entry, teamResults []TeamSimResult, payouts map[int]int, firstPlacePayout int) ([]SimulationResult, error) {
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
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].points > scores[j].points
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

func (s *Service) writeSimulationOutcomes(ctx context.Context, runID string, calcuttaEvaluationRunID string, results []SimulationResult) error {
	var err error
	if calcuttaEvaluationRunID != "" {
		_, err = s.pool.Exec(ctx, "DELETE FROM analytics.entry_simulation_outcomes WHERE calcutta_evaluation_run_id = $1", calcuttaEvaluationRunID)
	} else {
		_, err = s.pool.Exec(ctx, "DELETE FROM analytics.entry_simulation_outcomes WHERE run_id = $1", runID)
	}
	if err != nil {
		return err
	}

	// Batch insert new results
	batch := &pgx.Batch{}
	for _, r := range results {
		var evalID any
		if calcuttaEvaluationRunID != "" {
			evalID = calcuttaEvaluationRunID
		} else {
			evalID = nil
		}
		batch.Queue(`
			INSERT INTO analytics.entry_simulation_outcomes (run_id, entry_name, sim_id, points_scored, payout_points, rank, calcutta_evaluation_run_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, runID, r.EntryName, r.SimID, r.TotalPoints, r.PayoutCents, r.Rank, evalID)
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

func (s *Service) writePerformanceMetrics(ctx context.Context, runID string, calcuttaEvaluationRunID string, performance map[string]*EntryPerformance) error {
	var err error
	if calcuttaEvaluationRunID != "" {
		_, err = s.pool.Exec(ctx, "DELETE FROM analytics.entry_performance WHERE calcutta_evaluation_run_id = $1", calcuttaEvaluationRunID)
	} else {
		_, err = s.pool.Exec(ctx, "DELETE FROM analytics.entry_performance WHERE run_id = $1", runID)
	}
	if err != nil {
		return err
	}

	// Insert new performance metrics
	for _, p := range performance {
		var evalID any
		if calcuttaEvaluationRunID != "" {
			evalID = calcuttaEvaluationRunID
		} else {
			evalID = nil
		}
		_, err := s.pool.Exec(ctx, `
			INSERT INTO analytics.entry_performance (run_id, entry_name, mean_payout, median_payout, p_top1, p_in_money, calcutta_evaluation_run_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, runID, p.EntryName, p.MeanPayout, p.MedianPayout, p.PTop1, p.PInMoney, evalID)
		if err != nil {
			return err
		}
	}

	return nil
}
