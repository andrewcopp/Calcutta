package ports

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// SchoolRepository provides read access to schools.
type SchoolRepository interface {
	List(ctx context.Context) ([]models.School, error)
	GetByID(ctx context.Context, id string) (*models.School, error)
}
