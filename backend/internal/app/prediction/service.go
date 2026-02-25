package prediction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/winprob"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

// Service handles prediction generation and storage.
type Service struct {
	repo ports.PredictionRepository
}

// New creates a new prediction service.
func New(repo ports.PredictionRepository) *Service {
	return &Service{repo: repo}
}

// RunParams configures a prediction run.
type RunParams struct {
	TournamentID         string
	ProbabilitySourceKey string         // e.g., "kenpom"
	GameOutcomeSpec      *winprob.Model // KenPom parameters
	ThroughRound         *int           // Override checkpoint; nil = auto-detect from team progress
}

func (p *RunParams) applyDefaults() {
	if p.ProbabilitySourceKey == "" {
		p.ProbabilitySourceKey = "kenpom"
	}
	if p.GameOutcomeSpec == nil {
		p.GameOutcomeSpec = &winprob.Model{Kind: "kenpom", Sigma: 10.0}
	}
	p.GameOutcomeSpec.Normalize()
}

// RunResult holds the output of a prediction run.
type RunResult struct {
	BatchID              string
	TournamentID         string
	ProbabilitySourceKey string
	TeamCount            int
	Duration             time.Duration
}

// loadTournamentData loads teams, scoring rules, and final four config from the database.
func (s *Service) loadTournamentData(ctx context.Context, tournamentID string) (*TournamentData, error) {
	teams, err := s.repo.LoadTeams(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load teams: %w", err)
	}

	rules, err := s.repo.LoadScoringRules(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load scoring rules: %w", err)
	}
	if len(rules) == 0 {
		rules = DefaultScoringRules()
	}

	ffConfig, err := s.repo.LoadFinalFourConfig(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load final four config: %w", err)
	}

	return &TournamentData{
		Teams:    teams,
		Rules:    rules,
		FFConfig: ffConfig,
	}, nil
}

// runForCheckpoint generates predictions for a single checkpoint and stores them.
func (s *Service) runForCheckpoint(ctx context.Context, tournamentID string, data *TournamentData, spec *winprob.Model, probSourceKey string, throughRound int) (*RunResult, error) {
	start := time.Now()

	state := NewTournamentState(data, throughRound)

	teamValues, err := generatePredictions(state, spec)
	if err != nil {
		return nil, fmt.Errorf("prediction run failed for checkpoint %d: %w", throughRound, err)
	}

	specJSON, _ := json.Marshal(spec)
	batchID, err := s.repo.StorePredictions(ctx, tournamentID, probSourceKey, specJSON, teamValues, throughRound)
	if err != nil {
		return nil, fmt.Errorf("failed to store predictions for checkpoint %d: %w", throughRound, err)
	}

	if _, err := s.repo.PruneOldBatchesForCheckpoint(ctx, tournamentID, throughRound, 1); err != nil {
		slog.Warn("prediction_prune_failed", "tournament_id", tournamentID, "through_round", throughRound, "error", err)
	}

	return &RunResult{
		BatchID:              batchID,
		TournamentID:         tournamentID,
		ProbabilitySourceKey: probSourceKey,
		TeamCount:            len(teamValues),
		Duration:             time.Since(start),
	}, nil
}

// Run generates predictions for a tournament and stores them in the database.
func (s *Service) Run(ctx context.Context, p RunParams) (*RunResult, error) {
	if p.TournamentID == "" {
		return nil, errors.New("TournamentID is required")
	}
	p.applyDefaults()

	data, err := s.loadTournamentData(ctx, p.TournamentID)
	if err != nil {
		return nil, err
	}

	throughRound := detectThroughRoundFromTeams(data.Teams)
	if p.ThroughRound != nil {
		throughRound = *p.ThroughRound
	}

	return s.runForCheckpoint(ctx, p.TournamentID, data, p.GameOutcomeSpec, p.ProbabilitySourceKey, throughRound)
}

// RunAllCheckpoints generates prediction batches for every checkpoint from
// round 0 through the current tournament state. Loads tournament data once
// and reuses it across all checkpoints.
func (s *Service) RunAllCheckpoints(ctx context.Context, p RunParams) ([]RunResult, error) {
	if p.TournamentID == "" {
		return nil, errors.New("TournamentID is required")
	}
	p.applyDefaults()

	data, err := s.loadTournamentData(ctx, p.TournamentID)
	if err != nil {
		return nil, err
	}

	throughRound := detectThroughRoundFromTeams(data.Teams)

	var results []RunResult
	for cp := 0; cp <= throughRound; cp++ {
		result, err := s.runForCheckpoint(ctx, p.TournamentID, data, p.GameOutcomeSpec, p.ProbabilitySourceKey, cp)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}
	return results, nil
}

// DetectThroughRound loads teams for the tournament and returns the current checkpoint.
func (s *Service) DetectThroughRound(ctx context.Context, tournamentID string) (int, error) {
	teams, err := s.repo.LoadTeams(ctx, tournamentID)
	if err != nil {
		return 0, fmt.Errorf("failed to load teams: %w", err)
	}
	return detectThroughRoundFromTeams(teams), nil
}

// detectThroughRoundFromTeams computes the tournament checkpoint from team progress.
func detectThroughRoundFromTeams(teams []TeamInput) int {
	maxWins := 0
	throughRound := 0
	for _, t := range teams {
		if t.Wins > maxWins {
			maxWins = t.Wins
		}
	}
	if maxWins > 0 {
		for _, t := range teams {
			if p := t.Wins + t.Byes; p > throughRound {
				throughRound = p
			}
		}
	}
	return throughRound
}

// ListBatches returns all non-deleted prediction batches for a tournament, newest first.
func (s *Service) ListBatches(ctx context.Context, tournamentID string) ([]PredictionBatch, error) {
	return s.repo.ListBatches(ctx, tournamentID)
}

// GetLatestBatch returns the most recent prediction batch for a tournament.
func (s *Service) GetLatestBatch(ctx context.Context, tournamentID string) (*PredictionBatch, bool, error) {
	return s.repo.GetLatestBatch(ctx, tournamentID)
}

// GetLatestBatchID returns the most recent prediction batch ID for a tournament.
func (s *Service) GetLatestBatchID(ctx context.Context, tournamentID string) (string, bool, error) {
	batch, found, err := s.repo.GetLatestBatch(ctx, tournamentID)
	if err != nil || !found {
		return "", found, err
	}
	return batch.ID, true, nil
}

// GetBatchSummary returns the batch metadata for a specific batch ID.
func (s *Service) GetBatchSummary(ctx context.Context, batchID string) (*PredictionBatch, error) {
	return s.repo.GetBatchSummary(ctx, batchID)
}

// GetTeamValues returns predicted team values for a batch.
func (s *Service) GetTeamValues(ctx context.Context, batchID string) ([]PredictedTeamValue, error) {
	return s.repo.GetTeamValues(ctx, batchID)
}

// GetExpectedPointsMap returns a map of team_id -> expected_points for a tournament.
func (s *Service) GetExpectedPointsMap(ctx context.Context, tournamentID string) (map[string]float64, error) {
	batchID, found, err := s.GetLatestBatchID(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("getting latest batch id: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("no prediction batch found for tournament %s", tournamentID)
	}

	values, err := s.repo.GetTeamValues(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("getting team values: %w", err)
	}

	result := make(map[string]float64, len(values))
	for _, v := range values {
		result[v.TeamID] = v.ExpectedPoints
	}
	return result, nil
}

// BackfillMissing generates predictions for any tournament that has 68 teams
// with KenPom data and scoring rules but no prediction batch.
func (s *Service) BackfillMissing(ctx context.Context) int {
	tournamentIDs, err := s.repo.ListEligibleTournamentsForBackfill(ctx)
	if err != nil {
		slog.Warn("prediction_backfill_query_failed", "error", err)
		return 0
	}

	if len(tournamentIDs) == 0 {
		return 0
	}

	count := 0
	for _, tid := range tournamentIDs {
		result, err := s.Run(ctx, RunParams{
			TournamentID:         tid,
			ProbabilitySourceKey: "kenpom",
		})
		if err != nil {
			slog.Warn("prediction_backfill_failed", "tournament_id", tid, "error", err)
			continue
		}
		slog.Info("prediction_backfill_succeeded",
			"tournament_id", tid, "batch_id", result.BatchID,
			"team_count", result.TeamCount, "duration_ms", result.Duration.Milliseconds())
		count++
	}
	return count
}
