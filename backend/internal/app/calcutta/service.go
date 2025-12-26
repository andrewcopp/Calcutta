package calcutta

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/andrewcopp/Calcutta/backend/pkg/services"
)

type Service struct {
	svc *services.CalcuttaService
}

func New(svc *services.CalcuttaService) *Service {
	return &Service{svc: svc}
}

func (s *Service) GetAllCalcuttas(ctx context.Context) ([]*models.Calcutta, error) {
	res, err := s.svc.GetAllCalcuttas(ctx)
	return res, apperrors.Translate(err)
}

func (s *Service) CreateCalcuttaWithRounds(ctx context.Context, calcutta *models.Calcutta) error {
	return apperrors.Translate(s.svc.CreateCalcuttaWithRounds(ctx, calcutta))
}

func (s *Service) GetCalcuttaByID(ctx context.Context, id string) (*models.Calcutta, error) {
	res, err := s.svc.GetCalcuttaByID(ctx, id)
	return res, apperrors.Translate(err)
}

func (s *Service) UpdateCalcutta(ctx context.Context, calcutta *models.Calcutta) error {
	return apperrors.Translate(s.svc.UpdateCalcutta(ctx, calcutta))
}

func (s *Service) GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, error) {
	res, err := s.svc.GetEntries(ctx, calcuttaID)
	return res, apperrors.Translate(err)
}

func (s *Service) GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error) {
	res, err := s.svc.GetEntry(ctx, id)
	return res, apperrors.Translate(err)
}

func (s *Service) GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error) {
	res, err := s.svc.GetEntryTeams(ctx, entryID)
	return res, apperrors.Translate(err)
}

func (s *Service) ReplaceEntryTeams(ctx context.Context, entryID string, teams []*models.CalcuttaEntryTeam) error {
	return apperrors.Translate(s.svc.ReplaceEntryTeams(ctx, entryID, teams))
}

func (s *Service) ValidateEntry(entry *models.CalcuttaEntry, teams []*models.CalcuttaEntryTeam) error {
	return apperrors.Translate(s.svc.ValidateEntry(entry, teams))
}

func (s *Service) EnsurePortfoliosAndRecalculate(ctx context.Context, calcuttaID string) error {
	return apperrors.Translate(s.svc.EnsurePortfoliosAndRecalculate(ctx, calcuttaID))
}

func (s *Service) GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	res, err := s.svc.GetPortfoliosByEntry(ctx, entryID)
	return res, apperrors.Translate(err)
}

func (s *Service) GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error) {
	res, err := s.svc.GetPortfolioTeams(ctx, portfolioID)
	return res, apperrors.Translate(err)
}

func (s *Service) CalculatePortfolioScores(ctx context.Context, portfolioID string) error {
	return apperrors.Translate(s.svc.CalculatePortfolioScores(ctx, portfolioID))
}

func (s *Service) UpdatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error {
	return apperrors.Translate(s.svc.UpdatePortfolioTeam(ctx, team))
}

func (s *Service) UpdatePortfolioScores(ctx context.Context, portfolioID string, maximumPoints float64) error {
	return apperrors.Translate(s.svc.UpdatePortfolioScores(ctx, portfolioID, maximumPoints))
}

func (s *Service) GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error) {
	res, err := s.svc.GetCalcuttasByTournament(ctx, tournamentID)
	return res, apperrors.Translate(err)
}

func (s *Service) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	res, err := s.svc.GetTournamentTeam(ctx, id)
	return res, apperrors.Translate(err)
}
