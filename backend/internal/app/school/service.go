package school

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Service struct {
	repo ports.SchoolRepository
}

func New(repo ports.SchoolRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]models.School, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*models.School, error) {
	return s.repo.GetByID(ctx, id)
}
