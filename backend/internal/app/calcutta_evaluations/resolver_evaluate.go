package calcutta_evaluations

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"golang.org/x/sync/errgroup"
)

// LabEntryPerformance contains performance metrics for a single entry in an evaluation.
type LabEntryPerformance struct {
	EntryName  string
	MeanPayout float64
	PTop1      float64
	PInMoney   float64
	Rank       int
}

// LabEvaluationResult contains the performance metrics for a lab entry evaluation.
type LabEvaluationResult struct {
	MeanNormalizedPayout   float64
	MedianNormalizedPayout float64
	PTop1                  float64
	PInMoney               float64
	NSims                  int
	AllEntryResults        []LabEntryPerformance
}

// EvaluateLabEntry evaluates a lab entry against all other entries in a calcutta.
// It runs simulations and returns the performance metrics for the lab entry.
func (s *Service) EvaluateLabEntry(
	ctx context.Context,
	calcuttaID string,
	labEntryBids map[string]int, // team_id -> bid_points
	excludedEntryName string,
) (*LabEvaluationResult, error) {
	cc, err := s.getCalcuttaContext(ctx, calcuttaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get calcutta context: %w", err)
	}

	batchID, err := s.resolveSimulationBatchID(ctx, cc.TournamentID)
	if err != nil {
		return nil, err
	}

	payouts, firstPlacePayout, err := s.getPayoutStructure(ctx, cc.CalcuttaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payout structure: %w", err)
	}

	entries, err := s.getEntriesForLabEvaluation(ctx, cc, excludedEntryName, labEntryBids)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	simulations, err := s.getSimulations(ctx, cc, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get simulations: %w", err)
	}
	if len(simulations) == 0 {
		return nil, fmt.Errorf("no simulations available for tournament %s", cc.TournamentID)
	}

	allResults, err := s.runConcurrentEvaluations(ctx, entries, simulations, payouts, firstPlacePayout)
	if err != nil {
		return nil, err
	}

	return buildLabEvaluationResult(allResults, len(simulations))
}

// resolveSimulationBatchID returns the latest simulation batch ID for the
// tournament, running simulations on-demand if none exist yet.
func (s *Service) resolveSimulationBatchID(ctx context.Context, tournamentID string) (string, error) {
	batchID, ok, err := s.getLatestTournamentSimulationBatchID(ctx, tournamentID)
	if err != nil {
		return "", fmt.Errorf("failed to get tournament simulation batch: %w", err)
	}
	if ok {
		return batchID, nil
	}

	season, err := s.tournamentResolver.ResolveSeasonFromTournamentID(ctx, tournamentID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve season from tournament: %w", err)
	}

	simSvc := simulation.New(s.pool, simulation.WithTournamentResolver(s.tournamentResolver))
	result, err := simSvc.Run(ctx, simulation.RunParams{
		Season:               season,
		NSims:                10000,
		Seed:                 42,
		Workers:              runtime.GOMAXPROCS(0),
		BatchSize:            1000,
		ProbabilitySourceKey: "lab_pipeline",
		StartingStateKey:     "current",
	})
	if err != nil {
		return "", fmt.Errorf("failed to run simulations: %w", err)
	}
	return result.TournamentSimulationBatchID, nil
}

// runConcurrentEvaluations fans out simulation outcome calculations across
// goroutines with bounded concurrency and collects the results.
func (s *Service) runConcurrentEvaluations(
	ctx context.Context,
	entries map[string]*Entry,
	simulations map[int][]TeamSimResult,
	payouts map[int]int,
	firstPlacePayout int,
) ([]SimulationResult, error) {
	var (
		mu         sync.Mutex
		allResults []SimulationResult
	)
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	for simID := range simulations {
		sid := simID
		g.Go(func() error {
			simResults, err := CalculateSimulationOutcomes(sid, entries, simulations[sid], payouts, firstPlacePayout)
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
		return nil, err
	}
	return allResults, nil
}

// buildLabEvaluationResult aggregates simulation results into performance
// metrics, extracts the lab entry's metrics, and ranks all entries.
func buildLabEvaluationResult(allResults []SimulationResult, nSims int) (*LabEvaluationResult, error) {
	performance := CalculatePerformanceMetrics(allResults)

	ourPerformance, ok := performance[models.LabStrategyEntryName]
	if !ok {
		return nil, fmt.Errorf("failed to find performance for lab entry")
	}

	allEntryResults := RankEntryPerformance(performance)

	return &LabEvaluationResult{
		MeanNormalizedPayout:   ourPerformance.MeanPayout,
		MedianNormalizedPayout: ourPerformance.MedianPayout,
		PTop1:                  ourPerformance.PTop1,
		PInMoney:               ourPerformance.PInMoney,
		NSims:                  nSims,
		AllEntryResults:        allEntryResults,
	}, nil
}

// RankEntryPerformance converts the performance map into a sorted,
// ranked slice ordered by MeanPayout descending.
func RankEntryPerformance(performance map[string]*EntryPerformance) []LabEntryPerformance {
	results := make([]LabEntryPerformance, 0, len(performance))
	for _, perf := range performance {
		results = append(results, LabEntryPerformance{
			EntryName:  perf.EntryName,
			MeanPayout: perf.MeanPayout,
			PTop1:      perf.PTop1,
			PInMoney:   perf.PInMoney,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].MeanPayout > results[j].MeanPayout
	})
	for i := range results {
		results[i].Rank = i + 1
	}
	return results
}
