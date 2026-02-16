package calcutta

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// MockCalcuttaRepository is a mock implementation of CalcuttaRepository for testing
type MockCalcuttaRepository struct{}

func newTestCalcuttaService() *Service {
	mockRepo := &MockCalcuttaRepository{}
	return New(Ports{
		CalcuttaReader:  mockRepo,
		CalcuttaWriter:  mockRepo,
		EntryReader:     mockRepo,
		EntryWriter:     mockRepo,
		PayoutReader:    mockRepo,
		PortfolioReader: mockRepo,
		RoundReader:     mockRepo,
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

func (m *MockCalcuttaRepository) GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) GetRounds(ctx context.Context, calcuttaID string) ([]*models.CalcuttaRound, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) CreateEntry(ctx context.Context, entry *models.CalcuttaEntry) error {
	return nil
}

func (m *MockCalcuttaRepository) ReplaceEntryTeams(ctx context.Context, entryID string, teams []*models.CalcuttaEntryTeam) error {
	return nil
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

func (m *MockCalcuttaRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Calcutta, error) {
	return nil, nil
}

func (m *MockCalcuttaRepository) GetDistinctUserIDsByCalcutta(ctx context.Context, calcuttaID string) ([]string, error) {
	return nil, nil
}
