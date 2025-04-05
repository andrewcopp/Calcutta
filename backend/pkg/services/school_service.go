package services

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// SchoolService handles business logic for schools
type SchoolService struct {
	repo *SchoolRepository
}

// NewSchoolService creates a new SchoolService
func NewSchoolService(repo *SchoolRepository) *SchoolService {
	return &SchoolService{repo: repo}
}

// GetAllSchools returns all schools
func (s *SchoolService) GetAllSchools(ctx context.Context) ([]models.School, error) {
	return s.repo.GetAll(ctx)
}

// GetSchoolByID returns a school by ID
func (s *SchoolService) GetSchoolByID(ctx context.Context, id string) (models.School, error) {
	return s.repo.GetByID(ctx, id)
}
