package calcutta_evaluations

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (s *Service) EnqueueCalcuttaEvaluationRun(
	ctx context.Context,
	calcuttaID string,
	excludedEntryName string,
	tournamentSimulationBatchID *string,
) (string, string, error) {
	if calcuttaID == "" {
		return "", "", fmt.Errorf("calcuttaID is required")
	}

	runKey := uuid.NewString()

	cc, err := s.getCalcuttaContext(ctx, calcuttaID)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve calcutta context: %w", err)
	}

	batchID := ""
	if tournamentSimulationBatchID != nil {
		batchID = *tournamentSimulationBatchID
	}
	if batchID == "" {
		var ok bool
		batchID, ok, err = s.getLatestTournamentSimulationBatchID(ctx, cc.TournamentID)
		if err != nil {
			return "", "", fmt.Errorf("failed to resolve latest tournament simulation batch: %w", err)
		}
		if !ok {
			tournamentStateSnapshotID, err := s.createTournamentStateSnapshot(ctx, cc.TournamentID)
			if err != nil {
				return "", "", fmt.Errorf("failed to create tournament state snapshot: %w", err)
			}

			batchID, err = s.createTournamentSimulationBatch(ctx, cc.TournamentID, tournamentStateSnapshotID)
			if err != nil {
				return "", "", fmt.Errorf("failed to create tournament simulation batch: %w", err)
			}

			if err := s.attachSimulationBatchToSimulatedTournaments(ctx, cc.TournamentID, batchID); err != nil {
				return "", "", fmt.Errorf("failed to attach tournament_simulation_batch_id to simulated_tournaments: %w", err)
			}
		}
	}

	snapshotID, err := s.createCalcuttaSnapshot(ctx, cc.CalcuttaID, cc.TournamentID, runKey, excludedEntryName, nil, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create calcutta snapshot: %w", err)
	}

	paramsJSON, _ := json.Marshal(map[string]any{
		"workerize": true,
		"source":    "enqueue_calcutta_evaluation",
	})

	gitSHA := os.Getenv("GIT_SHA")
	var gitSHAParam any
	if gitSHA != "" {
		gitSHAParam = gitSHA
	} else {
		gitSHAParam = nil
	}

	var evalRunID string
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO derived.calcutta_evaluation_runs (
			run_key,
			simulated_tournament_id,
			calcutta_snapshot_id,
			purpose,
			params_json,
			git_sha
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, 'simulated_calcutta', $4::jsonb, $5)
		RETURNING id::text
	`, runKey, batchID, snapshotID, string(paramsJSON), gitSHAParam).Scan(&evalRunID); err != nil {
		return "", "", err
	}

	return evalRunID, runKey, nil
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

	_, _, err = s.calculateAndWriteCalcuttaEvaluation(ctx, cc, runID, calcuttaEvaluationRunID, tournamentSimulationBatchID, entries, payouts, firstPlacePayout)
	if err != nil {
		return "", err
	}

	log.Printf("Successfully calculated simulated calcutta for %d entries", len(entries))
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
	log.Printf("Found %d entries", len(entries))

	log.Printf("Found %d simulations", len(simulations))

	results := make(chan []SimulationResult, len(simulations))
	errors := make(chan error, len(simulations))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)

	for simID := range simulations {
		wg.Add(1)
		go func(sid int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			simResults, err := s.calculateSimulationOutcomes(ctx, sid, entries, simulations[sid], payouts, firstPlacePayout)
			if err != nil {
				errors <- fmt.Errorf("simulation %d: %w", sid, err)
				return
			}
			results <- simResults
		}(simID)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	var allResults []SimulationResult
	for simResults := range results {
		allResults = append(allResults, simResults...)
	}
	for err := range errors {
		if err != nil {
			return 0, 0, err
		}
	}

	log.Printf("Calculated %d total outcomes", len(allResults))

	persistDetails := os.Getenv("CALCUTTA_PERSIST_SIMULATION_DETAILS") == "true"
	if persistDetails {
		if err := s.writeSimulationOutcomes(ctx, runID, calcuttaEvaluationRunID, allResults); err != nil {
			return 0, 0, fmt.Errorf("failed to write simulation outcomes: %w", err)
		}
	} else {
		if err := s.deleteSimulationOutcomes(ctx, runID, calcuttaEvaluationRunID); err != nil {
			return 0, 0, fmt.Errorf("failed to clear simulation outcomes: %w", err)
		}
	}

	performance := s.calculatePerformanceMetrics(allResults)
	if err := s.writePerformanceMetrics(ctx, runID, calcuttaEvaluationRunID, performance); err != nil {
		return 0, 0, fmt.Errorf("failed to write performance metrics: %w", err)
	}

	return len(simulations), len(entries), nil
}
