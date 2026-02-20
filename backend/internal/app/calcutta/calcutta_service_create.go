package calcutta

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// CreateCalcuttaWithRounds creates a new calcutta and its associated rounds
func (s *Service) CreateCalcuttaWithRounds(ctx context.Context, calcutta *models.Calcutta, rounds []*models.CalcuttaRound) error {
	if err := s.ports.CalcuttaWriter.Create(ctx, calcutta); err != nil {
		return err
	}

	for _, r := range rounds {
		calcuttaRound := &models.CalcuttaRound{
			CalcuttaID: calcutta.ID,
			Round:      r.Round,
			Points:     r.Points,
		}
		if err := s.ports.RoundWriter.CreateRound(ctx, calcuttaRound); err != nil {
			return err
		}
	}

	return nil
}
