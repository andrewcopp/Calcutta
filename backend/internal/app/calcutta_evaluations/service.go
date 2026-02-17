package calcutta_evaluations

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

// Service handles simulated calcutta analysis
type Service struct {
	pool *pgxpool.Pool
}

// New creates a new simulated calcutta service
func New(pool *pgxpool.Pool) *Service {
	return &Service{
		pool: pool,
	}
}

func (s *Service) CalculateSimulatedCalcuttaForEvaluationRun(
	ctx context.Context,
	calcuttaID string,
	runID string,
	excludedEntryName string,
	tournamentSimulationBatchID *string,
) (string, error) {
	return s.calculateSimulatedCalcuttaInternal(ctx, calcuttaID, runID, excludedEntryName, tournamentSimulationBatchID, nil)
}

func (s *Service) calculateSimulatedCalcuttaInternal(
	ctx context.Context,
	calcuttaID string,
	runID string,
	excludedEntryName string,
	tournamentSimulationBatchIDOverride *string,
	strategyGenerationRunID *string,
) (string, error) {
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

	calcuttaSnapshotID, err := s.createCalcuttaSnapshot(ctx, cc.CalcuttaID, cc.TournamentID, runID, excludedEntryName, strategyGenerationRunID)
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

	// Get all entries and their bids
	entries, err := s.getEntries(ctx, cc, runID, excludedEntryName, strategyGenerationRunID)
	if err != nil {
		return "", fmt.Errorf("failed to get entries: %w", err)
	}

	_, _, err = s.calculateAndWriteCalcuttaEvaluation(ctx, cc, runID, calcuttaEvaluationRunID, tournamentSimulationBatchID, entries, payouts, firstPlacePayout)
	if err != nil {
		return "", err
	}

	return calcuttaEvaluationRunID, nil
}

func (s *Service) calculateAndWriteCalcuttaEvaluation(
	ctx context.Context,
	cc *calcuttaContext,
	runID string,
	calcuttaEvaluationRunID string,
	tournamentSimulationBatchID string,
	entries map[string]*Entry,
	payouts map[int]int,
	firstPlacePayout int,
) (int, int, error) {
	simulations, err := s.getSimulations(ctx, cc, tournamentSimulationBatchID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get simulations: %w", err)
	}

	return s.calculateAndWriteCalcuttaEvaluationWithSimulations(ctx, runID, calcuttaEvaluationRunID, simulations, entries, payouts, firstPlacePayout)
}

func (s *Service) calculateAndWriteCalcuttaEvaluationWithSimulations(
	ctx context.Context,
	runID string,
	calcuttaEvaluationRunID string,
	simulations map[int][]TeamSimResult,
	entries map[string]*Entry,
	payouts map[int]int,
	firstPlacePayout int,
) (int, int, error) {
	var (
		mu         sync.Mutex
		allResults []SimulationResult
	)

	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(10)

	for simID := range simulations {
		sid := simID
		g.Go(func() error {
			simResults, err := calculateSimulationOutcomes(sid, entries, simulations[sid], payouts, firstPlacePayout)
			if err != nil {
				return fmt.Errorf("simulation %d: %w", sid, err)
			}
			mu.Lock()
			allResults = append(allResults, simResults...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return 0, 0, err
	}

	performance := calculatePerformanceMetrics(allResults)
	if err := s.writePerformanceMetrics(ctx, runID, calcuttaEvaluationRunID, performance); err != nil {
		return 0, 0, fmt.Errorf("failed to write performance metrics: %w", err)
	}

	return len(simulations), len(entries), nil
}
