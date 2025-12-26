package calcutta

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// CreateCalcuttaWithRounds creates a new calcutta and its associated rounds
func (s *Service) CreateCalcuttaWithRounds(ctx context.Context, calcutta *models.Calcutta) error {
	if err := s.ports.CalcuttaWriter.Create(ctx, calcutta); err != nil {
		return err
	}

	rounds := []struct {
		round  int
		points int
	}{
		{1, 50},  // Round of 64
		{2, 100}, // Round of 32
		{3, 150}, // Sweet 16
		{4, 200}, // Elite 8
		{5, 250}, // Final 4
		{6, 300}, // Championship
	}

	for _, round := range rounds {
		calcuttaRound := &models.CalcuttaRound{
			CalcuttaID: calcutta.ID,
			Round:      round.round,
			Points:     round.points,
		}
		if err := s.ports.RoundWriter.CreateRound(ctx, calcuttaRound); err != nil {
			return err
		}
	}

	return nil
}
