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
	Rounds          ports.RoundRepository
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

// CreateCalcuttaWithRounds creates a new calcutta and its associated rounds
func (s *Service) CreateCalcuttaWithRounds(ctx context.Context, calcutta *models.Calcutta, rounds []*models.CalcuttaRound) error {
	if err := s.ports.Calcuttas.Create(ctx, calcutta); err != nil {
		return fmt.Errorf("creating calcutta: %w", err)
	}

	for _, r := range rounds {
		calcuttaRound := &models.CalcuttaRound{
			CalcuttaID: calcutta.ID,
			Round:      r.Round,
			Points:     r.Points,
		}
		if err := s.ports.Rounds.CreateRound(ctx, calcuttaRound); err != nil {
			return fmt.Errorf("creating calcutta round: %w", err)
		}
	}

	return nil
}

// GetPortfolioTeams retrieves all teams for a portfolio
func (s *Service) GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error) {
	return s.ports.PortfolioReader.GetPortfolioTeams(ctx, portfolioID)
}
