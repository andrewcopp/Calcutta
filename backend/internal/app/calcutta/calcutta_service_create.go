package calcutta

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// defaultScoringRules returns the standard scoring rules for a new calcutta
// when no template is available: 0/50/100/150/200/250/300 points for wins 1-7.
func defaultScoringRules() []*models.CalcuttaRound {
	points := []int{0, 50, 100, 150, 200, 250, 300}
	rounds := make([]*models.CalcuttaRound, len(points))
	for i, p := range points {
		rounds[i] = &models.CalcuttaRound{
			Round:  i + 1,
			Points: p,
		}
	}
	return rounds
}

// CreateCalcuttaWithRounds creates a new calcutta and its associated rounds
func (s *Service) CreateCalcuttaWithRounds(ctx context.Context, calcutta *models.Calcutta) error {
	if err := s.ports.CalcuttaWriter.Create(ctx, calcutta); err != nil {
		return err
	}

	rounds := s.findTemplateRounds(ctx, calcutta)
	if len(rounds) == 0 {
		rounds = defaultScoringRules()
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

// findTemplateRounds looks for scoring rules from an existing calcutta in the
// same tournament to use as a template. Returns nil if none found.
func (s *Service) findTemplateRounds(ctx context.Context, calcutta *models.Calcutta) []*models.CalcuttaRound {
	calcuttas, err := s.ports.CalcuttaReader.GetCalcuttasByTournament(
		ctx,
		calcutta.TournamentID,
	)
	if err != nil {
		return nil
	}

	var template *models.Calcutta
	for _, c := range calcuttas {
		if c == nil || c.ID == calcutta.ID {
			continue
		}
		if template == nil || c.Created.After(template.Created) {
			template = c
		}
	}
	if template == nil {
		return nil
	}

	rounds, err := s.ports.RoundReader.GetRounds(ctx, template.ID)
	if err != nil || len(rounds) == 0 {
		return nil
	}
	return rounds
}
