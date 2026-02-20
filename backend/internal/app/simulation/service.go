package simulation

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	dbadapter "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	appbracket "github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type RunParams struct {
	Season               int
	NSims                int
	Seed                 int
	Workers              int
	BatchSize            int
	ProbabilitySourceKey string
	StartingStateKey     string
	GameOutcomeRunID     *string
	GameOutcomeSpec      *simulation_game_outcomes.Spec
}

type RunResult struct {
	CoreTournamentID            string
	TournamentStateSnapshotID   string
	TournamentSimulationBatchID string
	NSims                       int
	RowsWritten                 int64
	LoadDuration                time.Duration
	SimulateWriteDuration       time.Duration
	OverallDuration             time.Duration
}

func (s *Service) Run(ctx context.Context, p RunParams) (*RunResult, error) {
	if p.Season <= 0 {
		return nil, errors.New("Season must be positive")
	}
	if p.NSims <= 0 {
		return nil, errors.New("NSims must be positive")
	}
	if p.Seed == 0 {
		return nil, errors.New("Seed must be non-zero")
	}
	if p.BatchSize <= 0 {
		return nil, errors.New("BatchSize must be positive")
	}
	if p.Workers <= 0 {
		p.Workers = runtime.GOMAXPROCS(0)
		if p.Workers <= 0 {
			p.Workers = 1
		}
	}
	if p.ProbabilitySourceKey == "" {
		p.ProbabilitySourceKey = "go_worker"
	}
	if strings.TrimSpace(p.StartingStateKey) == "" {
		p.StartingStateKey = "current"
	}
	if p.StartingStateKey != "current" && p.StartingStateKey != "post_first_four" {
		return nil, errors.New("StartingStateKey must be 'current' or 'post_first_four'")
	}

	overallStart := time.Now()

	loadStart := time.Now()
	coreTournamentID, err := dbadapter.ResolveCoreTournamentID(ctx, s.pool, p.Season)
	if err != nil {
		return nil, err
	}

	ff, err := dbadapter.LoadFinalFourConfig(ctx, s.pool, coreTournamentID)
	if err != nil {
		return nil, err
	}

	teams, err := s.loadTeams(ctx, coreTournamentID)
	if err != nil {
		return nil, err
	}

	br, err := appbracket.BuildBracketStructure(coreTournamentID, teams, ff)
	if err != nil {
		return nil, fmt.Errorf("failed to build bracket: %w", err)
	}

	var provider ProbabilityProvider
	var probs map[MatchupKey]float64
	if p.GameOutcomeSpec != nil {
		p.GameOutcomeSpec.Normalize()
		if err := p.GameOutcomeSpec.Validate(); err != nil {
			return nil, err
		}
		netByTeamID, err := s.loadKenPomNetByTeamID(ctx, coreTournamentID)
		if err != nil {
			return nil, err
		}
		if len(netByTeamID) == 0 {
			return nil, errors.New("no kenpom ratings available for tournament")
		}
		overrides := make(map[MatchupKey]float64)
		if p.StartingStateKey == "post_first_four" {
			if err := s.lockInFirstFourResults(ctx, br, overrides); err != nil {
				return nil, err
			}
		}
		provider = kenPomProvider{spec: p.GameOutcomeSpec, netByTeamID: netByTeamID, overrides: overrides}
	} else {
		selectedGameOutcomeRunID, loaded, nPredRows, err := s.loadPredictedGameOutcomesForTournament(ctx, coreTournamentID, p.GameOutcomeRunID)
		if err != nil {
			return nil, err
		}
		if nPredRows == 0 {
			if selectedGameOutcomeRunID != nil {
				return nil, fmt.Errorf("no predicted_game_outcomes found for run_id=%s", *selectedGameOutcomeRunID)
			}
			return nil, fmt.Errorf("no predicted_game_outcomes found for tournament_id=%s", coreTournamentID)
		}
		probs = loaded
		if p.StartingStateKey == "post_first_four" {
			if err := s.lockInFirstFourResults(ctx, br, probs); err != nil {
				return nil, err
			}
		}
		provider = nil
	}

	loadDur := time.Since(loadStart)

	snapshotID := ""
	if p.StartingStateKey == "post_first_four" {
		created, err := s.createTournamentStateSnapshotFromBracket(ctx, coreTournamentID, br, teams)
		if err != nil {
			return nil, err
		}
		snapshotID = created
	} else {
		created, err := s.createTournamentStateSnapshot(ctx, coreTournamentID)
		if err != nil {
			return nil, err
		}
		snapshotID = created
	}
	batchID, err := s.createTournamentSimulationBatch(ctx, coreTournamentID, snapshotID, p.NSims, p.Seed, p.ProbabilitySourceKey)
	if err != nil {
		return nil, err
	}

	simStart := time.Now()
	rowsWritten := int64(0)

	for offset := 0; offset < p.NSims; offset += p.BatchSize {
		n := p.BatchSize
		if offset+n > p.NSims {
			n = p.NSims - offset
		}

		batchSeed := int64(p.Seed) + int64(offset)*1_000_003
		var results []TeamSimulationResult
		if provider != nil {
			results, err = SimulateWithProvider(br, provider, n, batchSeed, Options{Workers: p.Workers})
		} else {
			results, err = Simulate(br, probs, n, batchSeed, Options{Workers: p.Workers})
		}
		if err != nil {
			return nil, err
		}

		inserted, err := s.copyInsertSimulatedTournaments(ctx, batchID, coreTournamentID, offset, results)
		if err != nil {
			return nil, err
		}
		rowsWritten += inserted
	}

	simDur := time.Since(simStart)
	overallDur := time.Since(overallStart)

	return &RunResult{
		CoreTournamentID:            coreTournamentID,
		TournamentStateSnapshotID:   snapshotID,
		TournamentSimulationBatchID: batchID,
		NSims:                       p.NSims,
		RowsWritten:                 rowsWritten,
		LoadDuration:                loadDur,
		SimulateWriteDuration:       simDur,
		OverallDuration:             overallDur,
	}, nil
}
