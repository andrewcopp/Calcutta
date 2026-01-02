package calcutta

import (
	"context"
	"errors"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// CreateCalcuttaWithRounds creates a new calcutta and its associated rounds
func (s *Service) CreateCalcuttaWithRounds(ctx context.Context, calcutta *models.Calcutta) error {
	if err := s.ports.CalcuttaWriter.Create(ctx, calcutta); err != nil {
		return err
	}

	calcuttas, err := s.ports.CalcuttaReader.GetCalcuttasByTournament(
		ctx,
		calcutta.TournamentID,
	)
	if err != nil {
		return err
	}

	var template *models.Calcutta
	for _, c := range calcuttas {
		if c == nil {
			continue
		}
		if c.ID == calcutta.ID {
			continue
		}
		if template == nil || c.Created.After(template.Created) {
			template = c
		}
	}
	if template == nil {
		return errors.New(
			"no template calcutta found for tournament; cannot initialize scoring rules",
		)
	}

	rounds, err := s.ports.RoundReader.GetRounds(ctx, template.ID)
	if err != nil {
		return err
	}
	if len(rounds) == 0 {
		return errors.New(
			"template calcutta has no scoring rules; cannot initialize scoring rules",
		)
	}

	for _, r := range rounds {
		if r == nil {
			continue
		}
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
