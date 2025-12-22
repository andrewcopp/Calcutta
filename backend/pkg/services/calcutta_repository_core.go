package services

import (
	"context"
	"database/sql"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// CalcuttaRepositoryInterface defines the interface that a Calcutta repository must implement

type CalcuttaReader interface {
	GetAll(ctx context.Context) ([]*models.Calcutta, error)
	GetByID(ctx context.Context, id string) (*models.Calcutta, error)
	GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error)
}

type CalcuttaWriter interface {
	Create(ctx context.Context, calcutta *models.Calcutta) error
	Update(ctx context.Context, calcutta *models.Calcutta) error
	Delete(ctx context.Context, id string) error
	CreateRound(ctx context.Context, round *models.CalcuttaRound) error
}

type EntryReader interface {
	GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, error)
	GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error)
	GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error)
}

type PortfolioReader interface {
	GetPortfolio(ctx context.Context, id string) (*models.CalcuttaPortfolio, error)
	GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error)
	GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error)
	GetPortfolios(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error)
}

type PortfolioWriter interface {
	CreatePortfolio(ctx context.Context, portfolio *models.CalcuttaPortfolio) error
	CreatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error
	UpdatePortfolio(ctx context.Context, portfolio *models.CalcuttaPortfolio) error
	UpdatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error
}

type RoundReader interface {
	GetRounds(ctx context.Context, calcuttaID string) ([]*models.CalcuttaRound, error)
}

type PayoutReader interface {
	GetPayouts(ctx context.Context, calcuttaID string) ([]*models.CalcuttaPayout, error)
}

type TournamentTeamReader interface {
	GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error)
}

type CalcuttaRepositoryInterface interface {
	CalcuttaReader
	CalcuttaWriter
	EntryReader
	PayoutReader
	PortfolioReader
	PortfolioWriter
	RoundReader
	TournamentTeamReader
}

// CalcuttaRepository handles data access for Calcutta entities
type CalcuttaRepository struct {
	db *sql.DB
}

// NewCalcuttaRepository creates a new CalcuttaRepository
func NewCalcuttaRepository(db *sql.DB) *CalcuttaRepository {
	return &CalcuttaRepository{db: db}
}
