package ports

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type PoolReader interface {
	GetAll(ctx context.Context) ([]*models.Pool, error)
	GetByID(ctx context.Context, id string) (*models.Pool, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Pool, error)
	GetPoolsByTournament(ctx context.Context, tournamentID string) ([]*models.Pool, error)
}

type PoolWriter interface {
	Create(ctx context.Context, pool *models.Pool) error
	Update(ctx context.Context, pool *models.Pool) error
}

type PoolRepository interface {
	PoolReader
	PoolWriter
}

type PortfolioReader interface {
	GetPortfolios(ctx context.Context, poolID string) ([]*models.Portfolio, map[string]float64, error)
	GetPortfolio(ctx context.Context, id string) (*models.Portfolio, error)
	GetInvestments(ctx context.Context, portfolioID string) ([]*models.Investment, error)
	GetInvestmentsByPortfolioIDs(ctx context.Context, portfolioIDs []string) (map[string][]*models.Investment, error)
	GetDistinctUserIDsByPool(ctx context.Context, poolID string) ([]string, error)
}

type PortfolioWriter interface {
	CreatePortfolio(ctx context.Context, portfolio *models.Portfolio) error
	ReplaceInvestments(ctx context.Context, portfolioID string, investments []*models.Investment) error
	UpdatePortfolioStatus(ctx context.Context, id string, status string) error
}

type PortfolioRepository interface {
	PortfolioReader
	PortfolioWriter
}

type OwnershipReader interface {
	GetOwnershipSummary(ctx context.Context, id string) (*models.OwnershipSummary, error)
	GetOwnershipDetails(ctx context.Context, portfolioID string) ([]*models.OwnershipDetail, error)
	GetOwnershipSummariesByPortfolio(ctx context.Context, portfolioID string) ([]*models.OwnershipSummary, error)
	GetOwnershipSummariesByPortfolioIDs(ctx context.Context, portfolioIDs []string) (map[string][]*models.OwnershipSummary, error)
	GetOwnershipDetailsByPortfolioIDs(ctx context.Context, portfolioIDs []string) (map[string][]*models.OwnershipDetail, error)
}

type ScoringRuleReader interface {
	GetScoringRules(ctx context.Context, poolID string) ([]*models.ScoringRule, error)
}

type ScoringRuleWriter interface {
	CreateScoringRule(ctx context.Context, rule *models.ScoringRule) error
}

type ScoringRuleRepository interface {
	ScoringRuleReader
	ScoringRuleWriter
}

type PayoutReader interface {
	GetPayouts(ctx context.Context, poolID string) ([]*models.PoolPayout, error)
}

type PayoutWriter interface {
	ReplacePayouts(ctx context.Context, poolID string, payouts []*models.PoolPayout) error
}

type PayoutRepository interface {
	PayoutReader
	PayoutWriter
}

type PoolInvitationReader interface {
	ListInvitations(ctx context.Context, poolID string) ([]*models.PoolInvitation, error)
	GetInvitationByPoolAndUser(ctx context.Context, poolID, userID string) (*models.PoolInvitation, error)
	GetPendingInvitationByPoolAndUser(ctx context.Context, poolID, userID string) (*models.PoolInvitation, error)
	ListPendingInvitationsByUserID(ctx context.Context, userID string) ([]*models.PoolInvitation, error)
}

type PoolInvitationWriter interface {
	CreateInvitation(ctx context.Context, invitation *models.PoolInvitation) error
	AcceptInvitation(ctx context.Context, id string) error
	RevokeInvitation(ctx context.Context, id string) error
}

type PoolInvitationRepository interface {
	PoolInvitationReader
	PoolInvitationWriter
}

type InvestmentSnapshotWriter interface {
	CreateInvestmentSnapshot(ctx context.Context, snapshot *models.InvestmentSnapshot) error
}

type TournamentTeamReader interface {
	GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error)
}
