package ports

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type CalcuttaReader interface {
	GetAll(ctx context.Context) ([]*models.Calcutta, error)
	GetByID(ctx context.Context, id string) (*models.Calcutta, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Calcutta, error)
	GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error)
}

type CalcuttaWriter interface {
	Create(ctx context.Context, calcutta *models.Calcutta) error
	Update(ctx context.Context, calcutta *models.Calcutta) error
}

type CalcuttaRepository interface {
	CalcuttaReader
	CalcuttaWriter
}

type EntryReader interface {
	GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, error)
	GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error)
	GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error)
	GetEntryTeamsByEntryIDs(ctx context.Context, entryIDs []string) (map[string][]*models.CalcuttaEntryTeam, error)
	GetDistinctUserIDsByCalcutta(ctx context.Context, calcuttaID string) ([]string, error)
}

type EntryWriter interface {
	CreateEntry(ctx context.Context, entry *models.CalcuttaEntry) error
	ReplaceEntryTeams(ctx context.Context, entryID string, teams []*models.CalcuttaEntryTeam) error
}

type EntryRepository interface {
	EntryReader
	EntryWriter
}

type PortfolioReader interface {
	GetPortfolio(ctx context.Context, id string) (*models.CalcuttaPortfolio, error)
	GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error)
	GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error)
	GetPortfoliosByEntryIDs(ctx context.Context, entryIDs []string) (map[string][]*models.CalcuttaPortfolio, error)
	GetPortfolioTeamsByPortfolioIDs(ctx context.Context, portfolioIDs []string) (map[string][]*models.CalcuttaPortfolioTeam, error)
}

type RoundReader interface {
	GetRounds(ctx context.Context, calcuttaID string) ([]*models.CalcuttaRound, error)
}

type RoundWriter interface {
	CreateRound(ctx context.Context, round *models.CalcuttaRound) error
}

type RoundRepository interface {
	RoundReader
	RoundWriter
}

type PayoutReader interface {
	GetPayouts(ctx context.Context, calcuttaID string) ([]*models.CalcuttaPayout, error)
}

type PayoutWriter interface {
	ReplacePayouts(ctx context.Context, calcuttaID string, payouts []*models.CalcuttaPayout) error
}

type PayoutRepository interface {
	PayoutReader
	PayoutWriter
}

type CalcuttaInvitationReader interface {
	ListInvitations(ctx context.Context, calcuttaID string) ([]*models.CalcuttaInvitation, error)
	GetInvitationByCalcuttaAndUser(ctx context.Context, calcuttaID, userID string) (*models.CalcuttaInvitation, error)
	GetPendingInvitationByCalcuttaAndUser(ctx context.Context, calcuttaID, userID string) (*models.CalcuttaInvitation, error)
	ListPendingInvitationsByUserID(ctx context.Context, userID string) ([]*models.CalcuttaInvitation, error)
}

type CalcuttaInvitationWriter interface {
	CreateInvitation(ctx context.Context, invitation *models.CalcuttaInvitation) error
	AcceptInvitation(ctx context.Context, id string) error
	RevokeInvitation(ctx context.Context, id string) error
}

type CalcuttaInvitationRepository interface {
	CalcuttaInvitationReader
	CalcuttaInvitationWriter
}

type TournamentTeamReader interface {
	GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error)
}
