package pool

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func (s *Service) InviteUser(ctx context.Context, invitation *models.PoolInvitation) error {
	return s.ports.PoolInvitations.CreateInvitation(ctx, invitation)
}

func (s *Service) ListInvitations(ctx context.Context, poolID string) ([]*models.PoolInvitation, error) {
	return s.ports.PoolInvitations.ListInvitations(ctx, poolID)
}

func (s *Service) AcceptInvitation(ctx context.Context, id string) error {
	return s.ports.PoolInvitations.AcceptInvitation(ctx, id)
}

func (s *Service) GetInvitationByPoolAndUser(ctx context.Context, poolID, userID string) (*models.PoolInvitation, error) {
	return s.ports.PoolInvitations.GetInvitationByPoolAndUser(ctx, poolID, userID)
}

func (s *Service) GetPendingInvitationByPoolAndUser(ctx context.Context, poolID, userID string) (*models.PoolInvitation, error) {
	return s.ports.PoolInvitations.GetPendingInvitationByPoolAndUser(ctx, poolID, userID)
}

func (s *Service) RevokeInvitation(ctx context.Context, id string) error {
	return s.ports.PoolInvitations.RevokeInvitation(ctx, id)
}

func (s *Service) ListPendingInvitationsByUserID(ctx context.Context, userID string) ([]*models.PoolInvitation, error) {
	return s.ports.PoolInvitations.ListPendingInvitationsByUserID(ctx, userID)
}
