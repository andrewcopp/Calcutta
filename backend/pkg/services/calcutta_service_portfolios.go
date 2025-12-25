package services

import (
	"context"
	"log"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// CalculateTotalBids calculates the total bids for a Calcutta entry
func (s *CalcuttaService) CalculateTotalBids(ctx context.Context, entryID string) (float64, error) {
	teams, err := s.ports.EntryReader.GetEntryTeams(ctx, entryID)
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
	// Get this entry's teams
	teams, err := s.ports.EntryReader.GetEntryTeams(ctx, entryID)
	if err != nil {
		return nil, err
	}

	// Find the calcutta ID for this entry
	entry, err := s.ports.EntryReader.GetEntry(ctx, entryID)
	if err != nil {
		return nil, err
	}

	// Build a list of all entry teams across the entire calcutta
	allEntries, err := s.ports.EntryReader.GetEntries(ctx, entry.CalcuttaID)
	if err != nil {
		return nil, err
	}

	var allTeams []*models.CalcuttaEntryTeam
	for _, e := range allEntries {
		ets, err := s.ports.EntryReader.GetEntryTeams(ctx, e.ID)
		if err != nil {
			return nil, err
		}
		allTeams = append(allTeams, ets...)
	}

	// Create the portfolio
	portfolio := &models.CalcuttaPortfolio{
		EntryID: entryID,
	}

	if err := s.ports.PortfolioWriter.CreatePortfolio(ctx, portfolio); err != nil {
		return nil, err
	}

	// For each team in this entry, calculate ownership as (entry bid for team) / (total bids for team across all entries)
	for _, t := range teams {
		totalBidsForTeam := 0.0
		for _, et := range allTeams {
			if et.TeamID == t.TeamID {
				totalBidsForTeam += float64(et.Bid)
			}
		}

		ownershipPercentage := 0.0
		if totalBidsForTeam > 0 {
			ownershipPercentage = float64(t.Bid) / totalBidsForTeam
		}

		portfolioTeam := &models.CalcuttaPortfolioTeam{
			PortfolioID:         portfolio.ID,
			TeamID:              t.TeamID,
			OwnershipPercentage: ownershipPercentage,
		}

		if err := s.ports.PortfolioWriter.CreatePortfolioTeam(ctx, portfolioTeam); err != nil {
			return nil, err
		}
	}

	return portfolio, nil
}

func (s *CalcuttaService) EnsurePortfoliosAndRecalculate(ctx context.Context, calcuttaID string) error {
	entries, err := s.ports.EntryReader.GetEntries(ctx, calcuttaID)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		portfolios, err := s.ports.PortfolioReader.GetPortfoliosByEntry(ctx, entry.ID)
		if err != nil {
			return err
		}
		if len(portfolios) == 0 {
			if _, err := s.CreatePortfolio(ctx, entry.ID); err != nil {
				return err
			}
		}
	}

	if err := s.RecalculatePortfolio(ctx, calcuttaID); err != nil {
		return err
	}

	for _, entry := range entries {
		portfolios, err := s.ports.PortfolioReader.GetPortfoliosByEntry(ctx, entry.ID)
		if err != nil {
			return err
		}
		for _, p := range portfolios {
			if err := s.CalculatePortfolioScores(ctx, p.ID); err != nil {
				log.Printf("Error calculating portfolio scores for portfolio %s: %v", p.ID, err)
			}
		}
	}

	return nil
}

// UpdatePortfolioScores updates the maximum possible score for a portfolio
func (s *CalcuttaService) UpdatePortfolioScores(ctx context.Context, portfolioID string, maxPoints float64) error {
	portfolio, err := s.ports.PortfolioReader.GetPortfolio(ctx, portfolioID)
	if err != nil {
		return err
	}

	portfolio.MaximumPoints = maxPoints
	portfolio.Updated = time.Now()

	return s.ports.PortfolioWriter.UpdatePortfolio(ctx, portfolio)
}

// CalculatePortfolioScores calculates and updates the scores for a portfolio
func (s *CalcuttaService) CalculatePortfolioScores(ctx context.Context, portfolioID string) error {
	teams, err := s.ports.PortfolioReader.GetPortfolioTeams(ctx, portfolioID)
	if err != nil {
		return err
	}

	var maxPoints float64
	now := time.Now()

	for _, portfolioTeam := range teams {
		// Get the tournament team information
		tournamentTeam, err := s.ports.TeamReader.GetTournamentTeam(ctx, portfolioTeam.TeamID)
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

		err = s.ports.PortfolioWriter.UpdatePortfolioTeam(ctx, portfolioTeam)
		if err != nil {
			return err
		}

		// Add to maximum possible score if team is not eliminated
		if !tournamentTeam.Eliminated {
			maxPoints += float64(tournamentTeam.Wins+tournamentTeam.Byes) * portfolioTeam.OwnershipPercentage
		}
	}

	// Update the portfolio's maximum possible score
	portfolio, err := s.ports.PortfolioReader.GetPortfolio(ctx, portfolioID)
	if err != nil {
		return err
	}

	portfolio.MaximumPoints = maxPoints
	portfolio.Updated = now

	return s.ports.PortfolioWriter.UpdatePortfolio(ctx, portfolio)
}

// GetPortfolioTeams retrieves all teams for a portfolio
func (s *CalcuttaService) GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error) {
	return s.ports.PortfolioReader.GetPortfolioTeams(ctx, portfolioID)
}

// UpdatePortfolioTeam updates a portfolio team
func (s *CalcuttaService) UpdatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error {
	return s.ports.PortfolioWriter.UpdatePortfolioTeam(ctx, team)
}
