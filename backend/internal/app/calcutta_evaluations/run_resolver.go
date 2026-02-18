package calcutta_evaluations

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"sync"

	dbadapter "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

type calcuttaContext struct {
	CalcuttaID   string
	TournamentID string
}

func (s *Service) getLatestTournamentSimulationBatchID(ctx context.Context, coreTournamentID string) (string, bool, error) {
	var batchID string
	err := s.pool.QueryRow(ctx, `
		SELECT b.id
		FROM derived.simulated_tournaments b
		WHERE b.tournament_id = $1
			AND b.deleted_at IS NULL
			AND EXISTS (
				SELECT 1
				FROM derived.simulated_teams st
				WHERE st.tournament_id = $1
					AND st.simulated_tournament_id = b.id
					AND st.deleted_at IS NULL
			)
		ORDER BY b.created_at DESC
		LIMIT 1
	`, coreTournamentID).Scan(&batchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return batchID, true, nil
}

func (s *Service) getCalcuttaContext(ctx context.Context, calcuttaID string) (*calcuttaContext, error) {
	// Resolve a single calcutta_id (and its tournament_id) for a given bronze tournament.
	// This intentionally isolates the season/name join in one place.
	query := `
		SELECT c.id, c.tournament_id
		FROM core.calcuttas c
		WHERE c.id = $1::uuid
			AND c.deleted_at IS NULL
		LIMIT 1
	`

	var resolvedCalcuttaID, tournamentID string
	if err := s.pool.QueryRow(ctx, query, calcuttaID).Scan(&resolvedCalcuttaID, &tournamentID); err != nil {
		return nil, err
	}
	return &calcuttaContext{CalcuttaID: resolvedCalcuttaID, TournamentID: tournamentID}, nil
}

func (s *Service) getSimulations(ctx context.Context, cc *calcuttaContext, tournamentSimulationBatchID string) (map[int][]TeamSimResult, error) {
	// Simulations are keyed by bronze tournaments; map bronze teams to core teams in the resolved tournament.
	// Compute points using canonical scoring rules.
	rules, err := s.loadCoreScoringRules(ctx, cc.CalcuttaID)
	if err != nil {
		return nil, err
	}
	return s.getSimulationsWithRules(ctx, cc.TournamentID, tournamentSimulationBatchID, rules)
}

func (s *Service) getSimulationsWithRules(ctx context.Context, tournamentID string, tournamentSimulationBatchID string, rules []scoring.Rule) (map[int][]TeamSimResult, error) {
	query := `
		SELECT
			sst.sim_id,
			sst.team_id,
			sst.wins::int,
			sst.byes::int
		FROM derived.simulated_teams sst
		WHERE sst.tournament_id = $1
			AND sst.simulated_tournament_id = $2
			AND sst.deleted_at IS NULL
		ORDER BY sst.sim_id, sst.team_id
	`

	rows, err := s.pool.Query(ctx, query, tournamentID, tournamentSimulationBatchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	simulations := make(map[int][]TeamSimResult)

	for rows.Next() {
		var simID int
		var teamID string
		var wins int
		var byes int
		if err := rows.Scan(&simID, &teamID, &wins, &byes); err != nil {
			return nil, err
		}
		points := scoring.PointsForProgress(rules, wins, byes)
		simulations[simID] = append(simulations[simID], TeamSimResult{
			TeamID: teamID,
			Points: points,
		})
	}

	return simulations, nil
}

func (s *Service) loadCoreScoringRules(ctx context.Context, calcuttaID string) ([]scoring.Rule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT win_index::int, points_awarded::int
		FROM core.calcutta_scoring_rules
		WHERE calcutta_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY win_index ASC
	`, calcuttaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]scoring.Rule, 0)
	for rows.Next() {
		var r scoring.Rule
		if err := rows.Scan(&r.WinIndex, &r.PointsAwarded); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rules, nil
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
	// Get calcutta context
	cc, err := s.getCalcuttaContext(ctx, calcuttaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get calcutta context: %w", err)
	}

	// Get or create tournament simulation batch
	batchID, ok, err := s.getLatestTournamentSimulationBatchID(ctx, cc.TournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament simulation batch: %w", err)
	}
	if !ok {
		// No simulations exist - run them on-demand using the simulation service
		season, err := dbadapter.ResolveSeasonFromTournamentID(ctx, s.pool, cc.TournamentID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve season from tournament: %w", err)
		}

		simSvc := simulation.New(s.pool)
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
			return nil, fmt.Errorf("failed to run simulations: %w", err)
		}
		batchID = result.TournamentSimulationBatchID
	}

	// Get payout structure
	payouts, firstPlacePayout, err := s.getPayoutStructure(ctx, cc.CalcuttaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payout structure: %w", err)
	}

	// Build entries map: existing entries + lab entry
	entries, err := s.getEntriesForLabEvaluation(ctx, cc, excludedEntryName, labEntryBids)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	// Get simulations
	simulations, err := s.getSimulations(ctx, cc, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get simulations: %w", err)
	}

	if len(simulations) == 0 {
		return nil, fmt.Errorf("no simulations available for tournament %s", cc.TournamentID)
	}

	// Run simulations
	var (
		mu         sync.Mutex
		allResults []SimulationResult
	)
	g, ctx := errgroup.WithContext(ctx)
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
		return nil, err
	}

	// Calculate performance metrics for all entries
	performance := calculatePerformanceMetrics(allResults)

	// Extract metrics for our lab entry
	ourPerformance, ok := performance[models.LabStrategyEntryName]
	if !ok {
		return nil, fmt.Errorf("failed to find performance for lab entry")
	}

	// Build sorted list of all entry results
	allEntryResults := make([]LabEntryPerformance, 0, len(performance))
	for _, perf := range performance {
		allEntryResults = append(allEntryResults, LabEntryPerformance{
			EntryName:  perf.EntryName,
			MeanPayout: perf.MeanPayout,
			PTop1:      perf.PTop1,
			PInMoney:   perf.PInMoney,
		})
	}
	// Sort by MeanPayout descending
	sort.Slice(allEntryResults, func(i, j int) bool {
		return allEntryResults[i].MeanPayout > allEntryResults[j].MeanPayout
	})
	// Assign ranks
	for i := range allEntryResults {
		allEntryResults[i].Rank = i + 1
	}

	return &LabEvaluationResult{
		MeanNormalizedPayout:   ourPerformance.MeanPayout,
		MedianNormalizedPayout: ourPerformance.MedianPayout,
		PTop1:                  ourPerformance.PTop1,
		PInMoney:               ourPerformance.PInMoney,
		NSims:                  len(simulations),
		AllEntryResults:        allEntryResults,
	}, nil
}

// getEntriesForLabEvaluation builds the entries map for lab evaluation.
// It includes all real entries (except excluded) plus the lab entry.
func (s *Service) getEntriesForLabEvaluation(
	ctx context.Context,
	cc *calcuttaContext,
	excludedEntryName string,
	labEntryBids map[string]int,
) (map[string]*Entry, error) {
	// Get real entries from core tables
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

	rows, err := s.pool.Query(ctx, query, cc.CalcuttaID, excludedEntryName)
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

	// Add lab entry
	if len(labEntryBids) > 0 {
		entries[models.LabStrategyEntryName] = &Entry{
			Name:  models.LabStrategyEntryName,
			Teams: labEntryBids,
		}
	}

	return entries, nil
}
