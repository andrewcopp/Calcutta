package calcutta

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func (s *Service) InviteUser(ctx context.Context, invitation *models.CalcuttaInvitation) error {
	return s.ports.Invitations.CreateInvitation(ctx, invitation)
}

func (s *Service) ListInvitations(ctx context.Context, calcuttaID string) ([]*models.CalcuttaInvitation, error) {
	return s.ports.Invitations.ListInvitations(ctx, calcuttaID)
}

func (s *Service) AcceptInvitation(ctx context.Context, id string) error {
	return s.ports.Invitations.AcceptInvitation(ctx, id)
}

func (s *Service) GetInvitationByCalcuttaAndUser(ctx context.Context, calcuttaID, userID string) (*models.CalcuttaInvitation, error) {
	return s.ports.Invitations.GetInvitationByCalcuttaAndUser(ctx, calcuttaID, userID)
}

func (s *Service) GetPendingInvitationByCalcuttaAndUser(ctx context.Context, calcuttaID, userID string) (*models.CalcuttaInvitation, error) {
	return s.ports.Invitations.GetPendingInvitationByCalcuttaAndUser(ctx, calcuttaID, userID)
}

func (s *Service) RevokeInvitation(ctx context.Context, id string) error {
	return s.ports.Invitations.RevokeInvitation(ctx, id)
}

func (s *Service) ListPendingInvitationsByUserID(ctx context.Context, userID string) ([]*models.CalcuttaInvitation, error) {
	return s.ports.Invitations.ListPendingInvitationsByUserID(ctx, userID)
}
