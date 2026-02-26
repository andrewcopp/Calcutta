package calcutta

import (
	"context"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Ports struct {
	Calcuttas       ports.CalcuttaRepository
	Entries         ports.EntryRepository
	Payouts         ports.PayoutRepository
	PortfolioReader ports.PortfolioReader
	ScoringRules    ports.ScoringRuleRepository
	TeamReader      ports.TournamentTeamReader
	Invitations     ports.CalcuttaInvitationRepository
}

// Service handles business logic for Calcutta auctions
type Service struct {
	ports Ports
}

// New creates a new Service
func New(ports Ports) *Service {
	return &Service{ports: ports}
}

// CreateCalcuttaWithScoringRules creates a new calcutta and its associated scoring rules.
func (s *Service) CreateCalcuttaWithScoringRules(ctx context.Context, calcutta *models.Calcutta, rules []*models.ScoringRule) error {
	if err := s.ports.Calcuttas.Create(ctx, calcutta); err != nil {
		return fmt.Errorf("creating calcutta: %w", err)
	}

	for _, r := range rules {
		rule := &models.ScoringRule{
			CalcuttaID:    calcutta.ID,
			WinIndex:      r.WinIndex,
			PointsAwarded: r.PointsAwarded,
		}
		if err := s.ports.ScoringRules.CreateScoringRule(ctx, rule); err != nil {
			return fmt.Errorf("creating scoring rule: %w", err)
		}
	}

	return nil
}

// GetPortfolioTeams retrieves all teams for a portfolio
func (s *Service) GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error) {
	return s.ports.PortfolioReader.GetPortfolioTeams(ctx, portfolioID)
}
