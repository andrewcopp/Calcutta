package ports

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByIDs(ctx context.Context, ids []string) ([]*models.User, error)
	GetByExternalProvider(ctx context.Context, provider, providerID string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
}

type UserMergeRepository interface {
	MergeUsers(ctx context.Context, sourceUserID, targetUserID, mergedBy string) (*models.UserMerge, error)
	BatchMergeUsers(ctx context.Context, sourceUserIDs []string, targetUserID, mergedBy string) ([]*models.UserMerge, error)
	ListStubUsers(ctx context.Context) ([]*models.User, error)
	FindMergeCandidates(ctx context.Context, userID string) ([]*models.User, error)
	ListMergeHistory(ctx context.Context, userID string) ([]*models.UserMerge, error)
}
