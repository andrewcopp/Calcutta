package pool

import (
	"context"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// ReinviteFromPool creates a new pool and invites all participants from a source pool.
func (s *Service) ReinviteFromPool(ctx context.Context, sourcePoolID string, newPool *models.Pool, invitedBy string) (*models.Pool, []*models.PoolInvitation, error) {
	source, err := s.ports.Pools.GetByID(ctx, sourcePoolID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting source pool: %w", err)
	}

	// Copy defaults from source if not set
	if newPool.MinTeams == 0 {
		newPool.MinTeams = source.MinTeams
	}
	if newPool.MaxTeams == 0 {
		newPool.MaxTeams = source.MaxTeams
	}
	if newPool.MaxInvestmentCredits == 0 {
		newPool.MaxInvestmentCredits = source.MaxInvestmentCredits
	}

	sourceScoringRules, err := s.ports.ScoringRules.GetScoringRules(ctx, sourcePoolID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting source scoring rules: %w", err)
	}

	if err := s.CreatePoolWithScoringRules(ctx, newPool, sourceScoringRules); err != nil {
		return nil, nil, fmt.Errorf("creating pool with scoring rules: %w", err)
	}

	userIDs, err := s.ports.Portfolios.GetDistinctUserIDsByPool(ctx, sourcePoolID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting user ids from source pool: %w", err)
	}

	invitations := make([]*models.PoolInvitation, 0, len(userIDs))
	for _, uid := range userIDs {
		inv := &models.PoolInvitation{
			PoolID:    newPool.ID,
			UserID:    uid,
			InvitedBy: invitedBy,
		}
		if err := s.ports.PoolInvitations.CreateInvitation(ctx, inv); err != nil {
			return nil, nil, fmt.Errorf("creating invitation for user %s: %w", uid, err)
		}
		invitations = append(invitations, inv)
	}

	return newPool, invitations, nil
}
