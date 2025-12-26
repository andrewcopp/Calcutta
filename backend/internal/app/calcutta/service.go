package calcutta

import (
	"context"

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
	return s.svc.GetAllCalcuttas(ctx)
}

func (s *Service) CreateCalcuttaWithRounds(ctx context.Context, calcutta *models.Calcutta) error {
	return s.svc.CreateCalcuttaWithRounds(ctx, calcutta)
}

func (s *Service) GetCalcuttaByID(ctx context.Context, id string) (*models.Calcutta, error) {
	return s.svc.GetCalcuttaByID(ctx, id)
}

func (s *Service) UpdateCalcutta(ctx context.Context, calcutta *models.Calcutta) error {
	return s.svc.UpdateCalcutta(ctx, calcutta)
}

func (s *Service) GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, error) {
	return s.svc.GetEntries(ctx, calcuttaID)
}

func (s *Service) GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error) {
	return s.svc.GetEntry(ctx, id)
}

func (s *Service) GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error) {
	return s.svc.GetEntryTeams(ctx, entryID)
}

func (s *Service) ReplaceEntryTeams(ctx context.Context, entryID string, teams []*models.CalcuttaEntryTeam) error {
	return s.svc.ReplaceEntryTeams(ctx, entryID, teams)
}

func (s *Service) ValidateEntry(entry *models.CalcuttaEntry, teams []*models.CalcuttaEntryTeam) error {
	return s.svc.ValidateEntry(entry, teams)
}

func (s *Service) EnsurePortfoliosAndRecalculate(ctx context.Context, calcuttaID string) error {
	return s.svc.EnsurePortfoliosAndRecalculate(ctx, calcuttaID)
}

func (s *Service) GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	return s.svc.GetPortfoliosByEntry(ctx, entryID)
}

func (s *Service) GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error) {
	return s.svc.GetPortfolioTeams(ctx, portfolioID)
}

func (s *Service) CalculatePortfolioScores(ctx context.Context, portfolioID string) error {
	return s.svc.CalculatePortfolioScores(ctx, portfolioID)
}

func (s *Service) UpdatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error {
	return s.svc.UpdatePortfolioTeam(ctx, team)
}

func (s *Service) UpdatePortfolioScores(ctx context.Context, portfolioID string, maximumPoints float64) error {
	return s.svc.UpdatePortfolioScores(ctx, portfolioID, maximumPoints)
}

func (s *Service) GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error) {
	return s.svc.GetCalcuttasByTournament(ctx, tournamentID)
}

func (s *Service) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	return s.svc.GetTournamentTeam(ctx, id)
}
