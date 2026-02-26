package pool

import (
	"context"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Ports struct {
	Pools           ports.PoolRepository
	Portfolios      ports.PortfolioRepository
	Payouts         ports.PayoutRepository
	OwnershipReader ports.OwnershipReader
	ScoringRules    ports.ScoringRuleRepository
	TeamReader      ports.TournamentTeamReader
	PoolInvitations ports.PoolInvitationRepository
}

// Service handles business logic for investment pools
type Service struct {
	ports Ports
}

// New creates a new Service
func New(ports Ports) *Service {
	return &Service{ports: ports}
}

// CreatePoolWithScoringRules creates a new pool and its associated scoring rules.
func (s *Service) CreatePoolWithScoringRules(ctx context.Context, pool *models.Pool, rules []*models.ScoringRule) error {
	if err := s.ports.Pools.Create(ctx, pool); err != nil {
		return fmt.Errorf("creating pool: %w", err)
	}

	for _, r := range rules {
		rule := &models.ScoringRule{
			PoolID:        pool.ID,
			WinIndex:      r.WinIndex,
			PointsAwarded: r.PointsAwarded,
		}
		if err := s.ports.ScoringRules.CreateScoringRule(ctx, rule); err != nil {
			return fmt.Errorf("creating scoring rule: %w", err)
		}
	}

	return nil
}

// GetOwnershipDetails retrieves all ownership details for an ownership summary
func (s *Service) GetOwnershipDetails(ctx context.Context, portfolioID string) ([]*models.OwnershipDetail, error) {
	return s.ports.OwnershipReader.GetOwnershipDetails(ctx, portfolioID)
}
