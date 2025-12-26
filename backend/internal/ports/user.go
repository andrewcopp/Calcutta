package ports

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type UserReader interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
}

type UserWriter interface {
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
}

type UserRepository interface {
	UserReader
	UserWriter
}
