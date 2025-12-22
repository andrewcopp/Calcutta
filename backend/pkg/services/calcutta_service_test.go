package services

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// MockCalcuttaRepository is a mock implementation of CalcuttaRepository for testing
type MockCalcuttaRepository struct{}

func newTestCalcuttaService() *CalcuttaService {
	mockRepo := &MockCalcuttaRepository{}
	return NewCalcuttaService(CalcuttaServicePorts{
		CalcuttaReader:  mockRepo,
		CalcuttaWriter:  mockRepo,
		EntryReader:     mockRepo,
		PayoutReader:    mockRepo,
		PortfolioReader: mockRepo,
		PortfolioWriter: mockRepo,
		RoundWriter:     mockRepo,
		TeamReader:      mockRepo,
	})
}

func (m *MockCalcuttaRepository) GetAll(ctx context.Context) ([]*models.Calcutta, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) GetByID(ctx context.Context, id string) (*models.Calcutta, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) Create(ctx context.Context, calcutta *models.Calcutta) error {
	return nil
}

func (m *MockCalcuttaRepository) Update(ctx context.Context, calcutta *models.Calcutta) error {
	return nil
}

func (m *MockCalcuttaRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockCalcuttaRepository) GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) GetPayouts(ctx context.Context, calcuttaID string) ([]*models.CalcuttaPayout, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) GetPortfolio(ctx context.Context, id string) (*models.CalcuttaPortfolio, error) {
	return &models.CalcuttaPortfolio{
		ID:      id,
		EntryID: "entry1",
	}, nil
}

func (m *MockCalcuttaRepository) GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) UpdatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error {
	return nil
}

func (m *MockCalcuttaRepository) GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) UpdatePortfolio(ctx context.Context, portfolio *models.CalcuttaPortfolio) error {
	return nil
}

func (m *MockCalcuttaRepository) GetRounds(ctx context.Context, calcuttaID string) ([]*models.CalcuttaRound, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) CreatePortfolio(ctx context.Context, portfolio *models.CalcuttaPortfolio) error {
	return nil
}

func (m *MockCalcuttaRepository) CreatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error {
	return nil
}

func (m *MockCalcuttaRepository) GetPortfolios(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) CreateRound(ctx context.Context, round *models.CalcuttaRound) error {
	return nil
}
