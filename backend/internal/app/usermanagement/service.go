package usermanagement

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Ports struct {
	Merges ports.UserMergeRepository
}

type Service struct {
	ports Ports
}

func New(p Ports) *Service {
	return &Service{ports: p}
}

func (s *Service) ListStubUsers(ctx context.Context) ([]*models.User, error) {
	return s.ports.Merges.ListStubUsers(ctx)
}

func (s *Service) FindMergeCandidates(ctx context.Context, userID string) ([]*models.User, error) {
	return s.ports.Merges.FindMergeCandidates(ctx, userID)
}

func (s *Service) MergeUsers(ctx context.Context, sourceUserID, targetUserID, mergedBy string) (*models.UserMerge, error) {
	return s.ports.Merges.MergeUsers(ctx, sourceUserID, targetUserID, mergedBy)
}

func (s *Service) ListMergeHistory(ctx context.Context, userID string) ([]*models.UserMerge, error) {
	return s.ports.Merges.ListMergeHistory(ctx, userID)
}
