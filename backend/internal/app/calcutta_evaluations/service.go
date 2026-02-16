package calcutta_evaluations

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type CalcuttaEvaluationResult struct {
	CalcuttaEvaluationRunID string
	RunKey                  string
	NSims                   int
	NEntries                int
}

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

func (s *Service) CalculateSimulatedCalcuttaForSimulatedCalcutta(
	ctx context.Context,
	simulatedCalcuttaID string,
	runID string,
	excludedEntryName string,
	tournamentSimulationBatchID *string,
) (string, error) {
	if simulatedCalcuttaID == "" {
		return "", fmt.Errorf("simulatedCalcuttaID is required")
	}

	runKey := runID
	if runKey == "" {
		runKey = uuid.NewString()
	}

	var tournamentID string
	if err := s.pool.QueryRow(ctx, `
		SELECT tournament_id::text
		FROM derived.simulated_calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, simulatedCalcuttaID).Scan(&tournamentID); err != nil {
		return "", err
	}

	batchID := ""
	if tournamentSimulationBatchID != nil {
		batchID = *tournamentSimulationBatchID
	}
	if batchID == "" {
		var ok bool
		batchID, ok, err := s.getLatestTournamentSimulationBatchID(ctx, tournamentID)
		if err != nil {
			return "", fmt.Errorf("failed to resolve latest tournament simulation batch: %w", err)
		}
		if !ok {
			tournamentStateSnapshotID, err := s.createTournamentStateSnapshot(ctx, tournamentID)
			if err != nil {
				return "", fmt.Errorf("failed to create tournament state snapshot: %w", err)
			}

			batchID, err = s.createTournamentSimulationBatch(ctx, tournamentID, tournamentStateSnapshotID)
			if err != nil {
				return "", fmt.Errorf("failed to create tournament simulation batch: %w", err)
			}

			if err := s.attachSimulationBatchToSimulatedTournaments(ctx, tournamentID, batchID); err != nil {
				return "", fmt.Errorf("failed to attach tournament_simulation_batch_id to simulated_tournaments: %w", err)
			}
		}
	}

	evalRunID, err := s.createCalcuttaEvaluationRunForSimulatedCalcutta(ctx, batchID, simulatedCalcuttaID)
	if err != nil {
		return "", err
	}

	payouts, firstPlacePayout, err := s.getPayoutStructureFromSimulatedCalcutta(ctx, simulatedCalcuttaID)
	if err != nil {
		return "", err
	}

	entries, err := s.getEntriesFromSimulatedCalcutta(ctx, simulatedCalcuttaID, excludedEntryName)
	if err != nil {
		return "", err
	}

	rules, err := s.loadSimulatedCalcuttaScoringRules(ctx, simulatedCalcuttaID)
	if err != nil {
		return "", err
	}

	simulations, err := s.getSimulationsWithRules(ctx, tournamentID, batchID, rules)
	if err != nil {
		return "", fmt.Errorf("failed to get simulations: %w", err)
	}

	_, _, err = s.calculateAndWriteCalcuttaEvaluationWithSimulations(ctx, runKey, evalRunID, simulations, entries, payouts, firstPlacePayout)
	if err != nil {
		return "", err
	}

	return evalRunID, nil
}

func (s *Service) EvaluateExistingCalcuttaEvaluationRun(ctx context.Context, calcuttaEvaluationRunID string) (*CalcuttaEvaluationResult, error) {
	if calcuttaEvaluationRunID == "" {
		return nil, fmt.Errorf("calcuttaEvaluationRunID is required")
	}

	var (
		runKey           string
		simulatedID      string
		snapshotID       string
		baseCalcuttaID   string
		coreTournamentID string
	)
	if err := s.pool.QueryRow(ctx, `
		SELECT
			cer.run_key::text,
			cer.simulated_tournament_id::text,
			cer.calcutta_snapshot_id::text,
			cs.base_calcutta_id::text,
			c.tournament_id::text
		FROM derived.calcutta_evaluation_runs cer
		JOIN core.calcutta_snapshots cs
			ON cs.id = cer.calcutta_snapshot_id
			AND cs.deleted_at IS NULL
		JOIN core.calcuttas c
			ON c.id = cs.base_calcutta_id
			AND c.deleted_at IS NULL
		WHERE cer.id = $1::uuid
			AND cer.deleted_at IS NULL
		LIMIT 1
	`, calcuttaEvaluationRunID).Scan(&runKey, &simulatedID, &snapshotID, &baseCalcuttaID, &coreTournamentID); err != nil {
		return nil, err
	}
	if runKey == "" {
		runKey = uuid.NewString()
	}

	cc := &calcuttaContext{CalcuttaID: baseCalcuttaID, TournamentID: coreTournamentID}

	payouts, firstPlacePayout, err := s.getPayoutStructureFromSnapshot(ctx, snapshotID)
	if err != nil {
		return nil, err
	}

	entries, err := s.getEntriesFromSnapshot(ctx, snapshotID)
	if err != nil {
		return nil, err
	}

	nSims, nEntries, err := s.calculateAndWriteCalcuttaEvaluation(ctx, cc, runKey, calcuttaEvaluationRunID, simulatedID, entries, payouts, firstPlacePayout)
	if err != nil {
		return nil, err
	}

	return &CalcuttaEvaluationResult{
		CalcuttaEvaluationRunID: calcuttaEvaluationRunID,
		RunKey:                  runKey,
		NSims:                   nSims,
		NEntries:                nEntries,
	}, nil
}

func (s *Service) getPayoutStructureFromSnapshot(ctx context.Context, snapshotID string) (map[int]int, int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT position, amount_cents
		FROM core.calcutta_snapshot_payouts
		WHERE calcutta_snapshot_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY position
	`, snapshotID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	payouts := make(map[int]int)
	firstPlace := 0
	for rows.Next() {
		var position, amountCents int
		if err := rows.Scan(&position, &amountCents); err != nil {
			return nil, 0, err
		}
		payouts[position] = amountCents
		if position == 1 {
			firstPlace = amountCents
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if firstPlace == 0 {
		return nil, 0, fmt.Errorf("no first place payout found")
	}
	return payouts, firstPlace, nil
}

func (s *Service) getEntriesFromSnapshot(ctx context.Context, snapshotID string) (map[string]*Entry, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			cse.display_name,
			cset.team_id::text,
			cset.bid_points
		FROM core.calcutta_snapshot_entries cse
		JOIN core.calcutta_snapshot_entry_teams cset
			ON cset.calcutta_snapshot_entry_id = cse.id
			AND cset.deleted_at IS NULL
		WHERE cse.calcutta_snapshot_id = $1::uuid
			AND cse.deleted_at IS NULL
		ORDER BY cse.created_at ASC
	`, snapshotID)
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
			entries[entryName] = &Entry{Name: entryName, Teams: make(map[string]int)}
		}
		entries[entryName].Teams[teamID] = bidPoints
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
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

	if err := s.deleteSimulationOutcomes(ctx, runID, calcuttaEvaluationRunID); err != nil {
		return 0, 0, fmt.Errorf("failed to clear simulation outcomes: %w", err)
	}

	performance := calculatePerformanceMetrics(allResults)
	if err := s.writePerformanceMetrics(ctx, runID, calcuttaEvaluationRunID, performance); err != nil {
		return 0, 0, fmt.Errorf("failed to write performance metrics: %w", err)
	}

	return len(simulations), len(entries), nil
}
