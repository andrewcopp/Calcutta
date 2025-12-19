package services

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

func (s *CalcuttaService) GetAllCalcuttas(ctx context.Context) ([]*models.Calcutta, error) {
	return s.ports.CalcuttaReader.GetAll(ctx)
}

func (s *CalcuttaService) GetCalcuttaByID(ctx context.Context, id string) (*models.Calcutta, error) {
	return s.ports.CalcuttaReader.GetByID(ctx, id)
}

func (s *CalcuttaService) CreateCalcutta(ctx context.Context, calcutta *models.Calcutta) error {
	return s.ports.CalcuttaWriter.Create(ctx, calcutta)
}

func (s *CalcuttaService) CreateRound(ctx context.Context, round *models.CalcuttaRound) error {
	return s.ports.RoundWriter.CreateRound(ctx, round)
}

func (s *CalcuttaService) GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, error) {
	return s.ports.EntryReader.GetEntries(ctx, calcuttaID)
}

func (s *CalcuttaService) GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error) {
	return s.ports.EntryReader.GetEntryTeams(ctx, entryID)
}

func (s *CalcuttaService) GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	return s.ports.PortfolioReader.GetPortfoliosByEntry(ctx, entryID)
}

func (s *CalcuttaService) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	return s.ports.TeamReader.GetTournamentTeam(ctx, id)
}

func (s *CalcuttaService) GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error) {
	return s.ports.CalcuttaReader.GetCalcuttasByTournament(ctx, tournamentID)
}
