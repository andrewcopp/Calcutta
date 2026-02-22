package ports

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByExternalProvider(ctx context.Context, provider, providerID string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
}
