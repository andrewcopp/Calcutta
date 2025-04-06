package services

import (
	"context"
	"errors"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// CalcuttaService handles business logic for Calcutta auctions
type CalcuttaService struct {
	repo CalcuttaRepositoryInterface
}

// NewCalcuttaService creates a new CalcuttaService
func NewCalcuttaService(repo CalcuttaRepositoryInterface) *CalcuttaService {
	return &CalcuttaService{repo: repo}
}

// ValidateEntry validates all bids for an entry according to the rules
func (s *CalcuttaService) ValidateEntry(entry *models.CalcuttaEntry, teams []*models.CalcuttaEntryTeam) error {
	// Rule 1: All bids are sealed until the tournament begins
	// This is handled at the API level, not in the service

	// Rule 2: Players must bid on a minimum of 3 teams
	if len(teams) < 3 {
		return errors.New("players must bid on a minimum of 3 teams")
	}

	// Rule 3: Players may bid on a maximum of 10 teams
	if len(teams) > 10 {
		return errors.New("players may bid on a maximum of 10 teams")
	}

	// Rule 4: Maximum bid on any single team is $50
	for _, team := range teams {
		if team.Bid > 50 {
			return errors.New("maximum bid on any single team is $50")
		}
	}

	// Rule 5: Total bids cannot exceed starting budget of $100
	totalBids := 0.0
	for _, team := range teams {
		totalBids += float64(team.Bid)
	}
	if totalBids > 100 {
		return errors.New("total bids cannot exceed starting budget of $100")
	}

	// Rule 6: Minimum bid on any team is $1
	for _, team := range teams {
		if team.Bid < 1 {
			return errors.New("minimum bid on any team is $1")
		}
	}

	// Rule 7: Players cannot bid on the same team multiple times
	teamBids := make(map[string]bool)
	for _, team := range teams {
		if teamBids[team.TeamID] {
			return errors.New("players cannot bid on the same team multiple times")
		}
		teamBids[team.TeamID] = true
	}

	return nil
}

// CalculateOwnershipPercentage calculates the ownership percentage for a team
func (s *CalcuttaService) CalculateOwnershipPercentage(team interface{}, allTeams []*models.CalcuttaEntryTeam) float64 {
	// Rule 8: Ownership percentage = (Player's bid on team) ÷ (Total bids on team)
	totalBids := 0.0
	var teamID string
	var bid float64

	// Extract team ID and bid based on type
	switch t := team.(type) {
	case *models.CalcuttaEntryTeam:
		teamID = t.TeamID
		bid = float64(t.Bid)
	case *models.CalcuttaPortfolioTeam:
		teamID = t.TeamID
		// For portfolio teams, we need to find the corresponding entry team to get the bid
		// First, get the portfolio to find the entry ID
		portfolio, err := s.repo.GetPortfolio(context.Background(), t.PortfolioID)
		if err != nil {
			return 0
		}
		// Now find the specific entry team for this entry
		for _, et := range allTeams {
			if et.TeamID == t.TeamID && et.EntryID == portfolio.EntryID {
				bid = float64(et.Bid)
				break
			}
		}
	default:
		return 0
	}

	// Calculate total bids for this team
	for _, t := range allTeams {
		if t.TeamID == teamID {
			totalBids += float64(t.Bid)
		}
	}

	if totalBids == 0 {
		return 0
	}

	return bid / totalBids
}

// CalculatePoints calculates the points earned by a team based on its performance
func (s *CalcuttaService) CalculatePoints(team *models.TournamentTeam, rounds []*models.CalcuttaRound) float64 {
	// Calculate total progress (wins + byes)
	totalProgress := team.Wins + team.Byes

	// Map total progress to points
	// 0 wins/byes = First Four Loss (0 points)
	// 1 win/bye = First Four Win or Bye (0 points)
	// 2 wins/byes = First Round Win (50 points)
	// 3 wins/byes = Round of 32 Win (150 points)
	// 4 wins/byes = Sweet 16 Win (300 points)
	// 5 wins/byes = Elite 8 Win (500 points)
	// 6 wins/byes = Final Four Win (750 points)
	// 7 wins/byes = Championship Game Win (1050 points)
	switch totalProgress {
	case 0:
		return 0
	case 1:
		return 0
	case 2:
		return 50 // First Round Winner
	case 3:
		return 150 // Round of 32 Winner
	case 4:
		return 300 // Sweet 16 Winner
	case 5:
		return 500 // Elite 8 Winner
	case 6:
		return 750 // Final Four Winner
	case 7:
		return 1050 // Tournament Winner
	default:
		// For any other value (shouldn't happen in a normal tournament)
		return 0
	}
}

// CalculatePlayerPoints calculates the points earned by a player based on their team ownerships
func (s *CalcuttaService) CalculatePlayerPoints(portfolio *models.CalcuttaPortfolio, portfolioTeams []*models.CalcuttaPortfolioTeam) float64 {
	totalPoints := 0.0
	for _, portfolioTeam := range portfolioTeams {
		if portfolioTeam.PortfolioID == portfolio.ID {
			totalPoints += portfolioTeam.ActualPoints
		}
	}
	return totalPoints
}

// CalculateTotalBids calculates the total bids for a Calcutta entry
func (s *CalcuttaService) CalculateTotalBids(ctx context.Context, entryID string) (float64, error) {
	teams, err := s.repo.GetEntryTeams(ctx, entryID)
	if err != nil {
		return 0, err
	}

	var totalBids float64
	for _, team := range teams {
		totalBids += float64(team.Bid)
	}

	return totalBids, nil
}

// CreatePortfolio creates a new portfolio for a Calcutta entry
func (s *CalcuttaService) CreatePortfolio(ctx context.Context, entryID string) (*models.CalcuttaPortfolio, error) {
	teams, err := s.repo.GetEntryTeams(ctx, entryID)
	if err != nil {
		return nil, err
	}

	totalBids, err := s.CalculateTotalBids(ctx, entryID)
	if err != nil {
		return nil, err
	}

	portfolio := &models.CalcuttaPortfolio{
		EntryID: entryID,
	}

	err = s.repo.CreatePortfolio(ctx, portfolio)
	if err != nil {
		return nil, err
	}

	for _, t := range teams {
		ownershipPercentage := float64(t.Bid) / totalBids
		portfolioTeam := &models.CalcuttaPortfolioTeam{
			PortfolioID:         portfolio.ID,
			TeamID:              t.TeamID,
			OwnershipPercentage: ownershipPercentage,
		}

		err = s.repo.CreatePortfolioTeam(ctx, portfolioTeam)
		if err != nil {
			return nil, err
		}
	}

	return portfolio, nil
}

// UpdatePortfolioScores updates the maximum possible score for a portfolio
func (s *CalcuttaService) UpdatePortfolioScores(ctx context.Context, portfolioID string, maxPoints float64) error {
	portfolio, err := s.repo.GetPortfolio(ctx, portfolioID)
	if err != nil {
		return err
	}

	portfolio.MaximumPoints = maxPoints
	portfolio.Updated = time.Now()

	return s.repo.UpdatePortfolio(ctx, portfolio)
}

// CalculatePortfolioScores calculates and updates the scores for a portfolio
func (s *CalcuttaService) CalculatePortfolioScores(ctx context.Context, portfolioID string) error {
	teams, err := s.repo.GetPortfolioTeams(ctx, portfolioID)
	if err != nil {
		return err
	}

	var maxPoints float64
	now := time.Now()

	for _, portfolioTeam := range teams {
		// Get the tournament team information
		tournamentTeam, err := s.repo.GetTournamentTeam(ctx, portfolioTeam.TeamID)
		if err != nil {
			return err
		}

		// Calculate expected value score based on current performance
		expectedScore := float64(tournamentTeam.Wins) * portfolioTeam.OwnershipPercentage

		// Calculate predicted future score based on team's seed and current performance
		predictedScore := float64(tournamentTeam.Wins+tournamentTeam.Byes-1) * portfolioTeam.OwnershipPercentage

		// Update the portfolio team's scores
		portfolioTeam.ExpectedPoints = expectedScore
		portfolioTeam.PredictedPoints = predictedScore
		portfolioTeam.Updated = now

		err = s.repo.UpdatePortfolioTeam(ctx, portfolioTeam)
		if err != nil {
			return err
		}

		// Add to maximum possible score if team is not eliminated
		if !tournamentTeam.Eliminated {
			maxPoints += float64(tournamentTeam.Wins+tournamentTeam.Byes) * portfolioTeam.OwnershipPercentage
		}
	}

	// Update the portfolio's maximum possible score
	portfolio, err := s.repo.GetPortfolio(ctx, portfolioID)
	if err != nil {
		return err
	}

	portfolio.MaximumPoints = maxPoints
	portfolio.Updated = now

	return s.repo.UpdatePortfolio(ctx, portfolio)
}

// GetPortfolioTeams retrieves all teams for a portfolio
func (s *CalcuttaService) GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error) {
	return s.repo.GetPortfolioTeams(ctx, portfolioID)
}

// UpdatePortfolioTeam updates a portfolio team
func (s *CalcuttaService) UpdatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error {
	return s.repo.UpdatePortfolioTeam(ctx, team)
}
