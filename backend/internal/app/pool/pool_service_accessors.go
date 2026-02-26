package pool

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func (s *Service) GetAllPools(ctx context.Context) ([]*models.Pool, error) {
	return s.ports.Pools.GetAll(ctx)
}

func (s *Service) GetPoolByID(ctx context.Context, id string) (*models.Pool, error) {
	return s.ports.Pools.GetByID(ctx, id)
}

func (s *Service) UpdatePool(ctx context.Context, pool *models.Pool) error {
	return s.ports.Pools.Update(ctx, pool)
}

func (s *Service) GetPortfolios(ctx context.Context, poolID string) ([]*models.Portfolio, []*models.PortfolioStanding, error) {
	portfolios, returnsByPortfolio, err := s.ports.Portfolios.GetPortfolios(ctx, poolID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting portfolios: %w", err)
	}

	payouts, err := s.ports.Payouts.GetPayouts(ctx, poolID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting payouts: %w", err)
	}

	standings := ComputeStandings(portfolios, returnsByPortfolio, payouts)
	return portfolios, standings, nil
}

// ComputeStandings computes finish positions and payouts from portfolios, their returns, and payout rules.
// Returns standings sorted by returns descending. Does not mutate portfolios.
func ComputeStandings(
	portfolios []*models.Portfolio,
	returnsByPortfolio map[string]float64,
	payouts []*models.PoolPayout,
) []*models.PortfolioStanding {
	if portfolios == nil {
		return nil
	}

	type portfolioWithReturns struct {
		portfolio *models.Portfolio
		returns   float64
	}

	sorted := make([]portfolioWithReturns, 0, len(portfolios))
	for _, p := range portfolios {
		if p == nil {
			continue
		}
		sorted = append(sorted, portfolioWithReturns{portfolio: p, returns: returnsByPortfolio[p.ID]})
	}

	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].returns == sorted[j].returns {
			return sorted[i].portfolio.CreatedAt.After(sorted[j].portfolio.CreatedAt)
		}
		return sorted[i].returns > sorted[j].returns
	})

	payoutByPosition := map[int]int{}
	for _, p := range payouts {
		if p == nil {
			continue
		}
		payoutByPosition[p.Position] = p.AmountCents
	}

	const epsilon = 0.0001
	standings := make([]*models.PortfolioStanding, len(sorted))

	position := 1
	for i := 0; i < len(sorted); {
		j := i + 1
		for j < len(sorted) {
			if math.Abs(sorted[j].returns-sorted[i].returns) >= epsilon {
				break
			}
			j++
		}

		groupSize := j - i
		isTied := groupSize > 1

		totalGroupPayout := 0
		for pos := position; pos < position+groupSize; pos++ {
			totalGroupPayout += payoutByPosition[pos]
		}

		base := 0
		remainder := 0
		if groupSize > 0 {
			base = totalGroupPayout / groupSize
			remainder = totalGroupPayout % groupSize
		}

		for k := 0; k < groupSize; k++ {
			payoutCents := base
			if remainder > 0 {
				payoutCents++
				remainder--
			}
			standings[i+k] = &models.PortfolioStanding{
				PortfolioID:    sorted[i+k].portfolio.ID,
				TotalReturns:   sorted[i+k].returns,
				FinishPosition: position,
				IsTied:         isTied,
				PayoutCents:    payoutCents,
				InTheMoney:     payoutCents > 0,
			}
		}

		position += groupSize
		i = j
	}

	return standings
}

func (s *Service) GetInvestments(ctx context.Context, portfolioID string) ([]*models.Investment, error) {
	return s.ports.Portfolios.GetInvestments(ctx, portfolioID)
}

func (s *Service) GetInvestmentsByPortfolioIDs(ctx context.Context, portfolioIDs []string) (map[string][]*models.Investment, error) {
	return s.ports.Portfolios.GetInvestmentsByPortfolioIDs(ctx, portfolioIDs)
}

func (s *Service) GetPortfolio(ctx context.Context, id string) (*models.Portfolio, error) {
	return s.ports.Portfolios.GetPortfolio(ctx, id)
}

func (s *Service) CreatePortfolio(ctx context.Context, portfolio *models.Portfolio) error {
	return s.ports.Portfolios.CreatePortfolio(ctx, portfolio)
}

func (s *Service) ReplaceInvestments(ctx context.Context, portfolioID string, investments []*models.Investment) error {
	return s.ports.Portfolios.ReplaceInvestments(ctx, portfolioID, investments)
}

func (s *Service) UpdatePortfolioStatus(ctx context.Context, id string, status string) error {
	return s.ports.Portfolios.UpdatePortfolioStatus(ctx, id, status)
}

func (s *Service) GetOwnershipSummariesByPortfolio(ctx context.Context, portfolioID string) ([]*models.OwnershipSummary, error) {
	return s.ports.OwnershipReader.GetOwnershipSummariesByPortfolio(ctx, portfolioID)
}

func (s *Service) GetOwnershipSummariesByPortfolioIDs(ctx context.Context, portfolioIDs []string) (map[string][]*models.OwnershipSummary, error) {
	return s.ports.OwnershipReader.GetOwnershipSummariesByPortfolioIDs(ctx, portfolioIDs)
}

func (s *Service) GetOwnershipDetailsByPortfolioIDs(ctx context.Context, portfolioIDs []string) (map[string][]*models.OwnershipDetail, error) {
	return s.ports.OwnershipReader.GetOwnershipDetailsByPortfolioIDs(ctx, portfolioIDs)
}

func (s *Service) GetOwnershipSummary(ctx context.Context, id string) (*models.OwnershipSummary, error) {
	return s.ports.OwnershipReader.GetOwnershipSummary(ctx, id)
}

func (s *Service) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	return s.ports.TeamReader.GetTournamentTeam(ctx, id)
}

func (s *Service) GetPoolsByUser(ctx context.Context, userID string) ([]*models.Pool, error) {
	return s.ports.Pools.GetByUserID(ctx, userID)
}

func (s *Service) GetScoringRules(ctx context.Context, poolID string) ([]*models.ScoringRule, error) {
	return s.ports.ScoringRules.GetScoringRules(ctx, poolID)
}

func (s *Service) GetPoolsByTournament(ctx context.Context, tournamentID string) ([]*models.Pool, error) {
	return s.ports.Pools.GetPoolsByTournament(ctx, tournamentID)
}

func (s *Service) GetDistinctUserIDsByPool(ctx context.Context, poolID string) ([]string, error) {
	return s.ports.Portfolios.GetDistinctUserIDsByPool(ctx, poolID)
}

func (s *Service) GetPayouts(ctx context.Context, poolID string) ([]*models.PoolPayout, error) {
	return s.ports.Payouts.GetPayouts(ctx, poolID)
}

func (s *Service) ReplacePayouts(ctx context.Context, poolID string, payouts []*models.PoolPayout) error {
	return s.ports.Payouts.ReplacePayouts(ctx, poolID, payouts)
}
