package school

import (
	"context"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type Service struct {
	repo *dbadapters.SchoolRepository
}

func New(repo *dbadapters.SchoolRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]models.School, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*models.School, error) {
	return s.repo.GetByID(ctx, id)
}
