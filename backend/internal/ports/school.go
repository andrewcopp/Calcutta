package ports

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type SchoolReader interface {
	GetAll(ctx context.Context) ([]models.School, error)
	GetByID(ctx context.Context, id string) (models.School, error)
}
