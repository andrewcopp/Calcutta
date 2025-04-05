package services

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// TournamentService handles business logic for tournaments
type TournamentService struct {
	repo *TournamentRepository
}

// NewTournamentService creates a new TournamentService
func NewTournamentService(repo *TournamentRepository) *TournamentService {
	return &TournamentService{repo: repo}
}

// GetAllTournaments returns all tournaments
func (s *TournamentService) GetAllTournaments(ctx context.Context) ([]models.Tournament, error) {
	return s.repo.GetAll(ctx)
}

// GetTournamentWithWinner returns a tournament with its winning team and school
func (s *TournamentService) GetTournamentWithWinner(ctx context.Context, tournamentID string) (*models.Tournament, *models.TournamentTeam, *models.School, error) {
	return s.repo.GetTournamentWithWinner(ctx, tournamentID)
}

// GetTournamentByID returns a tournament by ID
func (s *TournamentService) GetTournamentByID(ctx context.Context, id string) (*models.Tournament, error) {
	return s.repo.GetByID(ctx, id)
}

// GetWinningTeam returns the winning team for a tournament
func (s *TournamentService) GetWinningTeam(ctx context.Context, tournamentID string) (*models.TournamentTeam, error) {
	return s.repo.GetWinningTeam(ctx, tournamentID)
}
