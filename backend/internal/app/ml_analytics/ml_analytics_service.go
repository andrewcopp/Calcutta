package ml_analytics

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

// Service handles business logic for ML analytics
type Service struct {
	repo ports.MLAnalyticsRepository
}

// New creates a new ML analytics service
func New(repo ports.MLAnalyticsRepository) *Service {
	return &Service{repo: repo}
}

// GetTournamentSimStats retrieves simulation statistics for a tournament
func (s *Service) GetTournamentSimStats(ctx context.Context, year int) (*ports.TournamentSimStats, error) {
	return s.repo.GetTournamentSimStats(ctx, year)
}

// GetTeamPerformance retrieves performance metrics for a specific team
func (s *Service) GetTeamPerformance(ctx context.Context, year int, teamID string) (*ports.TeamPerformance, error) {
	return s.repo.GetTeamPerformance(ctx, year, teamID)
}

func (s *Service) GetTeamPerformanceByCalcutta(ctx context.Context, calcuttaID string, teamID string) (*ports.TeamPerformance, error) {
	return s.repo.GetTeamPerformanceByCalcutta(ctx, calcuttaID, teamID)
}

// GetTeamPredictions retrieves ML predictions for all teams in a tournament
func (s *Service) GetTeamPredictions(ctx context.Context, year int, runID *string) ([]ports.TeamPrediction, error) {
	return s.repo.GetTeamPredictions(ctx, year, runID)
}

// GetOurEntryDetails retrieves our optimized entry with portfolio and performance
func (s *Service) GetOurEntryDetails(ctx context.Context, year int, runID string) (*ports.OurEntryDetails, error) {
	return s.repo.GetOurEntryDetails(ctx, year, runID)
}

// GetEntryRankings retrieves entry rankings sorted by normalized payout
func (s *Service) GetEntryRankings(ctx context.Context, year int, runID string, limit, offset int) ([]ports.EntryRanking, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetEntryRankings(ctx, year, runID, limit, offset)
}

// GetEntrySimulations retrieves all simulation outcomes for a specific entry
func (s *Service) GetEntrySimulations(ctx context.Context, year int, runID string, entryKey string, limit, offset int) (*ports.EntrySimulationDrillDown, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 100
	}
	if limit > 5000 {
		limit = 5000
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetEntrySimulations(ctx, year, runID, entryKey, limit, offset)
}

// GetEntryPortfolio retrieves the team composition for any entry
func (s *Service) GetEntryPortfolio(ctx context.Context, year int, runID string, entryKey string) (*ports.EntryPortfolio, error) {
	return s.repo.GetEntryPortfolio(ctx, year, runID, entryKey)
}

// GetOptimizationRuns retrieves all available optimization runs for a year
func (s *Service) GetOptimizationRuns(ctx context.Context, year int) ([]ports.OptimizationRun, error) {
	return s.repo.GetOptimizationRuns(ctx, year)
}

func (s *Service) GetSimulatedCalcuttaEntryRankings(ctx context.Context, calcuttaID string) (string, []ports.SimulatedCalcuttaEntryRanking, error) {
	return s.repo.GetSimulatedCalcuttaEntryRankings(ctx, calcuttaID)
}
