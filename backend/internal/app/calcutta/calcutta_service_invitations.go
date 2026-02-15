package calcutta

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

func (s *Service) InviteUser(ctx context.Context, invitation *models.CalcuttaInvitation) error {
	return s.ports.InvitationWriter.CreateInvitation(ctx, invitation)
}

func (s *Service) ListInvitations(ctx context.Context, calcuttaID string) ([]*models.CalcuttaInvitation, error) {
	return s.ports.InvitationReader.ListInvitations(ctx, calcuttaID)
}

func (s *Service) AcceptInvitation(ctx context.Context, id string) error {
	return s.ports.InvitationWriter.AcceptInvitation(ctx, id)
}

func (s *Service) GetInvitationByCalcuttaAndUser(ctx context.Context, calcuttaID, userID string) (*models.CalcuttaInvitation, error) {
	return s.ports.InvitationReader.GetInvitationByCalcuttaAndUser(ctx, calcuttaID, userID)
}
