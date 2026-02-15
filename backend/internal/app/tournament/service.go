package tournament

import (
	"context"
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/google/uuid"
)

type Service struct {
	repo *dbadapters.TournamentRepository
}

func New(repo *dbadapters.TournamentRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]models.Tournament, error) {
	return s.repo.GetAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*models.Tournament, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, name string, rounds int) (*models.Tournament, error) {
	t := &models.Tournament{ID: uuid.New().String(), Name: name, Rounds: rounds}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) GetTeams(ctx context.Context, tournamentID string) ([]*models.TournamentTeam, error) {
	return s.repo.GetTeams(ctx, tournamentID)
}

func (s *Service) CreateTeam(ctx context.Context, team *models.TournamentTeam) error {
	return s.repo.CreateTeam(ctx, team)
}

func (s *Service) UpdateTournamentTeam(ctx context.Context, team *models.TournamentTeam) error {
	return s.repo.UpdateTournamentTeam(ctx, team)
}

func (s *Service) GetWinningTeam(ctx context.Context, tournamentID string) (*models.TournamentTeam, error) {
	return s.repo.GetWinningTeam(ctx, tournamentID)
}

func (s *Service) UpdateStartingAt(ctx context.Context, tournamentID string, startingAt *time.Time) error {
	return s.repo.UpdateStartingAt(ctx, tournamentID, startingAt)
}
