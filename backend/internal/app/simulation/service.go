package simulation

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	appbracket "github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TournamentResolver resolves tournament metadata without importing adapters.
type TournamentResolver interface {
	ResolveCoreTournamentID(ctx context.Context, season int) (string, error)
	LoadFinalFourConfig(ctx context.Context, coreTournamentID string) (*models.FinalFourConfig, error)
}

type Service struct {
	pool               *pgxpool.Pool
	tournamentResolver TournamentResolver
}

func New(pool *pgxpool.Pool, opts ...Option) *Service {
	s := &Service{pool: pool}
	for _, o := range opts {
		o(s)
	}
	return s
}

type Option func(*Service)

func WithTournamentResolver(r TournamentResolver) Option {
	return func(s *Service) { s.tournamentResolver = r }
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
	if err := validateAndDefaultParams(&p); err != nil {
		return nil, err
	}

	overallStart := time.Now()

	// Phase 1: Load data and build bracket/probabilities.
	loadStart := time.Now()
	setup, err := s.loadBracketAndProbabilities(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("loading bracket and probabilities: %w", err)
	}
	loadDur := time.Since(loadStart)

	// Phase 2: Persist snapshot and batch records.
	snapshotID, batchID, err := s.createSnapshotAndBatch(ctx, setup.coreTournamentID, setup.bracket, setup.teams, p)
	if err != nil {
		return nil, fmt.Errorf("creating snapshot and batch: %w", err)
	}

	// Phase 3: Run simulation batches and write results.
	simStart := time.Now()
	rowsWritten, err := s.runSimulationBatches(ctx, setup.bracket, setup.provider, setup.probs, batchID, setup.coreTournamentID, p)
	if err != nil {
		return nil, fmt.Errorf("running simulation batches: %w", err)
	}
	simDur := time.Since(simStart)

	overallDur := time.Since(overallStart)
	return &RunResult{
		CoreTournamentID:            setup.coreTournamentID,
		TournamentStateSnapshotID:   snapshotID,
		TournamentSimulationBatchID: batchID,
		NSims:                       p.NSims,
		RowsWritten:                 rowsWritten,
		LoadDuration:                loadDur,
		SimulateWriteDuration:       simDur,
		OverallDuration:             overallDur,
	}, nil
}

// validateAndDefaultParams validates required fields and fills in defaults for
// optional fields. It modifies p in place.
func validateAndDefaultParams(p *RunParams) error {
	if p.Season <= 0 {
		return errors.New("Season must be positive")
	}
	if p.NSims <= 0 {
		return errors.New("NSims must be positive")
	}
	if p.Seed == 0 {
		return errors.New("Seed must be non-zero")
	}
	if p.BatchSize <= 0 {
		return errors.New("BatchSize must be positive")
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
		return errors.New("StartingStateKey must be 'current' or 'post_first_four'")
	}
	return nil
}

// setupResult holds the outputs of the data-loading phase.
type setupResult struct {
	coreTournamentID string
	teams            []*models.TournamentTeam
	bracket          *models.BracketStructure
	provider         ProbabilityProvider
	probs            map[MatchupKey]float64
}

// loadBracketAndProbabilities resolves the tournament, loads teams, builds the
// bracket structure, and resolves the probability source (KenPom or predicted
// game outcomes).
func (s *Service) loadBracketAndProbabilities(ctx context.Context, p RunParams) (*setupResult, error) {
	coreTournamentID, err := s.tournamentResolver.ResolveCoreTournamentID(ctx, p.Season)
	if err != nil {
		return nil, fmt.Errorf("resolving tournament id for season %d: %w", p.Season, err)
	}

	ff, err := s.tournamentResolver.LoadFinalFourConfig(ctx, coreTournamentID)
	if err != nil {
		return nil, fmt.Errorf("loading final four config: %w", err)
	}

	teams, err := s.loadTeams(ctx, coreTournamentID)
	if err != nil {
		return nil, fmt.Errorf("loading teams: %w", err)
	}

	br, err := appbracket.BuildBracketStructure(coreTournamentID, teams, ff)
	if err != nil {
		return nil, fmt.Errorf("failed to build bracket: %w", err)
	}

	provider, probs, err := s.resolveProbabilities(ctx, coreTournamentID, br, p)
	if err != nil {
		return nil, fmt.Errorf("resolving probabilities: %w", err)
	}

	return &setupResult{
		coreTournamentID: coreTournamentID,
		teams:            teams,
		bracket:          br,
		provider:         provider,
		probs:            probs,
	}, nil
}

// resolveProbabilities builds either a KenPom-based provider or loads predicted
// game outcome probabilities from the database.
func (s *Service) resolveProbabilities(
	ctx context.Context,
	coreTournamentID string,
	br *models.BracketStructure,
	p RunParams,
) (ProbabilityProvider, map[MatchupKey]float64, error) {
	if p.GameOutcomeSpec != nil {
		return s.resolveKenPomProbabilities(ctx, coreTournamentID, br, p)
	}
	return s.resolvePredictedProbabilities(ctx, coreTournamentID, br, p)
}

func (s *Service) resolveKenPomProbabilities(
	ctx context.Context,
	coreTournamentID string,
	br *models.BracketStructure,
	p RunParams,
) (ProbabilityProvider, map[MatchupKey]float64, error) {
	p.GameOutcomeSpec.Normalize()
	if err := p.GameOutcomeSpec.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validating game outcome spec: %w", err)
	}
	netByTeamID, err := s.loadKenPomNetByTeamID(ctx, coreTournamentID)
	if err != nil {
		return nil, nil, fmt.Errorf("loading kenpom ratings: %w", err)
	}
	if len(netByTeamID) == 0 {
		return nil, nil, errors.New("no kenpom ratings available for tournament")
	}
	overrides := make(map[MatchupKey]float64)
	if p.StartingStateKey == "post_first_four" {
		if err := s.lockInFirstFourResults(ctx, br, overrides); err != nil {
			return nil, nil, fmt.Errorf("locking in first four results: %w", err)
		}
	}
	provider := kenPomProvider{spec: p.GameOutcomeSpec, netByTeamID: netByTeamID, overrides: overrides}
	return provider, nil, nil
}

func (s *Service) resolvePredictedProbabilities(
	ctx context.Context,
	coreTournamentID string,
	br *models.BracketStructure,
	p RunParams,
) (ProbabilityProvider, map[MatchupKey]float64, error) {
	selectedGameOutcomeRunID, loaded, nPredRows, err := s.loadPredictedGameOutcomesForTournament(ctx, coreTournamentID, p.GameOutcomeRunID)
	if err != nil {
		return nil, nil, fmt.Errorf("loading predicted game outcomes: %w", err)
	}
	if nPredRows == 0 {
		if selectedGameOutcomeRunID != nil {
			return nil, nil, fmt.Errorf("no predicted_game_outcomes found for run_id=%s", *selectedGameOutcomeRunID)
		}
		return nil, nil, fmt.Errorf("no predicted_game_outcomes found for tournament_id=%s", coreTournamentID)
	}
	if p.StartingStateKey == "post_first_four" {
		if err := s.lockInFirstFourResults(ctx, br, loaded); err != nil {
			return nil, nil, fmt.Errorf("locking in first four results: %w", err)
		}
	}
	return nil, loaded, nil
}

// createSnapshotAndBatch persists the tournament state snapshot and creates the
// simulation batch record.
func (s *Service) createSnapshotAndBatch(
	ctx context.Context,
	coreTournamentID string,
	br *models.BracketStructure,
	teams []*models.TournamentTeam,
	p RunParams,
) (string, string, error) {
	var snapshotID string
	var err error
	if p.StartingStateKey == "post_first_four" {
		snapshotID, err = s.createTournamentStateSnapshotFromBracket(ctx, coreTournamentID, br, teams)
	} else {
		snapshotID, err = s.createTournamentStateSnapshot(ctx, coreTournamentID)
	}
	if err != nil {
		return "", "", fmt.Errorf("creating tournament state snapshot: %w", err)
	}

	batchID, err := s.createTournamentSimulationBatch(ctx, coreTournamentID, snapshotID, p.NSims, p.Seed, p.ProbabilitySourceKey)
	if err != nil {
		return "", "", fmt.Errorf("creating simulation batch: %w", err)
	}

	return snapshotID, batchID, nil
}

// runSimulationBatches runs simulations in chunks of BatchSize, writing each
// chunk to the database via COPY. It returns the total number of rows written.
func (s *Service) runSimulationBatches(
	ctx context.Context,
	br *models.BracketStructure,
	provider ProbabilityProvider,
	probs map[MatchupKey]float64,
	batchID string,
	coreTournamentID string,
	p RunParams,
) (int64, error) {
	rowsWritten := int64(0)

	for offset := 0; offset < p.NSims; offset += p.BatchSize {
		n := p.BatchSize
		if offset+n > p.NSims {
			n = p.NSims - offset
		}

		batchSeed := int64(p.Seed) + int64(offset)*1_000_003
		var results []TeamSimulationResult
		var err error
		if provider != nil {
			results, err = SimulateWithProvider(br, provider, n, batchSeed, Options{Workers: p.Workers})
		} else {
			results, err = Simulate(br, probs, n, batchSeed, Options{Workers: p.Workers})
		}
		if err != nil {
			return 0, fmt.Errorf("simulating batch at offset %d: %w", offset, err)
		}

		inserted, err := s.copyInsertSimulatedTournaments(ctx, batchID, coreTournamentID, offset, results)
		if err != nil {
			return 0, fmt.Errorf("inserting simulated tournaments at offset %d: %w", offset, err)
		}
		rowsWritten += inserted
	}

	return rowsWritten, nil
}
