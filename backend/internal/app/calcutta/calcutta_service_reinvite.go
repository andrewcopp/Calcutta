package calcutta

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// ReinviteFromCalcutta creates a new calcutta and invites all participants from a source calcutta.
func (s *Service) ReinviteFromCalcutta(ctx context.Context, sourceCalcuttaID string, newCalcutta *models.Calcutta, invitedBy string) (*models.Calcutta, []*models.CalcuttaInvitation, error) {
	source, err := s.ports.CalcuttaReader.GetByID(ctx, sourceCalcuttaID)
	if err != nil {
		return nil, nil, err
	}

	// Copy defaults from source if not set
	if newCalcutta.MinTeams == 0 {
		newCalcutta.MinTeams = source.MinTeams
	}
	if newCalcutta.MaxTeams == 0 {
		newCalcutta.MaxTeams = source.MaxTeams
	}
	if newCalcutta.MaxBidPoints == 0 {
		newCalcutta.MaxBidPoints = source.MaxBidPoints
	}

	sourceRounds, err := s.ports.RoundReader.GetRounds(ctx, sourceCalcuttaID)
	if err != nil {
		return nil, nil, err
	}

	if err := s.CreateCalcuttaWithRounds(ctx, newCalcutta, sourceRounds); err != nil {
		return nil, nil, err
	}

	userIDs, err := s.ports.EntryReader.GetDistinctUserIDsByCalcutta(ctx, sourceCalcuttaID)
	if err != nil {
		return nil, nil, err
	}

	invitations := make([]*models.CalcuttaInvitation, 0, len(userIDs))
	for _, uid := range userIDs {
		inv := &models.CalcuttaInvitation{
			CalcuttaID: newCalcutta.ID,
			UserID:     uid,
			InvitedBy:  invitedBy,
		}
		if err := s.ports.InvitationWriter.CreateInvitation(ctx, inv); err != nil {
			return nil, nil, err
		}
		invitations = append(invitations, inv)
	}

	return newCalcutta, invitations, nil
}
