package bracket

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/andrewcopp/Calcutta/backend/pkg/services"
)

type Service struct {
	svc *services.BracketService
}

func New(svc *services.BracketService) *Service {
	return &Service{svc: svc}
}

func (s *Service) GetBracket(ctx context.Context, tournamentID string) (*models.BracketStructure, error) {
	res, err := s.svc.GetBracket(ctx, tournamentID)
	return res, apperrors.Translate(err)
}

func (s *Service) SelectWinner(ctx context.Context, tournamentID, gameID, winnerTeamID string) (*models.BracketStructure, error) {
	res, err := s.svc.SelectWinner(ctx, tournamentID, gameID, winnerTeamID)
	return res, apperrors.Translate(err)
}

func (s *Service) UnselectWinner(ctx context.Context, tournamentID, gameID string) (*models.BracketStructure, error) {
	res, err := s.svc.UnselectWinner(ctx, tournamentID, gameID)
	return res, apperrors.Translate(err)
}

func (s *Service) ValidateBracketSetup(ctx context.Context, tournamentID string) error {
	return apperrors.Translate(s.svc.ValidateBracketSetup(ctx, tournamentID))
}
