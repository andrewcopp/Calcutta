package simulated_calcutta

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

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

// CalculateSimulatedCalcutta calculates entry outcomes for all simulations
func (s *Service) CalculateSimulatedCalcutta(ctx context.Context, calcuttaID string, runID string) error {
	// Get excluded entry name from environment (e.g., "Andrew Copp")
	excludedEntryName := os.Getenv("EXCLUDED_ENTRY_NAME")
	_, err := s.calculateSimulatedCalcuttaInternal(ctx, calcuttaID, runID, excludedEntryName, nil, nil, nil)
	return err
}

func (s *Service) CalculateSimulatedCalcuttaForEvaluationRun(
	ctx context.Context,
	calcuttaID string,
	runID string,
	excludedEntryName string,
	tournamentSimulationBatchID *string,
) (string, error) {
	return s.calculateSimulatedCalcuttaInternal(ctx, calcuttaID, runID, excludedEntryName, tournamentSimulationBatchID, nil, nil)
}

func (s *Service) CalculateSimulatedCalcuttaForStrategyGenerationRun(
	ctx context.Context,
	calcuttaID string,
	runID string,
	excludedEntryName string,
	tournamentSimulationBatchID *string,
	strategyGenerationRunID string,
) (string, error) {
	if strategyGenerationRunID == "" {
		return "", fmt.Errorf("strategyGenerationRunID is required")
	}
	return s.calculateSimulatedCalcuttaInternal(ctx, calcuttaID, runID, excludedEntryName, tournamentSimulationBatchID, nil, &strategyGenerationRunID)
}

func (s *Service) CalculateSimulatedCalcuttaForEntryCandidate(
	ctx context.Context,
	calcuttaID string,
	runID string,
	excludedEntryName string,
	tournamentSimulationBatchID *string,
	entryCandidateID string,
) (string, error) {
	if entryCandidateID == "" {
		return "", fmt.Errorf("entryCandidateID is required")
	}
	return s.calculateSimulatedCalcuttaInternal(ctx, calcuttaID, runID, excludedEntryName, tournamentSimulationBatchID, &entryCandidateID, nil)
}

func (s *Service) calculateSimulatedCalcuttaInternal(
	ctx context.Context,
	calcuttaID string,
	runID string,
	excludedEntryName string,
	tournamentSimulationBatchIDOverride *string,
	entryCandidateID *string,
	strategyGenerationRunID *string,
) (string, error) {
	log.Printf("Calculating simulated calcutta for calcutta %s, run %s", calcuttaID, runID)
	if excludedEntryName != "" {
		log.Printf("Excluding entry name: %s", excludedEntryName)
	}

	cc, err := s.getCalcuttaContext(ctx, calcuttaID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve calcutta context: %w", err)
	}

	tournamentSimulationBatchID := ""
	if tournamentSimulationBatchIDOverride != nil {
		tournamentSimulationBatchID = *tournamentSimulationBatchIDOverride
	}

	if tournamentSimulationBatchID == "" {
		var ok bool
		tournamentSimulationBatchID, ok, err = s.getLatestTournamentSimulationBatchID(ctx, cc.TournamentID)
		if err != nil {
			return "", fmt.Errorf("failed to resolve latest tournament simulation batch: %w", err)
		}
		if !ok {
			tournamentStateSnapshotID, err := s.createTournamentStateSnapshot(ctx, cc.TournamentID)
			if err != nil {
				return "", fmt.Errorf("failed to create tournament state snapshot: %w", err)
			}

			tournamentSimulationBatchID, err = s.createTournamentSimulationBatch(ctx, cc.TournamentID, tournamentStateSnapshotID)
			if err != nil {
				return "", fmt.Errorf("failed to create tournament simulation batch: %w", err)
			}

			if err := s.attachSimulationBatchToSimulatedTournaments(ctx, cc.TournamentID, tournamentSimulationBatchID); err != nil {
				return "", fmt.Errorf("failed to attach tournament_simulation_batch_id to simulated_tournaments: %w", err)
			}
		}
	}

	calcuttaSnapshotID, err := s.createCalcuttaSnapshot(ctx, cc.CalcuttaID, cc.TournamentID, runID, excludedEntryName, entryCandidateID, strategyGenerationRunID)
	if err != nil {
		return "", fmt.Errorf("failed to create calcutta snapshot: %w", err)
	}

	calcuttaEvaluationRunID, err := s.createCalcuttaEvaluationRun(ctx, tournamentSimulationBatchID, calcuttaSnapshotID)
	if err != nil {
		return "", fmt.Errorf("failed to create calcutta evaluation run: %w", err)
	}

	// Get payout structure from database
	payouts, firstPlacePayout, err := s.getPayoutStructure(ctx, cc.CalcuttaID)
	if err != nil {
		return "", fmt.Errorf("failed to get payout structure: %w", err)
	}

	log.Printf("Found payout structure with %d positions, 1st place: %d cents", len(payouts), firstPlacePayout)

	// Get all entries and their bids
	entries, err := s.getEntries(ctx, cc, runID, excludedEntryName, entryCandidateID, strategyGenerationRunID)
	if err != nil {
		return "", fmt.Errorf("failed to get entries: %w", err)
	}

	log.Printf("Found %d entries", len(entries))

	simulations, err := s.getSimulations(ctx, cc, tournamentSimulationBatchID)
	if err != nil {
		return "", fmt.Errorf("failed to get simulations: %w", err)
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
			return "", err
		}
	}

	log.Printf("Calculated %d total outcomes", len(allResults))

	persistDetails := os.Getenv("CALCUTTA_PERSIST_SIMULATION_DETAILS") == "true"
	if persistDetails {
		// Write detailed per-simulation outcomes (debug mode)
		if err := s.writeSimulationOutcomes(ctx, runID, calcuttaEvaluationRunID, allResults); err != nil {
			return "", fmt.Errorf("failed to write simulation outcomes: %w", err)
		}
	} else {
		// Ensure we don't accumulate detailed rows across runs.
		if err := s.deleteSimulationOutcomes(ctx, runID, calcuttaEvaluationRunID); err != nil {
			return "", fmt.Errorf("failed to clear simulation outcomes: %w", err)
		}
	}

	// Calculate and write aggregated performance metrics
	performance := s.calculatePerformanceMetrics(allResults)
	if err := s.writePerformanceMetrics(ctx, runID, calcuttaEvaluationRunID, performance); err != nil {
		return "", fmt.Errorf("failed to write performance metrics: %w", err)
	}

	log.Printf("Successfully calculated simulated calcutta for %d entries", len(entries))
	return calcuttaEvaluationRunID, nil
}
