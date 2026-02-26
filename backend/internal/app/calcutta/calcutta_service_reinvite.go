package calcutta

import (
	"context"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// ReinviteFromCalcutta creates a new calcutta and invites all participants from a source calcutta.
func (s *Service) ReinviteFromCalcutta(ctx context.Context, sourceCalcuttaID string, newCalcutta *models.Calcutta, invitedBy string) (*models.Calcutta, []*models.CalcuttaInvitation, error) {
	source, err := s.ports.Calcuttas.GetByID(ctx, sourceCalcuttaID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting source calcutta: %w", err)
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

	sourceScoringRules, err := s.ports.ScoringRules.GetScoringRules(ctx, sourceCalcuttaID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting source scoring rules: %w", err)
	}

	if err := s.CreateCalcuttaWithScoringRules(ctx, newCalcutta, sourceScoringRules); err != nil {
		return nil, nil, fmt.Errorf("creating calcutta with scoring rules: %w", err)
	}

	userIDs, err := s.ports.Entries.GetDistinctUserIDsByCalcutta(ctx, sourceCalcuttaID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting user ids from source calcutta: %w", err)
	}

	invitations := make([]*models.CalcuttaInvitation, 0, len(userIDs))
	for _, uid := range userIDs {
		inv := &models.CalcuttaInvitation{
			CalcuttaID: newCalcutta.ID,
			UserID:     uid,
			InvitedBy:  invitedBy,
		}
		if err := s.ports.Invitations.CreateInvitation(ctx, inv); err != nil {
			return nil, nil, fmt.Errorf("creating invitation for user %s: %w", uid, err)
		}
		invitations = append(invitations, inv)
	}

	return newCalcutta, invitations, nil
}
